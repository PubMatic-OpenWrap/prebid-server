package config

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/stats"
)

// NewMetricsEngine initialises the stats-client and prometheus and return them as MultiMetricsEngine
func NewMetricsEngine(cfg config.Config) (MultiMetricsEngine, error) {

	// Create a list of metrics engines to use.
	engineList := make(MultiMetricsEngine, 0, 1)

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

	// TODO: Set up the Prometheus metrics engine.

	return engineList, nil
}

// MultiMetricsEngine logs metrics to multiple metrics databases The can be useful in transitioning
// an instance from one engine to another, you can run both in parallel to verify stats match up.
type MultiMetricsEngine []metrics.MetricsEngine

// RecordOpenWrapServerPanicStats across all engines
func (me *MultiMetricsEngine) RecordOpenWrapServerPanicStats() {
	for _, thisME := range *me {
		thisME.RecordOpenWrapServerPanicStats()
	}
}

// RecordPublisherPartnerStats across all engines
func (me *MultiMetricsEngine) RecordPublisherPartnerStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordPublisherPartnerStats(publisher, partner)
	}
}

// RecordPublisherPartnerImpStats across all engines
func (me *MultiMetricsEngine) RecordPublisherPartnerImpStats(publisher, partner string, impCount int) {
	for _, thisME := range *me {
		thisME.RecordPublisherPartnerImpStats(publisher, partner, impCount)
	}
}

// RecordPublisherPartnerNoCookieStats across all engines
func (me *MultiMetricsEngine) RecordPublisherPartnerNoCookieStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordPublisherPartnerNoCookieStats(publisher, partner)
	}
}

// RecordPartnerTimeoutErrorStats across all engines
func (me *MultiMetricsEngine) RecordPartnerTimeoutErrorStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordPartnerTimeoutErrorStats(publisher, partner)
	}
}

// RecordNobiderStatusErrorStats across all engines
func (me *MultiMetricsEngine) RecordNobiderStatusErrorStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordNobiderStatusErrorStats(publisher, partner)
	}
}

// RecordNobidErrorStats across all engines
func (me *MultiMetricsEngine) RecordNobidErrorStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordNobidErrorStats(publisher, partner)
	}
}

// RecordUnkownPrebidErrorStats across all engines
func (me *MultiMetricsEngine) RecordUnkownPrebidErrorStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordUnkownPrebidErrorStats(publisher, partner)
	}
}

// RecordSlotNotMappedErrorStats across all engines
func (me *MultiMetricsEngine) RecordSlotNotMappedErrorStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordSlotNotMappedErrorStats(publisher, partner)
	}
}

// RecordMisConfigurationErrorStats across all engines
func (me *MultiMetricsEngine) RecordMisConfigurationErrorStats(publisher, partner string) {
	for _, thisME := range *me {
		thisME.RecordMisConfigurationErrorStats(publisher, partner)
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

// RecordPublisherNoConsentRequests across all engines
func (me *MultiMetricsEngine) RecordPublisherNoConsentRequests(publisher string) {
}

// RecordPublisherNoConsentImpressions across all engines
func (me *MultiMetricsEngine) RecordPublisherNoConsentImpressions(publisher string, impCnt int) {
}

// RecordPublisherRequestStats across all engines
func (me *MultiMetricsEngine) RecordPublisherRequestStats(publisher string) {
	for _, thisME := range *me {
		thisME.RecordPublisherRequestStats(publisher)
	}
}

// RecordNobidErrPrebidServerRequests across all engines
func (me *MultiMetricsEngine) RecordNobidErrPrebidServerRequests(publisher string) {
	for _, thisME := range *me {
		thisME.RecordNobidErrPrebidServerRequests(publisher)
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

// RecordCTVIncompleteAdPodsCount across all engines
func (me *MultiMetricsEngine) RecordCTVIncompleteAdPodsCount(impCount int, reason string, publisher string) {
	for _, thisME := range *me {
		thisME.RecordCTVIncompleteAdPodsCount(impCount, reason, publisher)
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

// RecordAdPodSecondsMissedCount across all engines
func (me *MultiMetricsEngine) RecordAdPodSecondsMissedCount(seconds int, publisher string) {
	for _, thisME := range *me {
		thisME.RecordAdPodSecondsMissedCount(seconds, publisher)
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

// RecordCTVKeyBidDuration across all engines
func (me *MultiMetricsEngine) RecordCTVKeyBidDuration(duration int, publisher, profile string) {
	for _, thisME := range *me {
		thisME.RecordCTVKeyBidDuration(duration, publisher, profile)
	}
}

// RecordReqImpsWithAppContentCount across all engines
func (me *MultiMetricsEngine) RecordReqImpsWithAppContentCount(publisher string) {
	for _, thisME := range *me {
		thisME.RecordReqImpsWithAppContentCount(publisher)
	}
}

// RecordReqImpsWithSiteContentCount across all engines
func (me *MultiMetricsEngine) RecordReqImpsWithSiteContentCount(publisher string) {
	for _, thisME := range *me {
		thisME.RecordReqImpsWithSiteContentCount(publisher)
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

// Shutdown across all engines
func (me *MultiMetricsEngine) Shutdown() {
	for _, thisME := range *me {
		thisME.Shutdown()
	}
}
