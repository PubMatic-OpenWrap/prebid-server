package metrics

import (
	"testing"
	"time"

	"github.com/prebid/prebid-server/modules/moduledeps"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func createMetricsForTesting() *Metrics {
	cfg := moduledeps.ModuleDeps{Registry: prometheus.NewRegistry()}
	return NewMetricsEngine(cfg)
}
func TestRecordRequestTime(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordRequestTime("5890", "pubmatic", time.Millisecond*250)

	result := getHistogramFromHistogramVec(m.requestTime, "pub_id", "5890")
	assertHistogram(t, result, 1, 250)
}
func TestRecordRequestStatus(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordRequestStatus("5890", "pubmatic", "0")

	assertCounterVecValue(t, "Record_Request_Status", "Record_Request_Status_Success", m.requests, float64(1), prometheus.Labels{
		"pub_id": "5890",
		"bidder": "pubmatic",
		"status": "0",
	})
}

func assertCounterValue(t *testing.T, description, name string, counter prometheus.Counter, expected float64) {
	m := dto.Metric{}
	counter.Write(&m)
	actual := *m.GetCounter().Value

	assert.Equal(t, expected, actual, description)
}

func assertCounterVecValue(t *testing.T, description, name string, counterVec *prometheus.CounterVec, expected float64, labels prometheus.Labels) {
	counter := counterVec.With(labels)
	assertCounterValue(t, description, name, counter, expected)
}

func assertHistogram(t *testing.T, histogram dto.Histogram, expectedCount uint64, expectedSum float64) {
	assert.Equal(t, expectedCount, histogram.GetSampleCount())
	assert.Equal(t, expectedSum, histogram.GetSampleSum())
}
func getHistogramFromHistogramVec(histogram *prometheus.HistogramVec, labelKey, labelValue string) dto.Histogram {
	var result dto.Histogram
	processMetrics(histogram, func(m dto.Metric) {
		for _, label := range m.GetLabel() {
			if label.GetName() == labelKey && label.GetValue() == labelValue {
				result = *m.GetHistogram()
			}
		}
	})
	return result
}

func processMetrics(collector prometheus.Collector, handler func(m dto.Metric)) {
	collectorChan := make(chan prometheus.Metric)
	go func() {
		collector.Collect(collectorChan)
		close(collectorChan)
	}()

	for metric := range collectorChan {
		dtoMetric := dto.Metric{}
		metric.Write(&dtoMetric)
		handler(dtoMetric)
	}
}
