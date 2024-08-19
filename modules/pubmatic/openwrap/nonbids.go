package openwrap

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// prepareSeatNonBids forms the rctx.SeatNonBids map from rctx values
// currently, this function prepares and returns nonbids for partner-throttle and slot-not-mapped errors
// prepareSeatNonBids forms the rctx.SeatNonBids map from rctx values
// currently, this function prepares and returns nonbids for partner-throttle and slot-not-mapped errors
func prepareSeatNonBids(rctx models.RequestCtx) openrtb_ext.NonBidCollection {

	var seatNonBid openrtb_ext.NonBidCollection
	for impID, impCtx := range rctx.ImpBidCtx {
		// seat-non-bid for partner-throttled error
		for bidder := range rctx.AdapterThrottleMap {
			nonBid := openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{Bid: &openrtb2.Bid{ImpID: impID}, NonBidReason: int(nbr.RequestBlockedPartnerThrottle)})
			seatNonBid.AddBid(nonBid, bidder)

		}

		// seat-non-bid for slot-not-mapped error
		// Note : Throttled partner will not be a part of impCtx.NonMapped
		for bidder := range impCtx.NonMapped {
			nonBid := openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{Bid: &openrtb2.Bid{ImpID: impID}, NonBidReason: int(nbr.RequestBlockedSlotNotMapped)})
			seatNonBid.AddBid(nonBid, bidder)
		}
	}
	return seatNonBid
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

	for impID, impCtx := range rctx.ImpBidCtx {
		// Do not update the nbr in case of adpod bids
		if impCtx.AdpodConfig != nil {
			continue
		}

		_, ok := rctx.WinningBids[impID]
		if !ok {
			continue
		}

		for bidID, bidCtx := range impCtx.BidCtx {
			// do not update NonBR for winning bid
			if rctx.WinningBids.IsWinningBid(impID, bidID) {
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
			bidCtx.BidExt.Nbr = nbr.LossBidLostToDealBid.Ptr()
			rctx.ImpBidCtx[impID].BidCtx[bidID] = bidCtx
		}
	}
}

func getSeatNonBid(Bidders map[string]struct{}, payload hookstage.BeforeValidationRequestPayload) openrtb_ext.NonBidCollection {
	var seatNonBids openrtb_ext.NonBidCollection
	for bidderName := range Bidders {
		for _, imp := range payload.BidRequest.Imp {
			nonBid := openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
				Bid:          &openrtb2.Bid{ImpID: imp.ID},
				NonBidReason: int(nbr.RequestBlockedPartnerFiltered),
			})
			seatNonBids.AddBid(nonBid, bidderName)
		}
	}
	return seatNonBids
}
