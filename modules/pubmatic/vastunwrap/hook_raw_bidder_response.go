package vastunwrap

import (
	"sync"

	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/vastunwrap/models"

	"github.com/prebid/prebid-server/v2/hooks/hookstage"
)

func (m VastUnwrapModule) handleRawBidderResponseHook(
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
	unwrapURL string,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	vastRequestContext, ok := miCtx.ModuleContext[RequestContext].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
		return result, nil
	}
	if !vastRequestContext.VastUnwrapEnabled && !vastRequestContext.VastUnwrapStatsEnabled {
		result.DebugMessages = append(result.DebugMessages, "error: vast unwrap flag is not enabled in handleRawBidderResponseHook()")
		return result, nil
	}

	// Below code collects stats only
	if vastRequestContext.VastUnwrapStatsEnabled {
		for _, bid := range payload.Bids {
			if string(bid.BidType) == MediaTypeVideo {
				go func(bid *adapters.TypedBid) {
					m.doUnwrapandUpdateBid(vastRequestContext.VastUnwrapStatsEnabled, bid, vastRequestContext.UA, unwrapURL, miCtx.AccountID, payload.Bidder)
				}(bid)
			}
		}
	} else {
		wg := new(sync.WaitGroup)
		for _, bid := range payload.Bids {
			if string(bid.BidType) == MediaTypeVideo {
				wg.Add(1)
				go func(bid *adapters.TypedBid) {
					defer wg.Done()
					m.doUnwrapandUpdateBid(vastRequestContext.VastUnwrapStatsEnabled, bid, vastRequestContext.UA, unwrapURL, miCtx.AccountID, payload.Bidder)
				}(bid)
			}
		}
		wg.Wait()
		changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}
		changeSet.RawBidderResponse().Bids().Update(payload.Bids)
		result.ChangeSet = changeSet
	}
	return result, nil
}
