package openwrap

import (
	"context"
	"encoding/json"
	"log"

	"github.com/prebid/prebid-server/v3/exchange/entities"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

// handleAllProcessedBidResponsesHook will create unique id for each bid in bid Response. This hook is introduced
// because bidresponse should be updated in mutations and we need modified bidID at the start of auction response hook.
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

	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleAllProcessedBidResponsesHook()")
		return result, nil
	}

	updateWakandaHTTPCalls(&rCtx, payload)

	//Do not execute the module for requests processed in SSHB(8001)
	if rCtx.Sshb == "1" || rCtx.Endpoint == models.EndpointHybrid {
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
			seatBid.Bids[i].Bid.ID = utils.SetUniqueBidID(seatBid.Bids[i].Bid.ID, seatBid.Bids[i].GeneratedBidID)
		}
	}
}

func updateWakandaHTTPCalls(rCtx *models.RequestCtx, payload hookstage.AllProcessedBidResponsesPayload) {

	if rCtx.WakandaDebug != nil && rCtx.WakandaDebug.IsEnable() {

		bidderHttpCalls := make(map[openrtb_ext.BidderName][]*openrtb_ext.ExtHttpCall)
		for abc, http := range payload.Responses {
			bidderHttpCalls[abc] = append(bidderHttpCalls[abc], http.HttpCalls...)
		}

		wakandaDebugData, err := json.Marshal(bidderHttpCalls)
		if err != nil {
			log.Printf("Error marshaling bidderHttpCalls: %v", err)
		} else {
			rCtx.WakandaDebug.SetHttpCalls(json.RawMessage(wakandaDebugData))
		}
	}
}
