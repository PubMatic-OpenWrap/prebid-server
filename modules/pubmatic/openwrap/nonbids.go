package openwrap

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

// prepareSeatNonBids forms the rctx.SeatNonBids map from rctx values
// currently, this function prepares and returns nonbids for partner-throttle and slot-not-mapped errors
// prepareSeatNonBids forms the rctx.SeatNonBids map from rctx values
// currently, this function prepares and returns nonbids for partner-throttle and slot-not-mapped errors
func prepareSeatNonBids(rctx models.RequestCtx) openrtb_ext.SeatNonBidBuilder {

	var seatNonBid openrtb_ext.SeatNonBidBuilder
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

	// Update seat-non-bids with default-bids for the web-s2s endpoint
	// In other endpoints, default bids are added to response.seatbid, but for web-s2s, we must return a vanilla prebid	 response.
	if rctx.Endpoint == models.EndpointWebS2S {
		updateSeatNonBidsFromDefaultBids(rctx, &seatNonBid)
	}

	return seatNonBid
}

func updateSeatNonBidsFromDefaultBids(rctx models.RequestCtx, seatNonBid *openrtb_ext.SeatNonBidBuilder) {
	for impID, defaultBid := range rctx.DefaultBids {
		for seat, bids := range defaultBid {
			for _, bid := range bids {
				if rctx.ImpBidCtx != nil && rctx.ImpBidCtx[impID].BidCtx != nil && rctx.ImpBidCtx[impID].BidCtx[bid.ID].Nbr != nil {
					nbr := rctx.ImpBidCtx[impID].BidCtx[bid.ID].Nbr
					nonBid := openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{Bid: &openrtb2.Bid{ImpID: impID}, NonBidReason: int(*nbr)})
					seatNonBid.AddBid(nonBid, seat)
				}

			}
		}
	}
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

func getSeatNonBid(Bidders map[string]struct{}, bidRequest *openrtb2.BidRequest) openrtb_ext.SeatNonBidBuilder {
	var seatNonBids openrtb_ext.SeatNonBidBuilder
	for bidderName := range Bidders {
		for _, imp := range bidRequest.Imp {
			nonBid := openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
				Bid:          &openrtb2.Bid{ImpID: imp.ID},
				NonBidReason: int(nbr.RequestBlockedPartnerFiltered),
			})
			seatNonBids.AddBid(nonBid, bidderName)
		}
	}
	return seatNonBids
}
