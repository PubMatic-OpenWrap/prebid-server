package exchange

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/exchange/entities"
	"github.com/prebid/prebid-server/floors"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// RejectedBid defines the contract for bid rejection errors due to floors enforcement
type RejectedBid struct {
	Bid             *entities.PbsOrtbBid `json:"bid,omitempty"`
	RejectionReason int                  `json:"rejectreason,omitempty"`
	BidderName      string               `json:"biddername,omitempty"`
}

// Check for Floors enforcement for deals,
// In case bid wit DealID present and enforceDealFloors = false then bid floor enforcement should be skipped
func checkDealsForEnforcement(bid *entities.PbsOrtbBid, enforceDealFloors bool) *entities.PbsOrtbBid {
	if bid != nil && bid.Bid != nil && bid.Bid.DealID != "" && !enforceDealFloors {
		return bid
	}
	return nil
}

// Get conversion rate in case floor currency and seatBid currency are not same
func getCurrencyConversionRate(seatBidCur, reqImpCur string, conversions currency.Conversions) (float64, error) {
	rate := 1.0
	if seatBidCur != reqImpCur {
		return conversions.GetRate(seatBidCur, reqImpCur)
	} else {
		return rate, nil
	}
}

func updateBidExtWithFloors(reqImp *openrtb_ext.ImpWrapper, bid *entities.PbsOrtbBid, floorCurrency string) {

	impExt, err := reqImp.GetImpExt()
	if err != nil || impExt == nil {
		return
	}

	prebidExt := impExt.GetPrebid()
	if prebidExt == nil || prebidExt.Floors == nil {
		return
	}

	var bidExtFloors openrtb_ext.ExtBidFloors
	bidExtFloors.FloorRule = prebidExt.Floors.FloorRule
	bidExtFloors.FloorRuleValue = prebidExt.Floors.FloorRuleValue
	bidExtFloors.FloorValue = prebidExt.Floors.FloorValue
	bidExtFloors.FloorCurrency = floorCurrency
	bid.BidFloors = &bidExtFloors
}

// enforceFloorToBids function does floors enforcement for each bid.
//
//	The bids returned by each partner below bid floor price are rejected and remaining eligible bids are considered for further processing
func enforceFloorToBids(bidRequestWrapper *openrtb_ext.RequestWrapper, seatBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, conversions currency.Conversions, enforceDealFloors bool) (map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, []error, []analytics.RejectedBid) {
	errs := []error{}
	rejectedBids := []analytics.RejectedBid{}
	impMap := make(map[string]*openrtb_ext.ImpWrapper, bidRequestWrapper.LenImp())

	//Maintaining BidRequest Impression Map
	for _, v := range bidRequestWrapper.GetImp() {
		impMap[v.ID] = v
	}

	for bidderName, seatBid := range seatBids {
		eligibleBids := make([]*entities.PbsOrtbBid, 0, len(seatBid.Bids))
		for _, bid := range seatBid.Bids {
			retBid := checkDealsForEnforcement(bid, enforceDealFloors)
			if retBid != nil {
				eligibleBids = append(eligibleBids, retBid)
				continue
			}

			reqImp, ok := impMap[bid.Bid.ImpID]
			if !ok {
				continue
			}

			reqImpCur := reqImp.BidFloorCur
			if reqImpCur == "" {
				reqImpCur = "USD"
				if bidRequestWrapper.Cur != nil {
					reqImpCur = bidRequestWrapper.Cur[0]
				}
			}
			rate, err := getCurrencyConversionRate(seatBid.Currency, reqImpCur, conversions)
			if err != nil {
				errMsg := fmt.Errorf("error in rate conversion from = %s to %s with bidder %s for impression id %s and bid id %s", seatBid.Currency, reqImpCur, bidderName, bid.Bid.ImpID, bid.Bid.ID)
				glog.Errorf(errMsg.Error())
				errs = append(errs, errMsg)
				continue
			}

			bidPrice := rate * bid.Bid.Price
			if reqImp.BidFloor > bidPrice {
				rejectedBid := analytics.RejectedBid{
					Bid:  bid,
					Seat: seatBid.Seat,
				}
				rejectedBid.RejectionReason = openrtb3.LossBidBelowAuctionFloor
				if bid.Bid.DealID != "" {
					rejectedBid.RejectionReason = openrtb3.LossBidBelowDealFloor
				}
				rejectedBids = append(rejectedBids, rejectedBid)
				errs = append(errs, fmt.Errorf("bid rejected [bid ID: %s] reason: bid price value %.4f %s is less than bidFloor value %.4f %s for impression id %s bidder %s", bid.Bid.ID, bidPrice, reqImpCur, reqImp.BidFloor, reqImpCur, bid.Bid.ImpID, bidderName))
				continue
			}
			eligibleBids = append(eligibleBids, bid)
		}
		seatBids[bidderName].Bids = eligibleBids
	}
	return seatBids, errs, rejectedBids
}

// getFloorsFlagFromReqExt returns floors enabled flag,
// if floors enabled flag is not provided in request extesion, by default treated as true
func getFloorsFlagFromReqExt(prebidExt *openrtb_ext.ExtRequestPrebid) bool {
	floorEnabled := true
	if prebidExt == nil || prebidExt.Floors == nil || prebidExt.Floors.Enabled == nil {
		return floorEnabled
	}
	return *prebidExt.Floors.Enabled
}

func getEnforceDealsFlag(Floors *openrtb_ext.PriceFloorRules) bool {
	return Floors != nil && Floors.Enforcement != nil && Floors.Enforcement.FloorDeals != nil && *Floors.Enforcement.FloorDeals
}

// eneforceFloors function does floors enforcement
func enforceFloors(r *AuctionRequest, seatBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, floor config.PriceFloors, conversions currency.Conversions, responseDebugAllow bool) (map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, []error) {
	rejectionsErrs := []error{}
	if r == nil || r.BidRequestWrapper == nil {
		return seatBids, rejectionsErrs
	}

	requestExt, err := r.BidRequestWrapper.GetRequestExt()
	if err != nil {
		rejectionsErrs = append(rejectionsErrs, err)
		return seatBids, rejectionsErrs
	}
	prebidExt := requestExt.GetPrebid()
	reqFloorEnable := getFloorsFlagFromReqExt(prebidExt)
	if floor.Enabled && reqFloorEnable && r.Account.PriceFloors.Enabled {
		var enforceDealFloors bool
		var floorsEnfocement bool
		var updateReqExt bool
		updateBidExt(r.BidRequestWrapper, seatBids)
		floorsEnfocement = floors.RequestHasFloors(r.BidRequestWrapper.BidRequest)
		if prebidExt != nil && floorsEnfocement {
			if floorsEnfocement, updateReqExt = floors.ShouldEnforce(prebidExt.Floors, r.Account.PriceFloors.EnforceFloorRate, rand.Intn); floorsEnfocement {
				enforceDealFloors = r.Account.PriceFloors.EnforceDealFloors && getEnforceDealsFlag(prebidExt.Floors)
			}
		}

		if floorsEnfocement {
			rejectedBids := []analytics.RejectedBid{}
			seatBids, rejectionsErrs, rejectedBids = enforceFloorToBids(r.BidRequestWrapper, seatBids, conversions, enforceDealFloors)
			if r.LoggableObject != nil {
				r.LoggableObject.RejectedBids = append(r.LoggableObject.RejectedBids, rejectedBids...)
			}
		}

		if updateReqExt {
			requestExt.SetPrebid(prebidExt)
			err = r.BidRequestWrapper.RebuildRequestExt()
			if err != nil {
				rejectionsErrs = append(rejectionsErrs, err)
				return seatBids, rejectionsErrs
			}

			if responseDebugAllow {
				updatedBidReq, _ := json.Marshal(r.BidRequestWrapper.BidRequest)
				r.ResolvedBidRequest = updatedBidReq
			}
		}
	}

	return seatBids, rejectionsErrs
}

func updateBidExt(bidRequestWrapper *openrtb_ext.RequestWrapper, seatBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) {
	impMap := make(map[string]*openrtb_ext.ImpWrapper, bidRequestWrapper.LenImp())

	//Maintaining BidRequest Impression Map
	for _, v := range bidRequestWrapper.GetImp() {
		impMap[v.ID] = v
	}

	for _, seatBid := range seatBids {

		for _, bid := range seatBid.Bids {
			reqImp, ok := impMap[bid.Bid.ImpID]
			if !ok {
				continue
			}

			reqImpCur := reqImp.BidFloorCur
			if reqImpCur == "" {
				reqImpCur = "USD"
				if bidRequestWrapper.Cur != nil {
					reqImpCur = bidRequestWrapper.Cur[0]
				}
			}
			updateBidExtWithFloors(reqImp, bid, reqImpCur)
		}
	}
}
