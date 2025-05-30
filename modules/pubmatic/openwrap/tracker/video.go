package tracker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// Inject Trackers in Video Creative
func injectVideoCreativeTrackers(rctx models.RequestCtx, bid openrtb2.Bid, videoParams []models.OWTracker) (string, string, error) {
	if bid.AdM == "" || len(videoParams) == 0 {
		return "", bid.BURL, errors.New("bid is nil or tracker data is missing")
	}

	skipTracker := false
	if rctx.Endpoint == models.EndpointAppLovinMax || rctx.Endpoint == models.EndpointGoogleSDK {
		skipTracker = true
	}

	creative := bid.AdM
	strictVastMode := rctx.NewReqExt != nil && rctx.NewReqExt.Prebid.StrictVastMode
	if strings.HasPrefix(creative, models.HTTPProtocol) {
		creative = strings.Replace(models.VastWrapper, models.PartnerURLPlaceholder, creative, -1)
		if skipTracker {
			creative = strings.Replace(creative, models.VASTImpressionURLTemplate, "", -1)
		} else {
			creative = strings.Replace(creative, models.TrackerPlaceholder, videoParams[0].TrackerURL, -1)
		}
		creative = strings.Replace(creative, models.ErrorPlaceholder, videoParams[0].ErrorURL, -1)
		// Add advertiser domain and category tags if strictVastMode is enabled
		if strictVastMode {
			if len(bid.ADomain) > 0 {
				// Create advertiser domain tag and add it to VAST XML
				advertiserTag := fmt.Sprintf("<Advertiser><![CDATA[%s]]></Advertiser>", bid.ADomain[0])
				// Insert advertiser tag before </Wrapper>
				creative = strings.Replace(creative, "</Wrapper>", advertiserTag+"</Wrapper>", -1)
			}
			// Add category tag if categories are available
			if len(bid.Cat) > 0 {
				// Create category tag and add it to VAST XML
				categoryTag := fmt.Sprintf("<Category><![CDATA[%v]]></Category>", bid.Cat)
				// Insert category tag before </Wrapper>
				creative = strings.Replace(creative, "</Wrapper>", categoryTag+"</Wrapper>", -1)
			}
		}
		bid.AdM = creative
	} else {
		creative = strings.TrimSpace(creative)
		ti := GetTrackerInjector()
		if err := ti.Parse(creative); err != nil {
			//parsing failed
			return bid.AdM, bid.BURL, errors.New("invalid creative format")
		}
		creative, err := ti.Inject(videoParams, skipTracker)
		if strictVastMode {
			creative, err = ti.UpdateADMWithAdvCat(bid.ADomain[0], bid.Cat)
		}
		if err != nil {
			//injection failure
			return bid.AdM, bid.BURL, errors.New("invalid creative format")
		}
		bid.AdM = creative
	}

	if skipTracker && len(videoParams) > 0 {
		bid.BURL = getBURL(bid.BURL, videoParams[0].TrackerURL)
	}

	return bid.AdM, bid.BURL, nil
}
