package unitylevelplay

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/exchange/entities"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func getBids(bidResponse *openrtb2.BidResponse) []openrtb2.Bid {
	serializedResponse, err := jsoniterator.Marshal(bidResponse)
	if err != nil {
		return nil
	}

	bid := bidResponse.SeatBid[0].Bid[0]
	bid.AdM = string(serializedResponse)
	return []openrtb2.Bid{bid}
}

func ApplyUnityLevelPlayResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if rctx.Endpoint != models.EndpointUnityLevelPlay || bidResponse.NBR != nil || rctx.UnityLevelPlay.Reject {
		return bidResponse
	}

	bids := getBids(bidResponse)
	if len(bids) == 0 {
		return bidResponse
	}

	updatedResponse := openrtb2.BidResponse{
		ID:    bidResponse.ID,
		BidID: bidResponse.SeatBid[0].Bid[0].ID,
		Cur:   bidResponse.Cur,
		SeatBid: []openrtb2.SeatBid{
			{
				Bid: bids,
			},
		},
	}

	return &updatedResponse
}

func SetUnityLevelPlayResponseReject(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) bool {
	if rctx.Endpoint != models.EndpointUnityLevelPlay {
		return false
	}

	reject := false
	if bidResponse.NBR != nil {
		if !rctx.Debug {
			reject = true
		}
	} else if len(bidResponse.SeatBid) == 0 || len(bidResponse.SeatBid[0].Bid) == 0 {
		reject = true
	}
	return reject
}

func UpdateBidWithTestPrice(rctx models.RequestCtx, bidderResponses map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) {
	if rctx.Endpoint != models.EndpointUnityLevelPlay || rctx.IsTestRequest != 1 {
		return
	}

	for _, seatBid := range bidderResponses {
		for i := range seatBid.Bids {
			seatBid.Bids[i].Bid.Price = 99
		}
	}
}
