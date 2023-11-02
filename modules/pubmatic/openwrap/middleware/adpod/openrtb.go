package middleware

import (
	"encoding/json"

	"github.com/buger/jsonparser"
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

	// TODO: Do not merge the response, respond with 2.6 response
	mergedBidResponse := mergeSeatBids(bidResponse)
	data, err := json.Marshal(mergedBidResponse)
	if err != nil {
		return response
	}

	return data
}

func mergeSeatBids(bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return bidResponse
	}

	var seatBids []openrtb2.SeatBid
	bidArrayMap := make(map[string][]openrtb2.Bid)
	for _, seatBid := range bidResponse.SeatBid {
		//Copy seatBid and reset its bids
		videoSeatBid := seatBid
		videoSeatBid.Bid = nil
		for _, bid := range seatBid.Bid {
			if bid.Price == 0 {
				continue
			}

			adpodBid, _ := jsonparser.GetBoolean(bid.Ext, "adpod", "isAdpodBid")
			if !adpodBid {
				videoSeatBid.Bid = append(videoSeatBid.Bid, bid)
				continue
			}

			impId, _ := models.GetImpressionID(bid.ImpID)
			bids := bidArrayMap[impId]
			bids = append(bids, bid)
			bidArrayMap[impId] = bids
		}

		if len(videoSeatBid.Bid) > 0 {
			seatBids = append(seatBids, videoSeatBid)
		}
	}

	// Get Merged prebid_ctv bid
	ctvSeatBid := getPrebidCTVSeatBid(bidArrayMap)

	seatBids = append(seatBids, ctvSeatBid...)
	bidResponse.SeatBid = seatBids

	return bidResponse
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
