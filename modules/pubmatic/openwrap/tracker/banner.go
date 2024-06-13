package tracker

import (
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func injectBannerTracker(rctx models.RequestCtx, tracker models.OWTracker, bid openrtb2.Bid, seat string, pixels []adunitconfig.UniversalPixel) (string, string) {
	if rctx.Endpoint == models.EndpointAppLovinMax {
		return bid.AdM, getBURL(bid.BURL, tracker.TrackerURL)
	}

	var replacedTrackerStr, trackerFormat string
	trackerFormat = models.TrackerCallWrap
	if trackerWithOM(tracker, rctx.Platform, seat) {
		trackerFormat = models.TrackerCallWrapOMActive
	}
	replacedTrackerStr = strings.Replace(trackerFormat, "${escapedUrl}", tracker.TrackerURL, 1)
	adm := applyTBFFeature(rctx, bid, replacedTrackerStr)
	return appendUPixelinBanner(adm, pixels), bid.BURL
}

// append universal pixels in creative based on conditions
func appendUPixelinBanner(adm string, universalPixel []adunitconfig.UniversalPixel) string {
	if universalPixel == nil {
		return adm
	}

	for _, pixelVal := range universalPixel {
		if pixelVal.Pos == models.PixelPosAbove {
			adm = pixelVal.Pixel + adm
			continue
		}
		adm = adm + pixelVal.Pixel
	}
	return adm
}

// TrackerWithOM checks for OM active condition for DV360
func trackerWithOM(tracker models.OWTracker, platform, bidderCode string) bool {
	if platform == models.PLATFORM_APP && bidderCode == string(openrtb_ext.BidderPubmatic) {
		if tracker.DspId == models.DspId_DV360 {
			return true
		}
	}
	return false
}

// applyTBFFeature adds the tracker before or after the actual bid.Adm
// If TBF feature is applicable based on database-configuration for
// given pub-prof combination then injects the tracker before adm
// else injects the tracker after adm.
func applyTBFFeature(rctx models.RequestCtx, bid openrtb2.Bid, tracker string) string {
	if rctx.IsTBFFeatureEnabled {
		return tracker + bid.AdM
	}
	return bid.AdM + tracker
}
