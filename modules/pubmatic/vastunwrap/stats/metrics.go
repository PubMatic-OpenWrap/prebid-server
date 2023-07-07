package stats

import (
	"time"

	"github.com/prebid/prebid-server/modules/moduledeps"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsEngine is a generic interface to record metrics into the desired backend
type MetricsEngine interface {
	RecordRequestStatus(pubID, bidder, status string)
	RecordRequestTime(pubID string, bidder string, readTime time.Duration)
}

// Metrics defines the datatype which will implement MetricsEngine
type Metrics struct {
	Registry    *prometheus.Registry
	requests    *prometheus.CounterVec
	requestTime *prometheus.HistogramVec
}

// NewMetricsEngine reads the configuration and returns the appropriate metrics engine
// for this instance.
func NewMetricsEngine(cfg moduledeps.ModuleDeps) *Metrics {

	metrics := Metrics{}
	metrics.Registry = cfg.Registry

	metrics.requests = newCounter(cfg, metrics.Registry,
		"vastunwrap_status",
		"Count of vast unwrap requests labeled by publisher ID, bidder and status.",
		[]string{"pubID", "bidder", "status"})

	metrics.requestTime = newHistogramVec(cfg, metrics.Registry,
		"vastunwrap_request_time",
		"Time taken to serve the request in seconds", []string{"pubID", "bidder"},
		[]float64{0.05, 0.1, 0.15, 0.20, 0.25, 0.3, 0.4, 0.5, 0.75, 1})

	return &metrics
}

func newCounter(cfg moduledeps.ModuleDeps, registry *prometheus.Registry, name, help string, labels []string) *prometheus.CounterVec {
	opts := prometheus.CounterOpts{
		Namespace: cfg.PrometheusMetrics.Namespace,
		Subsystem: cfg.PrometheusMetrics.Subsystem,
		Name:      name,
		Help:      help,
	}
	counter := prometheus.NewCounterVec(opts, labels)
	registry.MustRegister(counter)
	return counter
}

func newHistogramVec(cfg moduledeps.ModuleDeps, registry *prometheus.Registry, name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	opts := prometheus.HistogramOpts{
		Namespace: cfg.PrometheusMetrics.Namespace,
		Subsystem: cfg.PrometheusMetrics.Subsystem,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}
	histogram := prometheus.NewHistogramVec(opts, labels)
	registry.MustRegister(histogram)
	return histogram
}

// RecordRequest record counter with vast unwrap status
func (m *Metrics) RecordRequestStatus(pubID, bidder, status string) {
	m.requests.With(prometheus.Labels{
		"pubID":  pubID,
		"bidder": bidder,
		"status": status,
	}).Inc()
}

// RecordRequestReadTime records time takent to complete vast unwrap
func (m *Metrics) RecordRequestTime(pubId string, bidder string, requestTime time.Duration) {
	m.requestTime.With(prometheus.Labels{
		"pubID":  pubId,
		"bidder": bidder,
	}).Observe(float64(requestTime.Milliseconds()))
}
