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
	payload hookstage.ExitpointPaylaod,
) (hookstage.HookResult[hookstage.ExitpointPaylaod], error) {
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

	result.ChangeSet.AddMutation(func(ep hookstage.ExitpointPaylaod) (hookstage.ExitpointPaylaod, error) {
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
) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.ExitpointPaylaod], bool) {
	result := hookstage.HookResult[hookstage.ExitpointPaylaod]{}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}

	rCtxInterface, ok := moduleCtx.ModuleContext.Get("rctx")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}
	rCtx, ok := rCtxInterface.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}

	endpointHookManagerInterface, ok := moduleCtx.ModuleContext.Get("endpointhookmanager")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}
	endpointHookManager, ok := endpointHookManagerInterface.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}
