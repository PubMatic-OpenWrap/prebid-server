package tracker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"
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
		if rctx.Endpoint == models.EndpointVAST {
			creative = UpdateAdvAndCatTags(creative, strictVastMode, bid.ADomain, bid.Cat, nil)
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
		if err != nil {
			//injection failure
			return bid.AdM, bid.BURL, errors.New("invalid creative format")
		}

		// Add advertiser domain and category tags if strictVastMode is enabled
		if rctx.Endpoint == models.EndpointVAST {
			creative = UpdateAdvAndCatTags(creative, strictVastMode, bid.ADomain, bid.Cat, ti)
		}

		bid.AdM = creative
	}

	if skipTracker && len(videoParams) > 0 {
		bid.BURL = getBURL(bid.BURL, videoParams[0].TrackerURL)
	}

	return bid.AdM, bid.BURL, nil
}

func UpdateAdvAndCatTags(creative string, strictVastMode bool, adDomain []string, adCat []string, ti trackerInjector) string {
	if !strictVastMode || (len(adDomain) == 0 && len(adCat) == 0) {
		return creative
	}

	if ti != nil {
		var domain string
		if len(adDomain) > 0 {
			domain = adDomain[0]
		}
		updatedCreative, err := ti.UpdateADMWithAdvCat(domain, adCat)
		if err != nil {
			glog.Errorf("creative [%s]: Error updating ADM with advertiser/category:  %s", creative, err.Error())
			return creative
		}
		return updatedCreative
	}

	// fallback for URL-based creative (string replace)
	if len(adDomain) > 0 {
		advertiserTag := fmt.Sprintf("<Advertiser><![CDATA[%s]]></Advertiser>", adDomain[0])
		creative = strings.Replace(creative, "</Wrapper>", advertiserTag+"</Wrapper>", -1)
	}
	if len(adCat) > 0 {
		categoryTag := fmt.Sprintf("<Category><![CDATA[%s]]></Category>", strings.Join(adCat, ","))
		creative = strings.Replace(creative, "</Wrapper>", categoryTag+"</Wrapper>", -1)
	}

	return creative
}
