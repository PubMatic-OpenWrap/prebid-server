package metrics

// MetricsEngine is a generic interface to record PBS metrics into the desired backend
type MetricsEngine interface {
	RecordOpenWrapServerPanicStats(hostName, method string)
	RecordPublisherPartnerNoCookieStats(publisher, partner string)
	RecordPartnerResponseErrors(publisherID, partner, err string)
	RecordPartnerConfigErrors(publisherID, profileID, partner, err string)
	RecordPublisherProfileRequests(publisher, profileID string)
	RecordPublisherInvalidProfileImpressions(publisher, profileID string, impCount int)
	RecordNobidErrPrebidServerRequests(publisher string, nbr int)
	RecordNobidErrPrebidServerResponse(publisher string)
	RecordPlatformPublisherPartnerReqStats(platform, publisher, partner string)
	RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner string)
	RecordPartnerResponseTimeStats(publisher, partner string, responseTime int)
	RecordPublisherResponseTimeStats(publisher string, responseTimeMs int)
	RecordPublisherWrapperLoggerFailure(publisher, profileID, versionID string)
	RecordPublisherInvalidProfileRequests(endpoint, publisher, profileID string)
	RecordBadRequests(endpoint string, errorCode int)
	RecordUidsCookieNotPresentErrorStats(publisher, profileID string)
	RecordVideoInstlImpsStats(publisher, profileID string)
	RecordImpDisabledViaConfigStats(impType, publisher, profileID string)
	RecordPublisherRequests(endpoint string, publisher string, platform string)
	RecordReqImpsWithContentCount(publisher, contentType string)
	RecordInjectTrackerErrorCount(adformat, publisher, partner string)

	// not-captured in openwrap module, dont provide enough insights
	RecordPBSAuctionRequestsStats()
	RecordInvalidCreativeStats(publisher, partner string)

	// not implemented in openwrap module yet
	RecordCacheErrorRequests(endpoint string, publisher string, profileID string)
	RecordPublisherResponseEncodingErrorStats(publisher string)
	RecordVideoImpDisabledViaConnTypeStats(publisher, profileID string)

	// not applicable for openwrap module
	RecordPrebidTimeoutRequests(publisher, profileID string)
	RecordSSTimeoutRequests(publisher, profileID string)
	RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder string)
	RecordPreProcessingTimeStats(publisher string, processingTime int)

	// CTV specific metrics (not implemented in openwrap module yet)
	RecordStatsKeyCTVPrebidFailedImpression(errorcode int, publisher string, profile string)
	RecordCTVRequests(endpoint string, platform string)
	RecordCTVHTTPMethodRequests(endpoint string, publisher string, method string)
	RecordCTVInvalidReasonCount(errorCode int, publisher string)
	RecordCTVReqImpsWithDbConfigCount(publisher string)
	RecordCTVReqImpsWithReqConfigCount(publisher string)
	RecordAdPodGeneratedImpressionsCount(impCount int, publisher string)
	RecordRequestAdPodGeneratedImpressionsCount(impCount int, publisher string)
	RecordAdPodImpressionYield(maxDuration int, minDuration int, publisher string)
	RecordCTVReqCountWithAdPod(publisherID, profileID string)

	// stats-server specific metrics
	RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId string)
	RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId string)

	Shutdown()
}
