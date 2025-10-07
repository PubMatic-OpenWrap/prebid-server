package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) handleBidderRequestHook(ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BidderRequestPayload) (hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	rCtx, endpointHookManager, result, ok := validateModuleContextBidderRequestHook(miCtx)
	if !ok {
		return result, nil
	}

	defer func() {
		miCtx.ModuleContext["rctx"] = rCtx
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

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	endpointHookManager, ok := moduleCtx.ModuleContext["endpointhookmanager"].(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}
