package vastunwrap

import (
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

type mediaTypes map[string]struct{}

func handleRawBidderResponseHook(
	payload hookstage.RawBidderResponsePayload,
	moduleCtx hookstage.ModuleContext, unwrapDefaultTimeout int, unwrapURL string,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {

	vastRequestContext, ok := moduleCtx[RequestContext].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
		return result, nil
	}
	defer func() {
		moduleCtx[RequestContext] = vastRequestContext
	}()

	for _, bid := range payload.Bids {
		bidMediaTypes := mediaTypesFromBid(bid)
		if _, ok := bidMediaTypes[MediaTypeVideo]; ok {
			go doUnwrap(bid, vastRequestContext.UA, unwrapDefaultTimeout, unwrapURL)

		}
	}
	changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}

	changeSet.RawBidderResponse().Bids().Update(payload.Bids)
	result.ChangeSet = changeSet

	return result, err
}

func mediaTypesFromBid(bid *adapters.TypedBid) mediaTypes {
	return mediaTypes{string(bid.BidType): struct{}{}}
}
