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
	Cfg               unWrapCfg.VastUnWrapCfg `mapstructure:"VastUnWrapCfg" json:"VastUnWrapCfg"`
	TrafficPercentage int                     `mapstructure:"traffic_percentage" json:"traffic_percentage"`
	Enabled           bool                    `mapstructure:"enabled" json:"enabled"`
}

func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
	return initVastUnrap(rawCfg, deps)
}

func initVastUnrap(rawCfg json.RawMessage, _ moduledeps.ModuleDeps) (VastUnwrapModule, error) {
	vastUnwrapModuleCfg := VastUnwrapModule{}

	err := json.Unmarshal(rawCfg, &vastUnwrapModuleCfg)
	if err != nil {
		return VastUnwrapModule{}, fmt.Errorf("invalid vastunwrap config: %v", err)
	}

	if vastUnwrapModuleCfg.Enabled {
		vastunwrap.InitUnWrapperConfig(vastUnwrapModuleCfg.Cfg)
	}

	return VastUnwrapModule{
		Cfg:               vastUnwrapModuleCfg.Cfg,
		TrafficPercentage: vastUnwrapModuleCfg.TrafficPercentage,
		Enabled:           vastUnwrapModuleCfg.Enabled,
	}, nil
}

// HandleRawBidderResponseHook fetches rCtx and check for vast unwrapper flag to enable/disable vast unwrapping feature
func (m VastUnwrapModule) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {

	if m.Enabled {
		return handleRawBidderResponseHook(payload, miCtx.ModuleContext, m.Cfg.APPConfig.UnwrapDefaultTimeout, UnwrapURL)
	}
	return hookstage.HookResult[hookstage.RawBidderResponsePayload]{}, nil
}

// HandleEntrypointHook retrieves vast un-wrapper flag and User-agent provided in request context
func (m VastUnwrapModule) HandleEntrypointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {

	if m.Enabled {
		return handleEntrypointHook(ctx, miCtx, payload, m)
	}
	return hookstage.HookResult[hookstage.EntrypointPayload]{}, nil
}
