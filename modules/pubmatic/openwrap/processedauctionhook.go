package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
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

	//Do not execute the module for requests processed in SSHB(8001)
	if rCtx.Sshb == "1" || rCtx.Endpoint == models.EndpointHybrid {
		return result, nil
	}

	defer func() {
		moduleCtx.ModuleContext.Set("rctx", rCtx)
	}()

	result, ok = endpointHookManager.HandleProcessedAuctionHook(&rCtx, payload, moduleCtx, result)
	if !ok {
		return result, nil
	}

	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in handleProcessedAuctionHook mutation")
			return parp, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		if parp.Request != nil && parp.Request.BidRequest.Device != nil && (parp.Request.BidRequest.Device.IP == "" && parp.Request.BidRequest.Device.IPv6 == "") {
			parp.Request.BidRequest.Device.IP = rCtx.DeviceCtx.IP
		}
		return parp, nil
	}, hookstage.MutationUpdate, "update-device-ip")

	return result, nil
}

func validateModuleContextProcessedAuctionHook(
	moduleCtx hookstage.ModuleInvocationContext,
) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], bool) {
	result := hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleProcessedAuctionHook()")
		return models.RequestCtx{}, &endpointmanager.NilEndpointManager{}, result, false
	}

	rCtxInterface, ok := moduleCtx.ModuleContext.Get(models.RequestContext)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleProcessedAuctionHook()")
		return models.RequestCtx{}, &endpointmanager.NilEndpointManager{}, result, false
	}
	rCtx, ok := rCtxInterface.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleProcessedAuctionHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}

	endpointHookManagerInterface, ok := moduleCtx.ModuleContext.Get(models.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleProcessedAuctionHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}
	endpointHookManager, ok := endpointHookManagerInterface.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleProcessedAuctionHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}

	return rCtx, endpointHookManager, result, true
}
