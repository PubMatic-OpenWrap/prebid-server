package monitor

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

/*
prometheusMonitor holds histogram required for measuring the
execution time
*/
type prometheusMonitor struct {
	IMonitor
	histogram *prometheus.HistogramVec
	scenario  string
}

/*
newPrometheusMonitor creates an instance of IMonitor with Prometheus
objects and register the monitor object with Prometheus
*/
func newPrometheusMonitor(algorithm string) *prometheusMonitor {

	// responseSize := metrics.NewHistogram(`response_size{path="/foo/bar"}`)

	monitor := &prometheusMonitor{
		// init for measuring execution time
		histogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: algorithm,
				Help: "Time taken by " + algorithm + " in nanoseconds",
				// 10000 (ns) 40000 (ns), 80000 (ns), 100000 (ns)
				Buckets: []float64{10000, 40000, 80000, 100000},
			}, []string{algorithm}),
	}

	// register with prometheus
	prometheus.MustRegister(monitor.histogram)
	return monitor
}

func (monitor prometheusMonitor) MeasureExecutionTime(start time.Time) {
	duration := time.Since(start)
	fmt.Println("Time Taken = ", duration.Nanoseconds())
	monitor.histogram.WithLabelValues(fmt.Sprintf("%sns", monitor.scenario)).Observe(float64(duration.Nanoseconds()))
}

func (monitor *prometheusMonitor) Scenario(scenario string) {
	monitor.scenario = scenario
}
