package stats

import (
	"fmt"

	"github.com/golang/glog"
)

type StatsTCP struct {
	statsClient *Client
}

func initTCPStatsClient(statIP, statPort string,
	pubInterval, pubThreshold, retries, dialTimeout, keepAliveDur, maxIdleConn, maxIdleConnPerHost int) (*StatsTCP, error) {

	cfg := Config{
		Host: statIP,
		Port: statPort,
		// Server: server,
		// DC:                  dc,
		PublishingInterval:  pubInterval,
		PublishingThreshold: pubThreshold,
		Retries:             retries,
		DialTimeout:         dialTimeout,
		KeepAliveDuration:   keepAliveDur,
		MaxIdleConns:        maxIdleConn,
		MaxIdleConnsPerHost: maxIdleConnPerHost,
	}

	sc, err := NewClient(&cfg)
	if err != nil {
		glog.Errorf("[stats_fail] Failed to initialize stats client : %v", err.Error())
		return nil, err
	}

	return &StatsTCP{statsClient: sc}, nil
}

func (st *StatsTCP) RecordOpenWrapServerPanicStats() {
	st.statsClient.PublishStat(statKeys[statsKeyOpenWrapServerPanic], 1)
}

func (st *StatsTCP) RecordPublisherPartnerStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPartnerRequests, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordPublisherPartnerImpStats(publisher, partner string, impCount int) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPartnerImpressions, publisher, partner), impCount)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerImpressions], publisher, partner), impCount)
}

func (st *StatsTCP) RecordPublisherPartnerNoCookieStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPartnerNoCookieRequests, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerNoCookieRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordPartnerTimeoutErrorStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPartnerTimeoutErrorRequests, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPartnerTimeoutErrorRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordNobiderStatusErrorStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyNobidderStatusErrorRequests, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidderStatusErrorRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordNobidErrorStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyNobidErrorRequests, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidErrorRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordUnkownPrebidErrorStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyUnknownPrebidErrorResponse, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyUnknownPrebidErrorResponse], publisher, partner), 1)
}

func (st *StatsTCP) RecordSlotNotMappedErrorStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeySlotunMappedErrorRequests, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeySlotunMappedErrorRequests], publisher, partner), 1)

}

func (st *StatsTCP) RecordMisConfigurationErrorStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyMisConfErrorRequests, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyMisConfErrorRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordPublisherProfileRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherProfileRequests, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherProfileRequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordPublisherInvalidProfileRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherInvProfileRequests, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileRequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordPublisherInvalidProfileImpressions(publisher, profileID string, impCount int) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherInvProfileImpressions, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileImpressions], publisher, profileID), impCount)
	//TODO @viral ;previously by 1 but now by impCount
}

func (st *StatsTCP) RecordPublisherNoConsentRequests(publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherNoConsentRequests, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherNoConsentRequests], publisher), 1)
}

func (st *StatsTCP) RecordPublisherNoConsentImpressions(publisher string, impCount int) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherNoConsentImpressions, publisher), impCount)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherNoConsentImpressions], publisher), impCount)
}

func (st *StatsTCP) RecordPublisherRequestStats(publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPrebidRequests, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPrebidRequests], publisher), 1)
}

func (st *StatsTCP) RecordNobidErrPrebidServerRequests(publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyNobidErrPrebidServerRequests, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidErrPrebidServerRequests], publisher), 1)
}

func (st *StatsTCP) RecordNobidErrPrebidServerResponse(publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyNobidErrPrebidServerResponse, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidErrPrebidServerResponse], publisher), 1)
}

func (st *StatsTCP) RecordInvalidCreativeStats(publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyInvalidCreatives, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyInvalidCreatives], publisher, partner), 1)
}

func (st *StatsTCP) RecordPlatformPublisherPartnerReqStats(platform, publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPlatformPublisherPartnerRequests, platform, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPlatformPublisherPartnerRequests], platform, publisher, partner), 1)
}

func (st *StatsTCP) RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPlatformPublisherPartnerResponses, platform, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPlatformPublisherPartnerResponses], platform, publisher, partner), 1)
}

func (st *StatsTCP) RecordPublisherResponseEncodingErrorStats(publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherResponseEncodingErrors, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherResponseEncodingErrors], publisher), 1)
}

func (st *StatsTCP) RecordPartnerResponseTimeStats(publisher, partner string, responseTime int) {
	statKeyIndex := getStatsKeyIndexForResponseTime(responseTime)
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statKeyIndex, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statKeyIndex], publisher, partner), 1)
}

func (st *StatsTCP) RecordPublisherResponseTimeStats(publisher string, responseTime int) {
	statKeyIndex := getStatsKeyIndexForResponseTime(responseTime)
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statKeyIndex, publisher, "overall"), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statKeyIndex], publisher, "overall"), 1)
}

func (st *StatsTCP) RecordPublisherWrapperLoggerFailure(publisher, profileID, versionID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyLoggerErrorRequests, publisher, profileID, versionID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyLoggerErrorRequests], publisher, profileID, versionID), 1)
}

func (st *StatsTCP) RecordAMPPublisherRequests(publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyAMPPublisherRequests, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyAMPPublisherRequests], publisher), 1)
}

func (st *StatsTCP) RecordAMPCacheErrorRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyAMPCacheError, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyAMPCacheError], publisher, profileID), 1)
}

func (st *StatsTCP) RecordPublisherInvalidProfileAMPRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherInvProfileAMPRequests, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileAMPRequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordVideoBadRequests() {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyVideoBadRequests), 1)
	st.statsClient.PublishStat(statKeys[statsKeyVideoBadRequests], 1)
}

func (st *StatsTCP) RecordVideoPublisherRequests(publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyVideoPublisherRequests, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoPublisherRequests], publisher), 1)
}

func (st *StatsTCP) RecordVideoCacheErrorRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyVideoCacheError, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoCacheError], publisher, profileID), 1)
}

func (st *StatsTCP) RecordPublisherInvalidProfileVideoRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherInvProfileVideoRequests, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileVideoRequests], publisher, profileID), 1)
}

func (st *StatsTCP) Record25BadRequests() {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKey25BadRequests), 1)
	st.statsClient.PublishStat(statKeys[statsKey25BadRequests], 1)
}

func (st *StatsTCP) RecordAMPBadRequests() {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyAMPBadRequests), 1)
	st.statsClient.PublishStat(statKeys[statsKeyAMPBadRequests], 1)
}

func (st *StatsTCP) Record25PublisherRequests(publisher, platform string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKey25PublisherRequests, GetPlatformForV25(request), publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKey25PublisherRequests], platform, publisher), 1)
}

func (st *StatsTCP) RecordPrebidTimeoutRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPrebidTORequests, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPrebidTORequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordSSTimeoutRequests(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeySsTORequests, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeySsTORequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordUidsCookieNotPresentErrorStats(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyNoUIDSErrorRequest, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNoUIDSErrorRequest], publisher, profileID), 1)
}

func (st *StatsTCP) RecordVideoInstlImpsStats(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyVideoInterstitialImpressions, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoInterstitialImpressions], publisher, profileID), 1)
}

func (st *StatsTCP) RecordVideoImpDisabledViaConfigStats(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyVideoImpDisabledViaConfig, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoImpDisabledViaConfig], publisher, profileID), 1)
}

func (st *StatsTCP) RecordBannerImpDisabledViaConfigStats(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyBannerImpDisabledViaConfig, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyBannerImpDisabledViaConfig], publisher, profileID), 1)
}

func (st *StatsTCP) RecordVideoImpDisabledViaConnTypeStats(publisher, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyVideoImpDisabledViaConnType, publisher, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoImpDisabledViaConnType], publisher, profileID), 1)
}

func (st *StatsTCP) RecordPreProcessingTimeStats(publisher string, processingTime int) {
	statKeyIndex := 0
	switch {
	case processingTime >= 100:
		statKeyIndex = statsKeyPrTimeAbv100
	case processingTime >= 50:
		statKeyIndex = statsKeyPrTimeAbv50
	case processingTime >= 10:
		statKeyIndex = statsKeyPrTimeAbv10
	case processingTime >= 1:
		statKeyIndex = statsKeyPrTimeAbv1
	default: // below 1ms
		statKeyIndex = statsKeyPrTimeBlw1
	}
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statKeyIndex, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statKeyIndex], publisher), 1)
}

func (st *StatsTCP) RecordStatsKeyCTVPrebidFailedImpression(errorcode int, publisher string, profile string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyCTVPrebidFailedImpression, errorcode, publisher, profile), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVPrebidFailedImpression], errorcode, publisher, profile), 1)
}

func (st *StatsTCP) RecordCTVRequests(endpoint string, platform string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyCTVRequests, endpoint, platform), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVRequests], endpoint, platform), 1)
}

func (st *StatsTCP) RecordCTVBadRequests(endpoint string, errorCode int) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyCTVBadRequests, endpoint, errorCode), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVBadRequests], endpoint, errorCode), 1)
}

func (st *StatsTCP) RecordCTVPublisherRequests(endpoint string, publisher string, platform string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyCTVPublisherRequests, endpoint, platform, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVPublisherRequests], endpoint, platform, publisher), 1)
}

func (st *StatsTCP) RecordCTVHTTPMethodRequests(endpoint string, publisher string, method string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyCTVHTTPMethodRequests, endpoint, publisher, method), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVHTTPMethodRequests], endpoint, publisher, method), 1)
}

func (st *StatsTCP) RecordCTVInvalidReasonCount(errorCode int, publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyCTVValidationErr, errorCode, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVValidationErr], errorCode, publisher), 1)
}

func (st *StatsTCP) RecordCTVIncompleteAdPodsCount(impCount int, reason string, publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyIncompleteAdPods, reason, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyIncompleteAdPods], reason, publisher), 1)
}

// func tcpIncrCTVReqImpsWithConfigCount(st *statsTCP, source string, publisher string) {
// 	st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyCTVReqImpstWithConfig, source, publisher), 1)
// }

func (st *StatsTCP) RecordCTVReqImpsWithDbConfigCount(publisher string) {
	// tcpIncrCTVReqImpsWithConfigCount(st, "db", publisher)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVReqImpstWithConfig], "db", publisher), 1)
}

func (st *StatsTCP) RecordCTVReqImpsWithReqConfigCount(publisher string) {
	// tcpIncrCTVReqImpsWithConfigCount(st, "req", publisher)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVReqImpstWithConfig], "req", publisher), 1)
}

func (st *StatsTCP) RecordAdPodGeneratedImpressionsCount(impCount int, publisher string) {
	var impRange string
	if impCount <= 3 {
		impRange = "1-3"
	} else if impCount <= 6 {
		impRange = "4-6"
	} else if impCount <= 9 {
		impRange = "7-9"
	} else {
		impRange = "9+"
	}
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyTotalAdPodImpression], impRange, publisher), 1)
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyTotalAdPodImpression, impRange, publisher), 1)
}

func (st *StatsTCP) RecordRequestAdPodGeneratedImpressionsCount(impCount int, publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyReqTotalAdPodImpression, publisher), impCount)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyReqTotalAdPodImpression], publisher), impCount)
}

func (st *StatsTCP) RecordAdPodSecondsMissedCount(seconds int, publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyAdPodSecondsMissed, publisher), seconds)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyAdPodSecondsMissed], publisher), seconds)
}

// func tcpIncrRequestContentObjectPresentCount(st *statsTCP, location string, publisher string) {
// 	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyContentObjectPresent, location, publisher), 1)
// }

func (st *StatsTCP) RecordReqImpsWithAppContentCount(publisher string) {
	// tcpIncrRequestContentObjectPresentCount(st, "app", publisher)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyContentObjectPresent], "app", publisher), 1)
}

func (st *StatsTCP) RecordReqImpsWithSiteContentCount(publisher string) {
	// tcpIncrRequestContentObjectPresentCount(st, "site", publisher)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyContentObjectPresent], "site", publisher), 1)
}

func (st *StatsTCP) RecordAdPodImpressionYield(maxDuration int, minDuration int, publisher string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyReqImpDurationYield, maxDuration, minDuration, publisher), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyReqImpDurationYield], maxDuration, minDuration, publisher), 1)
}

func (st *StatsTCP) RecordCTVReqCountWithAdPod(publisherID, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyReqWithAdPodCount, publisherID, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyReqWithAdPodCount], publisherID, profileID), 1)
}

func (st *StatsTCP) RecordCTVKeyBidDuration(duration int, publisherID, profileID string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyBidDuration, duration, publisherID, profileID), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyBidDuration], duration, publisherID, profileID), 1)
}

func (st *StatsTCP) RecordAdomainPresentStats(creativeType, publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPartnerAdomainPresent, creativeType, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerAdomainPresent], creativeType, publisher, partner), 1)
}

func (st *StatsTCP) RecordAdomainAbsentStats(creativeType, publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPartnerAdomainAbsent, creativeType, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerAdomainAbsent], creativeType, publisher, partner), 1)
}

func (st *StatsTCP) RecordCatPresentStats(creativeType, publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPartnerCatPresent, creativeType, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerCatPresent], creativeType, publisher, partner), 1)
}

func (st *StatsTCP) RecordCatAbsentStats(creativeType, publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPublisherPartnerCatAbsent, creativeType, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerCatAbsent], creativeType, publisher, partner), 1)
}

func (st *StatsTCP) RecordPBSAuctionRequestsStats() {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyPBSAuctionRequests), 1)
	st.statsClient.PublishStat(statKeys[statsKeyPBSAuctionRequests], 1)
}

func (st *StatsTCP) RecordInjectTrackerErrorCount(adformat, publisher, partner string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyInjectTrackerErrorCount, adformat, publisher, partner), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyInjectTrackerErrorCount], adformat, publisher, partner), 1)
}

func (st *StatsTCP) RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsBidResponsesByDealUsingPBS, publisher, profile, aliasBidder, dealId), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsBidResponsesByDealUsingPBS], publisher, profile, aliasBidder, dealId), 1)
}

func (st *StatsTCP) RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsBidResponsesByDealUsingHB, publisher, profile, aliasBidder, dealId), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsBidResponsesByDealUsingHB], publisher, profile, aliasBidder, dealId), 1)
}

func (st *StatsTCP) RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder string) {
	// st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsPartnerTimeoutInPBS, publisher, profile, aliasBidder), 1)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsPartnerTimeoutInPBS], publisher, profile, aliasBidder), 1)
}

// getStatsKeyIndexForResponseTime returns respective stats key for a given responsetime
func getStatsKeyIndexForResponseTime(responseTime int) int {
	statKey := 0
	switch {
	case responseTime >= 2000:
		statKey = statsKeyA2000
	case responseTime >= 1500:
		statKey = statsKeyA1500
	case responseTime >= 1000:
		statKey = statsKeyA1000
	case responseTime >= 900:
		statKey = statsKeyA900
	case responseTime >= 800:
		statKey = statsKeyA800
	case responseTime >= 700:
		statKey = statsKeyA700
	case responseTime >= 600:
		statKey = statsKeyA600
	case responseTime >= 500:
		statKey = statsKeyA500
	case responseTime >= 400:
		statKey = statsKeyA400
	case responseTime >= 300:
		statKey = statsKeyA300
	case responseTime >= 200:
		statKey = statsKeyA200
	case responseTime >= 100:
		statKey = statsKeyA100
	case responseTime >= 50:
		statKey = statsKeyA50
	default: // below 50 ms
		statKey = statsKeyL50
	}
	return statKey
}
