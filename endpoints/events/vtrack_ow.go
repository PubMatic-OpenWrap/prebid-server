package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/beevik/etree"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// standard VAST macros
// https://interactiveadvertisingbureau.github.io/vast/vast4macros/vast4-macros-latest.html#macro-spec-adcount
const (
	VASTAdTypeMacro    = "[ADTYPE]"
	VASTAppBundleMacro = "[APPBUNDLE]"
	VASTDomainMacro    = "[DOMAIN]"
	VASTPageURLMacro   = "[PAGEURL]"

	// PBS specific macros
	PBSEventIDMacro = "[EVENT_ID]" // macro for injecting PBS defined  video event tracker id
	//[PBS-ACCOUNT] represents publisher id / account id
	PBSAccountMacro = "[PBS-ACCOUNT]"
	// [PBS-BIDDER] represents bidder name
	PBSBidderMacro = "[PBS-BIDDER]"
	// [PBS-ORIG_BIDID] represents original bid id.
	PBSOrigBidIDMacro = "[PBS-ORIG_BIDID]"
	// [PBS-BIDID] represents bid id. If auction.generate-bid-id config is on, then resolve with response.seatbid.bid.ext.prebid.bidid. Else replace with response.seatbid.bid.id
	PBSBidIDMacro = "[PBS-BIDID]"
	// [ADERVERTISER_NAME] represents advertiser name
	PBSAdvertiserNameMacro = "[ADVERTISER_NAME]"
	// Pass imp.tagId using this macro
	PBSAdUnitIDMacro = "[AD_UNIT]"
	//PBSBidderCodeMacro represents an alias id or core bidder id.
	PBSBidderCodeMacro = "[BIDDER_CODE]"
)

var trackingEvents = []string{"start", "firstQuartile", "midpoint", "thirdQuartile", "complete"}

// PubMatic specific event IDs
// This will go in event-config once PreBid modular design is in place
var trackingEventIDMap = map[string]string{
	"start":         "2",
	"firstQuartile": "4",
	"midpoint":      "3",
	"thirdQuartile": "5",
	"complete":      "6",
}

// InjectVideoEventTrackers injects the video tracking events
// Returns VAST xml contains as first argument. Second argument indicates whether the trackers are injected and last argument indicates if there is any error in injecting the trackers
func InjectVideoEventTrackers(trackerURL, vastXML string, bid *openrtb2.Bid, prebidGenBidId, requestingBidder, bidderCoreName, accountID string, timestamp int64, bidRequest *openrtb2.BidRequest) (string, error) {
	// parse VAST
	doc := etree.NewDocument()
	err := doc.ReadFromString(vastXML)
	if nil != err {
		err = fmt.Errorf("account:[%s] bidder:[%s] err:[vast_xml_parsing_failed:%s] vast:[%s] ", accountID, requestingBidder, err.Error(), vastXML)
		glog.Error(err.Error())
		return vastXML, err // false indicates events trackers are not injected
	}

	//Maintaining BidRequest Impression Map (Copied from exchange.go#applyCategoryMapping)
	//TODO: It should be optimized by forming once and reusing
	impMap := make(map[string]*openrtb2.Imp)
	for i := range bidRequest.Imp {
		impMap[bidRequest.Imp[i].ID] = &bidRequest.Imp[i]
	}

	eventURLMap := GetVideoEventTracking(trackerURL, bid, prebidGenBidId, requestingBidder, bidderCoreName, accountID, timestamp, bidRequest, impMap)
	trackersInjected := false
	// return if if no tracking URL
	if len(eventURLMap) == 0 {
		return vastXML, errors.New("event URLs not found")
	}

	creatives := FindCreatives(doc)

	if adm := strings.TrimSpace(bid.AdM); adm == "" || strings.HasPrefix(adm, "http") {
		// determine which creative type to be created based on linearity
		if imp, ok := impMap[bid.ImpID]; ok && nil != imp.Video {
			// create creative object
			creatives = doc.FindElements("VAST/Ad/Wrapper/Creatives")
			// var creative *etree.Element
			// if len(creatives) > 0 {
			// 	creative = creatives[0] // consider only first creative
			// } else {
			creative := doc.CreateElement("Creative")
			creatives[0].AddChild(creative)

			// }

			switch imp.Video.Linearity {
			case adcom1.LinearityLinear:
				creative.AddChild(doc.CreateElement("Linear"))
			case adcom1.LinearityNonLinear:
				creative.AddChild(doc.CreateElement("NonLinearAds"))
			default: // create both type of creatives
				creative.AddChild(doc.CreateElement("Linear"))
				creative.AddChild(doc.CreateElement("NonLinearAds"))
			}
			creatives = creative.ChildElements() // point to actual cratives
		}
	}
	for _, creative := range creatives {
		trackingEventsXML := creative.SelectElement("TrackingEvents")
		if trackingEventsXML == nil {
			trackingEventsXML = creative.CreateElement("TrackingEvents")
			creative.AddChild(trackingEventsXML)
		}
		// Inject - using trackingEvents instead of map to keep output xml predictable. (sequencing in map is not guaranteed)
		for _, event := range trackingEvents {
			trackingEle := trackingEventsXML.CreateElement("Tracking")
			trackingEle.CreateAttr("event", event)
			trackingEle.SetText(eventURLMap[event])
			trackersInjected = true
		}
	}

	if trackersInjected {
		out, err := doc.WriteToBytes()
		if err != nil {
			glog.Errorf("%v", err.Error())
		}
		return string(out), err
	}
	return vastXML, nil
}

// GetVideoEventTracking returns map containing key as event name value as associaed video event tracking URL
// By default PBS will expect [EVENT_ID] macro in trackerURL to inject event information
// [EVENT_ID] will be injected with one of the following values
//
//	firstQuartile, midpoint, thirdQuartile, complete
//
// If your company can not use [EVENT_ID] and has its own macro. provide config.TrackerMacros implementation
// and ensure that your macro is part of trackerURL configuration
func GetVideoEventTracking(trackerURL string, bid *openrtb2.Bid, prebidGenBidId, requestingBidder string, bidderCoreName string, accountId string, timestamp int64, req *openrtb2.BidRequest, impMap map[string]*openrtb2.Imp) map[string]string {
	if strings.TrimSpace(trackerURL) == "" {
		return nil
	}

	// replace standard macros
	// NYC shall we put all macros with their default values here?
	macroMap := map[string]string{
		PBSAdUnitIDMacro:       "",
		PBSBidIDMacro:          bid.ID,
		PBSOrigBidIDMacro:      bid.ID,
		PBSBidderMacro:         bidderCoreName,
		PBSBidderCodeMacro:     requestingBidder,
		PBSAdvertiserNameMacro: "",
		VASTAdTypeMacro:        string(openrtb_ext.BidTypeVideo),
	}

	/* Use generated bidId if present, else use bid.ID */
	if len(prebidGenBidId) > 0 && prebidGenBidId != bid.ID {
		macroMap[PBSBidIDMacro] = prebidGenBidId
	}

	if impMap != nil {
		if imp, ok := impMap[bid.ImpID]; ok {
			macroMap[PBSAdUnitIDMacro] = imp.TagID
		} else {
			glog.Warningf("Setting empty value for %s macro, as failed to determine imp.TagID for bid.ImpID: %s", PBSAdUnitIDMacro, bid.ImpID)
		}
	}

	if len(bid.ADomain) > 0 {
		var err error
		//macroMap[PBSAdvertiserNameMacro] = strings.Join(bid.ADomain, ",")
		macroMap[PBSAdvertiserNameMacro], err = extractDomain(bid.ADomain[0])
		if err != nil {
			glog.Warningf("Unable to extract domain from '%s'. [%s]", bid.ADomain[0], err.Error())
		}
	}

	if req != nil {
		if req.App != nil {
			// macroMap[VASTAppBundleMacro] = req.App.Bundle
			macroMap[VASTDomainMacro] = req.App.Bundle
			if req.App.Publisher != nil {
				macroMap[PBSAccountMacro] = req.App.Publisher.ID
			}
		}
		if req.Site != nil {
			macroMap[VASTDomainMacro] = getDomain(req.Site)
			macroMap[VASTPageURLMacro] = req.Site.Page
			if req.Site.Publisher != nil {
				macroMap[PBSAccountMacro] = req.Site.Publisher.ID
			}
		}
	}

	// lookup in custom macros - keep this block at last for highest priority
	reqExt := new(openrtb_ext.ExtRequest)
	if req.Ext != nil {
		err := json.Unmarshal(req.Ext, &reqExt)
		if err != nil {
			glog.Warningf("Error in unmarshling req.Ext.Prebid.Vast: [%s]", err.Error())
		}
	}
	for key, value := range reqExt.Prebid.Macros {
		macroMap[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	eventURLMap := make(map[string]string)
	for name, id := range trackingEventIDMap { // NYC check if trackingEvents and macroMap can be clubbed
		// replace [EVENT_ID] macro with PBS defined event ID
		macroMap[PBSEventIDMacro] = id
		eventURLMap[name] = replaceMacros(trackerURL, macroMap)
	}
	return eventURLMap
}

func replaceMacros(trackerURL string, macroMap map[string]string) string {
	var builder strings.Builder

	for i := 0; i < len(trackerURL); i++ {
		if trackerURL[i] == '[' {
			found := false
			j := i + 1
			for ; j < len(trackerURL); j++ {
				if trackerURL[j] == ']' {
					found = true
					break
				}
			}
			if found {
				n := j + 1
				k := trackerURL[i:n]
				if v, ok := macroMap[k]; ok {
					v = url.QueryEscape(v) // NYC move QueryEscape while creating map, no need to do this everytime
					_, _ = builder.Write([]byte(v))
					i = j
					continue
				}
			}
		}
		_ = builder.WriteByte(trackerURL[i])
	}

	return builder.String()
}

// FindCreatives finds Linear, NonLinearAds fro InLine and Wrapper Type of creatives
// from input doc - VAST Document
// NOTE: This function is temporarily seperated to reuse in ctv_auction.go. Because, in case of ctv
// we generate bid.id
func FindCreatives(doc *etree.Document) []*etree.Element {
	// Find Creatives of Linear and NonLinear Type
	// Injecting Tracking Events for Companion is not supported here
	creatives := doc.FindElements("VAST/Ad/InLine/Creatives/Creative/Linear")
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/Linear")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/InLine/Creatives/Creative/NonLinearAds")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/NonLinearAds")...)
	return creatives
}

func extractDomain(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}
	// decode rawURL
	rawURL, err := url.QueryUnescape(rawURL)
	if nil != err {
		return "", err
	}
	url, err := url.Parse(rawURL)
	if nil != err {
		return "", err
	}
	// remove www if present
	return strings.TrimPrefix(url.Hostname(), "www."), nil
}

func getDomain(site *openrtb2.Site) string {
	if site.Domain != "" {
		return site.Domain
	}

	hostname := ""

	if site.Page != "" {
		pageURL, err := url.Parse(site.Page)
		if err == nil && pageURL != nil {
			hostname = pageURL.Host
		}
	}
	return hostname
}
