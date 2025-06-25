package unitylevelplay

import (
	"encoding/json"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func ApplyUnityLevelPlayResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if rctx.Endpoint != models.EndpointUnityLevelPlay || bidResponse.NBR != nil || rctx.UnityLevelPlay.Reject {
		return bidResponse
	}

	serializedResponse, err := json.Marshal(bidResponse)
	if err != nil {
		return bidResponse
	}

	updatedResponse := openrtb2.BidResponse{
		ID:    bidResponse.ID,
		BidID: bidResponse.SeatBid[0].Bid[0].ID,
		Cur:   bidResponse.Cur,
		SeatBid: []openrtb2.SeatBid{
			{
				Bid: []openrtb2.Bid{
					{
						ID:    bidResponse.SeatBid[0].Bid[0].ID,
						ImpID: bidResponse.SeatBid[0].Bid[0].ImpID,
						Price: bidResponse.SeatBid[0].Bid[0].Price,
						AdM:   string(serializedResponse),
						BURL:  bidResponse.SeatBid[0].Bid[0].BURL,
						Ext:   bidResponse.SeatBid[0].Bid[0].Ext,
					},
				},
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
