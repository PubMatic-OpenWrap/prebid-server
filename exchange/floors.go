package exchange

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/golang/glog"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/floors"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func EnforceFloorToBids(bidRequest *openrtb2.BidRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, conversions currency.Conversions, enforceDealFloors bool) (map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []string) {

	type bidFloor struct {
		bidFloorCur string
		bidFloor    float64
	}
	var rejections []string
	impMap := make(map[string]bidFloor)

	//Maintaining BidRequest Impression Map
	for i := range bidRequest.Imp {
		var bidfloor bidFloor
		bidfloor.bidFloorCur = bidRequest.Imp[i].BidFloorCur
		if bidfloor.bidFloorCur == "" {
			bidfloor.bidFloorCur = "USD"
		}
		bidfloor.bidFloor = bidRequest.Imp[i].BidFloor
		impMap[bidRequest.Imp[i].ID] = bidfloor
	}

	for bidderName, seatBid := range seatBids {
		eligibleBids := make([]*pbsOrtbBid, 0)
		for bidInd := range seatBid.bids {
			bid := seatBid.bids[bidInd]
			bidID := bid.bid.ID
			if bid.bid.DealID != "" && !enforceDealFloors {
				eligibleBids = append(eligibleBids, bid)
				continue
			}
			bidFloor, ok := impMap[bid.bid.ImpID]
			if !ok {
				continue
			}
			bidPrice := bid.bid.Price
			if seatBid.currency != bidFloor.bidFloorCur {
				rate, err := conversions.GetRate(seatBid.currency, bidFloor.bidFloorCur)
				if err != nil {
					glog.Warningf("error in rate conversion with bidder %s for impression id %s and bid id %s", bidderName, bid.bid.ImpID, bidID)
					continue
				}
				bidPrice = rate * bid.bid.Price
			}
			if bidFloor.bidFloor > bidPrice {
				rejections = updateRejections(rejections, bidID, fmt.Sprintf("bid price value %f is less than bidFloor value %f for impression id %s bidder %s", bidPrice, bidFloor.bidFloor, bid.bid.ImpID, bidderName))
				continue
			}
			eligibleBids = append(eligibleBids, bid)

		}
		seatBids[bidderName].bids = eligibleBids

	}

	return seatBids, rejections
}

func SignalFloors(r *AuctionRequest, floor floors.Floor, conversions currency.Conversions, responseDebugAllow bool) []error {

	var errs []error
	requestExt, err := r.BidRequestWrapper.GetRequestExt()
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	prebidExt := requestExt.GetPrebid()
	if floor != nil && floor.Enabled() && floors.IsRequestEnabledWithFloor(prebidExt.Floors) {
		errs = floors.UpdateImpsWithFloors(prebidExt.Floors, r.BidRequestWrapper.BidRequest, conversions)
		requestExt.SetPrebid(prebidExt)
	}

	return errs
}

func EnforceFloors(r *AuctionRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, floor floors.Floor, conversions currency.Conversions, responseDebugAllow bool) (map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []string) {

	var rejections []string
	requestExt, err := r.BidRequestWrapper.GetRequestExt()
	if err != nil {
		rejections = append(rejections, err.Error())
		return seatBids, rejections
	}
	prebidExt := requestExt.GetPrebid()
	if floor != nil && floor.Enabled() && floors.IsRequestEnabledWithFloor(prebidExt.Floors) && floors.ShouldEnforceFloors(r.BidRequestWrapper.BidRequest, prebidExt.Floors, floor.GetEnforceRate(), rand.Intn) {
		var enforceDealFloors bool
		if prebidExt.Floors.Enforcement != nil {
			enforceDealFloors = prebidExt.Floors.Enforcement.FloorDeals && floor.EnforceDealFloor()
		}
		seatBids, rejections = EnforceFloorToBids(r.BidRequestWrapper.BidRequest, seatBids, conversions, enforceDealFloors)
		requestExt.SetPrebid(prebidExt)
	}

	err = r.BidRequestWrapper.RebuildRequest()
	if err != nil {
		rejections = append(rejections, err.Error())
	}
	updatedBidReq, _ := json.Marshal(r.BidRequestWrapper.BidRequest)
	JLogf("Updated Floor Request after parsing floors", string(updatedBidReq))
	if responseDebugAllow {
		//save updated request
		r.UpdatedBidRequest = updatedBidReq
	}

	return seatBids, rejections
}
