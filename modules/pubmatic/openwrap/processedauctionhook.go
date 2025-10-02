package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) HandleProcessedAuctionHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.ProcessedAuctionRequestPayload,
) (hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	rCtx, endpointHookManager, result, ok := validateModuleContextProcessedAuctionHook(moduleCtx)
	if !ok {
		return result, nil
	}
	defer func() {
		moduleCtx.ModuleContext["rctx"] = rCtx
	}()

	//Do not execute the module for requests processed in SSHB(8001)
	if rCtx.Sshb == "1" || rCtx.Endpoint == models.EndpointHybrid {
		return result, nil
	}

	rCtx, result, err := endpointHookManager.HandleProcessedAuctionHook(payload, rCtx, result, moduleCtx)
	if err != nil {
		return result, err
	}

	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		rCtx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
		defer func() {
			moduleCtx.ModuleContext["rctx"] = rCtx
		}()
		if parp.Request != nil && parp.Request.BidRequest.Device != nil && (parp.Request.BidRequest.Device.IP == "" && parp.Request.BidRequest.Device.IPv6 == "") {
			parp.Request.BidRequest.Device.IP = rCtx.DeviceCtx.IP
		}
		return parp, nil
	}, hookstage.MutationUpdate, "update-device-ip")

	return result, nil
}

func validateModuleContextProcessedAuctionHook(moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], bool) {
	result := hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{}

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleProcessedAuctionHook()")
		return models.RequestCtx{}, nil, result, false
	}

	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleProcessedAuctionHook()")
		return models.RequestCtx{}, nil, result, false
	}

	endpointHookManager, ok := moduleCtx.ModuleContext["endpointhookmanager"].(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleProcessedAuctionHook()")
		return models.RequestCtx{}, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}
