package config

import (
	"github.com/prometheus/client_golang/prometheus"
	gometrics "github.com/rcrowley/go-metrics"
)

type RegistryType = string
type MetricsRegistry map[RegistryType]interface{}

const (
	PrometheusRegistry RegistryType = "prometheus"
	InfluxRegistry     RegistryType = "influx"
)

// NewMetricsRegistry returns the map of metrics-engine-name and its respective registry
func NewMetricsRegistry() MetricsRegistry {
	return MetricsRegistry{
		PrometheusRegistry: prometheus.NewRegistry(),
		InfluxRegistry:     gometrics.NewPrefixedRegistry("prebidserver."),
	}
}
