package tracker

import (
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func InjectTrackers(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	var errs error
	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			var errMsg string
			var err error
			tracker := rctx.Trackers[bid.ID]
			adformat := tracker.BidType
			if rctx.Platform == models.PLATFORM_VIDEO {
				adformat = "video"
			}

			switch adformat {
			case models.Banner:
				bidResponse.SeatBid[i].Bid[j].AdM = injectBannerTracker(rctx, tracker, bid, seatBid.Seat)
			case models.Video:
				trackers := []models.OWTracker{tracker}
				bidResponse.SeatBid[i].Bid[j].AdM, err = injectVideoCreativeTrackers(bid, trackers)
			case models.Native:
				if impBidCtx, ok := rctx.ImpBidCtx[bid.ImpID]; ok {
					bidResponse.SeatBid[i].Bid[j].AdM, err = injectNativeCreativeTrackers(impBidCtx.Native, bid.AdM, tracker)
				} else {
					errMsg = fmt.Sprintf("native obj not found for impid %s", bid.ImpID)
				}
			default:
				errMsg = fmt.Sprintf("Invalid adformat %s for bidid %s", adformat, bid.ID)
			}

			if err != nil {
				errMsg = fmt.Sprintf("failed to inject tracker for bidid %s with error %s", bid.ID, err.Error())
			}
			if errMsg != "" {
				rctx.MetricsEngine.RecordInjectTrackerErrorCount(adformat, rctx.PubIDStr, seatBid.Seat)
				errs = models.ErrorWrap(errs, errors.New(errMsg))
			}

		}
	}
	return bidResponse, errs
}
