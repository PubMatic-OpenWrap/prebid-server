package vastunwrap

import (
	"context"
	"math/rand"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
)

var getRandomNumber = func() int {
	return rand.Intn(100)
}

func getVastUnwrapperEnable(ctx context.Context, field string) bool {
	vastEnableUnwrapper, _ := ctx.Value(field).(string)
	return vastEnableUnwrapper == "1"
}
func handleEntrypointHook(
	_ context.Context,
	_ hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("body:[%s] Error:[%v] stacktrace:[%s]", string(payload.Body), r, string(debug.Stack()))
		}
	}()
	result := hookstage.HookResult[hookstage.EntrypointPayload]{}
	vastRequestContext := models.RequestCtx{}
	queryParams := payload.Request.URL.Query()
	source := queryParams.Get("source")
	if queryParams.Get("sshb") == "1" {
		vastRequestContext = models.RequestCtx{
			VastUnwrapEnabled: getVastUnwrapperEnable(payload.Request.Context(), VastUnwrapEnabled),
			Redirect:          true,
		}
	} else {
		endpoint := openwrap.GetEndpoint(payload.Request.URL.Path, source)
		requestExtWrapper, _ := openwrap.GetRequestWrapper(payload, result, endpoint)
		vastRequestContext = models.RequestCtx{
			ProfileID: requestExtWrapper.ProfileId,
			DisplayID: requestExtWrapper.VersionId,
			Endpoint:  endpoint,
		}
	}
	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext[RequestContext] = vastRequestContext
	return result, nil
}
