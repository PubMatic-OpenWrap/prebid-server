package middleware

import (
	"encoding/json"
	"errors"

	"github.com/prebid/openrtb/v19/openrtb2"
)

func FormVastResponse(response []byte) []byte {
	bidResponse := openrtb2.BidResponse{}

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		return response
	}

	vast, err := getVast(&bidResponse)
	if err != nil {
		return response
	}

	return []byte(vast)
}

func getVast(bidResponse *openrtb2.BidResponse) (string, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return "", errors.New("recieved invalid bidResponse")
	}

	bidArray := make([]*openrtb2.Bid, 0)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			bidArray = append(bidArray, &bid)
		}
	}

	creative, _ := getAdPodBidCreativeAndPrice(bidArray, true)
	if len(creative) == 0 {
		return "", errors.New("error while creating creative")
	}

	return creative, nil
}
