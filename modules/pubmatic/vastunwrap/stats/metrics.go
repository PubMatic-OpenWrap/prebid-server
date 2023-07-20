package metrics

import (
	"github.com/prebid/prebid-server/modules/moduledeps"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	bidderLabel = "bidder"
	pubIdLabel  = "pub_id"
	statusLabel = "status"
)

// MetricsEngine is a generic interface to record metrics into the desired backend
type MetricsEngine interface {
	RecordRequestStatus(pubID, bidder, status string)
	// RecordRequestTime(pubID string, bidder string, readTime time.Duration)
}

// Metrics defines the datatype which will implement MetricsEngine
type Metrics struct {
	Registry *prometheus.Registry
	requests *prometheus.CounterVec
	// requestTime *prometheus.HistogramVec
}

// NewMetricsEngine reads the configuration and returns the appropriate metrics engine
// for this instance.
func NewMetricsEngine(cfg moduledeps.ModuleDeps) *Metrics {
	metrics := Metrics{}
	metrics.Registry = cfg.Registry
	metrics.requests = newCounter(cfg, metrics.Registry,
		"vastunwrap_status",
		"Count of vast unwrap requests labeled by publisher ID, bidder and status.",
		[]string{pubIdLabel, bidderLabel, statusLabel})
	// metrics.requestTime = newHistogramVec(cfg, metrics.Registry,
	// 	"vastunwrap_request_time",
	// 	"Time taken to serve the vast unwrap request in Milliseconds", []string{pubIdLabel, bidderLabel},
	// 	[]float64{50, 100, 200, 300, 500})
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

// func newHistogramVec(cfg moduledeps.ModuleDeps, registry *prometheus.Registry, name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
// 	opts := prometheus.HistogramOpts{
// 		Namespace: cfg.PrometheusMetrics.Namespace,
// 		Subsystem: cfg.PrometheusMetrics.Subsystem,
// 		Name:      name,
// 		Help:      help,
// 		Buckets:   buckets,
// 	}
// 	histogram := prometheus.NewHistogramVec(opts, labels)
// 	registry.MustRegister(histogram)
// 	return histogram
// }

// RecordRequest record counter with vast unwrap status
func (m *Metrics) RecordRequestStatus(pub_id, bidder, status string) {
	m.requests.With(prometheus.Labels{
		pubIdLabel:  pub_id,
		bidderLabel: bidder,
		statusLabel: status,
	}).Inc()
}

// // RecordRequestReadTime records time takent to complete vast unwrap
// func (m *Metrics) RecordRequestTime(pub_id string, bidder string, requestTime time.Duration) {
// 	m.requestTime.With(prometheus.Labels{
// 		pubIdLabel:  pub_id,
// 		bidderLabel: bidder,
// 	}).Observe(float64(requestTime.Milliseconds()))
// }
