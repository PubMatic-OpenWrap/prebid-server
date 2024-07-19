package events

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/PubMatic-OpenWrap/prebid-server/v2/openrtb_ext"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
)

// PubMatic specific event IDs
// This will go in event-config once PreBid modular design is in place
var eventIDMap = map[string]string{
	"start":         "2",
	"firstQuartile": "4",
	"midpoint":      "3",
	"thirdQuartile": "5",
	"complete":      "6",
}

// standard VAST macros
// https://interactiveadvertisingbureau.github.io/vast/vast4macros/vast4-macros-latest.html#macro-spec-adcount
const (
	VASTAdTypeMacro    = "[ADTYPE]"
	VASTAppBundleMacro = "[APPBUNDLE]"
	VASTDomainMacro    = "[DOMAIN]"
	VASTPageURLMacro   = "[PAGEURL]"

	// PBS specific macros
	PBSEventIDMacro        = "[EVENT_ID]"        // PBSEventIDMacro macro for injecting PBS defined  video event tracker id
	PBSAccountMacro        = "[PBS-ACCOUNT]"     // PBSAccountMacro represents publisher id / account id
	PBSBidderMacro         = "[PBS-BIDDER]"      // PBSBidderMacro represents bidder name
	PBSOrigBidIDMacro      = "[PBS-ORIG_BIDID]"  // PBSOrigBidIDMacro represents original bid id.
	PBSBidIDMacro          = "[PBS-BIDID]"       // PBSBidIDMacro represents bid id. If auction.generate-bid-id config is on, then resolve with response.seatbid.bid.ext.prebid.bidid. Else replace with response.seatbid.bid.id
	PBSAdvertiserNameMacro = "[ADVERTISER_NAME]" // [ADERVERTISER_NAME] represents advertiser name
	PBSAdUnitIDMacro       = "[AD_UNIT]"         // PBSAdUnitIDMacro Pass imp.tagId using this macro
	PBSBidderCodeMacro     = "[BIDDER_CODE]"     // PBSBidderCodeMacro represents an alias id or core bidder id.
)

// GetVideoEventTracking returns map containing key as event name value as associaed video event tracking URL
// By default PBS will expect [EVENT_ID] macro in trackerURL to inject event information
// [EVENT_ID] will be injected with one of the following values
//
//	firstQuartile, midpoint, thirdQuartile, complete
//
// If your company can not use [EVENT_ID] and has its own macro. provide config.TrackerMacros implementation
// and ensure that your macro is part of trackerURL configuration
func GetVideoEventTracking(trackerURL string, bid *openrtb2.Bid, prebidGenBidId, requestingBidder string, bidderCoreName string, accountId string, timestamp int64, req *openrtb2.BidRequest, impMap map[string]*openrtb2.Imp) map[string]string {
	eventURLMap := make(map[string]string)
	if len(strings.TrimSpace(trackerURL)) == 0 {
		return eventURLMap
	}

	// lookup custom macros
	var customMacroMap map[string]string
	if nil != req.Ext {
		reqExt := new(openrtb_ext.ExtRequest)
		err := json.Unmarshal(req.Ext, &reqExt)
		if err == nil {
			customMacroMap = reqExt.Prebid.Macros
		} else {
			glog.Warningf("Error in unmarshling req.Ext.Prebid.Vast: [%s]", err.Error())
		}
	}

	for _, event := range cTrackingEvents {
		eventURL := trackerURL
		// lookup in custom macros
		if nil != customMacroMap {
			for customMacro, value := range customMacroMap {
				eventURL = replaceMacro(eventURL, customMacro, value)
			}
		}
		// replace standard macros
		eventURL = replaceMacro(eventURL, VASTAdTypeMacro, string(openrtb_ext.BidTypeVideo))
		if nil != req && nil != req.App {
			// eventURL = replaceMacro(eventURL, VASTAppBundleMacro, req.App.Bundle)
			eventURL = replaceMacro(eventURL, VASTDomainMacro, req.App.Bundle)
			if nil != req.App.Publisher {
				eventURL = replaceMacro(eventURL, PBSAccountMacro, req.App.Publisher.ID)
			}
		}
		if nil != req && nil != req.Site {
			eventURL = replaceMacro(eventURL, VASTDomainMacro, getDomain(req.Site))
			eventURL = replaceMacro(eventURL, VASTPageURLMacro, req.Site.Page)
			if nil != req.Site.Publisher {
				eventURL = replaceMacro(eventURL, PBSAccountMacro, req.Site.Publisher.ID)
			}
		}

		domain := ""
		if len(bid.ADomain) > 0 {
			var err error
			//eventURL = replaceMacro(eventURL, PBSAdvertiserNameMacro, strings.Join(bid.ADomain, ","))
			domain, err = extractDomain(bid.ADomain[0])
			if err != nil {
				glog.Warningf("Unable to extract domain from '%s'. [%s]", bid.ADomain[0], err.Error())
			}
		}

		eventURL = replaceMacro(eventURL, PBSAdvertiserNameMacro, domain)

		eventURL = replaceMacro(eventURL, PBSBidderMacro, bidderCoreName)
		eventURL = replaceMacro(eventURL, PBSBidderCodeMacro, requestingBidder)

		/* Use generated bidId if present, else use bid.ID */
		if len(prebidGenBidId) > 0 && prebidGenBidId != bid.ID {
			eventURL = replaceMacro(eventURL, PBSBidIDMacro, prebidGenBidId)
		} else {
			eventURL = replaceMacro(eventURL, PBSBidIDMacro, bid.ID)
		}
		eventURL = replaceMacro(eventURL, PBSOrigBidIDMacro, bid.ID)

		// replace [EVENT_ID] macro with PBS defined event ID
		eventURL = replaceMacro(eventURL, PBSEventIDMacro, eventIDMap[event])

		if imp, ok := impMap[bid.ImpID]; ok {
			eventURL = replaceMacro(eventURL, PBSAdUnitIDMacro, imp.TagID)
		} else {
			glog.Warningf("Setting empty value for %s macro, as failed to determine imp.TagID for bid.ImpID: %s", PBSAdUnitIDMacro, bid.ImpID)
			eventURL = replaceMacro(eventURL, PBSAdUnitIDMacro, "")
		}

		eventURLMap[event] = eventURL
	}
	return eventURLMap
}

func replaceMacro(trackerURL, macro, value string) string {
	macro = strings.TrimSpace(macro)
	trimmedValue := strings.TrimSpace(value)

	if strings.HasPrefix(macro, "[") && strings.HasSuffix(macro, "]") && len(trimmedValue) > 0 {
		trackerURL = strings.ReplaceAll(trackerURL, macro, url.QueryEscape(value))
	} else if strings.HasPrefix(macro, "[") && strings.HasSuffix(macro, "]") && len(trimmedValue) == 0 {
		trackerURL = strings.ReplaceAll(trackerURL, macro, url.QueryEscape(""))
	} else {
		glog.Warningf("Invalid macro '%v'. Either empty or missing prefix '[' or suffix ']", macro)
	}
	return trackerURL
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
