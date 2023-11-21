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
	_ hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload, config VastUnwrapModule,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("body:[%s] Error:[%v] stacktrace:[%s]", string(payload.Body), r, string(debug.Stack()))
		}
	}()
	result := hookstage.HookResult[hookstage.EntrypointPayload]{}
	vastRequestContext := models.RequestCtx{
		VastUnwrapEnabled: getVastUnwrapperEnable(payload.Request.Context(), VastUnwrapEnabled),
	}

	if !vastRequestContext.VastUnwrapEnabled {
		vastRequestContext.VastUnwrapStatsEnabled = getRandomNumber() < config.StatTrafficPercentage
	}
	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext[RequestContext] = vastRequestContext
	return result, nil
}
