package vastunwrap

import (
	"context"
	"fmt"
	"math/rand"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap"
	ow_models "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/vastunwrap/models"
)

var getRandomNumber = func() int {
	return rand.Intn(100)
}

// supportedEndpoints holds the list of endpoints which supports VAST-unwrap feature
var supportedEndpoints = map[string]struct{}{
	ow_models.EndpointVAST:  {},
	ow_models.EndpointVideo: {},
	ow_models.EndpointJson:  {},
	ow_models.EndpointV25:   {},
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
	result := hookstage.HookResult[hookstage.EntrypointPayload]{
		ModuleContext: make(hookstage.ModuleContext),
	}
	vastRequestContext := models.RequestCtx{}
	queryParams := payload.Request.URL.Query()
	source := queryParams.Get("source")
	if queryParams.Get("sshb") == "1" {
		vastRequestContext = models.RequestCtx{
			VastUnwrapEnabled: getVastUnwrapperEnable(payload.Request.Context(), VastUnwrapEnabled),
			Redirect:          true,
			UA:                openwrap.GetRequestUserAgent(payload.Body, payload.Request),
			IP:                openwrap.GetRequestIP(payload.Body, payload.Request),
		}
	} else {
		endpoint := openwrap.GetEndpoint(payload.Request.URL.Path, source)
		if _, ok := supportedEndpoints[endpoint]; !ok {
			result.DebugMessages = append(result.DebugMessages, fmt.Sprintf("%s endpoint does not support vast-unwrap feature", endpoint))
			return result, nil
		}
		requestExtWrapper, _ := openwrap.GetRequestWrapper(payload, result, endpoint)
		vastRequestContext = models.RequestCtx{
			ProfileID: requestExtWrapper.ProfileId,
			DisplayID: requestExtWrapper.VersionId,
			Endpoint:  endpoint,
			UA:        openwrap.GetRequestUserAgent(payload.Body, payload.Request),
			IP:        openwrap.GetRequestIP(payload.Body, payload.Request),
		}
	}
	result.ModuleContext[RequestContext] = vastRequestContext
	return result, nil
}
