package exchange

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/openrtb/v17/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/floors"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// RejectedBid defines the contract for bid rejection errors due to floors enforcement
type RejectedBid struct {
	Bid             *openrtb2.Bid `json:"bid,omitempty"`
	RejectionReason int           `json:"rejectreason,omitempty"`
	BidderName      string        `json:"biddername,omitempty"`
}

// Check for Floors enforcement for deals,
// In case bid wit DealID present and enforceDealFloors = false then bid floor enforcement should be skipped
func checkDealsForEnforcement(bid *pbsOrtbBid, enforceDealFloors bool) *pbsOrtbBid {
	if bid != nil && bid.bid != nil && bid.bid.DealID != "" && !enforceDealFloors {
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

// enforceFloorToBids function does floors enforcement for each bid.
//
//	The bids returned by each partner below bid floor price are rejected and remaining eligible bids are considered for further processing
func enforceFloorToBids(bidRequest *openrtb2.BidRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, conversions currency.Conversions, enforceDealFloors bool) (map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []error, []analytics.RejectedBid) {
	errs := []error{}
	rejectedBids := []analytics.RejectedBid{}
	impMap := make(map[string]openrtb2.Imp, len(bidRequest.Imp))

	//Maintaining BidRequest Impression Map
	for i := range bidRequest.Imp {
		impMap[bidRequest.Imp[i].ID] = bidRequest.Imp[i]
	}

	for bidderName, seatBid := range seatBids {
		eligibleBids := make([]*pbsOrtbBid, 0, len(seatBid.bids))
		for _, bid := range seatBid.bids {
			retBid := checkDealsForEnforcement(bid, enforceDealFloors)
			if retBid != nil {
				eligibleBids = append(eligibleBids, retBid)
				continue
			}

			reqImp, ok := impMap[bid.bid.ImpID]
			if ok {
				reqImpCur := reqImp.BidFloorCur
				if reqImpCur == "" {
					if bidRequest.Cur != nil {
						reqImpCur = bidRequest.Cur[0]
					} else {
						reqImpCur = "USD"
					}
				}
				rate, err := getCurrencyConversionRate(seatBid.currency, reqImpCur, conversions)
				if err == nil {
					bidPrice := rate * bid.bid.Price
					if reqImp.BidFloor > bidPrice {
						rejectedBid := analytics.RejectedBid{
							Bid:  bid.bid,
							Seat: seatBid.seat,
						}
						rejectedBid.RejectionReason = openrtb3.LossBidBelowAuctionFloor
						if bid.bid.DealID != "" {
							rejectedBid.RejectionReason = openrtb3.LossBidBelowDealFloor
						}
						rejectedBids = append(rejectedBids, rejectedBid)
						errs = append(errs, fmt.Errorf("bid rejected [bid ID: %s] reason: bid price value %.4f %s is less than bidFloor value %.4f %s for impression id %s bidder %s", bid.bid.ID, bidPrice, reqImpCur, reqImp.BidFloor, reqImpCur, bid.bid.ImpID, bidderName))
					} else {
						eligibleBids = append(eligibleBids, bid)
					}
				} else {
					errMsg := fmt.Errorf("Error in rate conversion from = %s to %s with bidder %s for impression id %s and bid id %s", seatBid.currency, reqImpCur, bidderName, bid.bid.ImpID, bid.bid.ID)
					glog.Errorf(errMsg.Error())
					errs = append(errs, errMsg)

				}
			}
		}
		seatBids[bidderName].bids = eligibleBids
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
func enforceFloors(r *AuctionRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, floor config.PriceFloors, conversions currency.Conversions, responseDebugAllow bool) (map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []error) {
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
		floorsEnfocement = floors.RequestHasFloors(r.BidRequestWrapper.BidRequest)
		if prebidExt != nil && floorsEnfocement {
			if floorsEnfocement, updateReqExt = floors.ShouldEnforce(r.BidRequestWrapper.BidRequest, prebidExt.Floors, r.Account.PriceFloors.EnforceFloorRate, rand.Intn); floorsEnfocement {
				enforceDealFloors = r.Account.PriceFloors.EnforceDealFloors && getEnforceDealsFlag(prebidExt.Floors)
			}
		}

		if floorsEnfocement {
			rejectedBids := []analytics.RejectedBid{}
			seatBids, rejectionsErrs, rejectedBids = enforceFloorToBids(r.BidRequestWrapper.BidRequest, seatBids, conversions, enforceDealFloors)
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
