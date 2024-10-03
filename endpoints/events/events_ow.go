package events

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// standard VAST macros
// https://interactiveadvertisingbureau.github.io/vast/vast4macros/vast4-macros-latest.html#macro-spec-adcount
const (
	VASTAdTypeMacro        = "[ADTYPE]"          //VASTAdTypeMacro openwrap macro for ADTYPE
	VASTAppBundleMacro     = "[APPBUNDLE]"       //VASTAppBundleMacro openwrap macro for APPBUNDLE
	VASTDomainMacro        = "[DOMAIN]"          //VASTDomainMacro openwrap macro for DOMAIN
	VASTPageURLMacro       = "[PAGEURL]"         //VASTPageURLMacro openwrap macro for PAGEURL
	PBSEventIDMacro        = "[EVENT_ID]"        // PBSEventIDMacro macro for injecting PBS defined  video event tracker id
	PBSAccountMacro        = "[PBS-ACCOUNT]"     // PBSAccountMacro represents publisher id / account id
	PBSBidderMacro         = "[PBS-BIDDER]"      // PBSBidderMacro represents bidder name
	PBSOrigBidIDMacro      = "[PBS-ORIG_BIDID]"  // PBSOrigBidIDMacro represents original bid id.
	PBSBidIDMacro          = "[PBS-BIDID]"       // PBSBidIDMacro represents bid id. If auction.generate-bid-id config is on, then resolve with response.seatbid.bid.ext.prebid.bidid. Else replace with response.seatbid.bid.id
	PBSAdvertiserNameMacro = "[ADVERTISER_NAME]" // [ADERVERTISER_NAME] represents advertiser name
	PBSAdUnitIDMacro       = "[AD_UNIT]"         // PBSAdUnitIDMacro Pass imp.tagId using this macro
	PBSBidderCodeMacro     = "[BIDDER_CODE]"     // PBSBidderCodeMacro represents an alias id or core bidder id.
)

// PubMatic specific event IDs
// This will go in event-config once PreBid modular design is in place
var trackingEventIDMap = map[string]string{
	"start":         "2",
	"firstQuartile": "4",
	"midpoint":      "3",
	"thirdQuartile": "5",
	"complete":      "6",
}

var trackingEvents = []string{"start", "firstQuartile", "midpoint", "thirdQuartile", "complete"}

// GetVideoEventTracking returns map containing key as event name value as associaed video event tracking URL
// By default PBS will expect [EVENT_ID] macro in trackerURL to inject event information
// [EVENT_ID] will be injected with one of the following values
//
//	firstQuartile, midpoint, thirdQuartile, complete
//
// If your company can not use [EVENT_ID] and has its own macro. provide config.TrackerMacros implementation
// and ensure that your macro is part of trackerURL configuration
// GetVideoEventTracking returns map containing key as event name value as associaed video event tracking URL
// By default PBS will expect [EVENT_ID] macro in trackerURL to inject event information
// [EVENT_ID] will be injected with one of the following values
//
//	firstQuartile, midpoint, thirdQuartile, complete
//
// If your company can not use [EVENT_ID] and has its own macro. provide config.TrackerMacros implementation
// and ensure that your macro is part of trackerURL configuration
func GetVideoEventTracking(
	req *openrtb2.BidRequest,
	imp *openrtb2.Imp,
	bid *openrtb2.Bid,
	trackerURL string,
	prebidGenBidId, requestingBidder, bidderCoreName string,
	timestamp int64) map[string]string {

	if req == nil || imp == nil || bid == nil || strings.TrimSpace(trackerURL) == "" {
		return nil
	}

	// replace standard macros
	// NYC shall we put all macros with their default values here?
	macroMap := map[string]string{
		PBSAdUnitIDMacro:       imp.TagID,
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

	if len(bid.ADomain) > 0 {
		var err error
		//macroMap[PBSAdvertiserNameMacro] = strings.Join(bid.ADomain, ",")
		macroMap[PBSAdvertiserNameMacro], err = extractDomain(bid.ADomain[0])
		if err != nil {
			glog.Warningf("Unable to extract domain from '%s'. [%s]", bid.ADomain[0], err.Error())
		}
	}

	if req.App != nil {
		// macroMap[VASTAppBundleMacro] = req.App.Bundle
		macroMap[VASTDomainMacro] = req.App.Bundle
		if req.App.Publisher != nil {
			macroMap[PBSAccountMacro] = req.App.Publisher.ID
		}
	} else if req.Site != nil {
		macroMap[VASTDomainMacro] = getDomain(req.Site)
		macroMap[VASTPageURLMacro] = req.Site.Page
		if req.Site.Publisher != nil {
			macroMap[PBSAccountMacro] = req.Site.Publisher.ID
		}
	}

	// lookup in custom macros - keep this block at last for highest priority
	var reqExt openrtb_ext.ExtRequest
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
