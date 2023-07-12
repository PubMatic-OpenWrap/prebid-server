package vastunwrap

import (
	"context"
	"encoding/json"
	"fmt"

	vastunwrap "git.pubmatic.com/vastunwrap"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/stats"
)

type VastUnwrapModule struct {
	Cfg               unWrapCfg.VastUnWrapCfg `mapstructure:"VastUnWrapCfg" json:"VastUnWrapCfg"`
	TrafficPercentage int                     `mapstructure:"TrafficPercentage" json:"TrafficPercentage"`
	Enabled           bool                    `mapstructure:"enabled" json:"enabled"`
	MetricsEngine     stats.MetricsEngine
}

func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
	return initVastUnrap(rawCfg, deps)
}

func initVastUnrap(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (VastUnwrapModule, error) {
	vastUnwrapModuleCfg := VastUnwrapModule{}

	err := json.Unmarshal(rawCfg, &vastUnwrapModuleCfg)
	if err != nil {
		return VastUnwrapModule{}, fmt.Errorf("invalid vastunwrap config: %v", err)
	}

	if vastUnwrapModuleCfg.Enabled {
		vastunwrap.InitUnWrapperConfig(vastUnwrapModuleCfg.Cfg)
	}
	metricEngine := stats.NewMetricsEngine(deps)

	return VastUnwrapModule{
		Cfg:               vastUnwrapModuleCfg.Cfg,
		TrafficPercentage: vastUnwrapModuleCfg.TrafficPercentage,
		Enabled:           vastUnwrapModuleCfg.Enabled,
		MetricsEngine:     metricEngine,
	}, nil
}

// HandleRawBidderResponseHook fetches rCtx and check for vast unwrapper flag to enable/disable vast unwrapping feature
func (m VastUnwrapModule) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {

	if m.Enabled {
		return handleRawBidderResponseHook(m, miCtx, payload, UnwrapURL)
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
