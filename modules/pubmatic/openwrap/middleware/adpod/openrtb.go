package middleware

import (
	"encoding/json"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func formOperRTBResponse(response []byte) []byte {
	var bidResponse *openrtb2.BidResponse

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		return response
	}

	// TODO : Do not merge the response, respond with 2.6 response
	mergedBidResponse, err := mergeSeatBids(bidResponse)
	if err != nil {
		return response
	}

	data, err := json.Marshal(mergedBidResponse)
	if err != nil {
		return response
	}

	return data
}

func mergeSeatBids(bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return nil, errors.New("recieved invalid bidResponse")
	}

	bidArrayMap := make(map[string][]openrtb2.Bid)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price > 0 {
				impId, _ := models.GetImpressionID(bid.ImpID)
				bids, ok := bidArrayMap[impId]
				if !ok {
					bids = make([]openrtb2.Bid, 0)
				}

				bids = append(bids, bid)
				bidArrayMap[impId] = bids
			}
		}
	}

	bidResponse.SeatBid = getPrebidCTVSeatBid(bidArrayMap)

	return bidResponse, nil
}

func getPrebidCTVSeatBid(bidsMap map[string][]openrtb2.Bid) []openrtb2.SeatBid {
	seatBids := []openrtb2.SeatBid{}

	for impId, bids := range bidsMap {
		bid := openrtb2.Bid{}
		bidID, err := uuid.NewV4()
		if err == nil {
			bid.ID = bidID.String()
		} else {
			bid.ID = bids[0].ID
		}
		creative, price := getAdPodBidCreativeAndPrice(bids)
		bid.AdM = creative
		bid.Price = price
		if len(bids) > 0 {
			bid.Cat = bids[0].Cat
			bid.ADomain = bids[0].ADomain
		}
		bid.ImpID = impId

		seatBid := openrtb2.SeatBid{}
		seatBid.Seat = "prebid_ctv"
		seatBid.Bid = append(seatBid.Bid, bid)

		seatBids = append(seatBids, seatBid)
	}

	return seatBids
}
