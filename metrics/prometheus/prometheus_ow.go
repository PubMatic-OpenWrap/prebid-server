package prometheusmetrics

import (
	"strconv"
	"time"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	pubIDLabel       = "pubid"
	bidderLabel      = "bidder"
	codeLabel        = "code"
	profileLabel     = "profileid"
	dealLabel        = "deal"
	vastTagTypeLabel = "type"
)

func newHttpCounter(cfg config.PrometheusMetrics, registry *prometheus.Registry) prometheus.Counter {
	httpCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of http requests.",
	})
	registry.MustRegister(httpCounter)
	return httpCounter
}

// RecordAdapterDuplicateBidID captures the  bid.ID collisions when adaptor
// gives the bid response with multiple bids containing  same bid.ID
// ensure collisions value is greater than 1. This function will not give any error
// if collisions = 1 is passed
func (m *Metrics) RecordAdapterDuplicateBidID(adaptor string, collisions int) {
	m.adapterDuplicateBidIDCounter.With(prometheus.Labels{
		adapterLabel: adaptor,
	}).Add(float64(collisions))
}

// RecordRequestHavingDuplicateBidID keeps count of request when duplicate bid.id is
// detected in partner's response
func (m *Metrics) RecordRequestHavingDuplicateBidID() {
	m.requestsDuplicateBidIDCounter.Inc()
}

// pod specific metrics

// recordAlgoTime is common method which handles algorithm time performance
func recordAlgoTime(timer *prometheus.HistogramVec, labels metrics.PodLabels, elapsedTime time.Duration) {

	pmLabels := prometheus.Labels{
		podAlgorithm: labels.AlgorithmName,
	}

	if labels.NoOfImpressions != nil {
		pmLabels[podNoOfImpressions] = strconv.Itoa(*labels.NoOfImpressions)
	}
	if labels.NoOfCombinations != nil {
		pmLabels[podTotalCombinations] = strconv.Itoa(*labels.NoOfCombinations)
	}
	if labels.NoOfResponseBids != nil {
		pmLabels[podNoOfResponseBids] = strconv.Itoa(*labels.NoOfResponseBids)
	}

	timer.With(pmLabels).Observe(elapsedTime.Seconds())
}

// RecordPodImpGenTime records number of impressions generated and time taken
// by underneath algorithm to generate them
func (m *Metrics) RecordPodImpGenTime(labels metrics.PodLabels, start time.Time) {
	elapsedTime := time.Since(start)
	recordAlgoTime(m.podImpGenTimer, labels, elapsedTime)
}

// RecordPodCombGenTime records number of combinations generated and time taken
// by underneath algorithm to generate them
func (m *Metrics) RecordPodCombGenTime(labels metrics.PodLabels, elapsedTime time.Duration) {
	recordAlgoTime(m.podCombGenTimer, labels, elapsedTime)
}

// RecordPodCompititveExclusionTime records number of combinations comsumed for forming
// final ad pod response and time taken by underneath algorithm to generate them
func (m *Metrics) RecordPodCompititveExclusionTime(labels metrics.PodLabels, elapsedTime time.Duration) {
	recordAlgoTime(m.podCompExclTimer, labels, elapsedTime)
}

// RecordAdapterVideoBidDuration records actual ad duration (>0) returned by the bidder
func (m *Metrics) RecordAdapterVideoBidDuration(labels metrics.AdapterLabels, videoBidDuration int) {
	if videoBidDuration > 0 {
		m.adapterVideoBidDuration.With(prometheus.Labels{adapterLabel: string(labels.Adapter)}).Observe(float64(videoBidDuration))
	}
}

// RecordRejectedBids records rejected bids labeled by pubid, bidder and reason code
func (m *Metrics) RecordRejectedBids(pubid, biddder, code string) {
	m.rejectedBids.With(prometheus.Labels{
		pubIDLabel:  pubid,
		bidderLabel: biddder,
		codeLabel:   code,
	}).Inc()
}

// RecordBids records bids labeled by pubid, profileid, bidder and deal
func (m *Metrics) RecordBids(pubid, profileid, biddder, deal string) {
	m.bids.With(prometheus.Labels{
		pubIDLabel:   pubid,
		profileLabel: profileid,
		bidderLabel:  biddder,
		dealLabel:    deal,
	}).Inc()
}

// RecordVastVersion record the count of vast version labelled by bidder and vast version
func (m *Metrics) RecordVastVersion(coreBiddder, vastVersion string) {
	m.vastVersion.With(prometheus.Labels{
		adapterLabel: coreBiddder,
		versionLabel: vastVersion,
	}).Inc()
}

func (m *Metrics) RecordHttpCounter() {
	m.httpCounter.Inc()
}

// RecordVASTTagType record the count of vast tags labeled by bidder and vast tag
func (m *Metrics) RecordVASTTagType(bidder, vastTagType string) {
	m.vastTagType.With(prometheus.Labels{
		bidderLabel:      bidder,
		vastTagTypeLabel: vastTagType,
	}).Inc()
}
