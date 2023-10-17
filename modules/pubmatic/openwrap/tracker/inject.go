package tracker

import (
	"fmt"

	"github.com/pkg/errors"
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
				bidResponse.SeatBid[i].Bid[j].AdM = injectBannerTracker(rctx, tracker, bidResponse.SeatBid[i].Bid[j], seatBid.Seat)
			case models.Video:
				// trackers := make([]models.OWTracker, 0, len(rctx.Trackers))
				// for _, tracker := range rctx.Trackers {
				// 	trackers = append(trackers, tracker)
				// }
				trackers := []models.OWTracker{tracker}

				bidResponse.SeatBid[i].Bid[j].AdM, err = injectVideoCreativeTrackers(bid, trackers)
				if err != nil {
					errMsg = fmt.Sprintf("failed to inject tracker for bidid %s with error %s", bid.ID, err.Error())
				}
			case models.Native:
				// bidResponse.SeatBid[i].Bid[j].AdM, err = injectNativeCreativeTrackers(rctx.ImpBidCtx[bid.ImpID].Native, bidResponse.SeatBid[i].Bid[j].AdM, tracker)
			default:
				errMsg = fmt.Sprintf("Invalid adformat %s for bidid %s", adformat, bid.ID)
			}

			if errMsg != "" {
				rctx.MetricsEngine.RecordInjectTrackerErrorCount(adformat, rctx.PubIDStr, seatBid.Seat)
				errs = errors.Wrap(errs, errMsg)
			}

		}
	}
	return bidResponse, errs
}
