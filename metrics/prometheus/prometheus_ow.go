package prometheusmetrics

import (
	"strconv"
	"time"

	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/metrics"
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

type OWMetrics struct {
	vastTagType *prometheus.CounterVec
	// Rejected Bids
	rejectedBids *prometheus.CounterVec
	bids         *prometheus.CounterVec
	vastVersion  *prometheus.CounterVec
	//rejectedBids         *prometheus.CounterVec
	accountRejectedBid   *prometheus.CounterVec
	accountFloorsRequest *prometheus.CounterVec

	//Dynamic Fetch Failure
	dynamicFetchFailure           *prometheus.CounterVec
	adapterDuplicateBidIDCounter  *prometheus.CounterVec
	requestsDuplicateBidIDCounter prometheus.Counter // total request having duplicate bid.id for given bidder
	adapterVideoBidDuration       *prometheus.HistogramVec

	// podImpGenTimer indicates time taken by impression generator
	// algorithm to generate impressions for given ad pod request
	podImpGenTimer *prometheus.HistogramVec

	// podImpGenTimer indicates time taken by combination generator
	// algorithm to generate combination based on bid response and ad pod request
	podCombGenTimer *prometheus.HistogramVec

	// podCompExclTimer indicates time taken by compititve exclusion
	// algorithm to generate final pod response based on bid response and ad pod request
	podCompExclTimer *prometheus.HistogramVec
	httpCounter      prometheus.Counter
}

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
func (m *OWMetrics) RecordAdapterDuplicateBidID(adaptor string, collisions int) {
	m.adapterDuplicateBidIDCounter.With(prometheus.Labels{
		adapterLabel: adaptor,
	}).Add(float64(collisions))
}

// RecordRequestHavingDuplicateBidID keeps count of request when duplicate bid.id is
// detected in partner's response
func (m *OWMetrics) RecordRequestHavingDuplicateBidID() {
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
func (m *OWMetrics) RecordPodImpGenTime(labels metrics.PodLabels, start time.Time) {
	elapsedTime := time.Since(start)
	recordAlgoTime(m.podImpGenTimer, labels, elapsedTime)
}

// RecordPodCombGenTime records number of combinations generated and time taken
// by underneath algorithm to generate them
func (m *OWMetrics) RecordPodCombGenTime(labels metrics.PodLabels, elapsedTime time.Duration) {
	recordAlgoTime(m.podCombGenTimer, labels, elapsedTime)
}

// RecordPodCompititveExclusionTime records number of combinations comsumed for forming
// final ad pod response and time taken by underneath algorithm to generate them
func (m *OWMetrics) RecordPodCompititveExclusionTime(labels metrics.PodLabels, elapsedTime time.Duration) {
	recordAlgoTime(m.podCompExclTimer, labels, elapsedTime)
}

// RecordAdapterVideoBidDuration records actual ad duration (>0) returned by the bidder
func (m *OWMetrics) RecordAdapterVideoBidDuration(labels metrics.AdapterLabels, videoBidDuration int) {
	if videoBidDuration > 0 {
		m.adapterVideoBidDuration.With(prometheus.Labels{adapterLabel: string(labels.Adapter)}).Observe(float64(videoBidDuration))
	}
}

// RecordRejectedBids records rejected bids labeled by pubid, bidder and reason code
func (m *OWMetrics) RecordRejectedBids(pubid, biddder, code string) {
	m.rejectedBids.With(prometheus.Labels{
		pubIDLabel:  pubid,
		bidderLabel: biddder,
		codeLabel:   code,
	}).Inc()
}

// RecordBids records bids labeled by pubid, profileid, bidder and deal
func (m *OWMetrics) RecordBids(pubid, profileid, biddder, deal string) {
	m.bids.With(prometheus.Labels{
		pubIDLabel:   pubid,
		profileLabel: profileid,
		bidderLabel:  biddder,
		dealLabel:    deal,
	}).Inc()
}

// RecordVastVersion record the count of vast version labelled by bidder and vast version
func (m *OWMetrics) RecordVastVersion(coreBiddder, vastVersion string) {
	m.vastVersion.With(prometheus.Labels{
		adapterLabel: coreBiddder,
		versionLabel: vastVersion,
	}).Inc()
}

// RecordVASTTagType record the count of vast tags labeled by bidder and vast tag
func (m *OWMetrics) RecordVASTTagType(bidder, vastTagType string) {
	m.vastTagType.With(prometheus.Labels{
		bidderLabel:      bidder,
		vastTagTypeLabel: vastTagType,
	}).Inc()
}
func (m *Metrics) RecordRejectedBidsForAccount(pubId string) {
	if pubId != metrics.PublisherUnknown {
		m.accountRejectedBid.With(prometheus.Labels{
			accountLabel: pubId,
		}).Inc()
	}
}

func (m *Metrics) RecordFloorsRequestForAccount(pubId string) {
	if pubId != metrics.PublisherUnknown {
		m.accountFloorsRequest.With(prometheus.Labels{
			accountLabel: pubId,
		}).Inc()
	}
}
func (m *Metrics) RecordFloorStatus(pubId, source, code string) {
	if pubId != metrics.PublisherUnknown {
		m.dynamicFetchFailure.With(prometheus.Labels{
			accountLabel: pubId,
			sourceLabel:  source,
			codeLabel:    code,
		}).Inc()
	}
}

func (m *Metrics) RecordHttpCounter() {
	m.httpCounter.Inc()
}

func (m *OWMetrics) init(cfg config.PrometheusMetrics, reg *prometheus.Registry) {
	m.httpCounter = newHttpCounter(cfg, reg)
	m.rejectedBids = newCounter(cfg, reg,
		"rejected_bids",
		"Count of rejected bids by publisher id, bidder and rejection reason code",
		[]string{pubIDLabel, bidderLabel, codeLabel})

	m.vastVersion = newCounter(cfg, reg,
		"vast_version",
		"Count of vast version by bidder and vast version",
		[]string{adapterLabel, versionLabel})

	m.vastTagType = newCounter(cfg, reg,
		"vast_tag_type",
		"Count of vast tag by bidder and vast tag type (Wrapper, InLine, URL, Unknown)",
		[]string{bidderLabel, vastTagTypeLabel})

	m.dynamicFetchFailure = newCounter(cfg, reg,
		"floors_account_status",
		"Count of failures in case of dynamic fetch labeled by account",
		[]string{accountLabel, codeLabel, sourceLabel})

	m.adapterDuplicateBidIDCounter = newCounter(cfg, reg,
		"duplicate_bid_ids",
		"Number of collisions observed for given adaptor",
		[]string{adapterLabel})

	m.requestsDuplicateBidIDCounter = newCounterWithoutLabels(cfg, reg,
		"requests_having_duplicate_bid_ids",
		"Count of number of request where bid collision is detected.")

	m.adapterVideoBidDuration = newHistogramVec(cfg, reg,
		"adapter_vidbid_dur",
		"Video Ad durations returned by the bidder", []string{adapterLabel},
		[]float64{4, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 120})

	m.bids = newCounter(cfg, reg,
		"bids",
		"Count of no of bids by publisher id, profile, bidder and deal",
		[]string{pubIDLabel, profileLabel, bidderLabel, dealLabel})

	m.accountRejectedBid = newCounter(cfg, reg,
		"floors_account_rejected_bid_requests",
		"Count of total requests to Prebid Server that have rejected bids due to floors enfocement labled by account",
		[]string{accountLabel})

	m.accountFloorsRequest = newCounter(cfg, reg,
		"floors_account_requests",
		"Count of total requests to Prebid Server that have non-zero imp.bidfloor labled by account",
		[]string{accountLabel})

	// adpod specific metrics
	m.podImpGenTimer = newHistogramVec(cfg, reg,
		"impr_gen",
		"Time taken by Ad Pod Impression Generator in seconds", []string{podAlgorithm, podNoOfImpressions},
		// 200 µS, 250 µS, 275 µS, 300 µS
		//[]float64{0.000200000, 0.000250000, 0.000275000, 0.000300000})
		// 100 µS, 200 µS, 300 µS, 400 µS, 500 µS,  600 µS,
		[]float64{0.000100000, 0.000200000, 0.000300000, 0.000400000, 0.000500000, 0.000600000})

	m.podCombGenTimer = newHistogramVec(cfg, reg,
		"comb_gen",
		"Time taken by Ad Pod Combination Generator in seconds", []string{podAlgorithm, podTotalCombinations},
		// 200 µS, 250 µS, 275 µS, 300 µS
		//[]float64{0.000200000, 0.000250000, 0.000275000, 0.000300000})
		[]float64{0.000100000, 0.000200000, 0.000300000, 0.000400000, 0.000500000, 0.000600000})

	m.podCompExclTimer = newHistogramVec(cfg, reg,
		"comp_excl",
		"Time taken by Ad Pod Compititve Exclusion in seconds", []string{podAlgorithm, podNoOfResponseBids},
		// 200 µS, 250 µS, 275 µS, 300 µS
		//[]float64{0.000200000, 0.000250000, 0.000275000, 0.000300000})
		[]float64{0.000100000, 0.000200000, 0.000300000, 0.000400000, 0.000500000, 0.000600000})

}
