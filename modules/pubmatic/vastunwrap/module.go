package vastunwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	vastunwrap "git.pubmatic.com/vastunwrap"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/stats"
)

type VastUnwrapModule struct {
	Cfg               unWrapCfg.VastUnWrapCfg `mapstructure:"vastunwrap_cfg" json:"vastunwrap_cfg"`
	TrafficPercentage int                     `mapstructure:"traffic_percentage" json:"traffic_percentage"`
	Enabled           bool                    `mapstructure:"enabled" json:"enabled"`
	MetricsEngine     metrics.MetricsEngine
}

func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
	glog.Infof("Starting Init vast unwrap.")
	return initVastUnwrap(rawCfg, deps)
}

func initVastUnwrap(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (VastUnwrapModule, error) {
	t := time.Now()
	defer glog.Infof("Time taken by initVastUnwrap---%v", time.Since(t).Milliseconds())
	vastUnwrapModuleCfg := VastUnwrapModule{}
	err := json.Unmarshal(rawCfg, &vastUnwrapModuleCfg)
	if err != nil {
		return vastUnwrapModuleCfg, fmt.Errorf("invalid vastunwrap config: %v", err)
	}
	vastunwrap.InitUnWrapperConfig(vastUnwrapModuleCfg.Cfg)
	metricEngine, err := metrics.NewMetricsEngine(deps)
	if err != nil {
		return vastUnwrapModuleCfg, fmt.Errorf("Prometheus registry is nil")
	}
	glog.Infof("Init vast unwrap completed.")

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
	t := time.Now()
	defer glog.Infof("Time taken by HandleRawBidderResponseHook---%v", time.Since(t).Milliseconds())
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
	t := time.Now()
	defer glog.Infof("Time taken by HandleEntrypointHook---%v", time.Since(t).Milliseconds())
	if m.Enabled {
		return handleEntrypointHook(ctx, miCtx, payload, m)
	}
	return hookstage.HookResult[hookstage.EntrypointPayload]{}, nil
}
