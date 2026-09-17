package openwrap

import (
	"github.com/prebid/prebid-server/v4/hooks/hookstage"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
)

func getRequestCtx(mc *hookstage.ModuleContext) (models.RequestCtx, bool) {
	if mc == nil {
		return models.RequestCtx{}, false
	}
	val, ok := mc.Get(models.RequestContext)
	if !ok {
		return models.RequestCtx{}, false
	}
	rCtx, ok := val.(models.RequestCtx)
	return rCtx, ok
}

func setRequestCtx(mc *hookstage.ModuleContext, rCtx models.RequestCtx) {
	if mc == nil {
		return
	}
	mc.Set(models.RequestContext, rCtx)
}

func hasModuleContext(mc *hookstage.ModuleContext) bool {
	if mc == nil {
		return false
	}
	return len(mc.GetAll()) > 0
}

func newModuleContextWithRequestCtx(rCtx models.RequestCtx) *hookstage.ModuleContext {
	mc := hookstage.NewModuleContext()
	mc.Set(models.RequestContext, rCtx)
	return mc
}

func moduleContextFromMap(m map[string]any) *hookstage.ModuleContext {
	mc := hookstage.NewModuleContext()
	for k, v := range m {
		mc.Set(k, v)
	}
	return mc
}
