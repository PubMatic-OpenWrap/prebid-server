package openwrap

import (
	"fmt"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (ow *OpenWrap) cacheRequest(rctx models.RequestCtx) {
	if rctx.IsTestRequest != 0 || rctx.ABTestConfig != 0 || rctx.ABTestConfigApplied != 0 || rctx.AdapterThrottleMap != nil || rctx.PageURL == "" {
		return
	}

	rctx.LoggerImpressionID = ""
	rctx.IP = ""
	rctx.StartTime = 0
	rctx.Trackers = nil
	rctx.ResponseExt = openrtb_ext.ExtBidResponse{}
	rctx.WinningBids = nil
	rctx.DroppedBids = nil
	rctx.DefaultBids = nil
	rctx.SeatNonBids = nil
	rctx.BidderResponseTimeMillis = nil
	rctx.MatchedImpression = nil
	rctx.CustomDimensions = nil
}

func (ow *OpenWrap) getCachedRequest(rctx models.RequestCtx) (models.RequestCtx, bool) {
	if rctx.Platform == models.PLATFORM_APP {
		storedRequestKey := fmt.Sprintf("%s%s%d%s%s", rctx.PubIDStr, rctx.ProfileIDStr, rctx.VersionID, rctx.PageURL, rctx.Source)
		storedRCtx, ok := ow.cache.Get(storedRequestKey)
		if ok {
			if newRctx, ok := storedRCtx.(models.RequestCtx); ok {
				newRctx.StartTime = rctx.StartTime
				newRctx.IP = rctx.IP
				newRctx.LoggerImpressionID = rctx.LoggerImpressionID
				return newRctx, true
			}
		}
	}
	return rctx, false
}
