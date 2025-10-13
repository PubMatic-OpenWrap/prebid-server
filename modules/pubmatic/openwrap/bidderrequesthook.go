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
	var err error
	rCtx, result, err = endpointHookManager.HandleBidderRequestHook(payload, rCtx, result, miCtx)
	if err != nil {
		return result, nil
	}

	return result, nil
}

func validateModuleContextBidderRequestHook(moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.BidderRequestPayload], bool) {
	result := hookstage.HookResult[hookstage.BidderRequestPayload]{}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	rCtxInterface, ok := moduleCtx.ModuleContext.Get("rctx")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}
	rCtx, ok := rCtxInterface.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	endpointHookManagerInterface, ok := moduleCtx.ModuleContext.Get("endpointhookmanager")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}
	endpointHookManager, ok := endpointHookManagerInterface.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}
