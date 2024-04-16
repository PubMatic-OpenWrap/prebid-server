package server

import (
	"net/http"

	"github.com/prebid/prebid-server/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func initPrometheusStatsEndpoint(cfg *config.Configuration, gatherer *prometheus.Registry, promMux *http.ServeMux) {
	promMux.Handle("/stats", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		ErrorLog:            loggerForPrometheus{},
		MaxRequestsInFlight: 5,
		Timeout:             cfg.Metrics.Prometheus.Timeout(),
	}))
}

func initPrometheusMetricsEndpoint(cfg *config.Configuration, promMux *http.ServeMux) {
	promMux.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorLog:            loggerForPrometheus{},
		MaxRequestsInFlight: 5,
		Timeout:             cfg.Metrics.Prometheus.Timeout(),
	}))
}

func getOpenWrapPrometheusServer(cfg *config.Configuration, gatherer *prometheus.Registry, server *http.Server) *http.Server {
	promMux := http.NewServeMux()
	initPrometheusStatsEndpoint(cfg, gatherer, promMux)
	initPrometheusMetricsEndpoint(cfg, promMux)
	server.Handler = promMux
	return server
}
