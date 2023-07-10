package metrics

// MetricsEngine is a generic interface to record PBS metrics into the desired backend
type MetricsEngine interface {
	RecordOpenWrapServerPanicStats()                               // DONE
	RecordPublisherPartnerNoCookieStats(publisher, partner string) // DONE
	RecordPartnerTimeoutErrorStats(publisher, partner string)      // DONE
	RecordNobidErrorStats(publisher, partner string)               // DONE
	RecordUnkownPrebidErrorStats(publisher, partner string)        // DONE

	RecordSlotNotMappedErrorStats(publisher, partner string)    // DONE
	RecordMisConfigurationErrorStats(publisher, partner string) // DONE

	RecordPublisherProfileRequests(publisher, profileID string)                         //DONE
	RecordPublisherInvalidProfileImpressions(publisher, profileID string, impCount int) // DONE
	RecordNobidErrPrebidServerRequests(publisher string)                                // DONE
	RecordNobidErrPrebidServerResponse(publisher string)                                // DONE

	RecordInvalidCreativeStats(publisher, partner string) // WONT_DO

	RecordPlatformPublisherPartnerReqStats(platform, publisher, partner string)      // DONE
	RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner string) // DONE

	RecordPublisherResponseEncodingErrorStats(publisher string)                 // CODE_NOT_AVL
	RecordPartnerResponseTimeStats(publisher, partner string, responseTime int) // SAME => RecordAdapterTime

	RecordPublisherResponseTimeStats(publisher string, responseTimeMs int)      // DONE
	RecordPublisherWrapperLoggerFailure(publisher, profileID, versionID string) // DONE

	RecordCacheErrorRequests(endpoint string, publisher string, profileID string) // CODE_NOT_AVL

	RecordPublisherInvalidProfileRequests(endpoint, publisher, profileID string) // DONE
	RecordBadRequests(endpoint string, errorCode int)                            // DONE

	RecordPrebidTimeoutRequests(publisher, profileID string)           // NOT_REQUIRED
	RecordSSTimeoutRequests(publisher, profileID string)               // NOT_REQUIRED
	RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder string)  // NOT_REQUIRED
	RecordPreProcessingTimeStats(publisher string, processingTime int) // NOT_REQUIRED

	RecordUidsCookieNotPresentErrorStats(publisher, profileID string)     // DONE
	RecordVideoInstlImpsStats(publisher, profileID string)                // DONE
	RecordImpDisabledViaConfigStats(impType, publisher, profileID string) // DONE
	RecordVideoImpDisabledViaConnTypeStats(publisher, profileID string)   // CODE_NOT_AVL

	RecordPublisherRequests(endpoint string, publisher string, platform string) //DONE

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

	RecordReqImpsWithAppContentCount(publisher string)  //DONE
	RecordReqImpsWithSiteContentCount(publisher string) //DONE

	RecordPBSAuctionRequestsStats()                                    // REALLY_NEED ?
	RecordInjectTrackerErrorCount(adformat, publisher, partner string) //DONE

	RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId string)
	RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId string)

	Shutdown()
}
