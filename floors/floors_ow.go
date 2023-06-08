package floors

import (
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/exchange/entities"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func RequestHasFloors(bidRequest *openrtb2.BidRequest) bool {
	for i := range bidRequest.Imp {
		if bidRequest.Imp[i].BidFloor > 0 {
			return true
		}
	}
	return false
}

func PbsOrtbBidToAnalyticsRejectedBid(pbsRejSeatBids []*entities.PbsOrtbSeatBid) []analytics.RejectedBid {
	var rejectedBid []analytics.RejectedBid
	for _, pbsRejSeatBid := range pbsRejSeatBids {
		for _, pbsRejBid := range pbsRejSeatBid.Bids {
			var rejectionReason = openrtb3.LossBidBelowAuctionFloor
			if pbsRejBid.Bid.DealID != "" {
				rejectionReason = openrtb3.LossBidBelowDealFloor
			}
			rejectedBid = append(rejectedBid, analytics.RejectedBid{
				Bid:             pbsRejBid,
				Seat:            pbsRejSeatBid.Seat,
				RejectionReason: rejectionReason,
			})
		}
	}
	return rejectedBid
}

// resolveFloorMin gets floorMin value from request and dynamic fetched data
func resolveFloorMinOW(reqFloors *openrtb_ext.PriceFloorRules, fetchFloors *openrtb_ext.PriceFloorRules, conversions currency.Conversions) Price {
	var requestFloorMinCur, providerFloorMinCur string
	var requestFloorMin, providerFloorMin float64

	if reqFloors != nil {
		requestFloorMin = reqFloors.FloorMin
		requestFloorMinCur = reqFloors.FloorMinCur
		if len(requestFloorMinCur) == 0 && reqFloors.Data != nil {
			requestFloorMinCur = reqFloors.Data.Currency
		}
	}

	if fetchFloors != nil {
		providerFloorMin = fetchFloors.FloorMin
		providerFloorMinCur = fetchFloors.FloorMinCur
		if len(providerFloorMinCur) == 0 && fetchFloors.Data != nil {
			providerFloorMinCur = fetchFloors.Data.Currency
		}
	}

	if len(requestFloorMinCur) > 0 {
		if requestFloorMin > 0 {
			return Price{FloorMin: requestFloorMin, FloorMinCur: requestFloorMinCur}
		}

		if providerFloorMin > 0 {
			if strings.Compare(providerFloorMinCur, requestFloorMinCur) == 0 || len(providerFloorMinCur) == 0 {
				return Price{FloorMin: providerFloorMin, FloorMinCur: requestFloorMinCur}
			}
			rate, err := conversions.GetRate(providerFloorMinCur, requestFloorMinCur)
			if err != nil {
				return Price{FloorMin: 0, FloorMinCur: requestFloorMinCur}
			}
			return Price{FloorMin: roundToFourDecimals(rate * providerFloorMin), FloorMinCur: requestFloorMinCur}
		}
	}

	if len(providerFloorMinCur) > 0 {
		if providerFloorMin > 0 {
			return Price{FloorMin: providerFloorMin, FloorMinCur: providerFloorMinCur}
		}
		if requestFloorMin > 0 {
			return Price{FloorMin: requestFloorMin, FloorMinCur: providerFloorMinCur}
		}
	}

	return Price{FloorMin: 0, FloorMinCur: ""}

}
