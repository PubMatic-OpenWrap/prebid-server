package openwrap

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

var getRandomNumber = func() int {
	return rand.Intn(100)
}

func (m OpenWrap) handleRawBidderResponseHook(
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
	unwrapURL string,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {

	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}

	if !rCtx.VastUnwrapEnabled {
		rCtx.VastUnwrapStatsEnabled = getRandomNumber() < m.cfg.Features.VASTUnwrapStatsPecent
	}

	if !rCtx.VastUnwrapEnabled && !rCtx.VastUnwrapStatsEnabled {
		result.DebugMessages = append(result.DebugMessages,
			fmt.Sprintf("error: vast unwrap flag is not enabled in handleRawBidderResponseHook() for pubid:[%d]", rCtx.PubID))
		return result, nil
	}

	// Below code collects stats only
	if rCtx.VastUnwrapStatsEnabled {
		for _, bid := range payload.Bids {
			if string(bid.BidType) == "video" { // TBDJ
				go func(bid *adapters.TypedBid) {
					m.doUnwrapandUpdateBid(rCtx.VastUnwrapStatsEnabled, bid, rCtx.UA, unwrapURL, fmt.Sprintf("%d", rCtx.PubID), payload.Bidder)
				}(bid)
			}
		}
	} else {
		wg := new(sync.WaitGroup)
		for _, bid := range payload.Bids {
			if string(bid.BidType) == "video" { // TBDJ
				wg.Add(1)
				go func(bid *adapters.TypedBid) {
					defer wg.Done()
					m.doUnwrapandUpdateBid(rCtx.VastUnwrapStatsEnabled, bid, rCtx.UA, unwrapURL, fmt.Sprintf("%d", rCtx.PubID), payload.Bidder)
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
