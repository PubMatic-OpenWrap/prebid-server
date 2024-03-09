package config

import (
	"testing"

	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/metrics"
	prometheusmetrics "github.com/prebid/prebid-server/v2/metrics/prometheus"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func TestGoMetricsEngineForNilRegistry(t *testing.T) {
	cfg := config.Configuration{}
	cfg.Metrics.Influxdb.Host = "localhost"
	adapterList := make([]openrtb_ext.BidderName, 0, 2)
	syncerKeys := []string{"keyA", "keyB"}
	testEngine := NewMetricsEngine(&cfg, nil, adapterList, syncerKeys, modulesStages)
	_, ok := testEngine.MetricsEngine.(*metrics.Metrics)
	if !ok {
		t.Error("Expected a Metrics as MetricsEngine, but didn't get it")
	}
}

func TestPrometheusMetricsEngine(t *testing.T) {

	adapterList := make([]openrtb_ext.BidderName, 0, 2)
	syncerKeys := []string{"keyA", "keyB"}

	type args struct {
		cfg             *config.Configuration
		metricsRegistry MetricsRegistry
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "nil_prometheus_registry",
			args: args{
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9090,
							Namespace:        "ow",
							Subsystem:        "pbs",
							TimeoutMillisRaw: 5,
						},
					},
				},
				metricsRegistry: MetricsRegistry{
					PrometheusRegistry: nil,
				},
			},
		},
		{
			name: "valid_prometheus_registry",
			args: args{
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9090,
							Namespace:        "ow",
							Subsystem:        "pbs",
							TimeoutMillisRaw: 5,
						},
					},
				},
				metricsRegistry: NewMetricsRegistry(),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			testEngine := NewMetricsEngine(test.args.cfg, test.args.metricsRegistry, adapterList, syncerKeys, modulesStages)
			_, ok := testEngine.MetricsEngine.(*prometheusmetrics.Metrics)
			if !ok {
				t.Error("Expected a Metrics as MetricsEngine, but didn't get it")
			}
		})
	}
}
