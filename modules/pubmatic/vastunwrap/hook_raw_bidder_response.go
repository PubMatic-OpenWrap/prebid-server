package vastunwrap

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"

	"github.com/prebid/prebid-server/hooks/hookstage"
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
	vastUnwrapEnabled := vastRequestContext.VastUnwrapEnabled
	if !vastRequestContext.Redirect {
		pubId, _ := strconv.Atoi(miCtx.AccountID)
		vastRequestContext.PubID = pubId
		vastUnwrapEnabled = m.getVastUnwrapEnabled(vastRequestContext, m.TrafficPercentage)
		result.DebugMessages = append(result.DebugMessages,
			fmt.Sprintf("found request without sshb=1 in handleRawBidderResponseHook() for pubid:[%d]", vastRequestContext.PubID))
	}

	vastRequestContext.VastUnwrapEnabled = vastUnwrapEnabled
	if vastRequestContext.VastUnwrapEnabled {
		// Do Unwrap and Update Adm
		wg := new(sync.WaitGroup)
		for _, bid := range payload.Bids {
			if string(bid.BidType) == MediaTypeVideo {
				wg.Add(1)
				go func(bid *adapters.TypedBid) {
					defer wg.Done()
					m.doUnwrapandUpdateBid(vastRequestContext.VastUnwrapStatsEnabled, bid, vastRequestContext.UA, vastRequestContext.IP, unwrapURL, miCtx.AccountID, payload.Bidder)
				}(bid)
			}
		}
		wg.Wait()
		changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}
		changeSet.RawBidderResponse().Bids().Update(payload.Bids)
		result.ChangeSet = changeSet
	} else {
		vastRequestContext.VastUnwrapStatsEnabled = openwrap.GetRandomNumberIn1To100() <= m.StatTrafficPercentage
		if vastRequestContext.VastUnwrapStatsEnabled {
			// Do Unwrap and Collect stats only
			for _, bid := range payload.Bids {
				if string(bid.BidType) == MediaTypeVideo {
					go func(bid *adapters.TypedBid) {
						m.doUnwrapandUpdateBid(vastRequestContext.VastUnwrapStatsEnabled, bid, vastRequestContext.UA, vastRequestContext.IP, unwrapURL, miCtx.AccountID, payload.Bidder)
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
