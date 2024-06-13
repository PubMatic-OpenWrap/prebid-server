package openwrap

import (
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"

	ow_metrics_prometheus "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type loggerForPrometheus struct{}

func (loggerForPrometheus) Println(v ...interface{}) {
	glog.Warningln(v...)
}

func PrometheusHandler(reg prometheus.Gatherer, timeDur time.Duration, endpoint string) http.Handler {
	handler := promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			ErrorLog:            loggerForPrometheus{},
			MaxRequestsInFlight: 5,
			Timeout:             timeDur,
		},
	)
	return http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		ow.metricEngine.RecordRequest(
			metrics.Labels{
				RType:         metrics.RequestType(endpoint),
				RequestStatus: ow_metrics_prometheus.RequestStatusOK,
			},
		)
		handler.ServeHTTP(rsp, req)
	})
}

func initPrometheusStatsEndpoint(cfg *config.Configuration, gatherer *prometheus.Registry, promMux *http.ServeMux) {
	promMux.Handle("/stats", PrometheusHandler(gatherer, cfg.Metrics.Prometheus.Timeout(), "stats"))
}

func initPrometheusMetricsEndpoint(cfg *config.Configuration, promMux *http.ServeMux) {
	promMux.Handle("/metrics", PrometheusHandler(prometheus.DefaultGatherer, cfg.Metrics.Prometheus.Timeout(), "metrics"))
}

func GetOpenWrapPrometheusServer(cfg *config.Configuration, gatherer *prometheus.Registry, server *http.Server) *http.Server {
	promMux := http.NewServeMux()
	initPrometheusStatsEndpoint(cfg, gatherer, promMux)
	initPrometheusMetricsEndpoint(cfg, promMux)
	server.Handler = promMux
	return server
}
