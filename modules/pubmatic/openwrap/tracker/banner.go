package tracker

import (
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func injectBannerTracker(rctx models.RequestCtx, tracker models.OWTracker, bid openrtb2.Bid, seat string) string {
	var replacedTrackerStr, trackerFormat string
	trackerFormat = models.TrackerCallWrap
	if trackerWithOM(tracker, rctx.Platform, seat) {
		trackerFormat = models.TrackerCallWrapOMActive
	}
	replacedTrackerStr = strings.Replace(trackerFormat, "${escapedUrl}", tracker.TrackerURL, 1)
	return bid.AdM + replacedTrackerStr
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
