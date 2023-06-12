package vastunwrap

import (
	"context"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
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
	defer func() {
		if r := recover(); r != nil {
			glog.Error("body:" + string(payload.Body) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	result := hookstage.HookResult[hookstage.EntrypointPayload]{
		Reject: true,
	}

	vastRequestContext := models.RequestCtx{
		IsVastUnwrapEnabled: getVastUnwrapperEnable(payload.Request.Context(), isVastUnWrapEnabled),
	}
	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext[RequestContext] = vastRequestContext

	result.Reject = false
	return result, nil
}
