package openwrap

import (
	"fmt"
	"sync"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

func (m OpenWrap) handleRawBidderResponseHook(
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	vastRequestContext, ok := miCtx.ModuleContext[models.RequestContext].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
		return result, nil
	}

	if vastRequestContext.VastUnwrapEnabled {
		// Do Unwrap and Update Adm
		wg := new(sync.WaitGroup)
		for _, bid := range payload.Bids {
			if string(bid.BidType) == models.MediaTypeVideo {
				wg.Add(1)
				go func(bid *adapters.TypedBid) {
					defer wg.Done()
					m.unwrap.Unwrap(miCtx.AccountID, payload.Bidder, bid, vastRequestContext.UA, vastRequestContext.IP, vastRequestContext.VastUnwrapStatsEnabled)
				}(bid)
			}
		}
		wg.Wait()
		changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}
		changeSet.RawBidderResponse().Bids().Update(payload.Bids)
		result.ChangeSet = changeSet
	} else {
		vastRequestContext.VastUnwrapStatsEnabled = GetRandomNumberIn1To100() <= m.cfg.Features.VASTUnwrapStatsPecent
		if vastRequestContext.VastUnwrapStatsEnabled {
			// Do Unwrap and Collect stats only
			for _, bid := range payload.Bids {
				if string(bid.BidType) == models.MediaTypeVideo {
					go func(bid *adapters.TypedBid) {
						m.unwrap.Unwrap(miCtx.AccountID, payload.Bidder, bid, vastRequestContext.UA, vastRequestContext.IP, vastRequestContext.VastUnwrapStatsEnabled)
					}(bid)
				}
			}
		}
	}

	if vastRequestContext.VastUnwrapEnabled || vastRequestContext.VastUnwrapStatsEnabled {
		result.DebugMessages = append(result.DebugMessages,
			fmt.Sprintf("For pubid:[%d] VastUnwrapEnabled: [%v] VastUnwrapStatsEnabled:[%v] ",
				vastRequestContext.PubID, vastRequestContext.VastUnwrapEnabled, vastRequestContext.VastUnwrapStatsEnabled))
	}

	return result, nil
}
