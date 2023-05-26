package stats

import (
	"sync"
	"time"
)

// AAA: temp code

// Stats models a single Stats.  It encapsulates the Stats format and
// corresponding configurations (like send threshold).
//
// Example-
//
//	Stats{
//	        Fmt: "GOPRO:REQ:%s:%s:%s",
//	        SendThresh: 50,
//	}

type Stats struct {
	Fmt              string
	SendThresh       int
	SendTimeInterval time.Duration
}

var (
	statKeys [maxStats]Stats
)

var once sync.Once
var owStats *StatsTCP
var owStatsErr error

// InitStat will be called 2 times
// 1 by OW-module and 1 by HB
// InitStat initializes stats client
func InitStat(host, server, dc string, interval int, criticalThreshold int, criticalInterval int, standardThreshold int, standardInterval int,
	portTCP string, pubInterval int, pubThreshold int, retries int, dialTimeout int, keepAliveDuration int, maxIdleConnes int, maxIdleConnesPerHost int) (*StatsTCP, error) {

	once.Do(func() {
		initStatKeys(criticalThreshold, criticalInterval, standardThreshold, standardInterval)
		owStats, owStatsErr = initTCPStatsClient(host, portTCP, server, dc, pubInterval, pubThreshold, retries, dialTimeout, keepAliveDuration, maxIdleConnes, maxIdleConnesPerHost)
	})

	return owStats, owStatsErr
}

func initStatKeys(criticalThreshold, criticalInterval, standardThreshold, standardInterval int) {
	//server level stats
	statKeys[statsKeyOpenWrapServerPanic] = Stats{Fmt: "hb:panic:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:panic:<dc>  :   Frequency:5M

	//publisher level stats
	statKeys[statsKeyPublisherNoConsentRequests] = Stats{Fmt: "hb:pubnocnsreq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pubnocnsreq:<pub>:<dc>
	statKeys[statsKeyPublisherNoConsentImpressions] = Stats{Fmt: "hb:pubnocnsimp:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pubnocnsimp:<pub>:<dc>
	statKeys[statsKeyPublisherPrebidRequests] = Stats{Fmt: "hb:pubrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pubrq:<pub>:<dc>
	statKeys[statsKeyNobidErrPrebidServerRequests] = Stats{Fmt: "hb:pubnbreq:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:pubnbreq:<pub>:<dc>
	statKeys[statsKeyNobidErrPrebidServerResponse] = Stats{Fmt: "hb:pubnbres:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:pubnbres:<pub>:<dc>
	statKeys[statsKeyContentObjectPresent] = Stats{Fmt: "hb:cnt:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:cnt:<app|site>:<pub>:<dc>:

	//publisher and profile level stats
	statKeys[statsKeyPublisherProfileRequests] = Stats{Fmt: "hb:pprofrq:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pprofrq:<pub>:<prof>:<dc>
	statKeys[statsKeyPublisherInvProfileRequests] = Stats{Fmt: "hb:pubinp:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:pubinp:<pub>:<prof>:<dc>
	statKeys[statsKeyPublisherInvProfileImpressions] = Stats{Fmt: "hb:pubinpimp:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:pubinpimp:<pub>:<prof>:<dc>
	statKeys[statsKeyPrebidTORequests] = Stats{Fmt: "hb:prebidto:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:prebidto:<pub>:<prof>:<dc>
	statKeys[statsKeySsTORequests] = Stats{Fmt: "hb:ssto:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:ssto:<pub>:<prof>:<dc>
	statKeys[statsKeyNoUIDSErrorRequest] = Stats{Fmt: "hb:nouids:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:nouids:<pub>:<prof>:<dc>
	statKeys[statsKeyVideoInterstitialImpressions] = Stats{Fmt: "hb:ppvidinstlimps:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:ppviimps:<pub>:<prof>:<dc>
	statKeys[statsKeyVideoImpDisabledViaConfig] = Stats{Fmt: "hb:ppdisimpcfg:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:ppdisimpcfg:<pub>:<prof>:<dc>
	statKeys[statsKeyVideoImpDisabledViaConnType] = Stats{Fmt: "hb:ppdisimpct:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:ppdisimpct:<pub>:<prof>:<dc>

	//publisher-partner level stats
	statKeys[statsKeyPublisherPartnerRequests] = Stats{Fmt: "hb:pprq:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pprq:<pub>:<partner>:<dc>
	statKeys[statsKeyPublisherPartnerImpressions] = Stats{Fmt: "hb:ppimp:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:ppimp:<pub>:<partner>:<dc>
	statKeys[statsKeyPublisherPartnerNoCookieRequests] = Stats{Fmt: "hb:ppnc:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:ppnc:<pub>:<partner>:<dc>
	statKeys[statsKeySlotunMappedErrorRequests] = Stats{Fmt: "hb:sler:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:sler:<pub>:<partner>:<dc>
	statKeys[statsKeyMisConfErrorRequests] = Stats{Fmt: "hb:cfer:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:cfer:<pub>:<partner>:<dc>
	statKeys[statsKeyPartnerTimeoutErrorRequests] = Stats{Fmt: "hb:toer:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:toer:<pub>:<partner>:<dc>
	statKeys[statsKeyUnknownPrebidErrorResponse] = Stats{Fmt: "hb:uner:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:uner:<pub>:<partner>:<dc>
	statKeys[statsKeyNobidErrorRequests] = Stats{Fmt: "hb:nber:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:nber:<pub>:<partner>:<dc>
	statKeys[statsKeyNobidderStatusErrorRequests] = Stats{Fmt: "hb:nbse:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:nbse:<pub>:<partner>:<dc>
	statKeys[statsKeyLoggerErrorRequests] = Stats{Fmt: "hb:wle:%s:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:nber:<pub>:<prof>:<version>:<dc>
	statKeys[statsKey24PublisherRequests] = Stats{Fmt: "hb:2.4:%s:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:2.4:(disp/app):pbrq:<pub>:<dc>
	statKeys[statsKey25BadRequests] = Stats{Fmt: "hb:2.5:badreq:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:2.5:badreq:<dc>
	statKeys[statsKey25PublisherRequests] = Stats{Fmt: "hb:2.5:%s:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:2.5:(disp/app):pbrq:<pub>:<dc>
	statKeys[statsKeyAMPBadRequests] = Stats{Fmt: "hb:amp:badreq:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:amp:badreq:<dc>
	statKeys[statsKeyAMPPublisherRequests] = Stats{Fmt: "hb:amp:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:amp:pbrq:<pub>:<dc>
	statKeys[statsKeyAMPCacheError] = Stats{Fmt: "hb:amp:ce::%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:amp:ce:<pub>:<prof>:<dc>
	statKeys[statsKeyPublisherInvProfileAMPRequests] = Stats{Fmt: "hb:amp:pubinp:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:amp:pubinp:<pub>:<prof>:<dc>
	statKeys[statsKeyVideoBadRequests] = Stats{Fmt: "hb:vid:badreq:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:vid:badreq:<dc>
	statKeys[statsKeyVideoPublisherRequests] = Stats{Fmt: "hb:vid:pbrq:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:vid:pbrq:<pub>:<dc>
	statKeys[statsKeyVideoCacheError] = Stats{Fmt: "hb:vid:ce:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:vid:ce:<pub>:<prof>:<dc>
	statKeys[statsKeyPublisherInvProfileVideoRequests] = Stats{Fmt: "hb:vid:pubinp:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:vid:pubinp:<pub>:<prof>:<dc>
	statKeys[statsKeyInvalidCreatives] = Stats{Fmt: "hb:invcr:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:invcr:<pub>:<partner>:<dc>
	statKeys[statsKeyPlatformPublisherPartnerRequests] = Stats{Fmt: "hb:pppreq:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pppreq:<platform>:<pub>:<partner>:<dc>
	statKeys[statsKeyPlatformPublisherPartnerResponses] = Stats{Fmt: "hb:pppres:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pppres:<platform>:<pub>:<partner>:<dc>
	statKeys[statsKeyPublisherResponseEncodingErrors] = Stats{Fmt: "hb:encerr:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:vid:encerr:<pub>:<dc>
	statKeys[statsKeyA2000] = Stats{Fmt: "hb:latabv_2000:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_2000:<pub>:<partner>:<dc>
	statKeys[statsKeyA1500] = Stats{Fmt: "hb:latabv_1500:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_1500:<pub>:<partner>:<dc>
	statKeys[statsKeyA1000] = Stats{Fmt: "hb:latabv_1000:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_1000:<pub>:<partner>:<dc>
	statKeys[statsKeyA900] = Stats{Fmt: "hb:latabv_900:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_900:<pub>:<partner>:<dc>
	statKeys[statsKeyA800] = Stats{Fmt: "hb:latabv_800:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_800:<pub>:<partner>:<dc>
	statKeys[statsKeyA700] = Stats{Fmt: "hb:latabv_800:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_700:<pub>:<partner>:<dc>
	statKeys[statsKeyA600] = Stats{Fmt: "hb:latabv_600:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_600:<pub>:<partner>:<dc>
	statKeys[statsKeyA500] = Stats{Fmt: "hb:latabv_500:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_500:<pub>:<partner>:<dc>
	statKeys[statsKeyA400] = Stats{Fmt: "hb:latabv_400:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_400:<pub>:<partner>:<dc>
	statKeys[statsKeyA300] = Stats{Fmt: "hb:latabv_300:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_300:<pub>:<partner>:<dc>
	statKeys[statsKeyA200] = Stats{Fmt: "hb:latabv_200:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_200:<pub>:<partner>:<dc>
	statKeys[statsKeyA100] = Stats{Fmt: "hb:latabv_100:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_100:<pub>:<partner>:<dc>
	statKeys[statsKeyA50] = Stats{Fmt: "hb:latabv_50:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latabv_50:<pub>:<partner>:<dc>
	statKeys[statsKeyL50] = Stats{Fmt: "hb:latblw_50:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:latblw_50:<pub>:<partner>:<dc>
	statKeys[statsKeyPrTimeAbv100] = Stats{Fmt: "hb:ptabv_100:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:ptabv_100:<pub>:<dc>
	statKeys[statsKeyPrTimeAbv50] = Stats{Fmt: "hb:ptabv_50:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:ptabv_50:<pub>:<dc>
	statKeys[statsKeyPrTimeAbv10] = Stats{Fmt: "hb:ptabv_10:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:ptabv_10:<pub>:<dc>
	statKeys[statsKeyPrTimeAbv1] = Stats{Fmt: "hb:ptabv_1:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:ptabv_1:<pub>:<dc>
	statKeys[statsKeyPrTimeBlw1] = Stats{Fmt: "hb:ptblw_1:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:ptblw_1:<pub>:<dc>
	statKeys[statsKeyBannerImpDisabledViaConfig] = Stats{Fmt: "hb:bnrdiscfg:%s:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:bnrdiscfg:<pub>:<prof>:<dc>

	//CTV Specific Keys
	statKeys[statsKeyCTVPrebidFailedImpression] = Stats{Fmt: "hb:lfv:badimp:%v:%v:%v:%v", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalThreshold)}
	//hb:lfv:badimp:<errorcode>:<pub>:<profile>:<dc>:
	statKeys[statsKeyCTVRequests] = Stats{Fmt: "hb:lfv:%v:%v:req:%v", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:<ortb/vast/json>:<platform>:req:<dc>:
	statKeys[statsKeyCTVBadRequests] = Stats{Fmt: "hb:lfv:%v:badreq:%d:%v", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalThreshold)}
	//hb:lfv:<ortb/vast/json>:badreq:<badreq-code>:<dc>:
	statKeys[statsKeyCTVPublisherRequests] = Stats{Fmt: "hb:lfv:%v:%v:pbrq:%v:%v", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:<ortb/vast/json>:<platform>:pbrq:<pub>:<dc>:
	statKeys[statsKeyCTVHTTPMethodRequests] = Stats{Fmt: "hb:lfv:%v:mtd:%v:%v:%v", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:<ortb/vast/json>:mtd:<pub>:<get/post>:<dc>:
	statKeys[statsKeyCTVValidationErr] = Stats{Fmt: "hb:lfv:ivr:%d:%s:%s", SendThresh: criticalThreshold, SendTimeInterval: time.Minute * time.Duration(criticalInterval)}
	//hb:lfv:ivr:<error_code>:<pub>:<dc>:
	statKeys[statsKeyIncompleteAdPods] = Stats{Fmt: "hb:lfv:nip:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:nip:<reason>:<pub>:<dc>:
	statKeys[statsKeyCTVReqImpstWithConfig] = Stats{Fmt: "hb:lfv:rwc:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:rwc:<req:db>:<pub>:<dc>:
	statKeys[statsKeyTotalAdPodImpression] = Stats{Fmt: "hb:lfv:tpi:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:tpi:<imp-range>:<pub>:<dc>:
	statKeys[statsKeyReqTotalAdPodImpression] = Stats{Fmt: "hb:lfv:rtpi:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:rtpi:<pub>:<dc>:
	statKeys[statsKeyAdPodSecondsMissed] = Stats{Fmt: "hb:lfv:sm:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:sm:<pub>:<dc>:
	statKeys[statsKeyReqImpDurationYield] = Stats{Fmt: "hb:lfv:impy:%d:%d:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:impy:<max_duration>:<min_duration>:<pub>:<dc>:
	statKeys[statsKeyReqWithAdPodCount] = Stats{Fmt: "hb:lfv:rwap:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:rwap:<pub>:<prof>:<dc>:
	statKeys[statsKeyBidDuration] = Stats{Fmt: "hb:lfv:dur:%d:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:lfv:dur:<duration>:<pub>:<prof>:<dc>:

	statKeys[statsKeyPublisherPartnerAdomainPresent] = Stats{Fmt: "hb:dompres:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:dompres:<creativeType>:<pub>:<partner>:<dc> - ADomain present in bid response
	statKeys[statsKeyPublisherPartnerAdomainAbsent] = Stats{Fmt: "hb:domabs:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:domabs:<creativeType>:<pub>:<partner>:<dc> - ADomain absent in bid response
	statKeys[statsKeyPublisherPartnerCatPresent] = Stats{Fmt: "hb:catpres:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:catpres:<creativeType>:<pub>:<partner>:<dc> - Category present in bid response
	statKeys[statsKeyPublisherPartnerCatAbsent] = Stats{Fmt: "hb:catabs:%s:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:catabs:<creativeType>:<pub>:<partner>:<dc> - Category absent in bid response
	statKeys[statsKeyPBSAuctionRequests] = Stats{Fmt: "hb:pbs:auc:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:pbs:auc:<dc> - no of PBS auction endpoint requests
	statKeys[statsKeyInjectTrackerErrorCount] = Stats{Fmt: "hb:mistrack:%s:%s:%s", SendThresh: standardThreshold, SendTimeInterval: time.Minute * time.Duration(standardInterval)}
	//hb:mistrack:<adformat>:<pubid>:<partner> - Error during Injecting Tracker
}
