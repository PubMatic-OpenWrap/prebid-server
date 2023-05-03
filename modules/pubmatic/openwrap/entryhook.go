package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

const (
	OpenWrapAuction        = "/pbs/openrtb2/auction"
	OpenWrapV25            = "/openrtb/2.5"
	OpenWrapVideo          = "/openrtb/video"
	OpenWrapAmp            = "/openrtb/amp"
	VastUnwrapperEnableKey = "enableVastUnwrapper"
	RequestContext         = "rctx"
)

func getVastUnwrapperEnable(ctx context.Context, field string) bool {
	vastEnableUnwrapper, _ := ctx.Value(field).(string)
	return vastEnableUnwrapper == "1"
}

func handleEntrypointHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {

	result := hookstage.HookResult[hookstage.EntrypointPayload]{Reject: true}

	rCtx := models.RequestCtx{
		VastUnwrapFlag: getVastUnwrapperEnable(payload.Request.Context(), VastUnwrapperEnableKey),
	}

	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext[RequestContext] = rCtx

	result.Reject = false
	return result, nil
}
