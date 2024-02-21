package config

import (
	"fmt"
	"time"

	cfg "github.com/prebid/prebid-server/v2/config"
	metrics_cfg "github.com/prebid/prebid-server/v2/metrics/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	ow_prometheus "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/prometheus"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/stats"
	"github.com/prometheus/client_golang/prometheus"
)

// NewMetricsEngine initialises the stats-client and prometheus and return them as MultiMetricsEngine
func NewMetricsEngine(cfg *config.Config, metricsCfg *cfg.Metrics, metricsRegistry metrics_cfg.MetricsRegistry) (MultiMetricsEngine, error) {

	// Create a list of metrics engines to use.
	engineList := make(MultiMetricsEngine, 0, 2)

	if cfg.Stats.Endpoint != "" {
		hostName := cfg.Stats.DefaultHostName // Dummy hostname N:P
		if cfg.Stats.UseHostName {
			hostName = cfg.Server.HostName // actual hostname node-name:pod-name
		}

		sc, err := stats.InitStatsClient(
			cfg.Stats.Endpoint,
			hostName,
			cfg.Server.HostName,
			cfg.Server.DCName,
			cfg.Stats.PublishInterval,
			cfg.Stats.PublishThreshold,
			cfg.Stats.Retries,
			cfg.Stats.DialTimeout,
			cfg.Stats.KeepAliveDuration,
			cfg.Stats.MaxIdleConnections,
			cfg.Stats.MaxIdleConnectionsPerHost,
			cfg.Stats.ResponseHeaderTimeout,
			cfg.Stats.MaxChannelLength,
			cfg.Stats.PoolMaxWorkers,
			cfg.Stats.PoolMaxCapacity)

		if err != nil {
			return nil, err
		}

		engineList = append(engineList, sc)
	}

	// Set up the Prometheus metrics engine.
	if metricsCfg != nil && metricsRegistry != nil && metricsRegistry[metrics_cfg.PrometheusRegistry] != nil {
		prometheusRegistry, ok := metricsRegistry[metrics_cfg.PrometheusRegistry].(*prometheus.Registry)
		if ok && prometheusRegistry != nil {
			prometheusEngine := ow_prometheus.NewMetrics(&metricsCfg.Prometheus, prometheusRegistry)
			engineList = append(engineList, prometheusEngine)
		}
	}

	if len(engineList) > 0 {
		return engineList, nil
	}
	return nil, fmt.Errorf("metric-engine is not configured")
}

// MultiMetricsEngine logs metrics to multiple metrics databases These can be useful in transitioning
// an instance from one engine to another, you can run both in parallel to verify stats match up.
type MultiMetricsEngine []metrics.MetricsEngine

// RecordOpenWrapServerPanicStats across all engines
func (me *MultiMetricsEngine) RecordOpenWrapServerPanicStats(host, method string) {
	for _, thisME := range *me {
		thisME.RecordOpenWrapServerPanicStats(host, method)
	}
}

// RecordPublisherPartnerNoCookieStats across all engines
func (me *MultiMetricsEngine) RecordPublisherPartnerNoCookieStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordPublisherPartnerNoCookieStats(publisher, partner)
	}
}

// RecordPartnerTimeoutErrorStats across all engines
func (me *MultiMetricsEngine) RecordPartnerResponseErrors(publisher, partner, err string) {
	for _, thisME := range *me {
		thisME.RecordPartnerResponseErrors(publisher, partner, err)
	}
}

// RecordMisConfigurationErrorStats across all engines
func (me *MultiMetricsEngine) RecordPartnerConfigErrors(publisher, profile, partner string, errcode int) {
	for _, thisME := range *me {
		thisME.RecordPartnerConfigErrors(publisher, profile, partner, errcode)
	}
}

// RecordPublisherProfileRequests across all engines
func (me *MultiMetricsEngine) RecordPublisherProfileRequests(publisher, profile string) {
	for _, thisME := range *me {
		thisME.RecordPublisherProfileRequests(publisher, profile)
	}
}

// RecordPublisherInvalidProfileImpressions across all engines
func (me *MultiMetricsEngine) RecordPublisherInvalidProfileImpressions(publisher, profileID string, impCount int) {
	for _, thisME := range *me {
		thisME.RecordPublisherInvalidProfileImpressions(publisher, profileID, impCount)
	}
}

// RecordNobidErrPrebidServerRequests across all engines
func (me *MultiMetricsEngine) RecordNobidErrPrebidServerRequests(publisher string, nbr int) {
	for _, thisME := range *me {
		thisME.RecordNobidErrPrebidServerRequests(publisher, nbr)
	}
}

// RecordNobidErrPrebidServerResponse across all engines
func (me *MultiMetricsEngine) RecordNobidErrPrebidServerResponse(publisher string) {
	for _, thisME := range *me {
		thisME.RecordNobidErrPrebidServerResponse(publisher)
	}
}

// RecordInvalidCreativeStats across all engines
func (me *MultiMetricsEngine) RecordInvalidCreativeStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordInvalidCreativeStats(publisher, partner)
	}
}

// RecordInvalidCreativeStats across all engines
func (me *MultiMetricsEngine) RecordPlatformPublisherPartnerReqStats(platform, publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordPlatformPublisherPartnerReqStats(platform, publisher, partner)
	}
}

// RecordInvalidCreativeStats across all engines
func (me *MultiMetricsEngine) RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner)
	}
}

// RecordPublisherResponseEncodingErrorStats across all engines
func (me *MultiMetricsEngine) RecordPublisherResponseEncodingErrorStats(publisher string) {
	for _, thisME := range *me {
		thisME.RecordPublisherResponseEncodingErrorStats(publisher)
	}
}

// RecordPartnerResponseTimeStats across all engines
func (me *MultiMetricsEngine) RecordPartnerResponseTimeStats(publisher, profileID string, responseTime int) {
	for _, thisME := range *me {
		thisME.RecordPartnerResponseTimeStats(publisher, profileID, responseTime)
	}
}

// RecordPublisherResponseTimeStats across all engines
func (me *MultiMetricsEngine) RecordPublisherResponseTimeStats(publisher string, responseTime int) {
	for _, thisME := range *me {
		thisME.RecordPublisherResponseTimeStats(publisher, responseTime)
	}
}

// RecordPublisherWrapperLoggerFailure across all engines
func (me *MultiMetricsEngine) RecordPublisherWrapperLoggerFailure(publisher, profileID, versionID string) {
	for _, thisME := range *me {
		thisME.RecordPublisherWrapperLoggerFailure(publisher, profileID, versionID)
	}
}

// RecordCacheErrorRequests across all engines
func (me *MultiMetricsEngine) RecordCacheErrorRequests(endpoint, publisher, profileID string) {
	for _, thisME := range *me {
		thisME.RecordCacheErrorRequests(endpoint, publisher, profileID)
	}
}

// RecordPublisherInvalidProfileRequests across all engines
func (me *MultiMetricsEngine) RecordPublisherInvalidProfileRequests(endpoint, publisher, profileID string) {
	for _, thisME := range *me {
		thisME.RecordPublisherInvalidProfileRequests(endpoint, publisher, profileID)
	}
}

// RecordBadRequests across all engines
func (me *MultiMetricsEngine) RecordBadRequests(endpoint string, errorCode int) {
	for _, thisME := range *me {
		thisME.RecordBadRequests(endpoint, errorCode)
	}
}

// RecordPrebidTimeoutRequests across all engines
func (me *MultiMetricsEngine) RecordPrebidTimeoutRequests(publisher, profileID string) {
	for _, thisME := range *me {
		thisME.RecordPrebidTimeoutRequests(publisher, profileID)
	}
}

// RecordSSTimeoutRequests across all engines
func (me *MultiMetricsEngine) RecordSSTimeoutRequests(publisher, profileID string) {
	for _, thisME := range *me {
		thisME.RecordSSTimeoutRequests(publisher, profileID)
	}
}

// RecordUidsCookieNotPresentErrorStats across all engines
func (me *MultiMetricsEngine) RecordUidsCookieNotPresentErrorStats(publisher, profileID string) {
	for _, thisME := range *me {
		thisME.RecordUidsCookieNotPresentErrorStats(publisher, profileID)
	}
}

// RecordVideoInstlImpsStats across all engines
func (me *MultiMetricsEngine) RecordVideoInstlImpsStats(publisher, profileID string) {
	for _, thisME := range *me {
		thisME.RecordVideoInstlImpsStats(publisher, profileID)
	}
}

// RecordImpDisabledViaConfigStats across all engines
func (me *MultiMetricsEngine) RecordImpDisabledViaConfigStats(impType, publisher, profileID string) {
	for _, thisME := range *me {
		thisME.RecordImpDisabledViaConfigStats(impType, publisher, profileID)
	}
}

// RecordPreProcessingTimeStats across all engines
func (me *MultiMetricsEngine) RecordPreProcessingTimeStats(publisher string, processingTime int) {
	for _, thisME := range *me {
		thisME.RecordPreProcessingTimeStats(publisher, processingTime)
	}
}

// RecordStatsKeyCTVPrebidFailedImpression across all engines
func (me *MultiMetricsEngine) RecordStatsKeyCTVPrebidFailedImpression(errorcode int, publisher string, profile string) {
	for _, thisME := range *me {
		thisME.RecordStatsKeyCTVPrebidFailedImpression(errorcode, publisher, profile)
	}
}

// RecordCTVRequests across all engines
func (me *MultiMetricsEngine) RecordCTVRequests(endpoint, platform string) {
	for _, thisME := range *me {
		thisME.RecordCTVRequests(endpoint, platform)
	}
}

// RecordPublisherRequests across all engines
func (me *MultiMetricsEngine) RecordPublisherRequests(endpoint, publisher, platform string) {
	for _, thisME := range *me {
		thisME.RecordPublisherRequests(endpoint, publisher, platform)
	}
}

// RecordCTVHTTPMethodRequests across all engines
func (me *MultiMetricsEngine) RecordCTVHTTPMethodRequests(endpoint, publisher, method string) {
	for _, thisME := range *me {
		thisME.RecordCTVHTTPMethodRequests(endpoint, publisher, method)
	}
}

// RecordCTVInvalidReasonCount across all engines
func (me *MultiMetricsEngine) RecordCTVInvalidReasonCount(errorCode int, publisher string) {
	for _, thisME := range *me {
		thisME.RecordCTVInvalidReasonCount(errorCode, publisher)
	}
}

// RecordCTVReqImpsWithDbConfigCount across all engines
func (me *MultiMetricsEngine) RecordCTVReqImpsWithDbConfigCount(publisher string) {
	for _, thisME := range *me {
		thisME.RecordCTVReqImpsWithDbConfigCount(publisher)
	}
}

// RecordCTVReqImpsWithReqConfigCount across all engines
func (me *MultiMetricsEngine) RecordCTVReqImpsWithReqConfigCount(publisher string) {
	for _, thisME := range *me {
		thisME.RecordCTVReqImpsWithReqConfigCount(publisher)
	}
}

// RecordAdPodGeneratedImpressionsCount across all engines
func (me *MultiMetricsEngine) RecordAdPodGeneratedImpressionsCount(impCount int, publisher string) {
	for _, thisME := range *me {
		thisME.RecordAdPodGeneratedImpressionsCount(impCount, publisher)
	}
}

// RecordRequestAdPodGeneratedImpressionsCount across all engines
func (me *MultiMetricsEngine) RecordRequestAdPodGeneratedImpressionsCount(impCount int, publisher string) {
	for _, thisME := range *me {
		thisME.RecordRequestAdPodGeneratedImpressionsCount(impCount, publisher)
	}
}

// RecordAdPodImpressionYield across all engines
func (me *MultiMetricsEngine) RecordAdPodImpressionYield(maxDuration int, minDuration int, publisher string) {
	for _, thisME := range *me {
		thisME.RecordAdPodImpressionYield(maxDuration, minDuration, publisher)
	}
}

// RecordCTVReqCountWithAdPod across all engines
func (me *MultiMetricsEngine) RecordCTVReqCountWithAdPod(publisher, profile string) {
	for _, thisME := range *me {
		thisME.RecordCTVReqCountWithAdPod(publisher, profile)
	}
}

// RecordReqImpsWithContentCount across all engines
func (me *MultiMetricsEngine) RecordReqImpsWithContentCount(publisher, contentType string) {
	for _, thisME := range *me {
		thisME.RecordReqImpsWithContentCount(publisher, contentType)
	}
}

// RecordPBSAuctionRequestsStats across all engines
func (me *MultiMetricsEngine) RecordPBSAuctionRequestsStats() {
	for _, thisME := range *me {
		thisME.RecordPBSAuctionRequestsStats()
	}
}

// RecordInjectTrackerErrorCount across all engines
func (me *MultiMetricsEngine) RecordInjectTrackerErrorCount(adformat, publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordInjectTrackerErrorCount(adformat, publisher, partner)
	}
}

// RecordBidResponseByDealCountInPBS across all engines
func (me *MultiMetricsEngine) RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId string) {
	for _, thisME := range *me {
		thisME.RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId)
	}
}

// RecordBidResponseByDealCountInHB across all engines
func (me *MultiMetricsEngine) RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId string) {
	for _, thisME := range *me {
		thisME.RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId)
	}
}

// RecordPartnerTimeoutInPBS across all engines
func (me *MultiMetricsEngine) RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder string) {
	for _, thisME := range *me {
		thisME.RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder)
	}
}

// RecordVideoImpDisabledViaConnTypeStats across all engines
func (me *MultiMetricsEngine) RecordVideoImpDisabledViaConnTypeStats(publisher, profile string) {
	for _, thisME := range *me {
		thisME.RecordVideoImpDisabledViaConnTypeStats(publisher, profile)
	}
}

// RecordGetProfileDataTime across all engines
func (me *MultiMetricsEngine) RecordGetProfileDataTime(requestType, profileid string, getTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordGetProfileDataTime(requestType, profileid, getTime)
	}
}

// RecordDBQueryFailure across all engines
func (me *MultiMetricsEngine) RecordDBQueryFailure(queryType, publisher, profile string) {
	for _, thisME := range *me {
		thisME.RecordDBQueryFailure(queryType, publisher, profile)
	}
}

// Shutdown across all engines
func (me *MultiMetricsEngine) Shutdown() {
	for _, thisME := range *me {
		thisME.Shutdown()
	}
}

// RecordRequest log openwrap request type
func (me *MultiMetricsEngine) RecordRequest(labels metrics.Labels) {
	for _, thisME := range *me {
		thisME.RecordRequest(labels)
	}
}

// RecordLurlSent log lurl status
func (me *MultiMetricsEngine) RecordLurlSent(labels metrics.LurlStatusLabels) {
	for _, thisME := range *me {
		thisME.RecordLurlSent(labels)
	}
}

// RecordLurlBatchSent log lurl batch status
func (me *MultiMetricsEngine) RecordLurlBatchSent(labels metrics.LurlBatchStatusLabels) {
	for _, thisME := range *me {
		thisME.RecordLurlBatchSent(labels)
	}
}

// RecordBids record ow bids
func (me *MultiMetricsEngine) RecordBids(pubid, profileid, biddder, deal string) {
	for _, thisME := range *me {
		thisME.RecordBids(pubid, profileid, biddder, deal)
	}
}

// RecordPartnerTimeoutRequests log request partner request timeout
func (me *MultiMetricsEngine) RecordPartnerTimeoutRequests(pubid, profileid, bidder string) {
	for _, thisME := range *me {
		thisME.RecordPartnerTimeoutRequests(pubid, profileid, bidder)
	}
}

// RecordCtvUaAccuracy log ctv UA accuracy
func (me *MultiMetricsEngine) RecordCtvUaAccuracy(pubId, status string) {
	for _, thisME := range *me {
		thisME.RecordCtvUaAccuracy(pubId, status)
	}
}

// RecordSendLoggerDataTime across all engines
func (me *MultiMetricsEngine) RecordSendLoggerDataTime(endpoint, profile string, sendTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordSendLoggerDataTime(endpoint, profile, sendTime)
	}
}

// RecordRequestTime record ow request time
func (me *MultiMetricsEngine) RecordRequestTime(requestType string, requestTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordRequestTime(requestType, requestTime)
	}
}

// RecordOWServerPanic record OW panics
func (me *MultiMetricsEngine) RecordOWServerPanic(endpoint, methodName, nodeName, podName string) {
	for _, thisME := range *me {
		thisME.RecordOWServerPanic(endpoint, methodName, nodeName, podName)
	}
}
