package vastunwrap

import (
	"context"
	"encoding/json"
	"fmt"

	vastunwrap "git.pubmatic.com/vastunwrap"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
)

type VastUnwrapModule struct {
	cfg VastUnwrapModuleCfg
}

// VastUnwrapModuleCfg contains the values read from the config file  for vast unwrapper module at boot time
type VastUnwrapModuleCfg struct {
	VastUnWrapCfg unWrapCfg.VastUnWrapCfg
}

func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
	return initVastUnrap(rawCfg, deps)
}

func initVastUnrap(rawCfg json.RawMessage, _ moduledeps.ModuleDeps) (VastUnwrapModule, error) {
	cfg := VastUnwrapModuleCfg{}

	err := json.Unmarshal(rawCfg, &cfg)
	if err != nil {
		return VastUnwrapModule{}, fmt.Errorf("invalid unwrap config: %v", err)
	}

	if cfg.VastUnWrapCfg.Enabled {
		vastunwrap.InitUnWrapperConfig(cfg.VastUnWrapCfg)
	}

	return VastUnwrapModule{
		cfg: cfg,
	}, nil

}

// HandleRawBidderResponseHook fetches rCtx and check for vast unwrapper flag to enable/disable vast unwrapping feature
func (m VastUnwrapModule) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {

	if m.cfg.VastUnWrapCfg.Enabled {
		return handleRawBidderResponseHook(payload, miCtx.ModuleContext, m.cfg.VastUnWrapCfg.APPConfig.UnwrapDefaultTimeout)
	}
	return hookstage.HookResult[hookstage.RawBidderResponsePayload]{}, nil
}

// HandleEntrypointHook retrieves vast un-wrapper flag and User-agent provided in request context
func (m VastUnwrapModule) HandleEntrypointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {

	if m.cfg.VastUnWrapCfg.Enabled {
		return handleEntrypointHook(ctx, miCtx, payload)
	}
	return hookstage.HookResult[hookstage.EntrypointPayload]{}, nil
}
