package stats

import (
	"sync"
)

type statKeyName = string

var (
	statKeys   [maxNumOfStats]statKeyName
	once       sync.Once
	owStats    *StatsTCP
	owStatsErr error
)

// stat represents a single stat-key along with its value
type stat struct {
	Key   string
	Value int
}

// InitStatsClient initializes stats client
func InitStatsClient(endpoint, defaultHost, actualHost, dcName string,
	pubInterval, pubThreshold, retries, dialTimeout, keepAliveDuration,
	maxIdleConnes, maxIdleConnesPerHost, respHeaderTimeout, maxChannelLength,
	poolMaxWorkers, poolMaxCapacity int) (*StatsTCP, error) {

	once.Do(func() {
		initStatKeys(dcName+":"+defaultHost, dcName+":"+actualHost)
		owStats, owStatsErr = initTCPStatsClient(endpoint, pubInterval, pubThreshold,
			retries, dialTimeout, keepAliveDuration, maxIdleConnes, maxIdleConnesPerHost,
			respHeaderTimeout, maxChannelLength, poolMaxWorkers, poolMaxCapacity)
	})

	return owStats, owStatsErr
}

// initStatKeys sets the key-name for all stats
// defaultServerName will be "actualDCName:N:P"
// actualServerName will be "actualDCName:actualNode:actualPod"
func initStatKeys(defaultServerName, actualServerName string) {

	//server level stats
	statKeys[statsKeyOpenWrapServerPanic] = "hb:panic:" + actualServerName
	//hb:panic:<dc:node:pod>

	//publisher level stats
	statKeys[statsKeyPublisherNoConsentRequests] = "hb:pubnocnsreq:%s:" + defaultServerName
	//hb:pubnocnsreq:<pub>:<dc:node:pod>

	statKeys[statsKeyPublisherNoConsentImpressions] = "hb:pubnocnsimp:%s:" + defaultServerName
	//hb:pubnocnsimp:<pub>:<dc:node:pod>

	statKeys[statsKeyPublisherPrebidRequests] = "hb:pubrq:%s:" + defaultServerName

	// statKeys[statsKeyNobidErrPrebidServerRequests] = "hb:pubnbreq:%s:", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyNobidErrPrebidServerRequests] = "hb:pubnbreq:%s:" + defaultServerName
	//hb:pubnbreq:<pub>:<dc:node:pod>

	statKeys[statsKeyNobidErrPrebidServerResponse] = "hb:pubnbres:%s:" + defaultServerName
	//hb:pubnbres:<pub>:<dc:node:pod>

	statKeys[statsKeyContentObjectPresent] = "hb:cnt:%s:%s:" + defaultServerName
	//hb:cnt:<app|site>:<pub>:<dc:node:pod>

	//publisher and profile level stats
	statKeys[statsKeyPublisherProfileRequests] = "hb:pprofrq:%s:%s:" + defaultServerName
	//hb:pprofrq:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyPublisherInvProfileRequests] = "hb:pubinp:%s:%s:" + defaultServerName
	//hb:pubinp:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyPublisherInvProfileImpressions] = "hb:pubinpimp:%s:%s:" + defaultServerName
	//hb:pubinpimp:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyPrebidTORequests] = "hb:prebidto:%s:%s:" + defaultServerName
	//hb:prebidto:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeySsTORequests] = "hb:ssto:%s:%s:" + defaultServerName
	//hb:ssto:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyNoUIDSErrorRequest] = "hb:nouids:%s:%s:" + defaultServerName
	//hb:nouids:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyVideoInterstitialImpressions] = "hb:ppvidinstlimps:%s:%s:" + defaultServerName
	//hb:ppvidinstlimps:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyVideoImpDisabledViaConfig] = "hb:ppdisimpcfg:%s:%s:" + defaultServerName
	//hb:ppdisimpcfg:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyVideoImpDisabledViaConnType] = "hb:ppdisimpct:%s:%s:" + defaultServerName
	//hb:ppdisimpct:<pub>:<prof>:<dc:node:pod>

	//publisher-partner level stats
	statKeys[statsKeyPublisherPartnerRequests] = "hb:pprq:%s:%s:" + defaultServerName
	//hb:pprq:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyPublisherPartnerImpressions] = "hb:ppimp:%s:%s:" + defaultServerName
	//hb:ppimp:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyPublisherPartnerNoCookieRequests] = "hb:ppnc:%s:%s:" + defaultServerName
	//hb:ppnc:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeySlotunMappedErrorRequests] = "hb:sler:%s:%s:" + defaultServerName
	//hb:sler:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyMisConfErrorRequests] = "hb:cfer:%s:%s:" + defaultServerName
	//hb:cfer:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyPartnerTimeoutErrorRequests] = "hb:toer:%s:%s:" + defaultServerName
	//hb:toer:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyUnknownPrebidErrorResponse] = "hb:uner:%s:%s:" + defaultServerName
	//hb:uner:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyNobidErrorRequests] = "hb:nber:%s:%s:" + defaultServerName
	//hb:nber:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyNobidderStatusErrorRequests] = "hb:nbse:%s:%s:" + defaultServerName
	//hb:nbse:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyLoggerErrorRequests] = "hb:wle:%s:%s:%s:" + defaultServerName
	//hb:nber:<pub>:<prof>:<version>:<dc:node:pod>

	statKeys[statsKey24PublisherRequests] = "hb:2.4:%s:pbrq:%s:" + defaultServerName
	//hb:2.4:<disp/app>:pbrq:<pub>:<dc:node:pod>

	statKeys[statsKey25BadRequests] = "hb:2.5:badreq:" + defaultServerName
	//hb:2.5:badreq:<dc:node:pod>

	statKeys[statsKey25PublisherRequests] = "hb:2.5:%s:pbrq:%s:" + defaultServerName
	//hb:2.5:<disp/app>:pbrq:<pub>:<dc:node:pod>

	statKeys[statsKeyAMPBadRequests] = "hb:amp:badreq:" + defaultServerName
	//hb:amp:badreq:<dc:node:pod>

	statKeys[statsKeyAMPPublisherRequests] = "hb:amp:pbrq:%s:" + defaultServerName
	//hb:amp:pbrq:<pub>:<dc:node:pod>

	statKeys[statsKeyAMPCacheError] = "hb:amp:ce:%s:%s:" + defaultServerName
	//hb:amp:ce:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyPublisherInvProfileAMPRequests] = "hb:amp:pubinp:%s:%s:" + defaultServerName
	//hb:amp:pubinp:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyVideoBadRequests] = "hb:vid:badreq:" + defaultServerName
	//hb:vid:badreq:<dc:node:pod>

	statKeys[statsKeyVideoPublisherRequests] = "hb:vid:pbrq:%s:" + defaultServerName
	//hb:vid:pbrq:<pub>:<dc:node:pod>

	statKeys[statsKeyVideoCacheError] = "hb:vid:ce:%s:%s:" + defaultServerName
	//hb:vid:ce:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyPublisherInvProfileVideoRequests] = "hb:vid:pubinp:%s:%s:" + defaultServerName
	//hb:vid:pubinp:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyInvalidCreatives] = "hb:invcr:%s:%s:" + defaultServerName
	//hb:invcr:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyPlatformPublisherPartnerRequests] = "hb:pppreq:%s:%s:%s:" + defaultServerName
	//hb:pppreq:<platform>:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyPlatformPublisherPartnerResponses] = "hb:pppres:%s:%s:%s:" + defaultServerName
	//hb:pppres:<platform>:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyPublisherResponseEncodingErrors] = "hb:encerr:%s:" + defaultServerName
	//hb:vid:encerr:<pub>:<dc:node:pod>

	statKeys[statsKeyA2000] = "hb:latabv_2000:%s:%s:" + defaultServerName
	//hb:latabv_2000:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA1500] = "hb:latabv_1500:%s:%s:" + defaultServerName
	//hb:latabv_1500:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA1000] = "hb:latabv_1000:%s:%s:" + defaultServerName
	//hb:latabv_1000:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA900] = "hb:latabv_900:%s:%s:" + defaultServerName
	//hb:latabv_900:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA800] = "hb:latabv_800:%s:%s:" + defaultServerName
	//hb:latabv_800:<pub>:<partner>:<dc:node:pod>

	// TBD : @viral key-change ???
	// statKeys[statsKeyA800] = statsclient.Stats{Fmt: "hb:latabv_800:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_800:<pub>:<partner>:<dc>
	// statKeys[statsKeyA700] = statsclient.Stats{Fmt: "hb:latabv_800:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_700:<pub>:<partner>:<dc>
	statKeys[statsKeyA700] = "hb:latabv_700:%s:%s:" + defaultServerName
	//hb:latabv_700:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA600] = "hb:latabv_600:%s:%s:" + defaultServerName
	//hb:latabv_600:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA500] = "hb:latabv_500:%s:%s:" + defaultServerName
	//hb:latabv_500:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA400] = "hb:latabv_400:%s:%s:" + defaultServerName
	//hb:latabv_400:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA300] = "hb:latabv_300:%s:%s:" + defaultServerName
	//hb:latabv_300:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA200] = "hb:latabv_200:%s:%s:" + defaultServerName
	//hb:latabv_200:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA100] = "hb:latabv_100:%s:%s:" + defaultServerName
	//hb:latabv_100:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyA50] = "hb:latabv_50:%s:%s:" + defaultServerName
	//hb:latabv_50:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyL50] = "hb:latblw_50:%s:%s:" + defaultServerName
	//hb:latblw_50:<pub>:<partner>:<dc:node:pod>

	statKeys[statsKeyPrTimeAbv100] = "hb:ptabv_100:%s:" + defaultServerName
	//hb:ptabv_100:<pub>:<dc:node:pod>

	statKeys[statsKeyPrTimeAbv50] = "hb:ptabv_50:%s:" + defaultServerName
	//hb:ptabv_50:<pub>:<dc:node:pod>

	statKeys[statsKeyPrTimeAbv10] = "hb:ptabv_10:%s:" + defaultServerName
	//hb:ptabv_10:<pub>:<dc:node:pod>

	statKeys[statsKeyPrTimeAbv1] = "hb:ptabv_1:%s:" + defaultServerName
	//hb:ptabv_1:<pub>:<dc:node:pod>

	statKeys[statsKeyPrTimeBlw1] = "hb:ptblw_1:%s:" + defaultServerName
	//hb:ptblw_1:<pub>:<dc:node:pod>

	statKeys[statsKeyBannerImpDisabledViaConfig] = "hb:bnrdiscfg:%s:%s:" + defaultServerName
	//hb:bnrdiscfg:<pub>:<prof>:<dc:node:pod>

	//CTV Specific Keys

	statKeys[statsKeyCTVPrebidFailedImpression] = "hb:lfv:badimp:%v:%v:%v:" + defaultServerName
	//hb:lfv:badimp:<errorcode>:<pub>:<profile>:<dc:node:pod>

	statKeys[statsKeyCTVRequests] = "hb:lfv:%v:%v:req:" + defaultServerName
	//hb:lfv:<ortb/vast/json>:<platform>:req:<dc:node:pod>

	statKeys[statsKeyCTVBadRequests] = "hb:lfv:%v:badreq:%d:" + defaultServerName
	//hb:lfv:<ortb/vast/json>:badreq:<badreq-code>:<dc:node:pod>

	statKeys[statsKeyCTVPublisherRequests] = "hb:lfv:%v:%v:pbrq:%v:" + defaultServerName
	//hb:lfv:<ortb/vast/json>:<platform>:pbrq:<pub>:<dc:node:pod>

	statKeys[statsKeyCTVHTTPMethodRequests] = "hb:lfv:%v:mtd:%v:%v:" + defaultServerName
	//hb:lfv:<ortb/vast/json>:mtd:<pub>:<get/post>:<dc:node:pod>

	statKeys[statsKeyCTVValidationErr] = "hb:lfv:ivr:%d:%s:" + defaultServerName
	//hb:lfv:ivr:<error_code>:<pub>:<dc:node:pod>

	statKeys[statsKeyIncompleteAdPods] = "hb:lfv:nip:%s:%s:" + defaultServerName
	//hb:lfv:nip:<reason>:<pub>:<dc:node:pod>

	statKeys[statsKeyCTVReqImpstWithConfig] = "hb:lfv:rwc:%s:%s:" + defaultServerName
	//hb:lfv:rwc:<req:db>:<pub>:<dc:node:pod>

	statKeys[statsKeyTotalAdPodImpression] = "hb:lfv:tpi:%s:%s:" + defaultServerName
	//hb:lfv:tpi:<imp-range>:<pub>:<dc:node:pod>

	statKeys[statsKeyReqTotalAdPodImpression] = "hb:lfv:rtpi:%s:" + defaultServerName
	//hb:lfv:rtpi:<pub>:<dc:node:pod>

	statKeys[statsKeyAdPodSecondsMissed] = "hb:lfv:sm:%s:" + defaultServerName
	//hb:lfv:sm:<pub>:<dc:node:pod>

	statKeys[statsKeyReqImpDurationYield] = "hb:lfv:impy:%d:%d:%s:" + defaultServerName
	//hb:lfv:impy:<max_duration>:<min_duration>:<pub>:<dc:node:pod>

	statKeys[statsKeyReqWithAdPodCount] = "hb:lfv:rwap:%s:%s:" + defaultServerName
	//hb:lfv:rwap:<pub>:<prof>:<dc:node:pod>

	statKeys[statsKeyBidDuration] = "hb:lfv:dur:%d:%s:%s:" + defaultServerName
	//hb:lfv:dur:<duration>:<pub>:<prof>:<dc:node:pod>:

	statKeys[statsKeyPBSAuctionRequests] = "hb:pbs:auc:" + defaultServerName
	//hb:pbs:auc:<dc:node:pod> - no of PBS auction endpoint requests

	statKeys[statsKeyInjectTrackerErrorCount] = "hb:mistrack:%s:%s:%s:" + defaultServerName
	//hb:mistrack:<adformat>:<pubid>:<partner>:<dc:node:pod> - Error during Injecting Tracker

	statKeys[statsBidResponsesByDealUsingPBS] = "hb:pbs:dbc:%s:%s:%s:%s:" + defaultServerName
	//hb:pbs:dbc:<pub>:<profile>:<aliasbidder>:<dealid>:<dc:node:pod> - PubMatic-OpenWrap to count number of responses received from aliasbidder per publisher profile

	statKeys[statsBidResponsesByDealUsingHB] = "hb:dbc:%s:%s:%s:%s:" + defaultServerName
	//hb:dbc:<pub>:<profile>:<aliasbidder>:<dealid>:<dc:node:pod> - header-bidding to count number of responses received from aliasbidder per publisher profile

	statKeys[statsPartnerTimeoutInPBS] = "hb:pbs:pto:%s:%s:%s:" + defaultServerName
	//hb:pbs:pto:<pub>:<profile>:<aliasbidder>:<dc:node:pod> - count timeout by aliasbidder per publisher profile
}
