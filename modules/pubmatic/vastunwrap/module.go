package vastunwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	vastunwrap "git.pubmatic.com/vastunwrap"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	openwrap "github.com/prebid/prebid-server/modules/pubmatic/openwrap"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/stats"
)

type VastUnwrapModule struct {
	Cfg                   unWrapCfg.VastUnWrapCfg `mapstructure:"vastunwrap_cfg" json:"vastunwrap_cfg"`
	TrafficPercentage     int                     `mapstructure:"traffic_percentage" json:"traffic_percentage"`
	StatTrafficPercentage int                     `mapstructure:"stat_traffic_percentage" json:"stat_traffic_percentage"`
	Enabled               bool                    `mapstructure:"enabled" json:"enabled"`
	MetricsEngine         metrics.MetricsEngine
	unwrapRequest         func(w http.ResponseWriter, r *http.Request)
	getVastUnwrapEnable   func(rctx models.RequestCtx) bool
}

func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
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

	if vastUnwrapModuleCfg.Cfg.StatConfig.UseHostName {
		vastUnwrapModuleCfg.Cfg.ServerConfig.ServerName = openwrap.GetHostName()
	}
	vastunwrap.InitUnWrapperConfig(vastUnwrapModuleCfg.Cfg)
	metricEngine, err := metrics.NewMetricsEngine(deps)
	if err != nil {
		return vastUnwrapModuleCfg, fmt.Errorf("Prometheus registry is nil")
	}
	return VastUnwrapModule{
		Cfg:                   vastUnwrapModuleCfg.Cfg,
		TrafficPercentage:     vastUnwrapModuleCfg.TrafficPercentage,
		StatTrafficPercentage: vastUnwrapModuleCfg.StatTrafficPercentage,
		Enabled:               vastUnwrapModuleCfg.Enabled,
		MetricsEngine:         metricEngine,
		unwrapRequest:         vastunwrap.UnwrapRequest,
		getVastUnwrapEnable:   openwrap.GetVastUnwrapEnabled,
	}, nil
}

// HandleRawBidderResponseHook fetches rCtx and check for vast unwrapper flag to enable/disable vast unwrapping feature
func (m VastUnwrapModule) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	if m.Enabled {
		return m.handleRawBidderResponseHook(miCtx, payload, UnwrapURL)
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
		return handleEntrypointHook(ctx, miCtx, payload)
	}
	return hookstage.HookResult[hookstage.EntrypointPayload]{}, nil
}
