package prometheus

import (
	"github.com/prebid/prebid-server/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics defines the Prometheus metrics backing the MetricsEngine implementation.
type Metrics struct {
	Registerer prometheus.Registerer
	Gatherer   *prometheus.Registry

	// general metrics
	panics *prometheus.CounterVec
	// pbsAuctionRequests *prometheus.CounterVec  //TODO - do we really need this ?

	// publisher-partner level metrics
	pubPartnerNoCookie            *prometheus.CounterVec
	pubPartnerRespErrors          *prometheus.CounterVec // pubPartnerNoBids + pubPartnerUnknownErrs + pubPartnerTimeouts
	pubPartnerSlotNotMappedErrors *prometheus.CounterVec //TODO  club ? pubPartnerSlotNotMappedError + pubPartnerMisConfigError
	pubPartnerMisConfigErrors     *prometheus.CounterVec
	pubPartnerInjectTrackerErrors *prometheus.CounterVec

	// publisher-profile level metrics
	pubProfRequests             *prometheus.CounterVec
	pubProfInvalidImps          *prometheus.CounterVec
	pubProfUidsCookieAbsent     *prometheus.CounterVec // TODO - really need this ?
	pubProfVidInstlImps         *prometheus.CounterVec // TODO - really need this ?
	pubProfImpDisabledViaConfig *prometheus.CounterVec

	// publisher level metrics
	pubRequestValidationErrors *prometheus.CounterVec // TODO : should we add profiles + error as label ?
	pubNoBidResponseError      *prometheus.CounterVec
	pubResponseTime            *prometheus.HistogramVec
	pubImpsWithContent         *prometheus.CounterVec

	// publisher-partner-platform level metrics
	pubPartnerPlatformRequests  *prometheus.CounterVec
	pubPartnerPlatformResponses *prometheus.CounterVec

	// publisher-profile-version level metrics
	pubProfVersionLoggerFailure *prometheus.CounterVec

	// publisher-profile-endpoint level metrics
	pubProfEndpointInvalidRequts *prometheus.CounterVec

	// endpoint level metrics
	endpointBadRequest *prometheus.CounterVec //TODO: should we add pub+prof labels ; also NBR is INT should it be string

	// publisher-platform-endpoint level metrics
	pubPlatformEndpointRequests *prometheus.CounterVec

	//TODO -should we add "prefix" in metrics-name to differentiate it from prebid-core ?
}

const (
	pubIDLabel     = "pub_id"
	profileIDLabel = "prof_id"
	versionIDLabel = "version_id"
	partnerLable   = "partner"
	platformLabel  = "platform"
	endpointLabel  = "endpoint"
	impTypeLabel   = "imp_type"
	adFormatLabel  = "adformat"
	nbrLabel       = "nbr"
	errorLabel     = "error"
)

// NewMetrics initializes a new Prometheus metrics instance.
func NewMetrics(cfg config.PrometheusMetrics) *Metrics {

	metrics := Metrics{}
	reg := prometheus.NewRegistry() // TODO - use prebid-core registry

	// general metrics
	metrics.panics = newCounter(cfg, reg,
		"panics",
		"Count of prebid server panics in openwrap module.",
		[]string{"node", "pod", "method"},
	)

	// metrics.pbsAuctionRequests = newCounter(cfg, reg,
	// 	"pbs_auction_requests",
	// 	"Count /pbs/auction requests.",
	// 	[]string{"node", "pod", "method"},
	// )

	// publisher-partner level metrics
	// TODO : check description of this
	metrics.pubPartnerNoCookie = newCounter(cfg, reg,
		"pub_partner_no_cookie",
		"Count requests without cookie at publisher, partner level.",
		[]string{pubIDLabel, partnerLable},
	)

	metrics.pubPartnerRespErrors = newCounter(cfg, reg,
		"pub_partner_response_error",
		"Count publisher requests where partner responded with error.",
		[]string{pubIDLabel, partnerLable, errorLabel},
	)

	metrics.pubPartnerSlotNotMappedErrors = newCounter(cfg, reg,
		"pub_partner_slot_not_map",
		"Count unmapped slot impressions for respective publisher, partner.",
		[]string{pubIDLabel, partnerLable},
	)

	metrics.pubPartnerMisConfigErrors = newCounter(cfg, reg,
		"pub_partner_missing_config",
		"Count missing configuration impressions at publisher, partner level.",
		[]string{pubIDLabel, partnerLable},
	)

	metrics.pubPartnerInjectTrackerErrors = newCounter(cfg, reg,
		"inject_tracker_errors",
		"Count of errors while injecting trackers at publisher, partner level.",
		[]string{pubIDLabel, partnerLable, adFormatLabel},
	)

	// publisher-profile level metrics
	metrics.pubProfRequests = newCounter(cfg, reg,
		"pub_profile_requests",
		"Count total number of requests at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfInvalidImps = newCounter(cfg, reg,
		"pub_profile_invalid_imps",
		"Count impressions having invalid profile-id for respective publisher.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfUidsCookieAbsent = newCounter(cfg, reg,
		"pub_profile_uids_cookie_absent",
		"Count requests for which uids cookie is absent at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfVidInstlImps = newCounter(cfg, reg,
		"pub_profile_vid_instl_imps",
		"Count video interstitial impressions at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel},
	)

	metrics.pubProfImpDisabledViaConfig = newCounter(cfg, reg,
		"imps_disabled_via_config",
		"Count banner/video impressions disabled via config at publisher, profile level.",
		[]string{pubIDLabel, profileIDLabel, impTypeLabel},
	)

	// publisher level metrics
	metrics.pubRequestValidationErrors = newCounter(cfg, reg,
		"pub_request_validation_error",
		"Count request validation failures at publisher level.",
		[]string{pubIDLabel},
	)

	metrics.pubNoBidResponseError = newCounter(cfg, reg,
		"pub_no_bid_response",
		"Count request for which bid response is empty at publisher level.",
		[]string{pubIDLabel},
	)

	metrics.pubResponseTime = newHistogramVec(cfg, reg,
		"pub_response_time",
		"Total time taken by request in milli-seconds at publisher level.",
		[]string{pubIDLabel},
		[]float64{50, 100, 200, 300, 500, 1000},
		//TODO- decide buckets
	)

	metrics.pubImpsWithContent = newCounter(cfg, reg,
		"imps_with_content",
		"Count impressions having app/site content at publisher level.",
		[]string{pubIDLabel},
	)

	// publisher-partner-platform metrics
	metrics.pubPartnerPlatformRequests = newCounter(cfg, reg,
		"pub_partner_platform_requests",
		"Count request at publisher, partner, platform level.",
		[]string{pubIDLabel, partnerLable, platformLabel},
	)
	metrics.pubPartnerPlatformResponses = newCounter(cfg, reg,
		"pub_partner_platform_responses",
		"Count response at publisher, partner, platform level.",
		[]string{pubIDLabel, partnerLable, platformLabel},
	)

	// publisher-profile-version level metrics
	metrics.pubProfVersionLoggerFailure = newCounter(cfg, reg,
		"owlogger_failures",
		"Count failures while sending owlogger at publisher, profile, version level.",
		[]string{pubIDLabel, profileIDLabel, versionIDLabel},
	)

	// publisher-profile-endpoint level metrics
	metrics.pubProfEndpointInvalidRequts = newCounter(cfg, reg,
		"pub_prof_invalid_requests",
		"Count invalid request at publisher, profile, endpoint level.",
		[]string{pubIDLabel, profileIDLabel, endpointLabel},
	)

	// endpoint level metrics
	metrics.endpointBadRequest = newCounter(cfg, reg,
		"bad_requests",
		"Count bad requests along with NBR code at endpoint level.",
		[]string{endpointLabel, nbrLabel},
	)

	// publisher platform endpoint level metrics
	metrics.pubPlatformEndpointRequests = newCounter(cfg, reg,
		"pub_platform_endpoint_requests",
		"Count requests at publisher, platform, endpoint level.",
		[]string{pubIDLabel, platformLabel, endpointLabel},
	)

	return &metrics
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
