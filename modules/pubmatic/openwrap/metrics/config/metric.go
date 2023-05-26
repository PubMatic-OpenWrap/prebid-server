package config

import (
	"github.com/pm-nilesh-chate/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/pm-nilesh-chate/prebid-server/modules/pubmatic/openwrap/metrics/stats"
	ow_cfg "github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
)

// NewMetricsEngine reads the configuration and returns the appropriate metrics engine
// for this instance.
func NewMetricsEngine(cfg ow_cfg.Metrics, host string) *DetailedMetricsEngine {
	// Create a list of metrics engines to use.
	// Capacity of 2, as unlikely to have more than 2 metrics backends, and in the case
	// of 1 we won't use the list so it will be garbage collected.
	engineList := make(MultiMetricsEngine, 0, 2)
	returnEngine := DetailedMetricsEngine{}

	if cfg.Stats.StatsHost != "" {
		// setup stats-server

		returnEngine, err := stats.InitStat(host, cfg.Stats.DefaultHostName, cfg.Stats.StatsHost,
			cfg.Stats.StandardInterval, cfg.Stats.CriticalThreshold, cfg.Stats.CriticalInterval,
			cfg.Stats.StandardThreshold, cfg.Stats.StandardInterval, cfg.Stats.StatsPort,
			cfg.Stats.PublishInterval, cfg.Stats.PublishThreshold, cfg.Stats.Retries, cfg.Stats.DialTimeout,
			cfg.Stats.KeepAliveDuration, cfg.Stats.MaxIdleConnections, cfg.Stats.MaxIdleConnectionsPerHost,
		)

		if err == nil && returnEngine != nil {
			engineList = append(engineList, returnEngine)
		}

	}
	if cfg.Prometheus.Port != 0 {
		// Set up the Prometheus metrics.
		// returnEngine.PrometheusMetrics = prometheusmetrics.NewMetrics(cfg.Prometheus, cfg.Disabled, syncerKeys, moduleStageNames)
		// engineList = append(engineList, returnEngine.PrometheusMetrics)
	}

	// Now return the proper metrics engine
	if len(engineList) > 1 {
		returnEngine.MetricsEngine = &engineList
	} else if len(engineList) == 1 {
		returnEngine.MetricsEngine = engineList[0]
	} else {
		returnEngine.MetricsEngine = &NilMetricsEngine{}
	}

	return &returnEngine
}

// DetailedMetricsEngine is a MultiMetricsEngine that preserves links to underlying metrics engines.
type DetailedMetricsEngine struct {
	metrics.MetricsEngine
	// GoMetrics         *metrics.Metrics
	// PrometheusMetrics *prometheusmetrics.Metrics
	Stats *stats.StatsTCP
}

// MultiMetricsEngine logs metrics to multiple metrics databases The can be useful in transitioning
// an instance from one engine to another, you can run both in parallel to verify stats match up.
type MultiMetricsEngine []metrics.MetricsEngine

// RecordRequest across all engines
func (me *MultiMetricsEngine) RecordOpenWrapServerPanicStats() {
	for _, thisME := range *me {
		thisME.RecordOpenWrapServerPanicStats()
	}
}

// NilMetricsEngine implements the MetricsEngine interface where no metrics are actually captured. This is
// used if no metric backend is configured and also for tests.
type NilMetricsEngine struct{}

func (me *NilMetricsEngine) RecordOpenWrapServerPanicStats() {}
