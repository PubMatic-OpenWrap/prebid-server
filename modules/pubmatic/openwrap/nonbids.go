package openwrap

import (
	"github.com/prebid/prebid-server/exchange"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// newProxyNonBid create and returns proxy nonbid for given impid and reason combination
func newProxyNonBid(impId string, nonBidReason int) openrtb_ext.NonBid {
	return openrtb_ext.NonBid{
		ImpId:      impId,
		StatusCode: nonBidReason,
	}
}

// prepareSeatNonBids forms the rctx.SeatNonBids map from rctx values
// currently, this function prepares nonbids for partner-throttle and slot-not-mapped errors
func prepareSeatNonBids(rctx models.RequestCtx) {

	for impID, impCtx := range rctx.ImpBidCtx {

		// seat-non-bid for partner-throttled error
		for bidder := range rctx.AdapterThrottleMap {
			if rctx.SeatNonBids[bidder] == nil {
				rctx.SeatNonBids[bidder] = []openrtb_ext.NonBid{}
			}
			nonBid := newProxyNonBid(impID, int(exchange.RequestBlockedPartnerThrottle))
			rctx.SeatNonBids[bidder] = append(rctx.SeatNonBids[bidder], nonBid)
		}

		// seat-non-bid for slot-not-mapped error
		// Note : Throttled partner will not be a part of impCtx.NonMapped
		for bidder := range impCtx.NonMapped {
			if rctx.SeatNonBids[bidder] == nil {
				rctx.SeatNonBids[bidder] = []openrtb_ext.NonBid{}
			}
			nonBid := newProxyNonBid(impID, int(exchange.RequestBlockedSlotNotMapped))
			rctx.SeatNonBids[bidder] = append(rctx.SeatNonBids[bidder], nonBid)
		}
	}
}

// addSeatNonBidsInResponseExt adds the rctx.SeatNonBids in the response-ext
func addSeatNonBidsInResponseExt(rctx models.RequestCtx, responseExt *openrtb_ext.ExtBidResponse) {
	if responseExt == nil || len(rctx.SeatNonBids) == 0 {
		return
	}

	if responseExt.Prebid == nil {
		responseExt.Prebid = &openrtb_ext.ExtResponsePrebid{}
	}

	if responseExt.Prebid.SeatNonBid == nil {
		responseExt.Prebid.SeatNonBid = make([]openrtb_ext.SeatNonBid, 0)
	}

	for index, seatnonbid := range responseExt.Prebid.SeatNonBid {
		// if response-ext contains list of nonbids for bidder then
		// add the rctx-nonbids to the same list
		if nonBids, found := rctx.SeatNonBids[seatnonbid.Seat]; found {
			responseExt.Prebid.SeatNonBid[index].NonBid = append(responseExt.Prebid.SeatNonBid[index].NonBid, nonBids...)
			delete(rctx.SeatNonBids, seatnonbid.Seat)
		}
	}

	// at this point, rctx.SeatNonBids will contain nonbids for only those seat/bidder which are not part of response-ext
	for seat, nonBids := range rctx.SeatNonBids {
		responseExt.Prebid.SeatNonBid = append(responseExt.Prebid.SeatNonBid,
			openrtb_ext.SeatNonBid{
				Seat:   seat,
				NonBid: nonBids,
			})
	}
}
