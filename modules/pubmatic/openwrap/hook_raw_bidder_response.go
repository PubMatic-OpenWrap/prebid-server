package openwrap

import (
	"fmt"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
)

type mediaTypes map[string]struct{}

func handleRawBidderResponseHook(

	payload hookstage.RawBidderResponsePayload,
	moduleCtx hookstage.ModuleContext,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	bidder := payload.Bidder

	// allowedBids will store all bids that have passed the attribute check
	allowedBids := make([]*adapters.TypedBid, 0)
	for _, bid := range payload.Bids {

		bidMediaTypes := mediaTypesFromBid(bid)

		fmt.Printf("\n Receieved bid ID = %v from = %v type = %v price = %v ", bid.Bid.ID, bidder, bidMediaTypes, bid.Bid.Price)

		//	addAllowedAnalyticTag(&result, bidder, bid.Bid.ImpID)
		allowedBids = append(allowedBids, bid)

	}

	changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}
	if len(payload.Bids) != len(allowedBids) {
		changeSet.RawBidderResponse().Bids().Update(allowedBids)
		result.ChangeSet = changeSet
	}

	return result, err
}

func mediaTypesFromBid(bid *adapters.TypedBid) mediaTypes {
	return mediaTypes{string(bid.BidType): struct{}{}}
}
