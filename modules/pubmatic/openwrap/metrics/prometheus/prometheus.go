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
	pubBidRecoveryStatus       *prometheus.CounterVec
	pubBidRecoveryTime         *prometheus.HistogramVec

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

	loggerFailure *prometheus.CounterVec

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
	signalStatus          *prometheus.CounterVec
	pbsAuctionResponse    *prometheus.CounterVec

	// VAST Unwrap
	requests       *prometheus.CounterVec
	wrapperCount   *prometheus.CounterVec
	requestTime    *prometheus.HistogramVec
	unwrapRespTime *prometheus.HistogramVec

	//CTV
	ctvRequests                    *prometheus.CounterVec
	ctvHTTPMethodRequests          *prometheus.CounterVec
	ctvInvalidReasonCount          *prometheus.CounterVec
	ctvReqImpsWithDbConfigCount    *prometheus.CounterVec
	ctvReqImpsWithReqConfigCount   *prometheus.CounterVec
	adPodGeneratedImpressionsCount *prometheus.CounterVec
	ctvReqCountWithAdPod           *prometheus.CounterVec
	cacheWriteTime                 *prometheus.HistogramVec

	//VMAP adrule
	pubProfAdruleEnabled           *prometheus.CounterVec
	pubProfAdruleValidationfailure *prometheus.CounterVec

	//ApplovinMax
	failedParsingItuneId *prometheus.CounterVec
	endpointResponseSize *prometheus.HistogramVec

	//IBV request
	ibvRequests *prometheus.CounterVec

	//geo lookup
	geoLookUpFailure *prometheus.CounterVec
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
	signalTypeLabel    = "signal_status"
	successLabel       = "success"
	adpodImpCountLabel = "adpod_imp_count"
	bidderCodeLabel    = "bidder_code"
	adapterCodeLabel   = "adapter_code"
)

var standardTimeBuckets = []float64{0.05, 0.1, 0.3, 0.75, 1}
var responseSizeBuckets = []float64{0, 4, 7, 10, 15}
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
	cacheWriteTimeBuckets := []float64{10, 25, 50, 100}

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

	metrics.pbsAuctionResponse = newCounter(cfg, promRegistry,
		"pbs_auction_response",
		"Count of errors while injecting trackers at publisher, partner level.",
		[]string{pubIDLabel, partnerLabel, bidderCodeLabel, adapterCodeLabel},
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
		[]string{pubIDLabel, endpointLabel, nbrLabel},
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

	metrics.cacheWriteTime = newHistogramVec(cfg, promRegistry,
		"cache_write_time",
		"Seconds to write to Prebid Cache labeled by success or failure. Failure timing is limited by Prebid Server enforced timeouts.",
		[]string{successLabel},
		cacheWriteTimeBuckets)

	metrics.loggerFailure = newCounter(cfg, promRegistry,
		"logger_send_failed",
		"Count of failures to send the logger to analytics endpoint at publisher and profile level",
		[]string{pubIDLabel},
	)
	metrics.analyticsThrottle = newCounter(cfg, promRegistry,
		"analytics_throttle",
		"Count of throttled analytics logger and tracker requestss",
		[]string{pubIDLabel, profileIDLabel, analyticsTypeLabel})

	metrics.signalStatus = newCounter(cfg, promRegistry,
		"signal_status",
		"Count signal status for applovinmax requests",
		[]string{pubIDLabel, profileIDLabel, signalTypeLabel})

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

	metrics.ctvRequests = newCounter(cfg, promRegistry,
		"ctv_requests",
		"Count of ctv requests",
		[]string{endpointLabel, platformLabel},
	)

	metrics.ctvHTTPMethodRequests = newCounter(cfg, promRegistry,
		"ctv_http_method_requests",
		"Count of ctv requests specific to http methods",
		[]string{endpointLabel, pubIDLabel, methodLabel},
	)

	metrics.ctvInvalidReasonCount = newCounter(cfg, promRegistry,
		"ctv_invalid_reason",
		"Count of bad ctv requests with code",
		[]string{pubIdLabel, nbrLabel},
	)

	metrics.ctvReqImpsWithDbConfigCount = newCounter(cfg, promRegistry,
		"ctv_imps_db_config",
		"Count of ctv requests having adpod configs from database",
		[]string{pubIdLabel},
	)

	metrics.ctvReqImpsWithReqConfigCount = newCounter(cfg, promRegistry,
		"ctv_imps_req_config",
		"Count of ctv requests having adpod configs from request",
		[]string{pubIdLabel},
	)

	metrics.adPodGeneratedImpressionsCount = newCounter(cfg, promRegistry,
		"adpod_imps",
		"Count of impressions generated from adpod configs",
		[]string{pubIdLabel, adpodImpCountLabel},
	)

	metrics.ctvReqCountWithAdPod = newCounter(cfg, promRegistry,
		"ctv_requests_with_adpod",
		"Count of ctv request with adpod object",
		[]string{pubIdLabel, profileIDLabel},
	)

	metrics.failedParsingItuneId = newCounter(cfg, promRegistry,
		"failed_parsing_itune_id",
		"Count of failed parsing itune id",
		[]string{pubIdLabel, profileIDLabel},
	)

	metrics.endpointResponseSize = newHistogramVec(cfg, promRegistry,
		"endpoint_response_size",
		"Size of response",
		[]string{endpointLabel},
		responseSizeBuckets,
	)

	metrics.ibvRequests = newCounter(cfg, promRegistry,
		"ibv_requests",
		"Count of in-banner video requests",
		[]string{pubIDLabel, profileIDLabel})
	metrics.pubBidRecoveryTime = newHistogramVec(cfg, promRegistry,
		"bid_recovery_response_time",
		"Total time taken by request for secondary auction in ms at publisher profile level.",
		[]string{pubIDLabel, profileIDLabel},
		[]float64{100, 200, 300, 400},
	)

	metrics.pubBidRecoveryStatus = newCounter(cfg, promRegistry,
		"bid_recovery_response_status",
		"Count bid recovery status for secondary auction",
		[]string{pubIDLabel, profileIDLabel, successLabel},
	)

	metrics.geoLookUpFailure = newCounter(cfg, promRegistry,
		"geo_lookup_fail",
		"Count of geo lookup failures",
		[]string{endpointLabel})

	newSSHBMetrics(&metrics, cfg, promRegistry)

	return &metrics
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

func (m *Metrics) RecordPrebidAuctionBidResponse(publisher string, partnerName string, bidderCode string, adapterCode string) {
	m.pbsAuctionResponse.With(prometheus.Labels{
		pubIDLabel:       publisher,
		partnerLabel:     partnerName,
		bidderCodeLabel:  bidderCode,
		adapterCodeLabel: adapterCode,
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

func (m *Metrics) RecordBadRequests(endpoint, publisherID string, errorCode int) {
	m.endpointBadRequest.With(prometheus.Labels{
		endpointLabel: endpoint,
		pubIDLabel:    publisherID,
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
func (m *Metrics) RecordPublisherWrapperLoggerFailure(publisher string) {
	m.loggerFailure.With(prometheus.Labels{
		pubIDLabel: publisher,
	}).Inc()
}

// RecordAnalyticsTrackingThrottled record analytics throttling at publisher profile level
func (m *Metrics) RecordAnalyticsTrackingThrottled(pubid, profileid, analyticsType string) {
	m.analyticsThrottle.With(prometheus.Labels{
		pubIDLabel:         pubid,
		profileIDLabel:     profileid,
		analyticsTypeLabel: analyticsType,
	}).Inc()
}

// RecordSignalDataStatus record signaldata status(invalid,missing) at publisher level
func (m *Metrics) RecordSignalDataStatus(pubid, profileid, signalType string) {
	m.signalStatus.With(prometheus.Labels{
		pubIDLabel:      pubid,
		profileIDLabel:  profileid,
		signalTypeLabel: signalType,
	}).Inc()
}

// RecordFailedParsingItuneID to record failed parsing itune id
func (m *Metrics) RecordFailedParsingItuneID(pubid, profileid string) {
	m.failedParsingItuneId.With(prometheus.Labels{
		pubIDLabel:     pubid,
		profileIDLabel: profileid,
	}).Inc()
}

// RecordIBVRequest to record IBV request
func (m *Metrics) RecordIBVRequest(pubid, profileid string) {
	m.ibvRequests.With(prometheus.Labels{
		pubIDLabel:     pubid,
		profileIDLabel: profileid,
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
func (m *Metrics) RecordCTVRequests(endpoint string, platform string) {
	m.ctvRequests.With(prometheus.Labels{
		endpointLabel: endpoint,
		platformLabel: platform,
	}).Inc()
}

func (m *Metrics) RecordCTVHTTPMethodRequests(endpoint string, publisherID string, method string) {
	m.ctvHTTPMethodRequests.With(prometheus.Labels{
		endpointLabel: endpoint,
		pubIDLabel:    publisherID,
		methodLabel:   method,
	}).Inc()
}

func (m *Metrics) RecordCTVInvalidReasonCount(errorCode int, publisherID string) {
	m.ctvInvalidReasonCount.With(prometheus.Labels{
		pubIDLabel: publisherID,
		nbrLabel:   strconv.Itoa(errorCode),
	}).Inc()
}

func (m *Metrics) RecordCTVReqImpsWithDbConfigCount(publisherID string) {
	m.ctvReqImpsWithDbConfigCount.With(prometheus.Labels{
		pubIdLabel: publisherID,
	}).Inc()
}

func (m *Metrics) RecordCTVReqImpsWithReqConfigCount(publisherID string) {
	m.ctvReqImpsWithReqConfigCount.With(prometheus.Labels{
		pubIdLabel: publisherID,
	}).Inc()
}

func (m *Metrics) RecordAdPodGeneratedImpressionsCount(impCount int, publisherID string) {
	m.adPodGeneratedImpressionsCount.With(prometheus.Labels{
		pubIDLabel:         publisherID,
		adpodImpCountLabel: strconv.Itoa(impCount),
	}).Inc()
}

func (m *Metrics) RecordCTVReqCountWithAdPod(publisherID, profileID string) {
	m.ctvReqCountWithAdPod.With(prometheus.Labels{
		pubIdLabel:     publisherID,
		profileIDLabel: profileID,
	}).Inc()
}

func (m *Metrics) RecordRequestAdPodGeneratedImpressionsCount(impCount int, publisherID string)    {}
func (m *Metrics) RecordAdPodImpressionYield(maxDuration int, minDuration int, publisherID string) {}
func (m *Metrics) RecordStatsKeyCTVPrebidFailedImpression(errorcode int, publisherID string, profile string) {
}

func (m *Metrics) Shutdown() {}

func (m *Metrics) RecordPrebidCacheRequestTime(success bool, length time.Duration) {
	m.cacheWriteTime.With(prometheus.Labels{
		successLabel: strconv.FormatBool(success),
	}).Observe(float64(length.Milliseconds()))
}

func (m *Metrics) RecordBidRecoveryStatus(publisherID, profileID string, success bool) {
	m.pubBidRecoveryStatus.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
		successLabel:   strconv.FormatBool(success),
	}).Inc()
}

func (m *Metrics) RecordBidRecoveryResponseTime(publisherID, profileID string, responseTime time.Duration) {
	m.pubBidRecoveryTime.With(prometheus.Labels{
		pubIDLabel:     publisherID,
		profileIDLabel: profileID,
	}).Observe(float64(responseTime.Milliseconds()))
}

func (m *Metrics) RecordEndpointResponseSize(endpoint string, bodySize float64) {
	m.endpointResponseSize.With(prometheus.Labels{
		endpointLabel: endpoint,
	}).Observe(float64(bodySize) / 1024)
}

func (m *Metrics) RecordGeoLookupFailure(endpoint string) {
	m.geoLookUpFailure.With(prometheus.Labels{
		endpointLabel: endpoint,
	}).Inc()
}
