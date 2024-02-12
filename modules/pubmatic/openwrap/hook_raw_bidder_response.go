package openwrap

import (
	"sync"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

func (ow OpenWrap) handleRawBidderResponseHook(
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
	unwrapURL string,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	// vastRequestContext, ok := miCtx.ModuleContext[RequestContext].(models.RequestCtx)
	// if !ok {
	// 	result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
	// 	return result, nil
	// }
	// if !vastRequestContext.VastUnwrapEnabled && !vastRequestContext.VastUnwrapStatsEnabled {
	// 	result.DebugMessages = append(result.DebugMessages, "error: vast unwrap flag is not enabled in handleRawBidderResponseHook()")
	// 	return result, nil
	// }

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleRawBidderResponseHook()")
		return
	}

	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok || rCtx.PartnerConfigMap == nil || rCtx.PartnerConfigMap[-1] == nil {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
		return
	}

	if rCtx.PartnerConfigMap[-1]["enableVastUnwrapper"] != "0" {
		result.DebugMessages = append(result.DebugMessages, "error: vast unwrap flag is not enabled in handleRawBidderResponseHook()")
		return
	}

	// Below code collects stats only
	if GetRandomNumberBelow100() < ow.cfg.VastUnwrapModule.StatTrafficPercentage {
		for _, bid := range payload.Bids {
			if string(bid.BidType) == models.Video {
				go func(bid *adapters.TypedBid) {
					vastunwrap.DoUnwrapandUpdateBid(true, bid, rCtx.UA, unwrapURL, rCtx.PubID, payload.Bidder)
				}(bid)
			}
		}
	} else {
		wg := new(sync.WaitGroup)
		for _, bid := range payload.Bids {
			if string(bid.BidType) == models.Video {
				wg.Add(1)
				go func(bid *adapters.TypedBid) {
					defer wg.Done()
					vastunwrap.DoUnwrapandUpdateBid(false, bid, rCtx.UA, unwrapURL, rCtx.PubID, payload.Bidder)
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
