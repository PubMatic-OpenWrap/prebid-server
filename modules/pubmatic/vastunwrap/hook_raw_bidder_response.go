package vastunwrap

import (
	"sync"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

type mediaTypes map[string]struct{}

func handleRawBidderResponseHook(
	payload hookstage.RawBidderResponsePayload,
	moduleCtx hookstage.ModuleContext, unwrapDefaultTimeout int, unwrapURL string,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	wg := new(sync.WaitGroup)
	vastRequestContext, ok := moduleCtx[RequestContext].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
		return result, nil
	}
	defer func() {
		moduleCtx[RequestContext] = vastRequestContext
	}()

	if !vastRequestContext.IsVastUnwrapEnabled {
		result.DebugMessages = append(result.DebugMessages, "error: vast unwrap flag is not enabled in handleRawBidderResponseHook()")
		return result, nil
	}

	for _, bid := range payload.Bids {
		bidMediaTypes := mediaTypesFromBid(bid)
		if _, ok := bidMediaTypes[MediaTypeVideo]; ok {
			wg.Add(1)
			go func(bid *adapters.TypedBid) {
				defer wg.Done()
				doUnwrap(bid, vastRequestContext.UA, unwrapDefaultTimeout, unwrapURL)
			}(bid)
		}
	}
	wg.Wait()
	changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}

	changeSet.RawBidderResponse().Bids().Update(payload.Bids)
	result.ChangeSet = changeSet

	return result, err
}

func mediaTypesFromBid(bid *adapters.TypedBid) mediaTypes {
	return mediaTypes{string(bid.BidType): struct{}{}}
}
