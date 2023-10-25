package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/exchange/entities"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (m OpenWrap) handleAllProcessedBidResponsesHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.AllProcessedBidResponsesPayload,
) (hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error) {
	result := hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]{
		ChangeSet: hookstage.ChangeSet[hookstage.AllProcessedBidResponsesPayload]{},
	}

	// absence of rctx at this hook means the first hook failed!. Do nothing
	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleAllProcessedBidResponsesHook()")
		return result, nil
	}

	result.ChangeSet.AddMutation(func(apbrp hookstage.AllProcessedBidResponsesPayload) (hookstage.AllProcessedBidResponsesPayload, error) {
		updateBidIds(apbrp.Responses)
		return apbrp, nil
	}, hookstage.MutationUpdate, "update-bid-id")

	return result, nil
}

func updateBidIds(bidderResponses map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) {
	for _, seatBid := range bidderResponses {
		for i := range seatBid.Bids {
			seatBid.Bids[i].Bid.ID = seatBid.Bids[i].Bid.ID + models.BidIdSeparator + seatBid.Bids[i].GeneratedBidID
		}
	}
}
