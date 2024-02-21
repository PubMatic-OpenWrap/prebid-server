package openwrap

import (
	"fmt"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (ow *OpenWrap) cacheRequest(rctx models.RequestCtx) {
	if !ow.cfg.Features.AppRequestCache ||
		rctx.IsTestRequest != 0 || rctx.ABTestConfig != 0 || rctx.ABTestConfigApplied != 0 || rctx.AdapterThrottleMap != nil ||
		rctx.PageURL == "" || rctx.App == nil || rctx.App.Bundle == "" || rctx.App.ID == "" {
		return
	}

	rctx.StartTime = 0
	rctx.IP = ""
	rctx.LoggerImpressionID = ""
	rctx.Trackers = make(map[string]models.OWTracker)
	rctx.ResponseExt = openrtb_ext.ExtBidResponse{}
	rctx.WinningBids = make(map[string]models.OwBid)
	rctx.DroppedBids = make(map[string][]openrtb2.Bid)
	rctx.DefaultBids = make(map[string]map[string][]openrtb2.Bid)
	rctx.SeatNonBids = make(map[string][]openrtb_ext.NonBid)
	rctx.BidderResponseTimeMillis = make(map[string]int)
	rctx.MatchedImpression = make(map[string]int)
	rctx.CustomDimensions = make(map[string]models.CustomDimension)

	ow.cache.Set(ow.getCachedRequestKey(rctx), rctx)
}

func (ow *OpenWrap) getCachedRequest(rctx models.RequestCtx) (models.RequestCtx, bool) {
	if ow.cfg.Features.AppRequestCache && rctx.Platform == models.PLATFORM_APP {

		storedRCtx, ok := ow.cache.Get(ow.getCachedRequestKey(rctx))
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

func (ow *OpenWrap) getCachedRequestKey(rctx models.RequestCtx) string {
	return fmt.Sprintf("%s%s%d%s%s%s", rctx.PubIDStr, rctx.ProfileIDStr, rctx.VersionID, rctx.PageURL, rctx.App.ID, rctx.App.Bundle)
}
