package hookexecution

import (
	"sync"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/privacy"
)

// executionContext holds information passed to module's hook during hook execution.
type executionContext struct {
	endpoint        string
	stage           string
	accountID       string
	account         *config.Account
	moduleContexts  *moduleContexts
	activityControl privacy.ActivityControl
}

func (ctx executionContext) getModuleContext(moduleName string) hookstage.ModuleInvocationContext {
	moduleInvocationCtx := hookstage.ModuleInvocationContext{Endpoint: ctx.endpoint}
	if ctx.moduleContexts != nil {
		if mc, ok := ctx.moduleContexts.get(moduleName); ok {
			moduleInvocationCtx.ModuleContext = mc
		}
	}

	if ctx.account != nil {
		cfg, err := ctx.account.Hooks.Modules.ModuleConfig(moduleName)
		if err != nil {
			glog.Warningf("Failed to get account config for %s module: %s", moduleName, err)
		}

		moduleInvocationCtx.AccountID = ctx.accountID
		moduleInvocationCtx.AccountConfig = cfg
		moduleInvocationCtx.AccountID = ctx.account.ID

	}

	return moduleInvocationCtx
}

// moduleContexts preserves data the module wants to pass to itself from earlier stages to later stages.
type moduleContexts struct {
	sync.RWMutex
	ctxs map[string]*hookstage.ModuleContext // format: {"module_name": hookstage.ModuleContext}
}

func (mc *moduleContexts) put(moduleName string, mCtx *hookstage.ModuleContext) {
	if mc == nil {
		return
	}

	mc.Lock()
	defer mc.Unlock()

	existingCtx, ok := mc.ctxs[moduleName]
	if !ok {
		// If there's no existing context, just store the new one directly
		mc.ctxs[moduleName] = mCtx
		return
	}

	// If there is an existing context, update it with values from the new context
	// This preserves the existing mutex while updating the data
	for k, v := range mCtx.GetAll() {
		existingCtx.Set(k, v)
	}
}

func (mc *moduleContexts) get(moduleName string) (*hookstage.ModuleContext, bool) {
	mc.RLock()
	defer mc.RUnlock()
	mCtx, ok := mc.ctxs[moduleName]

	return mCtx, ok
}

type stageModuleContext struct {
	groupCtx []groupModuleContext
}

type groupModuleContext map[string]*hookstage.ModuleContext
