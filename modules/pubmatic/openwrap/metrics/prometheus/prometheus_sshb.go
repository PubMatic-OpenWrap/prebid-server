package prometheus

import (
	"time"

	"github.com/prebid/prebid-server/config"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	bidderLabel  = "bidder"
	profileLabel = "profileid"
	dealLabel    = "deal"
	nodeal       = "nodeal"
)

const (
	requestStatusLabel = "request_status"
	requestTypeLabel   = "request_type"
	pubIdLabel         = "pub_id"
	partnerLable       = "partner"
	statusLabel        = "status"
	nodeNameLabel      = "node_name"
	podNameLabel       = "pod_name"
	methodNameLabel    = "method_name"
)

// Labels defines the labels that can be attached to the metrics.
type Labels struct {
	RType         RequestType
	RequestStatus RequestStatus
}

// RequestType : Request type enumeration
type RequestType string

// RequestStatus : The request return status
type RequestStatus string

// LurlStatusLabels defines labels applicable for LURL sent
type LurlStatusLabels struct {
	PublisherID string
	Partner     string
	Status      string
}

// LurlBatchStatusLabels defines labels applicable for LURL batche sent
type LurlBatchStatusLabels struct {
	Status string
}

// The request types (endpoints)
const (
	ReqTypeORTB25Web RequestType = "openrtb25-web"
	ReqTypeORTB25App RequestType = "openrtb25-app"
	ReqTypeAMP       RequestType = "amp"
	ReqTypeVideo     RequestType = "video"
)

// Request/return status
const (
	RequestStatusOK       RequestStatus = "ok"
	RequestStatusBadInput RequestStatus = "badinput"
	RequestStatusErr      RequestStatus = "err"
)

// RequestTypes returns all possible values for RequestType
func RequestTypes() []RequestType {
	return []RequestType{
		ReqTypeORTB25Web,
		ReqTypeORTB25App,
		ReqTypeAMP,
		ReqTypeVideo,
	}
}

// RequestStatuses return all possible values for RequestStatus
func RequestStatuses() []RequestStatus {
	return []RequestStatus{
		RequestStatusOK,
		RequestStatusBadInput,
		RequestStatusErr,
	}
}

// MetricsEngine is a generic interface to record header-bidding metrics into the desired backend
type MetricsEngine interface {
	RecordRequest(labels Labels) // ignores adapter. only statusOk and statusErr fom status
	RecordLurlSent(labels LurlStatusLabels)
	RecordLurlBatchSent(labels LurlBatchStatusLabels)
	RecordBids(pubid, profileid, biddder, deal string)
	RecordPrebidTimeoutRequests(pubid, profileid string)
	RecordPartnerTimeoutRequests(pubid, profileid, bidder string)
	RecordCtvUaAccuracy(pubId, status string)
	RecordSendLoggerDataTime(requestType, profileid string, sendTime time.Duration)
	RecordRequestTime(requestType string, requestTime time.Duration)
	RecordOWServerPanic(endpoint, methodName, nodeName, podName string)
}

// newSSHBMetrics initializes a new Prometheus metrics instance with preloaded label values for SSHB service
func newSSHBMetrics(metrics *Metrics, cfg *config.PrometheusMetrics, promRegistry *prometheus.Registry) {
	metrics.owRequests = newCounter(cfg, promRegistry,
		"requests",
		"Count of total requests to header-bidding server labeled by type and status.",
		[]string{requestTypeLabel, requestStatusLabel})

	metrics.sendLoggerData = newHistogramVec(cfg, promRegistry,
		"logger_data_send_time",
		"Time taken to send the wrapper logger body in seconds", []string{endpointLabel, profileIDLabel},
		standardTimeBuckets)

	metrics.owRequestTime = newHistogramVec(cfg, promRegistry,
		"sshb_request_time",
		"Time taken to serve the request in seconds", []string{apiTypeLabel},
		[]float64{0.05, 0.1, 0.15, 0.20, 0.25, 0.3, 0.4, 0.5, 0.75, 1})

	metrics.lurlSent = newCounter(cfg, promRegistry, "lurl_sent", "Count of lurl success, fail, drop and channel_full request sent labeled by publisherID, partner", []string{pubIdLabel, partnerLable, statusLabel})

	metrics.lurlBatchSent = newCounter(cfg, promRegistry, "lurl_batch_sent", "Count of lurl Batch success, fail and drop request sent to wtracker ", []string{statusLabel})

	metrics.bids = newCounter(cfg, promRegistry,
		"bids",
		"Count of bids by publisher id, profile, bidder and deal",
		[]string{pubIDLabel, profileLabel, bidderLabel, dealLabel})

	metrics.prebidTimeoutRequests = newCounter(cfg, promRegistry,
		"request_prebid_timeout",
		"count no of requests in which prebid timeouts",
		[]string{pubIDLabel, profileLabel})

	metrics.partnerTimeoutRequest = newCounter(cfg, promRegistry,
		"request_partner_timeout",
		"count no of requests in which partner timeouts",
		[]string{pubIDLabel, profileLabel, bidderLabel})

	metrics.ctvUaAccuracy = newCounter(cfg, promRegistry,
		"ctv_user_agent_accuracy",
		"Count of requests detected by Ctv user agent regex labeled by pub id and status.",
		[]string{pubIdLabel, statusLabel})

	metrics.panicCounts = newCounter(cfg, promRegistry,
		"panic",
		"Counts the header-bidding server panic.",
		[]string{nodeNameLabel, podNameLabel, methodNameLabel, endpointLabel})

	preloadLabelValues(metrics)
}

// RecordRequest across all engines
func (m *Metrics) RecordRequest(labels Labels) {
	m.owRequests.With(prometheus.Labels{
		requestTypeLabel:   string(labels.RType),
		requestStatusLabel: string(labels.RequestStatus),
	}).Inc()
}

// RecordLurlSent records lurl status success, fail, drop and channel_fool
func (m *Metrics) RecordLurlSent(labels LurlStatusLabels) {
	m.lurlSent.With(prometheus.Labels{
		pubIdLabel:   labels.PublisherID,
		partnerLable: labels.Partner,
		statusLabel:  labels.Status,
	}).Inc()
}

// RecordLurlBatchSent records lurl batchs sent to wtracker
func (m *Metrics) RecordLurlBatchSent(labels LurlBatchStatusLabels) {
	m.lurlBatchSent.With(prometheus.Labels{
		statusLabel: labels.Status,
	}).Inc()
}

// RecordBids records count of  bids labeled by pubid, profileid, bidder and deal
func (m *Metrics) RecordBids(pubid, profileid, bidder, deal string) {
	m.bids.With(prometheus.Labels{
		pubIDLabel:   pubid,
		profileLabel: profileid,
		bidderLabel:  bidder,
		dealLabel:    deal,
	}).Inc()
}

// RecordPrebidTimeoutRequests records count of request in which prebid timedout based on pubid and profileid
func (m *Metrics) RecordPrebidTimeoutRequests(pubid, profileid string) {
	m.prebidTimeoutRequests.With(prometheus.Labels{
		pubIDLabel:   pubid,
		profileLabel: profileid,
	}).Inc()
}

// RecordPartnerTimeoutRequests records count of Parnter timeout based on pubid, profileid and bidder
func (m *Metrics) RecordPartnerTimeoutRequests(pubid, profileid, bidder string) {
	m.partnerTimeoutRequest.With(prometheus.Labels{
		pubIDLabel:   pubid,
		profileLabel: profileid,
		bidderLabel:  bidder,
	}).Inc()
}

// RecordCtvUaAccuracy records accuracy of the ctv user agents
func (m *Metrics) RecordCtvUaAccuracy(pubId, status string) {
	m.ctvUaAccuracy.With(prometheus.Labels{
		pubIdLabel:  pubId,
		statusLabel: status,
	}).Inc()
}

// RecordSendLoggerDataTime as a noop
func (m *Metrics) RecordSendLoggerDataTime(endpoint, profileID string, sendTime time.Duration) {
	m.sendLoggerData.With(prometheus.Labels{
		endpointLabel:  endpoint,
		profileIDLabel: profileID,
	}).Observe(float64(sendTime.Seconds()))
}

// RecordSendLoggerDataTime as a noop
func (m *Metrics) RecordRequestTime(requestType string, requestTime time.Duration) {
	m.owRequestTime.With(prometheus.Labels{
		apiTypeLabel: requestType,
	}).Observe(float64(requestTime.Seconds()))
}

// RecordOWServerPanic counts the hb server panic
func (m *Metrics) RecordOWServerPanic(endpoint, methodName, nodeName, podName string) {
	m.panicCounts.With(prometheus.Labels{
		endpointLabel:   endpoint,
		methodNameLabel: methodName,
		nodeNameLabel:   nodeName,
		podNameLabel:    podName,
	}).Inc()
}

func preloadLabelValues(m *Metrics) {
	var (
		requestStatusValues = requestStatusesAsString()
		requestTypeValues   = requestTypesAsString()
	)

	preloadLabelValuesForCounter(m.owRequests, map[string][]string{
		requestTypeLabel:   requestTypeValues,
		requestStatusLabel: requestStatusValues,
	})
}

func preloadLabelValuesForCounter(counter *prometheus.CounterVec, labelsWithValues map[string][]string) {
	registerLabelPermutations(labelsWithValues, func(labels prometheus.Labels) {
		counter.With(labels)
	})
}

func registerLabelPermutations(labelsWithValues map[string][]string, register func(prometheus.Labels)) {
	if len(labelsWithValues) == 0 {
		return
	}

	keys := make([]string, 0, len(labelsWithValues))
	values := make([][]string, 0, len(labelsWithValues))
	for k, v := range labelsWithValues {
		keys = append(keys, k)
		values = append(values, v)
	}

	labels := prometheus.Labels{}
	registerLabelPermutationsRecursive(0, keys, values, labels, register)
}

func registerLabelPermutationsRecursive(depth int, keys []string, values [][]string, labels prometheus.Labels, register func(prometheus.Labels)) {
	label := keys[depth]
	isLeaf := depth == len(keys)-1

	if isLeaf {
		for _, v := range values[depth] {
			labels[label] = v
			register(cloneLabels(labels))
		}
	} else {
		for _, v := range values[depth] {
			labels[label] = v
			registerLabelPermutationsRecursive(depth+1, keys, values, labels, register)
		}
	}
}

func cloneLabels(labels prometheus.Labels) prometheus.Labels {
	clone := prometheus.Labels{}
	for k, v := range labels {
		clone[k] = v
	}
	return clone
}

func requestStatusesAsString() []string {
	values := RequestStatuses()
	valuesAsString := make([]string, len(values))
	for i, v := range values {
		valuesAsString[i] = string(v)
	}
	return valuesAsString
}

func requestTypesAsString() []string {
	values := RequestTypes()
	valuesAsString := make([]string, len(values))
	for i, v := range values {
		valuesAsString[i] = string(v)
	}
	return valuesAsString
}
