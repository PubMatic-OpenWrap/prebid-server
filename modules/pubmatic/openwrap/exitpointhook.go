package openwrap

import (
	"context"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) handleExitpointHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.ExitpointPayload,
) (hookstage.HookResult[hookstage.ExitpointPayload], error) {
	// validate module context
	rCtx, endpointManager, result, ok := validateModuleContextExitpointHook(miCtx)
	if !ok {
		return result, nil
	}

	defer func() {
		miCtx.ModuleContext.Set("rctx", rCtx)
	}()

	result, ok = endpointManager.HandleExitpointHook(&rCtx, payload, miCtx, result)
	if !ok {
		return result, nil
	}

	result.ChangeSet.AddMutation(func(ep hookstage.ExitpointPayload) (hookstage.ExitpointPayload, error) {
		ortbResponse, ok := ep.Response.(*openrtb2.BidResponse)
		if ok {
			resetBidIdtoOriginal(ortbResponse)
			ep.Response = ortbResponse
		}
		return ep, nil
	}, hookstage.MutationUpdate, "reset-bid-id-to-original")

	return result, nil
}

// validateModuleContext validates that required context is available
func validateModuleContextExitpointHook(
	moduleCtx hookstage.ModuleInvocationContext,
) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.ExitpointPayload], bool) {
	result := hookstage.HookResult[hookstage.ExitpointPayload]{}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleExitpointHook()")
		return models.RequestCtx{}, &endpointmanager.NilEndpointManager{}, result, false
	}

	rCtxInterface, ok := moduleCtx.ModuleContext.Get(models.RequestContext)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleExitpointHook()")
		return models.RequestCtx{}, &endpointmanager.NilEndpointManager{}, result, false
	}
	rCtx, ok := rCtxInterface.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleExitpointHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}

	endpointHookManagerInterface, ok := moduleCtx.ModuleContext.Get(models.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleExitpointHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}
	endpointHookManager, ok := endpointHookManagerInterface.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleExitpointHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}

	return rCtx, endpointHookManager, result, true
}
