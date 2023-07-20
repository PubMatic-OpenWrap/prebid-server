package vastunwrap

import (
	"context"
	"math/rand"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
)

func getVastUnwrapperEnable(ctx context.Context, field string) bool {
	vastEnableUnwrapper, _ := ctx.Value(field).(string)
	return vastEnableUnwrapper == "1"
}

var getRandomNumber = func() int {
	return rand.Intn(100)
}

func handleEntrypointHook(
	_ context.Context,
	_ hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload, config VastUnwrapModule,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("body:" + string(payload.Body) + ". stacktrace:" + string(debug.Stack()))
		}
	}()
	result := hookstage.HookResult[hookstage.EntrypointPayload]{}
	vastRequestContext := models.RequestCtx{
		VastUnwrapEnabled: getVastUnwrapperEnable(payload.Request.Context(), VastUnwrapEnabled) && getRandomNumber() < config.TrafficPercentage,
	}
	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext[RequestContext] = vastRequestContext
	return result, nil
}
