package stats

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
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

func (st *StatsTCP) RecordOpenWrapServerPanicStats(host, method string) {
	st.statsClient.PublishStat(statKeys[statsKeyOpenWrapServerPanic], 1)
}

func (st *StatsTCP) RecordPublisherPartnerNoCookieStats(publisher, partner string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPublisherPartnerNoCookieRequests], publisher, partner), 1)
}

func (st *StatsTCP) RecordPartnerResponseErrors(publisher, partner, err string) {
	switch err {
	case models.PartnerErrTimeout:
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyPartnerTimeoutErrorRequests], publisher, partner), 1)
	case models.PartnerErrNoBid:
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyNobidErrorRequests], publisher, partner), 1)
	case models.PartnerErrUnknownPrebidError:
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyUnknownPrebidErrorResponse], publisher, partner), 1)
	}
}

func (st *StatsTCP) RecordPartnerConfigErrors(publisher, profile, partner string, errcode int) {
	switch errcode {
	case models.PartnerErrMisConfig:
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyMisConfErrorRequests], publisher, partner), 1)
	case models.PartnerErrSlotNotMapped:
		st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeySlotunMappedErrorRequests], publisher, partner), 1)
	}
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

func (st *StatsTCP) RecordNobidErrPrebidServerRequests(publisher string, nbr int) {
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

func (st *StatsTCP) RecordPublisherWrapperLoggerFailure(publisher string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyLoggerErrorRequests], publisher), 1)
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

func (st *StatsTCP) RecordBadRequests(endpoint, publisher string, errorCode int) {
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

func (st *StatsTCP) RecordReqImpsWithContentCount(publisher, contentType string) {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyContentObjectPresent], contentType, publisher), 1)
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

	if platform == models.PLATFORM_APP {
		platform = models.HB_PLATFORM_APP
	}
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

func (st *StatsTCP) RecordGetProfileDataTime(getTime time.Duration) {}

func (st *StatsTCP) RecordDBQueryFailure(queryType, publisher, profile string) {}

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

func (st *StatsTCP) RecordRequest(labels metrics.Labels)                                      {}
func (st *StatsTCP) RecordLurlSent(labels metrics.LurlStatusLabels)                           {}
func (st *StatsTCP) RecordLurlBatchSent(labels metrics.LurlBatchStatusLabels)                 {}
func (st *StatsTCP) RecordBids(pubid, profileid, biddder, deal string)                        {}
func (st *StatsTCP) RecordPartnerTimeoutRequests(pubid, profileid, bidder string)             {}
func (st *StatsTCP) RecordCtvUaAccuracy(pubId, status string)                                 {}
func (st *StatsTCP) RecordSendLoggerDataTime(sendTime time.Duration)                          {}
func (st *StatsTCP) RecordRequestTime(requestType string, requestTime time.Duration)          {}
func (st *StatsTCP) RecordOWServerPanic(endpoint, methodName, nodeName, podName string)       {}
func (st *StatsTCP) RecordAmpVideoRequests(pubid, profileid string)                           {}
func (st *StatsTCP) RecordAmpVideoResponses(pubid, profileid string)                          {}
func (st *StatsTCP) RecordUnwrapRequestStatus(accountId, bidder, status string)               {}
func (st *StatsTCP) RecordUnwrapWrapperCount(accountId, bidder, wrapper_count string)         {}
func (st *StatsTCP) RecordUnwrapRequestTime(accountId, bidder string, respTime time.Duration) {}
func (st *StatsTCP) RecordUnwrapRespTime(accountId, wraperCnt string, respTime time.Duration) {}
func (st *StatsTCP) RecordAnalyticsTrackingThrottled(pubid, profileid, analyticsType string)  {}
func (st *StatsTCP) RecordAdruleEnabled(pubId, profId string)                                 {}
func (st *StatsTCP) RecordAdruleValidationFailure(pubId, profId string)                       {}
func (st *StatsTCP) RecordSignalDataStatus(pubid, profileid, signalType string)               {}
func (st *StatsTCP) RecordMBMFRequests(endpoint, pubId string, errorCode int)                 {}
func (st *StatsTCP) RecordPrebidCacheRequestTime(success bool, length time.Duration)          {}
func (st *StatsTCP) RecordBidRecoveryStatus(pubID string, profile string, success bool)       {}
func (st *StatsTCP) RecordBidRecoveryResponseTime(pubID string, profile string, responseTime time.Duration) {
}
func (st *StatsTCP) RecordPrebidAuctionBidResponse(publisher string, partnerName string, bidderCode string, adapterCode string) {
}
func (st *StatsTCP) RecordFailedParsingItuneID(pubId, profId string)              {}
func (st *StatsTCP) RecordEndpointResponseSize(endpoint string, bodySize float64) {}
func (st *StatsTCP) RecordIBVRequest(pubId, profId string)                        {}
func (st *StatsTCP) RecordGeoLookupFailure(endpoint string)                       {}
func (st *StatsTCP) RecordPartnerThrottledRequests(publisher, bidder string)      {}
