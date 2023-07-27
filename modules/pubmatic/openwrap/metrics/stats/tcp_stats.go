package stats

import (
	"fmt"

	"github.com/golang/glog"
)

type StatsTCP struct {
	statsClient *Client
}

func initTCPStatsClient(endpoint string,
	pubInterval, pubThreshold, retries, dialTimeout, keepAliveDur, maxIdleConn,
	maxIdleConnPerHost, respHeaderTimeout, maxChannelLength, poolMaxWorkers, poolMaxCapacity int) (*StatsTCP, error) {

	cfg := config{
		Endpoint:              endpoint,
		PublishingInterval:    pubInterval,
		PublishingThreshold:   pubThreshold,
		Retries:               retries,
		DialTimeout:           dialTimeout,
		KeepAliveDuration:     keepAliveDur,
		MaxIdleConns:          maxIdleConn,
		MaxIdleConnsPerHost:   maxIdleConnPerHost,
		ResponseHeaderTimeout: respHeaderTimeout,
		MaxChannelLength:      maxChannelLength,
		PoolMaxWorkers:        poolMaxWorkers,
		PoolMaxCapacity:       poolMaxCapacity,
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

func (st *StatsTCP) RecordPublisherPartnerNoCookieStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerNoCookieRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordPartnerTimeoutErrorStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPartnerTimeoutErrorRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordNobidErrorStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidErrorRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordUnkownPrebidErrorStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyUnknownPrebidErrorResponse], publisher, partner), 1)
}

func (st *StatsTCP) RecordSlotNotMappedErrorStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeySlotunMappedErrorRequests], publisher, partner), 1)

}

func (st *StatsTCP) RecordMisConfigurationErrorStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyMisConfErrorRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordPublisherProfileRequests(publisher, profileID string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherProfileRequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordPublisherInvalidProfileRequests(endpoint, publisher, profileID string) {
	switch endpoint {
	case "video":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileVideoRequests], publisher, profileID), 1)
	case "amp":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileAMPRequests], publisher, profileID), 1)
	default:
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileRequests], publisher, profileID), 1)
	}
}

func (st *StatsTCP) RecordPublisherInvalidProfileImpressions(publisher, profileID string, impCount int) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherInvProfileImpressions], publisher, profileID), impCount)
	//TODO @viral ;previously by 1 but now by impCount
}

func (st *StatsTCP) RecordNobidErrPrebidServerRequests(publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidErrPrebidServerRequests], publisher), 1)
}

func (st *StatsTCP) RecordNobidErrPrebidServerResponse(publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidErrPrebidServerResponse], publisher), 1)
}

func (st *StatsTCP) RecordInvalidCreativeStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyInvalidCreatives], publisher, partner), 1)
}

func (st *StatsTCP) RecordPlatformPublisherPartnerReqStats(platform, publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPlatformPublisherPartnerRequests], platform, publisher, partner), 1)
}

func (st *StatsTCP) RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPlatformPublisherPartnerResponses], platform, publisher, partner), 1)
}

func (st *StatsTCP) RecordPublisherResponseEncodingErrorStats(publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherResponseEncodingErrors], publisher), 1)
}

func (st *StatsTCP) RecordPartnerResponseTimeStats(publisher, partner string, responseTime int) {
	statKeyIndex := getStatsKeyIndexForResponseTime(responseTime)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statKeyIndex], publisher, partner), 1)
}

func (st *StatsTCP) RecordPublisherResponseTimeStats(publisher string, responseTime int) {
	statKeyIndex := getStatsKeyIndexForResponseTime(responseTime)
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statKeyIndex], publisher, "overall"), 1)
}

func (st *StatsTCP) RecordPublisherWrapperLoggerFailure(publisher, profileID, versionID string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyLoggerErrorRequests], publisher, profileID, versionID), 1)
}

func (st *StatsTCP) RecordPrebidTimeoutRequests(publisher, profileID string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPrebidTORequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordSSTimeoutRequests(publisher, profileID string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeySsTORequests], publisher, profileID), 1)
}

func (st *StatsTCP) RecordUidsCookieNotPresentErrorStats(publisher, profileID string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNoUIDSErrorRequest], publisher, profileID), 1)
}

func (st *StatsTCP) RecordVideoInstlImpsStats(publisher, profileID string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoInterstitialImpressions], publisher, profileID), 1)
}

func (st *StatsTCP) RecordImpDisabledViaConfigStats(impType, publisher, profileID string) {
	switch impType {
	case "video":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoImpDisabledViaConfig], publisher, profileID), 1)
	case "banner":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyBannerImpDisabledViaConfig], publisher, profileID), 1)
	}
}

func (st *StatsTCP) RecordVideoImpDisabledViaConnTypeStats(publisher, profileID string) {
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
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statKeyIndex], publisher), 1)
}

func (st *StatsTCP) RecordStatsKeyCTVPrebidFailedImpression(errorcode int, publisher string, profile string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVPrebidFailedImpression], errorcode, publisher, profile), 1)
}

func (st *StatsTCP) RecordCTVRequests(endpoint string, platform string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVRequests], endpoint, platform), 1)
}

func (st *StatsTCP) RecordBadRequests(endpoint string, errorCode int) {
	switch endpoint {
	case "amp":
		st.statsClient.PublishStat(statKeys[statsKeyAMPBadRequests], 1)
	case "video":
		st.statsClient.PublishStat(statKeys[statsKeyVideoBadRequests], 1)
	case "v25":
		st.statsClient.PublishStat(statKeys[statsKey25BadRequests], 1)
	case "vast", "ortb", "json", "openwrap":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVBadRequests], endpoint, errorCode), 1)
	}
}

func (st *StatsTCP) RecordCTVHTTPMethodRequests(endpoint string, publisher string, method string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVHTTPMethodRequests], endpoint, publisher, method), 1)
}

func (st *StatsTCP) RecordCTVInvalidReasonCount(errorCode int, publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVValidationErr], errorCode, publisher), 1)
}

func (st *StatsTCP) RecordCTVReqImpsWithDbConfigCount(publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVReqImpstWithConfig], "db", publisher), 1)
}

func (st *StatsTCP) RecordCTVReqImpsWithReqConfigCount(publisher string) {
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
}

func (st *StatsTCP) RecordRequestAdPodGeneratedImpressionsCount(impCount int, publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyReqTotalAdPodImpression], publisher), impCount)
}

func (st *StatsTCP) RecordReqImpsWithAppContentCount(publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyContentObjectPresent], "app", publisher), 1)
}

func (st *StatsTCP) RecordReqImpsWithSiteContentCount(publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyContentObjectPresent], "site", publisher), 1)
}

func (st *StatsTCP) RecordAdPodImpressionYield(maxDuration int, minDuration int, publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyReqImpDurationYield], maxDuration, minDuration, publisher), 1)
}

func (st *StatsTCP) RecordCTVReqCountWithAdPod(publisherID, profileID string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyReqWithAdPodCount], publisherID, profileID), 1)
}

func (st *StatsTCP) RecordPBSAuctionRequestsStats() {
	st.statsClient.PublishStat(statKeys[statsKeyPBSAuctionRequests], 1)
}

func (st *StatsTCP) RecordInjectTrackerErrorCount(adformat, publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyInjectTrackerErrorCount], adformat, publisher, partner), 1)
}

func (st *StatsTCP) RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsBidResponsesByDealUsingPBS], publisher, profile, aliasBidder, dealId), 1)
}

func (st *StatsTCP) RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsBidResponsesByDealUsingHB], publisher, profile, aliasBidder, dealId), 1)
}

func (st *StatsTCP) RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsPartnerTimeoutInPBS], publisher, profile, aliasBidder), 1)
}

func (st *StatsTCP) RecordPublisherRequests(endpoint, publisher, platform string) {

	switch endpoint {
	case "amp":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyAMPPublisherRequests], publisher), 1)
	case "video":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoPublisherRequests], publisher), 1)
	case "v25":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKey25PublisherRequests], platform, publisher), 1)
	case "vast", "ortb", "json":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyCTVPublisherRequests], endpoint, platform, publisher), 1)
	}
}

func (st *StatsTCP) RecordCacheErrorRequests(endpoint, publisher, profileID string) {
	switch endpoint {
	case "amp":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyAMPCacheError], publisher, profileID), 1)
	case "video":
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyVideoCacheError], publisher, profileID), 1)
	}
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

func (st *StatsTCP) Shutdown() {
	st.statsClient.ShutdownProcess()
}
