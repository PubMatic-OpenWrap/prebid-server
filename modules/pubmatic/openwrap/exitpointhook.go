package openwrap

import (
	"context"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) handleExitpointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.ExitpointPaylaod,
) (result hookstage.HookResult[hookstage.Exitpoint], err error) {
	// validate module context
	rCtx, endpointManager, result, ok := validateModuleContextExitpointHook(miCtx)
	if !ok {
		return result, nil
	}

	defer func() {
		miCtx.ModuleContext["rctx"] = rCtx
	}()

	// result, ok = validateExitpointPayload(&rCtx, result, payload)
	// if !ok {
	// 	return result, nil
	// }

	rCtx, result, err = endpointManager.HandleExitpointHook(payload, rCtx, result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// validateModuleContext validates that required context is available
func validateModuleContextExitpointHook(moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.Exitpoint], bool) {
	result := hookstage.HookResult[hookstage.Exitpoint]{}

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}

	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}

	endpointHookManager, ok := moduleCtx.ModuleContext["endpointhookmanager"].(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleExitpointHook()")
		return models.RequestCtx{}, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}

func validateExitpointPayload(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.Exitpoint], payload hookstage.ExitpointPaylaod) (hookstage.HookResult[hookstage.Exitpoint], bool) {
	response, ok := payload.Response.(*openrtb2.BidResponse)
	if !ok {
		result.Errors = append(result.Errors, "invalid response format while processing exitpoint hook")
		return result, false
	}

	if response.NBR != nil {
		return result, false
	}

	if len(response.SeatBid) == 0 {
		return result, false
	}

	return result, true
}
