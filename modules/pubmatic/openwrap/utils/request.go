package utils

import (
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func GetRequestContext(invocationContext hookstage.ModuleInvocationContext) (models.RequestCtx, bool) {
	if invocationContext.ModuleContext == nil {
		return models.RequestCtx{}, false
	}

	requestContext, ok := invocationContext.ModuleContext.Get(models.RequestContext)
	if !ok {
		return models.RequestCtx{}, false
	}

	requestContextValue, ok := requestContext.(models.RequestCtx)
	if !ok {
		return models.RequestCtx{}, false
	}

	return requestContextValue, true
}
