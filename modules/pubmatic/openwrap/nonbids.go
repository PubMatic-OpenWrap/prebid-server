package openwrap

import (
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// prepareSeatNonBids forms the rctx.SeatNonBids map from rctx values
// currently, this function prepares and returns nonbids for partner-throttle and slot-not-mapped errors
func prepareSeatNonBids(rctx models.RequestCtx) map[string][]openrtb_ext.NonBid {

	seatNonBids := make(map[string][]openrtb_ext.NonBid, 0)
	for impID, impCtx := range rctx.ImpBidCtx {
		// seat-non-bid for partner-throttled error
		for bidder := range rctx.AdapterThrottleMap {
			seatNonBids[bidder] = append(seatNonBids[bidder], openrtb_ext.NonBid{
				ImpId:      impID,
				StatusCode: int(exchange.RequestBlockedPartnerThrottle),
			})
		}
		// seat-non-bid for slot-not-mapped error
		// Note : Throttled partner will not be a part of impCtx.NonMapped
		for bidder := range impCtx.NonMapped {
			seatNonBids[bidder] = append(seatNonBids[bidder], openrtb_ext.NonBid{
				ImpId:      impID,
				StatusCode: int(exchange.RequestBlockedSlotNotMapped),
			})
		}
	}
	return seatNonBids
}

// addSeatNonBidsInResponseExt adds the rctx.SeatNonBids in the response-ext
func addSeatNonBidsInResponseExt(rctx models.RequestCtx, responseExt *openrtb_ext.ExtBidResponse) {
	if len(rctx.SeatNonBids) == 0 {
		return
	}

	if responseExt.Prebid == nil {
		responseExt.Prebid = new(openrtb_ext.ExtResponsePrebid)
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

// addLostToDealBidNonBRCode function sets the NonBR code of all lost-bids not satisfying dealTier to LossBidLostToDealBid
func addLostToDealBidNonBRCode(rctx *models.RequestCtx) {
	if !rctx.SupportDeals {
		return
	}

	for impID := range rctx.ImpBidCtx {
		winBid, ok := rctx.WinningBids[impID]
		if !ok {
			continue
		}

		for bidID, bidCtx := range rctx.ImpBidCtx[impID].BidCtx {
			// do not update NonBR for winning bid
			if winBid.ID == bidID {
				continue
			}

			bidDealTierSatisfied := false
			if bidCtx.BidExt.Prebid != nil {
				bidDealTierSatisfied = bidCtx.BidExt.Prebid.DealTierSatisfied
			}
			// do not update NonBr if lost-bid satisfies dealTier
			// because it can have NonBr reason as LossBidLostToHigherBid
			if bidDealTierSatisfied {
				continue
			}
			bidCtx.BidExt.Nbr = GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid)
			rctx.ImpBidCtx[impID].BidCtx[bidID] = bidCtx
		}
	}
}
