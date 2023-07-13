package metrics

// MetricsEngine is a generic interface to record PBS metrics into the desired backend
type MetricsEngine interface {
	RecordOpenWrapServerPanicStats(hostName, method string)        // DONE
	RecordPublisherPartnerNoCookieStats(publisher, partner string) // DONE

	RecordPartnerResponseErrors(publisherID, partner, err string)
	// RecordPartnerTimeoutErrorStats(publisher, partner string)      // DONE - pubPartnerRespErrors
	// RecordNobidErrorStats(publisher, partner string)               // DONE - pubPartnerRespErrors
	// RecordUnkownPrebidErrorStats(publisher, partner string)        // DONE - pubPartnerRespErrors

	// RecordSlotNotMappedErrorStats(publisher, partner string)    // DONE
	// RecordMisConfigurationErrorStats(publisher, partner string) // DONE
	RecordPartnerConfigErrors(publisherID, partner, err string)

	RecordPublisherProfileRequests(publisher, profileID string)                         //DONE
	RecordPublisherInvalidProfileImpressions(publisher, profileID string, impCount int) // DONE
	RecordNobidErrPrebidServerRequests(publisher string, nbr int)                       // DONE
	RecordNobidErrPrebidServerResponse(publisher string)                                // DONE

	RecordInvalidCreativeStats(publisher, partner string) // WONT_DO

	RecordPlatformPublisherPartnerReqStats(platform, publisher, partner string)      // DONE
	RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner string) // DONE
	RecordPartnerResponseTimeStats(publisher, partner string, responseTime int)      //DONE

	RecordPublisherResponseEncodingErrorStats(publisher string) // CODE_NOT_AVL

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

	RecordVideoImpDisabledViaConnTypeStats(publisher, profileID string) // CODE_NOT_AVL

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

	RecordReqImpsWithContentCount(publisher, contentType string) //DONE
	// RecordReqImpsWithAppContentCount(publisher string)  //DONE
	// RecordReqImpsWithSiteContentCount(publisher string) //DONE

	RecordPBSAuctionRequestsStats()                                    // REALLY_NEED ?
	RecordInjectTrackerErrorCount(adformat, publisher, partner string) //DONE

	RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId string)
	RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId string)

	Shutdown()
}
