package tradplus

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
)

func getBids(bidResponse *openrtb2.BidResponse) []openrtb2.Bid {
	serializedResponse, err := jsoniterator.Marshal(bidResponse)
	if err != nil {
		return nil
	}

	bid := bidResponse.SeatBid[0].Bid[0]
	bid.AdM = string(serializedResponse)
	bid.Ext = nil
	return []openrtb2.Bid{bid}
}

func ApplyTradPlusResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if rctx.Endpoint != models.EndpointTradPlus || bidResponse == nil {
		return bidResponse
	}

	// Reject is set before default bids are appended. Those placeholder seatbids must not
	// be returned; TradPlus no-bid is a response-level nbr and no seatbid.
	if rctx.TradPlus.Reject {
		nbr := bidResponse.NBR
		if nbr == nil {
			nbr = openrtb3.NoBidUnknownError.Ptr()
		}
		noBid := openrtb2.BidResponse{
			ID:  bidResponse.ID,
			NBR: nbr,
		}
		// debug=1 puts ext.debug (and related logger fields) on the response. Keep it
		// on the nbr body; non-debug responses stay seatbid-free and ext-free.
		if rctx.Debug {
			noBid.Ext = bidResponse.Ext
		}
		*bidResponse = noBid
		return bidResponse
	}

	if bidResponse.NBR != nil {
		return bidResponse
	}

	bids := getBids(bidResponse)
	if len(bids) == 0 {
		return bidResponse
	}

	*bidResponse = openrtb2.BidResponse{
		ID:    bidResponse.ID,
		BidID: bidResponse.SeatBid[0].Bid[0].ID,
		Cur:   bidResponse.Cur,
		SeatBid: []openrtb2.SeatBid{
			{
				Bid: bids,
			},
		},
	}

	return bidResponse
}

func SetTradPlusResponseReject(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) bool {
	if rctx.Endpoint != models.EndpointTradPlus {
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
