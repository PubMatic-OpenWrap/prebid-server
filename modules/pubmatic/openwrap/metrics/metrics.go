package metrics

import "time"

// MetricsEngine is a generic interface to record PBS metrics into the desired backend
type MetricsEngine interface {
	RecordOpenWrapServerPanicStats(hostName, method string)
	RecordPublisherPartnerNoCookieStats(publisher, partner string)
	RecordPartnerResponseErrors(publisherID, partner, err string)
	RecordPartnerConfigErrors(publisherID, profileID, partner string, errcode int)
	RecordPublisherProfileRequests(publisher, profileID string)
	RecordNobidErrPrebidServerRequests(publisher string, nbr int)
	RecordNobidErrPrebidServerResponse(publisher string)
	RecordPlatformPublisherPartnerReqStats(platform, publisher, partner string)
	RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner string)
	RecordPartnerResponseTimeStats(publisher, partner string, responseTime int)
	RecordPublisherResponseTimeStats(publisher string, responseTimeMs int)
	RecordPublisherWrapperLoggerFailure(publisher string)
	RecordPublisherInvalidProfileRequests(endpoint, publisher, profileID string)
	RecordBadRequests(endpoint, publisher string, errorCode int)
	RecordUidsCookieNotPresentErrorStats(publisher, profileID string)
	RecordVideoInstlImpsStats(publisher, profileID string)
	RecordImpDisabledViaConfigStats(impType, publisher, profileID string)
	RecordPublisherRequests(endpoint string, publisher string, platform string)
	RecordReqImpsWithContentCount(publisher, contentType string)
	RecordInjectTrackerErrorCount(adformat, publisher, partner string)
	RecordBidRecoveryStatus(publisher, profile string, success bool)
	RecordBidRecoveryResponseTime(publisher, profile string, responseTime time.Duration)

	// RecordRequests(endpoint string, publisher string, profile string)

	// not-captured in openwrap module, dont provide enough insights
	RecordPBSAuctionRequestsStats()
	RecordPrebidAuctionBidResponse(publisher string, partnerName string, bidderCode string, adapterCode string)
	RecordInvalidCreativeStats(publisher, partner string)

	// not implemented in openwrap module yet
	RecordCacheErrorRequests(endpoint string, publisher string, profileID string)
	RecordPublisherResponseEncodingErrorStats(publisher string)
	RecordVideoImpDisabledViaConnTypeStats(publisher, profileID string)

	// not applicable for openwrap module
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

	RecordGetProfileDataTime(getTime time.Duration)
	RecordDBQueryFailure(queryType, publisher, profile string)

	Shutdown()

	// temporary sshb metrics
	RecordRequest(labels Labels) // ignores adapter. only statusOk and statusErr fom status
	RecordLurlSent(labels LurlStatusLabels)
	RecordLurlBatchSent(labels LurlBatchStatusLabels)
	RecordBids(pubid, profileid, biddder, deal string)
	RecordPrebidTimeoutRequests(pubid, profileid string)
	RecordPartnerTimeoutRequests(pubid, profileid, bidder string)
	RecordCtvUaAccuracy(pubId, status string)
	RecordSendLoggerDataTime(sendTime time.Duration)
	RecordRequestTime(requestType string, requestTime time.Duration)
	RecordOWServerPanic(endpoint, methodName, nodeName, podName string)

	// prebid metrics
	RecordPrebidCacheRequestTime(success bool, length time.Duration)

	// AMP metrics
	RecordAmpVideoRequests(pubid, profileid string)
	RecordAmpVideoResponses(pubid, profileid string)
	RecordAnalyticsTrackingThrottled(pubid, profileid, analyticsType string)
	RecordSignalDataStatus(pubid, profileid, signalType string)
	RecordMBMFRequests(endpoint, pubId string, errorCode int)

	// VAST Unwrap metrics
	RecordUnwrapRequestStatus(accountId, bidder, status string)
	RecordUnwrapWrapperCount(accountId, bidder string, wrapper_count string)
	RecordUnwrapRequestTime(accountId, bidder string, respTime time.Duration)
	RecordUnwrapRespTime(accountId, wraperCnt string, respTime time.Duration)

	//VMAP-adrule
	RecordAdruleEnabled(pubId, profId string)
	RecordAdruleValidationFailure(pubId, profId string)

	//AppLovinMax metrics
	RecordFailedParsingItuneID(pubId, profId string)
	RecordEndpointResponseSize(endpoint string, bodySize float64)
	RecordGeoLookupFailure(endpoint string)

	//IBV metric
	RecordIBVRequest(pubId, profId string)
	RecordPartnerThrottledRequests(publisher, bidder, featureID string)
	RecordCountryLevelPartnerThrottledRequests(endpoint, bidder, country string)

	//Request with schain removed
	RecordRequestWithSchainABTestEnabled()
}
