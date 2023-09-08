package prometheus

import (
	"time"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"

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

// The request types (endpoints)
const (
	ReqTypeORTB25Web metrics.RequestType = "openrtb25-web"
	ReqTypeORTB25App metrics.RequestType = "openrtb25-app"
	ReqTypeAMP       metrics.RequestType = "amp"
	ReqTypeVideo     metrics.RequestType = "video"
)

// Request/return status
const (
	RequestStatusOK       metrics.RequestStatus = "ok"
	RequestStatusBadInput metrics.RequestStatus = "badinput"
	RequestStatusErr      metrics.RequestStatus = "err"
)

// RequestTypes returns all possible values for metrics.RequestType
func RequestTypes() []metrics.RequestType {
	return []metrics.RequestType{
		ReqTypeORTB25Web,
		ReqTypeORTB25App,
		ReqTypeAMP,
		ReqTypeVideo,
	}
}

// RequestStatuses return all possible values for metrics.RequestStatus
func RequestStatuses() []metrics.RequestStatus {
	return []metrics.RequestStatus{
		RequestStatusOK,
		RequestStatusBadInput,
		RequestStatusErr,
	}
}

// newSSHBMetrics initializes a new Prometheus metrics instance with preloaded label values for SSHB service
func newSSHBMetrics(metrics *Metrics, cfg *config.PrometheusMetrics, promRegistry *prometheus.Registry) {
	metrics.owRequests = newCounter(cfg, promRegistry,
		"sshb_requests",
		"Count of total requests to header-bidding server labeled by type and status.",
		[]string{requestTypeLabel, requestStatusLabel})

	metrics.sendLoggerData = newHistogramVec(cfg, promRegistry,
		"sshb_logger_data_send_time",
		"Time taken to send the wrapper logger body in seconds", []string{endpointLabel, profileIDLabel},
		standardTimeBuckets)

	metrics.owRequestTime = newHistogramVec(cfg, promRegistry,
		"sshb_request_time",
		"Time taken to serve the request in seconds", []string{apiTypeLabel},
		[]float64{0.05, 0.1, 0.15, 0.20, 0.25, 0.3, 0.4, 0.5, 0.75, 1})

	metrics.lurlSent = newCounter(cfg, promRegistry, "sshb_lurl_sent", "Count of lurl success, fail, drop and channel_full request sent labeled by publisherID, partner", []string{pubIdLabel, partnerLable, statusLabel})

	metrics.lurlBatchSent = newCounter(cfg, promRegistry, "sshb_lurl_batch_sent", "Count of lurl Batch success, fail and drop request sent to wtracker ", []string{statusLabel})

	metrics.bids = newCounter(cfg, promRegistry,
		"sshb_bids",
		"Count of bids by publisher id, profile, bidder and deal",
		[]string{pubIDLabel, profileLabel, bidderLabel, dealLabel})

	metrics.prebidTimeoutRequests = newCounter(cfg, promRegistry,
		"sshb_request_prebid_timeout",
		"count no of requests in which prebid timeouts",
		[]string{pubIDLabel, profileLabel})

	metrics.partnerTimeoutRequest = newCounter(cfg, promRegistry,
		"sshb_request_partner_timeout",
		"count no of requests in which partner timeouts",
		[]string{pubIDLabel, profileLabel, bidderLabel})

	metrics.ctvUaAccuracy = newCounter(cfg, promRegistry,
		"sshb_ctv_user_agent_accuracy",
		"Count of requests detected by Ctv user agent regex labeled by pub id and status.",
		[]string{pubIdLabel, statusLabel})

	metrics.panicCounts = newCounter(cfg, promRegistry,
		"sshb_panic",
		"Counts the header-bidding server panic.",
		[]string{nodeNameLabel, podNameLabel, methodNameLabel, endpointLabel})

	metrics.country = newCounter(cfg, promRegistry,
		"sshb_country",
		"Count of requests received with publishers Country by publisher id.",
		[]string{pubIDLabel})

	preloadLabelValues(metrics)
}

// RecordRequest across all engines
func (m *Metrics) RecordRequest(labels metrics.Labels) {
	m.owRequests.With(prometheus.Labels{
		requestTypeLabel:   string(labels.RType),
		requestStatusLabel: string(labels.RequestStatus),
	}).Inc()
}

// RecordLurlSent records lurl status success, fail, drop and channel_fool
func (m *Metrics) RecordLurlSent(labels metrics.LurlStatusLabels) {
	m.lurlSent.With(prometheus.Labels{
		pubIdLabel:   labels.PublisherID,
		partnerLable: labels.Partner,
		statusLabel:  labels.Status,
	}).Inc()
}

// RecordLurlBatchSent records lurl batchs sent to wtracker
func (m *Metrics) RecordLurlBatchSent(labels metrics.LurlBatchStatusLabels) {
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

// RecordCountry records count of requests received with req.device.geo.country
func (m *Metrics) RecordCountry(pubId string) {
	m.country.With(prometheus.Labels{
		pubIDLabel: pubId,
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
