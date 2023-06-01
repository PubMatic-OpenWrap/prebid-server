package stats

import (
	"sync"
)

type statKeyName = string

var (
	statKeys [maxNumOfStats]statKeyName
)

var once sync.Once
var owStats *StatsTCP
var owStatsErr error

// InitStat initializes stats client
func InitStat(defaultHost, actualHost, dcName string, interval int, criticalThreshold int, criticalInterval int, standardThreshold int, standardInterval int,
	portTCP string, pubInterval int, pubThreshold int, retries int, dialTimeout int, keepAliveDuration int, maxIdleConnes int, maxIdleConnesPerHost int) (*StatsTCP, error) {

	once.Do(func() {
		// initStatKeys(criticalThreshold, criticalInterval, standardThreshold, standardInterval)
		// owStats, owStatsErr = initTCPStatsClient(host, portTCP, server, dc, pubInterval, pubThreshold, retries, dialTimeout, keepAliveDuration, maxIdleConnes, maxIdleConnesPerHost)
	})

	return owStats, owStatsErr
}

// initStatKeys sets the key-name for all stats
// defaultHost will be "actualDCName:N:P"
// actualHost will be "actualHost:actualNode:actualPod"
func initStatKeys(defaultHost, actualHost string) {

	//server level stats
	// statKeys[statsKeyOpenWrapServerPanic] = Stats{Fmt: "hb:panic:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyOpenWrapServerPanic] = "hb:panic:" + actualHost
	//hb:panic:<dc:node:pod>

	//publisher level stats
	// statKeys[statsKeyPublisherNoConsentRequests] = Stats{Fmt: "hb:pubnocnsreq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherNoConsentRequests] = "hb:pubnocnsreq:%s:" + defaultHost
	//hb:pubnocnsreq:<pub>:<dc:node:pod>

	// statKeys[statsKeyPublisherNoConsentImpressions] = Stats{Fmt: "hb:pubnocnsimp:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherNoConsentImpressions] = "hb:pubnocnsimp:%s:" + defaultHost
	//hb:pubnocnsimp:<pub>:<dc:node:pod>

	// statKeys[statsKeyPublisherPrebidRequests] = Stats{Fmt: "hb:pubrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherPrebidRequests] = "hb:pubrq:%s:" + defaultHost

	// statKeys[statsKeyNobidErrPrebidServerRequests] = "hb:pubnbreq:%s:", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyNobidErrPrebidServerRequests] = "hb:pubnbreq:%s:" + defaultHost
	//hb:pubnbreq:<pub>:<dc:node:pod>

	// statKeys[statsKeyNobidErrPrebidServerResponse] = Stats{Fmt: "hb:pubnbres:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyNobidErrPrebidServerResponse] = "hb:pubnbres:%s:" + defaultHost
	//hb:pubnbres:<pub>:<dc:node:pod>

	// statKeys[statsKeyContentObjectPresent] = Stats{Fmt: "hb:cnt:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyContentObjectPresent] = "hb:cnt:%s:%s:" + defaultHost
	//hb:cnt:<app|site>:<pub>:<dc:node:pod>

	//publisher and profile level stats
	// statKeys[statsKeyPublisherProfileRequests] = Stats{Fmt: "hb:pprofrq:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherProfileRequests] = "hb:pprofrq:%s:%s:" + defaultHost
	//hb:pprofrq:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyPublisherInvProfileRequests] = "hb:pubinp:%s:%s:", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPublisherInvProfileRequests] = "hb:pubinp:%s:%s:" + defaultHost
	//hb:pubinp:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyPublisherInvProfileImpressions] = Stats{Fmt: "hb:pubinpimp:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPublisherInvProfileImpressions] = "hb:pubinpimp:%s:%s:" + defaultHost
	//hb:pubinpimp:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyPrebidTORequests] = Stats{Fmt: "hb:prebidto:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPrebidTORequests] = "hb:prebidto:%s:%s:" + defaultHost
	//hb:prebidto:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeySsTORequests] = Stats{Fmt: "hb:ssto:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeySsTORequests] = "hb:ssto:%s:%s:" + defaultHost
	//hb:ssto:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyNoUIDSErrorRequest] = Stats{Fmt: "hb:nouids:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyNoUIDSErrorRequest] = "hb:nouids:%s:%s:" + defaultHost
	//hb:nouids:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyVideoInterstitialImpressions] = Stats{Fmt: "hb:ppvidinstlimps:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyVideoInterstitialImpressions] = "hb:ppvidinstlimps:%s:%s:" + defaultHost
	//hb:ppvidinstlimps:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyVideoImpDisabledViaConfig] = Stats{Fmt: "hb:ppdisimpcfg:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyVideoImpDisabledViaConfig] = "hb:ppdisimpcfg:%s:%s:" + defaultHost
	//hb:ppdisimpcfg:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyVideoImpDisabledViaConnType] = Stats{Fmt: "hb:ppdisimpct:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyVideoImpDisabledViaConnType] = "hb:ppdisimpct:%s:%s:" + defaultHost
	//hb:ppdisimpct:<pub>:<prof>:<dc:node:pod>

	//publisher-partner level stats
	// statKeys[statsKeyPublisherPartnerRequests] = Stats{Fmt: "hb:pprq:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherPartnerRequests] = "hb:pprq:%s:%s:" + defaultHost
	//hb:pprq:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyPublisherPartnerImpressions] = Stats{Fmt: "hb:ppimp:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherPartnerImpressions] = "hb:ppimp:%s:%s:" + defaultHost
	//hb:ppimp:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyPublisherPartnerNoCookieRequests] = Stats{Fmt: "hb:ppnc:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPublisherPartnerNoCookieRequests] = "hb:ppnc:%s:%s:" + defaultHost
	//hb:ppnc:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeySlotunMappedErrorRequests] = Stats{Fmt: "hb:sler:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeySlotunMappedErrorRequests] = "hb:sler:%s:%s:" + defaultHost
	//hb:sler:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyMisConfErrorRequests] = Stats{Fmt: "hb:cfer:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyMisConfErrorRequests] = "hb:cfer:%s:%s:" + defaultHost
	//hb:cfer:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyPartnerTimeoutErrorRequests] = Stats{Fmt: "hb:toer:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPartnerTimeoutErrorRequests] = "hb:toer:%s:%s:" + defaultHost
	//hb:toer:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyUnknownPrebidErrorResponse] = Stats{Fmt: "hb:uner:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyUnknownPrebidErrorResponse] = "hb:uner:%s:%s:" + defaultHost
	//hb:uner:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyNobidErrorRequests] = Stats{Fmt: "hb:nber:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyNobidErrorRequests] = "hb:nber:%s:%s:" + defaultHost
	//hb:nber:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyNobidderStatusErrorRequests] = Stats{Fmt: "hb:nbse:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyNobidderStatusErrorRequests] = "hb:nbse:%s:%s:" + defaultHost
	//hb:nbse:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyLoggerErrorRequests] = Stats{Fmt: "hb:wle:%s:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyLoggerErrorRequests] = "hb:wle:%s:%s:%s:" + defaultHost
	//hb:nber:<pub>:<prof>:<version>:<dc:node:pod>

	// statKeys[statsKey24PublisherRequests] = Stats{Fmt: "hb:2.4:%s:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKey24PublisherRequests] = "hb:2.4:%s:pbrq:%s:" + defaultHost
	//hb:2.4:<disp/app>:pbrq:<pub>:<dc:node:pod>

	// statKeys[statsKey25BadRequests] = Stats{Fmt: "hb:2.5:badreq:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKey25BadRequests] = "hb:2.5:badreq:" + defaultHost
	//hb:2.5:badreq:<dc:node:pod>

	// statKeys[statsKey25PublisherRequests] = Stats{Fmt: "hb:2.5:%s:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKey25PublisherRequests] = "hb:2.5:%s:pbrq:%s:" + defaultHost
	//hb:2.5:<disp/app>:pbrq:<pub>:<dc:node:pod>

	// statKeys[statsKeyAMPBadRequests] = Stats{Fmt: "hb:amp:badreq:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyAMPBadRequests] = "hb:amp:badreq:" + defaultHost
	//hb:amp:badreq:<dc:node:pod>

	// statKeys[statsKeyAMPPublisherRequests] = Stats{Fmt: "hb:amp:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyAMPPublisherRequests] = "hb:amp:pbrq:%s:" + defaultHost
	//hb:amp:pbrq:<pub>:<dc:node:pod>

	// statKeys[statsKeyAMPCacheError] = Stats{Fmt: "hb:amp:ce::%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyAMPCacheError] = "hb:amp:ce::%s:%s:" + defaultHost
	//hb:amp:ce:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyPublisherInvProfileAMPRequests] = Stats{Fmt: "hb:amp:pubinp:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPublisherInvProfileAMPRequests] = "hb:amp:pubinp:%s:%s:" + defaultHost
	//hb:amp:pubinp:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyVideoBadRequests] = Stats{Fmt: "hb:vid:badreq:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyVideoBadRequests] = "hb:vid:badreq:" + defaultHost
	//hb:vid:badreq:<dc:node:pod>

	// statKeys[statsKeyVideoPublisherRequests] = Stats{Fmt: "hb:vid:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyVideoPublisherRequests] = "hb:vid:pbrq:%s:" + defaultHost
	//hb:vid:pbrq:<pub>:<dc:node:pod>

	// statKeys[statsKeyVideoCacheError] = Stats{Fmt: "hb:vid:ce:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyVideoCacheError] = "hb:vid:ce:%s:%s:" + defaultHost
	//hb:vid:ce:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyPublisherInvProfileVideoRequests] = Stats{Fmt: "hb:vid:pubinp:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPublisherInvProfileVideoRequests] = "hb:vid:pubinp:%s:%s:" + defaultHost
	//hb:vid:pubinp:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyInvalidCreatives] = Stats{Fmt: "hb:invcr:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyInvalidCreatives] = "hb:invcr:%s:%s:" + defaultHost
	//hb:invcr:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyPlatformPublisherPartnerRequests] = Stats{Fmt: "hb:pppreq:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPlatformPublisherPartnerRequests] = "hb:pppreq:%s:%s:%s:" + defaultHost
	//hb:pppreq:<platform>:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyPlatformPublisherPartnerResponses] = Stats{Fmt: "hb:pppres:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPlatformPublisherPartnerResponses] = "hb:pppres:%s:%s:%s:" + defaultHost
	//hb:pppres:<platform>:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyPublisherResponseEncodingErrors] = Stats{Fmt: "hb:encerr:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyPublisherResponseEncodingErrors] = "hb:encerr:%s:" + defaultHost
	//hb:vid:encerr:<pub>:<dc:node:pod>

	// statKeys[statsKeyA2000] = Stats{Fmt: "hb:latabv_2000:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA2000] = "hb:latabv_2000:%s:%s:" + defaultHost
	//hb:latabv_2000:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA1500] = Stats{Fmt: "hb:latabv_1500:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA1500] = "hb:latabv_1500:%s:%s:" + defaultHost
	//hb:latabv_1500:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA1000] = Stats{Fmt: "hb:latabv_1000:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA1000] = "hb:latabv_1000:%s:%s:" + defaultHost
	//hb:latabv_1000:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA900] = Stats{Fmt: "hb:latabv_900:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA900] = "hb:latabv_900:%s:%s:" + defaultHost
	//hb:latabv_900:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA800] = Stats{Fmt: "hb:latabv_800:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA800] = "hb:latabv_800:%s:%s:" + defaultHost
	//hb:latabv_800:<pub>:<partner>:<dc:node:pod>

	// TBD : @viral key-change ???
	// statKeys[statsKeyA700] = Stats{Fmt: "hb:latabv_800:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA700] = "hb:latabv_700:%s:%s:" + defaultHost
	//hb:latabv_700:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA600] = Stats{Fmt: "hb:latabv_600:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA600] = "hb:latabv_600:%s:%s:" + defaultHost
	//hb:latabv_600:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA500] = Stats{Fmt: "hb:latabv_500:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA500] = "hb:latabv_500:%s:%s:" + defaultHost
	//hb:latabv_500:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA400] = Stats{Fmt: "hb:latabv_400:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA400] = "hb:latabv_400:%s:%s:" + defaultHost
	//hb:latabv_400:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA300] = Stats{Fmt: "hb:latabv_300:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA300] = "hb:latabv_300:%s:%s:" + defaultHost
	//hb:latabv_300:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA200] = Stats{Fmt: "hb:latabv_200:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA200] = "hb:latabv_200:%s:%s:" + defaultHost
	//hb:latabv_200:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA100] = Stats{Fmt: "hb:latabv_100:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA100] = "hb:latabv_100:%s:%s:" + defaultHost
	//hb:latabv_100:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyA50] = Stats{Fmt: "hb:latabv_50:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyA50] = "hb:latabv_50:%s:%s:" + defaultHost
	//hb:latabv_50:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyL50] = Stats{Fmt: "hb:latblw_50:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyL50] = "hb:latblw_50:%s:%s:" + defaultHost
	//hb:latblw_50:<pub>:<partner>:<dc:node:pod>

	// statKeys[statsKeyPrTimeAbv100] = Stats{Fmt: "hb:ptabv_100:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPrTimeAbv100] = "hb:ptabv_100:%s:" + defaultHost
	//hb:ptabv_100:<pub>:<dc:node:pod>

	// statKeys[statsKeyPrTimeAbv50] = Stats{Fmt: "hb:ptabv_50:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPrTimeAbv50] = "hb:ptabv_50:%s:" + defaultHost
	//hb:ptabv_50:<pub>:<dc:node:pod>

	// statKeys[statsKeyPrTimeAbv10] = Stats{Fmt: "hb:ptabv_10:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPrTimeAbv10] = "hb:ptabv_10:%s:" + defaultHost
	//hb:ptabv_10:<pub>:<dc:node:pod>

	// statKeys[statsKeyPrTimeAbv1] = Stats{Fmt: "hb:ptabv_1:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPrTimeAbv1] = "hb:ptabv_1:%s:" + defaultHost
	//hb:ptabv_1:<pub>:<dc:node:pod>

	// statKeys[statsKeyPrTimeBlw1] = Stats{Fmt: "hb:ptblw_1:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPrTimeBlw1] = "hb:ptblw_1:%s:" + defaultHost
	//hb:ptblw_1:<pub>:<dc:node:pod>

	// statKeys[statsKeyBannerImpDisabledViaConfig] = Stats{Fmt: "hb:bnrdiscfg:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyBannerImpDisabledViaConfig] = "hb:bnrdiscfg:%s:%s:" + defaultHost
	//hb:bnrdiscfg:<pub>:<prof>:<dc:node:pod>

	//CTV Specific Keys

	// statKeys[statsKeyCTVPrebidFailedImpression] = Stats{Fmt: "hb:lfv:badimp:%v:%v:%v:%v", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalThreshold)}
	statKeys[statsKeyCTVPrebidFailedImpression] = "hb:lfv:badimp:%v:%v:%v:" + defaultHost
	//hb:lfv:badimp:<errorcode>:<pub>:<profile>:<dc:node:pod>

	// statKeys[statsKeyCTVRequests] = Stats{Fmt: "hb:lfv:%v:%v:req:%v", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyCTVRequests] = "hb:lfv:%v:%v:req:" + defaultHost
	//hb:lfv:<ortb/vast/json>:<platform>:req:<dc:node:pod>

	// statKeys[statsKeyCTVBadRequests] = Stats{Fmt: "hb:lfv:%v:badreq:%d:%v", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalThreshold)}
	statKeys[statsKeyCTVBadRequests] = "hb:lfv:%v:badreq:%d:" + defaultHost
	//hb:lfv:<ortb/vast/json>:badreq:<badreq-code>:<dc:node:pod>

	// statKeys[statsKeyCTVPublisherRequests] = Stats{Fmt: "hb:lfv:%v:%v:pbrq:%v:%v", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyCTVPublisherRequests] = "hb:lfv:%v:%v:pbrq:%v:" + defaultHost
	//hb:lfv:<ortb/vast/json>:<platform>:pbrq:<pub>:<dc:node:pod>

	// statKeys[statsKeyCTVHTTPMethodRequests] = Stats{Fmt: "hb:lfv:%v:mtd:%v:%v:%v", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyCTVHTTPMethodRequests] = "hb:lfv:%v:mtd:%v:%v:" + defaultHost
	//hb:lfv:<ortb/vast/json>:mtd:<pub>:<get/post>:<dc:node:pod>

	// statKeys[statsKeyCTVValidationErr] = Stats{Fmt: "hb:lfv:ivr:%d:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	statKeys[statsKeyCTVValidationErr] = "hb:lfv:ivr:%d:%s:" + defaultHost
	//hb:lfv:ivr:<error_code>:<pub>:<dc:node:pod>

	// statKeys[statsKeyIncompleteAdPods] = Stats{Fmt: "hb:lfv:nip:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyIncompleteAdPods] = "hb:lfv:nip:%s:%s:" + defaultHost
	//hb:lfv:nip:<reason>:<pub>:<dc:node:pod>

	// statKeys[statsKeyCTVReqImpstWithConfig] = Stats{Fmt: "hb:lfv:rwc:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyCTVReqImpstWithConfig] = "hb:lfv:rwc:%s:%s:" + defaultHost
	//hb:lfv:rwc:<req:db>:<pub>:<dc:node:pod>

	// statKeys[statsKeyTotalAdPodImpression] = Stats{Fmt: "hb:lfv:tpi:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyTotalAdPodImpression] = "hb:lfv:tpi:%s:%s:" + defaultHost
	//hb:lfv:tpi:<imp-range>:<pub>:<dc:node:pod>

	// statKeys[statsKeyReqTotalAdPodImpression] = Stats{Fmt: "hb:lfv:rtpi:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyReqTotalAdPodImpression] = "hb:lfv:rtpi:%s:" + defaultHost
	//hb:lfv:rtpi:<pub>:<dc:node:pod>

	// statKeys[statsKeyAdPodSecondsMissed] = Stats{Fmt: "hb:lfv:sm:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyAdPodSecondsMissed] = "hb:lfv:sm:%s:" + defaultHost
	//hb:lfv:sm:<pub>:<dc:node:pod>

	// statKeys[statsKeyReqImpDurationYield] = Stats{Fmt: "hb:lfv:impy:%d:%d:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyReqImpDurationYield] = "hb:lfv:impy:%d:%d:%s:" + defaultHost
	//hb:lfv:impy:<max_duration>:<min_duration>:<pub>:<dc:node:pod>

	// statKeys[statsKeyReqWithAdPodCount] = Stats{Fmt: "hb:lfv:rwap:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyReqWithAdPodCount] = "hb:lfv:rwap:%s:%s:" + defaultHost
	//hb:lfv:rwap:<pub>:<prof>:<dc:node:pod>

	// statKeys[statsKeyBidDuration] = Stats{Fmt: "hb:lfv:dur:%d:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyBidDuration] = "hb:lfv:dur:%d:%s:%s:" + defaultHost
	//hb:lfv:dur:<duration>:<pub>:<prof>:<dc:node:pod>:

	//

	// statKeys[statsKeyPublisherPartnerAdomainPresent] = Stats{Fmt: "hb:dompres:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherPartnerAdomainPresent] = "hb:dompres:%s:%s:%s:" + defaultHost
	//hb:dompres:<creativeType>:<pub>:<partner>:<dc:node:pod> - ADomain present in bid response

	// statKeys[statsKeyPublisherPartnerAdomainAbsent] = Stats{Fmt: "hb:domabs:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherPartnerAdomainAbsent] = "hb:domabs:%s:%s:%s:" + defaultHost
	//hb:domabs:<creativeType>:<pub>:<partner>:<dc:node:pod> - ADomain absent in bid response

	// statKeys[statsKeyPublisherPartnerCatPresent] = Stats{Fmt: "hb:catpres:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherPartnerCatPresent] = "hb:catpres:%s:%s:%s:" + defaultHost
	//hb:catpres:<creativeType>:<pub>:<partner>:<dc:node:pod> - Category present in bid response

	// statKeys[statsKeyPublisherPartnerCatAbsent] = Stats{Fmt: "hb:catabs:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPublisherPartnerCatAbsent] = "hb:catabs:%s:%s:%s:" + defaultHost
	//hb:catabs:<creativeType>:<pub>:<partner>:<dc:node:pod> - Category absent in bid response

	// statKeys[statsKeyPBSAuctionRequests] = Stats{Fmt: "hb:pbs:auc:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyPBSAuctionRequests] = "hb:pbs:auc:" + defaultHost
	//hb:pbs:auc:<dc:node:pod> - no of PBS auction endpoint requests

	// statKeys[statsKeyInjectTrackerErrorCount] = Stats{Fmt: "hb:mistrack:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	statKeys[statsKeyInjectTrackerErrorCount] = "hb:mistrack:%s:%s:%s:" + defaultHost
	//hb:mistrack:<adformat>:<pubid>:<partner>:<dc:node:pod> - Error during Injecting Tracker

	statKeys[statsBidResponsesByDealUsingPBS] = "hb:pbs:dbc:%s:%s:%s:%s:" + defaultHost
	//hb:pbs:dbc:<pub>:<profile>:<aliasbidder>:<dealid>:<dc:node:pod> - PubMatic-OpenWrap to count number of responses received from aliasbidder per publisher profile

	statKeys[statsBidResponsesByDealUsingHB] = "hb:dbc:%s:%s:%s:%s:" + defaultHost
	//hb:dbc:<pub>:<profile>:<aliasbidder>:<dealid>:<dc:node:pod> - header-bidding to count number of responses received from aliasbidder per publisher profile

	statKeys[statsPartnerTimeoutInPBS] = "hb:pbs:pto:%s:%s:%s:" + defaultHost
	//hb:pbs:pto:<pub>:<profile>:<aliasbidder>:<dc:node:pod> - count timeout by aliasbidder per publisher profile
}
