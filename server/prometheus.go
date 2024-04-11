package server

import (
	"net/http"
	"strconv"

	"github.com/golang/glog"

	"github.com/prebid/prebid-server/config"
	metricsconfig "github.com/prebid/prebid-server/metrics/config"
)

func newPrometheusServer(cfg *config.Configuration, metrics *metricsconfig.DetailedMetricsEngine) *http.Server {
	proMetrics := metrics.PrometheusMetrics

	if proMetrics == nil {
		glog.Fatal("Prometheus metrics configured, but a Prometheus metrics engine was not found. Cannot set up a Prometheus listener.")
	}
	return &http.Server{
		Addr: cfg.Host + ":" + strconv.Itoa(cfg.Metrics.Prometheus.Port),
	}
}

type loggerForPrometheus struct{}

func (loggerForPrometheus) Println(v ...interface{}) {
	glog.Warningln(v...)
}
