package events

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
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

	// parse VAST
	doc := etree.NewDocument()
	err := doc.ReadFromString(vastXML)
	if nil != err {
		err = fmt.Errorf("account:[%s] bidder:[%s] err:[vast_xml_parsing_failed:%s] vast:[%s] ", accountID, requestingBidder, err.Error(), vastXML)
		glog.Error(err.Error())
		return []byte(vastXML), false, err // false indicates events trackers are not injected
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
		out, err := doc.WriteToString()
		if err != nil {
			glog.Errorf("%v", err.Error())
		}
		return out, err
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

func injectVideoEventsETree(vastXML string, eventURLMap map[string]string, nurl bool, linearity adcom1.LinearityMode) (string, error) {
	var trackersInjected bool

	// parse VAST
	doc := etree.NewDocument()
	if err := doc.ReadFromString(vastXML); err != nil {
		return vastXML, err
	}

	creatives := FindCreatives(doc)
	if nurl {
		// create creative object
		creatives = doc.FindElements("VAST/Ad/Wrapper/Creatives")
		creative := doc.CreateElement("Creative")
		creatives[0].AddChild(creative)

		switch linearity {
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

	for _, creative := range creatives {
		trackingEvents := creative.SelectElement("TrackingEvents")
		if nil == trackingEvents {
			trackingEvents = creative.CreateElement("TrackingEvents")
			creative.AddChild(trackingEvents)
		}
		// Inject
		for _, event := range cTrackingEvents {
			if url, ok := eventURLMap[event]; ok {
				trackingEle := trackingEvents.CreateElement("Tracking")
				trackingEle.CreateAttr("event", event)
				trackingEle.SetText(url)
				trackersInjected = true
			}
		}
	}

	if !trackersInjected {
		return vastXML, nil
	}

	out, err := doc.WriteToBytes()
	if err != nil {
		return vastXML, err
	}
	return string(out), err
}

func injectVideoEventsFastXML(vastXML string, eventURLMap map[string]string, nurl bool, linearity adcom1.LinearityMode) (string, error) {

	//parse vast xml
	doc := fastxml.NewXMLReader(nil)
	if err := doc.Parse([]byte(vastXML)); err != nil {
		return vastXML, err
	}

	xu := fastxml.NewXMLUpdater([]byte(vastXML))

	// Find creatives
	creatives := doc.SelectElements(nil, "VAST", "Ad", "InLine", "Creatives", "Creative")
	creatives = append(creatives, doc.SelectElements(nil, "VAST", "Ad", "Wrapper", "Creatives", "Creative")...)

	for _, creative := range creatives {
		childs := doc.Childrens(creative)

		found := false
		for _, linearityElement := range childs {
			name := doc.Name(linearityElement)
			if !(name == "Linear" || name == "NonLinearAds") {
				continue
			}
			found = true

			createTrackingEvents := false
			parent := doc.SelectElement(linearityElement, "TrackingEvents")
			if parent == nil {
				createTrackingEvents = true
				parent = linearityElement //Linear/NonLinearAds
			}

			xu.AppendElement(parent, getTrackingEvents(createTrackingEvents, eventURLMap))
		}

		if !found || (len(childs) == 0 && nurl) {
			switch linearity {
			case adcom1.LinearityLinear:
				xu.AppendElement(creative, fastxml.CreateElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))

			case adcom1.LinearityNonLinear:
				xu.AppendElement(creative, fastxml.CreateElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))

			default:
				xu.AppendElement(creative, fastxml.CreateElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))
				xu.AppendElement(creative, fastxml.CreateElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))
			}
			continue
		}
	}

	// wrap cdata
	doc.Iterate(func(element *fastxml.Element) {
		if doc.IsLeaf(element) && !doc.IsCDATA(element) {
			if text := doc.RawText(element); len(text) > 0 {
				xu.UpdateText(element, text, true, fastxml.XMLUnescapeMode)
			}
		}
	})

	var buf bytes.Buffer
	xu.Build(&buf)
	return buf.String(), nil
}

func getTrackingEvents(createTrackingEvents bool, eventURLMap map[string]string) *fastxml.XMLElement {
	te := fastxml.CreateElement("")
	if createTrackingEvents {
		te.SetName("TrackingEvents")
	}

	for _, event := range cTrackingEvents {
		if url, ok := eventURLMap[event]; ok {
			tracking := fastxml.CreateElement("Tracking").AddAttribute("", "event", event).SetText(url, true)
			te.AddChild(tracking)
		}
	}
	return te
}
