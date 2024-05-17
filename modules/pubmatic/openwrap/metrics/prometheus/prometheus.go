package prometheus

import (
	"strconv"
	"sync"
	"time"

	"github.com/prebid/prebid-server/v2/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics defines the Prometheus metrics backing the MetricsEngine implementation.
type Metrics struct {

	// general metrics
	panics *prometheus.CounterVec

	// publisher-partner level metrics
	pubPartnerNoCookie            *prometheus.CounterVec
	pubPartnerRespErrors          *prometheus.CounterVec
	pubPartnerConfigErrors        *prometheus.CounterVec
	pubPartnerInjectTrackerErrors *prometheus.CounterVec
	pubPartnerResponseTimeSecs    *prometheus.HistogramVec

	// publisher-profile level metrics
	pubProfRequests             *prometheus.CounterVec
	pubProfInvalidImps          *prometheus.CounterVec
	pubProfUidsCookieAbsent     *prometheus.CounterVec // TODO - really need this ?
	pubProfVidInstlImps         *prometheus.CounterVec // TODO - really need this ?
	pubProfImpDisabledViaConfig *prometheus.CounterVec

	// publisher level metrics
	pubRequestValidationErrors *prometheus.CounterVec // TODO : should we add profiles as label ?
	pubNoBidResponseErrors     *prometheus.CounterVec
	pubResponseTime            *prometheus.HistogramVec
	pubImpsWithContent         *prometheus.CounterVec

	// publisher-partner-platform level metrics
	pubPartnerPlatformRequests  *prometheus.CounterVec
	pubPartnerPlatformResponses *prometheus.CounterVec

	// publisher-profile-endpoint level metrics
	pubProfEndpointInvalidRequests *prometheus.CounterVec

	// endpoint level metrics
	endpointBadRequest *prometheus.CounterVec //TODO: should we add pub+prof labels ; also NBR is INT should it be string

	// publisher-platform-endpoint level metrics
	pubPlatformEndpointRequests *prometheus.CounterVec

	getProfileData prometheus.Histogram
	sendLoggerData prometheus.Histogram

	dbQueryError *prometheus.CounterVec

	loggerFailure   *prometheus.CounterVec
	geoDBInitStatus *prometheus.GaugeVec

	//TODO -should we add "prefix" in metrics-name to differentiate it from prebid-core ?

	// sshb temporary
	owRequests            *prometheus.CounterVec
	lurlSent              *prometheus.CounterVec
	lurlBatchSent         *prometheus.CounterVec
	ctvUaAccuracy         *prometheus.CounterVec
	bids                  *prometheus.CounterVec
	prebidTimeoutRequests *prometheus.CounterVec
	partnerTimeoutRequest *prometheus.CounterVec
	panicCounts           *prometheus.CounterVec
	owRequestTime         *prometheus.HistogramVec
	ampVideoRequests      *prometheus.CounterVec
	ampVideoResponses     *prometheus.CounterVec
	analyticsThrottle     *prometheus.CounterVec

	// VAST Unwrap
	requests       *prometheus.CounterVec
	wrapperCount   *prometheus.CounterVec
	requestTime    *prometheus.HistogramVec
	unwrapRespTime *prometheus.HistogramVec
}

const (
	pubIDLabel         = "pub_id"
	profileIDLabel     = "profile_id"
	partnerLabel       = "partner"
	platformLabel      = "platform"
	endpointLabel      = "endpoint" // TODO- apiTypeLabel ?
	apiTypeLabel       = "api_type"
	impFormatLabel     = "imp_format" //TODO -confirm ?
	adFormatLabel      = "ad_format"
	sourceLabel        = "source" //TODO -confirm ?
	nbrLabel           = "nbr"    // TODO - errcode ?
	errorLabel         = "error"
	hostLabel          = "host" // combination of node:pod
	methodLabel        = "method"
	queryTypeLabel     = "query_type"
	analyticsTypeLabel = "an_type"
)

var standardTimeBuckets = []float64{0.1, 0.3, 0.75, 1}
var once sync.Once
var metric *Metrics

// NewMetrics initializes a new Prometheus metrics instance.
func NewMetrics(cfg *config.PrometheusMetrics, promRegistry *prometheus.Registry) *Metrics {
	once.Do(func() {
		metric = newMetrics(cfg, promRegistry)
	})
	return metric
}

func newMetrics(cfg *config.PrometheusMetrics, promRegistry *prometheus.Registry) *Metrics {
	metrics := Metrics{}

	// general metrics
	metrics.panics = newCounter(cfg, promRegistry,
		"panics",
		"Count of prebid server panics in openwrap module.",
		[]string{hostLabel, methodLabel},
	)

	// publisher-partner level metrics
	// TODO : check description of this
	metrics.pubPartnerNoCookie = newCounter(cfg, promRegistry,
		"no_cookie",
		"Count requests without cookie at publisher, partner level.",
		[]string{pubIDLabel, partnerLabel},
	)

	metrics.pubPartnerRespErrors = newCounter(cfg, promRegistry,
		"partner_response_error",
		"Count publisher requests where partner responded with error.",
		[]string{pubIDLabel, partnerLabel, errorLabel},
	)

	metrics.pubPartnerConfigErrors = newCounter(cfg, promRegistry,
		"partner_config_errors",
		"Count partner configuration errors at publisher, profile, partner level.",
		[]string{pubIDLabel, profileIDLabel, partnerLabel, errorLabel},
	)

	metrics.pubPartnerInjectTrackerErrors = newCounter(cfg, promRegistry,
		"inject_tracker_errors",
		"Count of errors while injecting trackers at publisher, partner level.",
		[]string{pubIDLabel, partnerLabel, adFormatLabel},
	)

	metrics.pubPartnerResponseTimeSecs = newHistogramVec(cfg, promRegistry,
		"partner_response_time",
		"Time taken by each partner to respond in seconds labeled by publisher.",
		[]string{pubIDLabel, partnerLabel},
		standardTimeBuckets,
	)

	// publisher-profile level metrics
	metrics.pubProfRequests = newCounter(cfg, promRegistry,
		"pub_profile_requests",
		"Count total number of requests at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfInvalidImps = newCounter(cfg, promRegistry,
		"invalid_imps",
		"Count impressions having invalid profile-id for respective publisher.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfUidsCookieAbsent = newCounter(cfg, promRegistry,
		"uids_cookie_absent",
		"Count requests for which uids cookie is absent at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfVidInstlImps = newCounter(cfg, promRegistry,
		"vid_instl_imps",
		"Count video interstitial impressions at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfImpDisabledViaConfig = newCounter(cfg, promRegistry,
		"imps_disabled_via_config",
		"Count banner/video impressions disabled via config at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel, impFormatLabel},
	)

	// publisher level metrics
	metrics.pubRequestValidationErrors = newCounter(cfg, promRegistry,
		"request_validation_errors",
		"Count request validation failures along with NBR at publisher level.",
		[]string{pubIDLabel, nbrLabel},
	)

	metrics.pubNoBidResponseErrors = newCounter(cfg, promRegistry,
		"no_bid",
		"Count of zero bid responses at publisher level.",
		[]string{pubIDLabel},
	)

	metrics.pubResponseTime = newHistogramVec(cfg, promRegistry,
		"pub_response_time",
		"Total time taken by request in seconds at publisher level.",
		[]string{pubIDLabel},
		standardTimeBuckets,
	)

	metrics.pubImpsWithContent = newCounter(cfg, promRegistry,
		"imps_with_content",
		"Count impressions having app/site content at publisher level.",
		[]string{pubIDLabel, sourceLabel},
		//TODO - contentLabel ??
	)

	// publisher-partner-platform metrics
	metrics.pubPartnerPlatformRequests = newCounter(cfg, promRegistry,
		"platform_requests",
		"Count requests at publisher, partner, platform level.",
		[]string{pubIDLabel, partnerLabel, platformLabel},
	)
	metrics.pubPartnerPlatformResponses = newCounter(cfg, promRegistry,
		"platform_responses",
		"Count responses at publisher, partner, platform level.",
		[]string{pubIDLabel, partnerLabel, platformLabel},
	)

	// publisher-profile-endpoint level metrics
	metrics.pubProfEndpointInvalidRequests = newCounter(cfg, promRegistry,
		"invalid_requests",
		"Count invalid requests at publisher, profile, endpoint level.",
		[]string{pubIDLabel, profileIDLabel, endpointLabel},
	)

	// endpoint level metrics
	metrics.endpointBadRequest = newCounter(cfg, promRegistry,
		"bad_requests",
		"Count bad requests along with NBR code at endpoint level.",
		[]string{endpointLabel, nbrLabel},
	)

	// publisher platform endpoint level metrics
	metrics.pubPlatformEndpointRequests = newCounter(cfg, promRegistry,
		"endpoint_requests",
		"Count requests at publisher, platform, endpoint level.",
		[]string{pubIDLabel, platformLabel, endpointLabel},
	)

	metrics.getProfileData = newHistogram(cfg, promRegistry,
		"profile_data_get_time",
		"Time taken to get the profile data in seconds", standardTimeBuckets)

	metrics.dbQueryError = newCounter(cfg, promRegistry,
		"db_query_failed",
		"Count failed db calls at profile, version level",
		[]string{queryTypeLabel, pubIDLabel, profileIDLabel},
	)

	metrics.loggerFailure = newCounter(cfg, promRegistry,
		"logger_send_failed",
		"Count of failures to send the logger to analytics endpoint at publisher and profile level",
		[]string{pubIDLabel, profileIDLabel},
	)
	metrics.analyticsThrottle = newCounter(cfg, promRegistry,
		"analytics_throttle",
		"Count of throttled analytics logger and tracker requestss",
		[]string{pubIDLabel, profileIDLabel, analyticsTypeLabel})

	metrics.requests = newCounter(cfg, promRegistry,
		"vastunwrap_status",
		"Count of vast unwrap requests labeled by status",
		[]string{pubIdLabel, bidderLabel, statusLabel})
	metrics.wrapperCount = newCounter(cfg, promRegistry,
		"vastunwrap_wrapper_count",
		"Count of vast unwrap levels labeled by bidder",
		[]string{pubIdLabel, bidderLabel, wrapperCountLabel})
	metrics.requestTime = newHistogramVec(cfg, promRegistry,
		"vastunwrap_request_time",
		"Time taken to serve the vast unwrap request in Milliseconds", []string{pubIdLabel, bidderLabel},
		[]float64{50, 100, 200, 300, 500})
	metrics.unwrapRespTime = newHistogramVec(cfg, promRegistry,
		"vastunwrap_resp_time",
		"Time taken to serve the vast unwrap request in Milliseconds at wrapper count level", []string{pubIdLabel, wrapperCountLabel},
		[]float64{50, 100, 150, 200})

	metrics.geoDBInitStatus = newGauge(cfg, promRegistry,
		"geodb_status",
		"An indicator to identify the GeoDB database's state of failure. 1 indicates failure and 0 indicates healthy.",
		[]string{dcNameLabel, nodeNameLabel, podNameLabel})

	newSSHBMetrics(&metrics, cfg, promRegistry)

	return &metrics
}

func newGauge(cfg *config.PrometheusMetrics, registry *prometheus.Registry, name, help string, labels []string) *prometheus.GaugeVec {
	opts := prometheus.GaugeOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      name,
		Help:      help,
	}
	gauge := prometheus.NewGaugeVec(opts, labels)
	registry.MustRegister(gauge)
	return gauge
}

func newCounter(cfg *config.PrometheusMetrics, registry *prometheus.Registry, name, help string, labels []string) *prometheus.CounterVec {
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

func newHistogram(cfg *config.PrometheusMetrics, registry *prometheus.Registry, name, help string, buckets []float64) prometheus.Histogram {
	opts := prometheus.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}
	histogram := prometheus.NewHistogram(opts)
	registry.MustRegister(histogram)
	return histogram
}

func newHistogramVec(cfg *config.PrometheusMetrics, registry *prometheus.Registry, name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
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

func (m *Metrics) RecordOpenWrapServerPanicStats(hostName, method string) {
	m.panics.With(prometheus.Labels{
		hostLabel:   hostName,
		methodLabel: method,
	}).Inc()
}

func (m *Metrics) RecordPublisherPartnerNoCookieStats(publisherID, partner string) {
	m.pubPartnerNoCookie.With(prometheus.Labels{
		pubIDLabel:   publisherID,
		partnerLabel: partner,
	}).Inc()
}

func (m *Metrics) RecordPartnerResponseErrors(publisherID, partner, err string) {
	m.pubPartnerRespErrors.With(prometheus.Labels{
		pubIDLabel:   publisherID,
		partnerLabel: partner,
		errorLabel:   err,
	}).Inc()
}

func (m *Metrics) RecordPartnerConfigErrors(publisherID, profileID, partner string, errcode int) {
	m.pubPartnerConfigErrors.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
		partnerLabel:   partner,
		errorLabel:     strconv.Itoa(errcode),
	}).Inc()
}

func (m *Metrics) RecordPublisherProfileRequests(publisherID, profileID string) {
	m.pubProfRequests.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
	}).Inc()
}

func (m *Metrics) RecordPublisherInvalidProfileImpressions(publisherID, profileID string, impCount int) {
	m.pubProfInvalidImps.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
	}).Add(float64(impCount))
}

func (m *Metrics) RecordNobidErrPrebidServerRequests(publisherID string, nbr int) {
	m.pubRequestValidationErrors.With(prometheus.Labels{
		pubIDLabel: publisherID,
		nbrLabel:   strconv.Itoa(nbr),
	}).Inc()
}

func (m *Metrics) RecordNobidErrPrebidServerResponse(publisherID string) {
	m.pubNoBidResponseErrors.With(prometheus.Labels{
		pubIDLabel: publisherID,
	}).Inc()
}

func (m *Metrics) RecordPlatformPublisherPartnerReqStats(platform, publisherID, partner string) {
	m.pubPartnerPlatformRequests.With(prometheus.Labels{
		platformLabel: platform,
		pubIDLabel:    publisherID,
		partnerLabel:  partner,
	}).Inc()
}

func (m *Metrics) RecordPlatformPublisherPartnerResponseStats(platform, publisherID, partner string) {
	m.pubPartnerPlatformResponses.With(prometheus.Labels{
		platformLabel: platform,
		pubIDLabel:    publisherID,
		partnerLabel:  partner,
	}).Inc()
}

func (m *Metrics) RecordPartnerResponseTimeStats(publisherID, partner string, responseTimeMs int) {
	m.pubPartnerResponseTimeSecs.With(prometheus.Labels{
		pubIDLabel:   publisherID,
		partnerLabel: partner,
	}).Observe(float64(responseTimeMs) / 1000)
}

func (m *Metrics) RecordPublisherResponseTimeStats(publisherID string, responseTimeMs int) {
	m.pubResponseTime.With(prometheus.Labels{
		pubIDLabel: publisherID,
	}).Observe(float64(responseTimeMs) / 1000)
}

func (m *Metrics) RecordPublisherInvalidProfileRequests(endpoint, publisherID, profileID string) {
	m.pubProfEndpointInvalidRequests.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
		endpointLabel:  endpoint,
	}).Inc()
}

func (m *Metrics) RecordBadRequests(endpoint string, errorCode int) {
	m.endpointBadRequest.With(prometheus.Labels{
		endpointLabel: endpoint,
		nbrLabel:      strconv.Itoa(errorCode),
	}).Inc()
}

func (m *Metrics) RecordUidsCookieNotPresentErrorStats(publisherID, profileID string) {
	m.pubProfUidsCookieAbsent.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
	}).Inc()
}

func (m *Metrics) RecordVideoInstlImpsStats(publisherID, profileID string) {
	m.pubProfVidInstlImps.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
	}).Inc()
}

func (m *Metrics) RecordImpDisabledViaConfigStats(impType, publisherID, profileID string) {
	m.pubProfImpDisabledViaConfig.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
		impFormatLabel: impType,
	}).Inc()
}

func (m *Metrics) RecordPublisherRequests(endpoint string, publisherID string, platform string) {
	m.pubPlatformEndpointRequests.With(prometheus.Labels{
		pubIDLabel:    publisherID,
		platformLabel: platform,
		endpointLabel: endpoint,
	}).Inc()
}

func (m *Metrics) RecordReqImpsWithContentCount(publisherID, content string) {
	m.pubImpsWithContent.With(prometheus.Labels{
		pubIDLabel:  publisherID,
		sourceLabel: content,
	}).Inc()
}

func (m *Metrics) RecordInjectTrackerErrorCount(adformat, publisherID, partner string) {
	m.pubPartnerInjectTrackerErrors.With(prometheus.Labels{
		adFormatLabel: adformat,
		pubIDLabel:    publisherID,
		partnerLabel:  partner,
	}).Inc()
}

// RecordGetProfileDataTime as a noop
func (m *Metrics) RecordGetProfileDataTime(getTime time.Duration) {
	m.getProfileData.Observe(float64(getTime.Seconds()))
}

// RecordDBQueryFailure as a noop
func (m *Metrics) RecordDBQueryFailure(queryType, publisher, profile string) {
	m.dbQueryError.With(prometheus.Labels{
		queryTypeLabel: queryType,
		pubIDLabel:     publisher,
		profileIDLabel: profile,
	}).Inc()
}

// RecordPublisherWrapperLoggerFailure to record count of owlogger failures
func (m *Metrics) RecordPublisherWrapperLoggerFailure(publisher, profile, version string) {
	m.loggerFailure.With(prometheus.Labels{
		pubIDLabel:     publisher,
		profileIDLabel: profile,
	}).Inc()
}

// RecordGeoDBInitStatus to record status of geodb initialisation status
func (m *Metrics) RecordGeoDBInitStatus(dcName, nodeName, podName string, value float64) {
	m.geoDBInitStatus.With(prometheus.Labels{
		dcNameLabel:   dcName,
		nodeNameLabel: nodeName,
		podNameLabel:  podName,
	}).Set(value)
}

// RecordAnalyticsTrackingThrottled record analytics throttling at publisher profile level
func (m *Metrics) RecordAnalyticsTrackingThrottled(pubid, profileid, analyticsType string) {
	m.analyticsThrottle.With(prometheus.Labels{
		pubIDLabel:         pubid,
		profileIDLabel:     profileid,
		analyticsTypeLabel: analyticsType,
	}).Inc()
}

// TODO - really need ?
func (m *Metrics) RecordPBSAuctionRequestsStats() {}

// TODO - empty because only stats are used currently
func (m *Metrics) RecordBidResponseByDealCountInPBS(publisherID, profile, aliasBidder, dealId string) {
}
func (m *Metrics) RecordBidResponseByDealCountInHB(publisherID, profile, aliasBidder, dealId string) {
}

// TODO - remove this functions once we are completely migrated from Header-bidding to module
func (m *Metrics) RecordSSTimeoutRequests(publisherID, profileID string)               {}
func (m *Metrics) RecordPartnerTimeoutInPBS(publisherID, profile, aliasBidder string)  {}
func (m *Metrics) RecordPreProcessingTimeStats(publisherID string, processingTime int) {}
func (m *Metrics) RecordInvalidCreativeStats(publisherID, partner string)              {}

// Code is not migrated yet
func (m *Metrics) RecordVideoImpDisabledViaConnTypeStats(publisherID, profileID string)           {}
func (m *Metrics) RecordCacheErrorRequests(endpoint string, publisherID string, profileID string) {}
func (m *Metrics) RecordPublisherResponseEncodingErrorStats(publisherID string)                   {}

// CTV_specific metrics
func (m *Metrics) RecordCTVRequests(endpoint string, platform string)                              {}
func (m *Metrics) RecordCTVHTTPMethodRequests(endpoint string, publisherID string, method string)  {}
func (m *Metrics) RecordCTVInvalidReasonCount(errorCode int, publisherID string)                   {}
func (m *Metrics) RecordCTVReqImpsWithDbConfigCount(publisherID string)                            {}
func (m *Metrics) RecordCTVReqImpsWithReqConfigCount(publisherID string)                           {}
func (m *Metrics) RecordAdPodGeneratedImpressionsCount(impCount int, publisherID string)           {}
func (m *Metrics) RecordRequestAdPodGeneratedImpressionsCount(impCount int, publisherID string)    {}
func (m *Metrics) RecordAdPodImpressionYield(maxDuration int, minDuration int, publisherID string) {}
func (m *Metrics) RecordCTVReqCountWithAdPod(publisherID, profileID string)                        {}
func (m *Metrics) RecordStatsKeyCTVPrebidFailedImpression(errorcode int, publisherID string, profile string) {
}

func (m *Metrics) Shutdown() {}
