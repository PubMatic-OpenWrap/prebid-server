package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) handleBidderRequestHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BidderRequestPayload) (hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	rCtx, endpointHookManager, result, ok := validateModuleContextBidderRequestHook(miCtx)
	if !ok {
		return result, nil
	}

	defer func() {
		miCtx.ModuleContext.Set("rctx", rCtx)
	}()

	// Execute Endpoint specific bidder request hook
	result, ok = endpointHookManager.HandleBidderRequestHook(&rCtx, payload, miCtx, result)
	if !ok {
		return result, nil
	}

	return result, nil
}

func validateModuleContextBidderRequestHook(
	moduleCtx hookstage.ModuleInvocationContext,
) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.BidderRequestPayload], bool) {
	result := hookstage.HookResult[hookstage.BidderRequestPayload]{}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, &endpointmanager.NilEndpointManager{}, result, false
	}

	rCtxInterface, ok := moduleCtx.ModuleContext.Get(models.RequestContext)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, &endpointmanager.NilEndpointManager{}, result, false
	}
	rCtx, ok := rCtxInterface.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, &endpointmanager.NilEndpointManager{}, result, false
	}

	endpointHookManagerInterface, ok := moduleCtx.ModuleContext.Get(models.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBidderRequestHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}
	endpointHookManager, ok := endpointHookManagerInterface.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBidderRequestHook()")
		return rCtx, &endpointmanager.NilEndpointManager{}, result, false
	}

	return rCtx, endpointHookManager, result, true
}
