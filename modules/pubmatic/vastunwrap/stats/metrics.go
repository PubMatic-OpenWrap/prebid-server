package metrics

import (
	"errors"
	"time"

	"github.com/prebid/prebid-server/config"
	metrics_cfg "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/modules/moduledeps"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	bidderLabel       = "bidder"
	pubIdLabel        = "pub_id"
	statusLabel       = "status"
	wrapperCountLabel = "wrapper_count"
)

// MetricsEngine is a generic interface to record metrics into the desired backend
type MetricsEngine interface {
	RecordRequestStatus(bidder, status string)
	RecordWrapperCount(bidder string, wrapper_count string)
	RecordRequestTime(bidder string, readTime time.Duration)
}

// Metrics defines the datatype which will implement MetricsEngine
type Metrics struct {
	Registry     *prometheus.Registry
	requests     *prometheus.CounterVec
	wrapperCount *prometheus.CounterVec
	requestTime  *prometheus.HistogramVec
}

// NewMetricsEngine reads the configuration and returns the appropriate metrics engine
// for this instance.
func NewMetricsEngine(cfg moduledeps.ModuleDeps) (*Metrics, error) {
	metrics := Metrics{}
	// Set up the Prometheus metrics engine.
	if cfg.MetricsCfg != nil && cfg.MetricsRegistry != nil && cfg.MetricsRegistry[metrics_cfg.PrometheusRegistry] != nil {
		prometheusRegistry, ok := cfg.MetricsRegistry[metrics_cfg.PrometheusRegistry].(*prometheus.Registry)
		if prometheusRegistry == nil {
			return &metrics, errors.New("Prometheus registry is nil")
		}
		if ok && prometheusRegistry != nil {
			metrics.Registry = prometheusRegistry
		}
	}
	metrics.requests = newCounter(cfg.MetricsCfg.Prometheus, metrics.Registry,
		"vastunwrap_status",
		"Count of vast unwrap requests labeled by status",
		[]string{bidderLabel, statusLabel})
	metrics.wrapperCount = newCounter(cfg.MetricsCfg.Prometheus, metrics.Registry,
		"vastunwrap_wrapper_count",
		"Count of vast unwrap levels labeled by bidder",
		[]string{bidderLabel, wrapperCountLabel})
	metrics.requestTime = newHistogramVec(cfg.MetricsCfg.Prometheus, metrics.Registry,
		"vastunwrap_request_time",
		"Time taken to serve the vast unwrap request in Milliseconds", []string{bidderLabel},
		[]float64{50, 100, 200, 300, 500})
	return &metrics, nil
}

func newCounter(cfg config.PrometheusMetrics, registry *prometheus.Registry, name, help string, labels []string) *prometheus.CounterVec {
	opts := prometheus.CounterOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      name,
		Help:      help,
	}
	counter := prometheus.NewCounterVec(opts, labels)
	registry.MustRegister(counter)
	return counter
}

func newHistogramVec(cfg config.PrometheusMetrics, registry *prometheus.Registry, name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	opts := prometheus.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}
	histogram := prometheus.NewHistogramVec(opts, labels)
	registry.MustRegister(histogram)
	return histogram
}

// RecordRequest record counter with vast unwrap status
func (m *Metrics) RecordRequestStatus(bidder, status string) {
	m.requests.With(prometheus.Labels{
		bidderLabel: bidder,
		statusLabel: status,
	}).Inc()
}

// RecordWrapperCount record counter of wrapper levels
func (m *Metrics) RecordWrapperCount(bidder, wrapper_count string) {
	m.wrapperCount.With(prometheus.Labels{
		bidderLabel:       bidder,
		wrapperCountLabel: wrapper_count,
	}).Inc()
}

// RecordRequestReadTime records time takent to complete vast unwrap
func (m *Metrics) RecordRequestTime(bidder string, requestTime time.Duration) {
	m.requestTime.With(prometheus.Labels{
		bidderLabel: bidder,
	}).Observe(float64(requestTime.Milliseconds()))
}
