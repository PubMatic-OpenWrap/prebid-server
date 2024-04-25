package metrics

import "time"

// MetricsEngine is a generic interface to record PBS metrics into the desired backend
type MetricsEngine interface {
	RecordOpenWrapServerPanicStats(hostName, method string)
	RecordPublisherPartnerNoCookieStats(publisher, partner string)
	RecordPartnerResponseErrors(publisherID, partner, err string)
	RecordPartnerConfigErrors(publisherID, profileID, partner string, errcode int)
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
	RecordHTTPCounter()

	// not-captured in openwrap module, dont provide enough insights
	RecordPBSAuctionRequestsStats()
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
	RecordAmpVideoRequests(pubid, profileid string)
	RecordAmpVideoResponses(pubid, profileid string)

	// VAST Unwrap metrics
	RecordUnwrapRequestStatus(accountId, bidder, status string)
	RecordUnwrapWrapperCount(accountId, bidder string, wrapper_count string)
	RecordUnwrapRequestTime(accountId, bidder string, respTime time.Duration)
	RecordUnwrapRespTime(accountId, wraperCnt string, respTime time.Duration)
}
