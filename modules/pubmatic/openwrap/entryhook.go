package openwrap

import (
	"context"
	"strings"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

const (
	OpenWrapAuction = "/pbs/openrtb2/auction"
	OpenWrapV25     = "/openrtb/2.5"
	OpenWrapVideo   = "/openrtb/video"
	OpenWrapAmp     = "/openrtb/amp"
)

func getContextValueForField(ctx context.Context, field string) bool {
	vastEnableUnwrapper, _ := ctx.Value(field).(string)
	if vastEnableUnwrapper == "1" || strings.ToLower(vastEnableUnwrapper) == "true" {
		return true
	}
	return false
}

func handleEntrypointHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {

	result := hookstage.HookResult[hookstage.EntrypointPayload]{}

	rCtx := RequestCtx{
		VastUnwrapFlag: getContextValueForField(payload.Request.Context(), "enableVastUnwrapper"),
	}

	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext["rctx"] = rCtx

	result.Reject = false
	return result, nil
}
