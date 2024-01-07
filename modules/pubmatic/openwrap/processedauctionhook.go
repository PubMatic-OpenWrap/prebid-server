package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) HandleProcessedAuctionHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.ProcessedAuctionRequestPayload,
) (hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	result := hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{}

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleProcessedAuctionHook()")
		return result, nil
	}
	rctx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleProcessedAuctionHook()")
		return result, nil
	}

	//Do not execute the module for requests processed in SSHB(8001)
	if rctx.Sshb == "1" || rctx.Endpoint == models.EndpointHybrid {
		return result, nil
	}

	ip := rctx.IP

	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		if parp.RequestWrapper != nil && parp.RequestWrapper.BidRequest.Device != nil && (parp.RequestWrapper.BidRequest.Device.IP == "" && parp.RequestWrapper.BidRequest.Device.IPv6 == "") {
			parp.RequestWrapper.BidRequest.Device.IP = ip
		}
		return parp, nil
	}, hookstage.MutationUpdate, "update-device-ip")

	return result, nil
}
