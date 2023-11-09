package vastunwrap

import (
	"context"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
)

func getValueFromContext(ctx context.Context) (int, int) {
	profileId := ctx.Value(ProfileId).(int)
	versionId := ctx.Value(VersionId).(int)
	return profileId, versionId
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
	profileId, versionId := getValueFromContext(payload.Request.Context())
	vastRequestContext := models.RequestCtx{
		ProfileID: profileId,
		VersionID: versionId,
		// DisplayID: versionId,
		Endpoint: payload.Request.URL.Path,
	}
	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext[RequestContext] = vastRequestContext
	return result, nil
}
