package stats

const (
	minPublishingInterval = 2  // In minutes
	maxPublishingInterval = 5  // In minutes
	minRetryDuration      = 35 // In seconds

	minDialTimeout       = 2  // In seconds
	minKeepAliveDuration = 15 // In seconds

	contentType     = "Content-Type"
	applicationJSON = "application/json;charset=utf-8"

	minPublishingThreshold   = 1000
	minResponseHeaderTimeout = 30
	minChannelLength         = 1000
	minPoolWorker            = 10
	minPoolCapacity          = 1000
)

const (
	// ADD NEW STATS ID HERE

	//statsKeyOpenWrapServerPanic stats Key for Server Panic Hits
	statsKeyOpenWrapServerPanic = iota

	//statsKeyPublisherNoConsentRequests stats Key for Counting requests for Publisher with no GDPR consent request respective publisher
	statsKeyPublisherNoConsentRequests

	//statsKeyPublisherNoConsentImpressions stats Key for Counting number of impressions lost in request due to no  GDPR consent for respective publisher
	statsKeyPublisherNoConsentImpressions

	//statsKeyPublisherPrebidRequests stats Key to count Requests to Prebid Server for respective publisher
	statsKeyPublisherPrebidRequests

	//statsKeyNobidErrPrebidServerRequests stats Key to count  Prebid Server Requests with  No AdUnit for respective publisher
	statsKeyNobidErrPrebidServerRequests

	//statsKeyNobidErrPrebidServerResponse stats Key to count  requests with No bid from prebid server response for respective publisher
	statsKeyNobidErrPrebidServerResponse

	//statsKeyContentObjectPresent for tracking the usage of content object in requests
	statsKeyContentObjectPresent

	//statsKeyPublisherProfileRequests stats Key for Counting requests for a Profile Id for respective publisher
	statsKeyPublisherProfileRequests

	//statsKeyPublisherInvProfileRequests stats Key for Counting requests with Invalid Profile Id for respective publisher
	statsKeyPublisherInvProfileRequests

	//statsKeyPublisherInvProfileImpressions stats Key for Counting number of impressions lost in request with Invalid Profile Id for respective publisher
	statsKeyPublisherInvProfileImpressions

	//statsKeyPrebidTORequests stats Key to count no of requests in which prebid timeouts
	statsKeyPrebidTORequests

	//statsKeySsTORequests stats Key for Counting requests in which server side timeouts
	statsKeySsTORequests

	//statsKeyNoUIDSErrorRequest stats Key for Counting requests with uids cookie not present
	statsKeyNoUIDSErrorRequest

	//statsKeyVideoInterstitialImpressions stats Key for Counting video interstitial impressions for a publisher/profile
	statsKeyVideoInterstitialImpressions

	//statsKeyVideoImpDisabledViaConfig stats Key for Counting video interstitial impressions that are disabled via config for a publisher/profile
	statsKeyVideoImpDisabledViaConfig

	//statsKeyVideoImpDisabledViaConnType stats Key for Counting video interstitial impressions that are disabled because of connection type for a publisher/profile
	statsKeyVideoImpDisabledViaConnType

	//statsKeyPublisherPartnerRequests stats Key for counting Publisher Partner level Requests
	statsKeyPublisherPartnerRequests

	//statsKeyPublisherPartnerImpressions stats Key for counting Publisher Partner level Impressions
	statsKeyPublisherPartnerImpressions

	//statsKeyPublisherPartnerNoCookieRequests stats Key for counting requests without cookie at Publisher Partner level
	statsKeyPublisherPartnerNoCookieRequests

	//statsKeySlotunMappedErrorRequests stats Key for counting Unmapped Slot impressions  for respective Publisher Partner
	statsKeySlotunMappedErrorRequests

	//statsKeyMisConfErrorRequests stats Key for counting  missing configuration impressions for Publisher Partner
	statsKeyMisConfErrorRequests

	//statsKeyPartnerTimeoutErrorRequests stats Key for counting Partner Timeout Requests for Parnter
	statsKeyPartnerTimeoutErrorRequests

	//statsKeyUnknownPrebidErrorResponse stats Key for counting Unknown Error from Prebid Server for respective partner
	statsKeyUnknownPrebidErrorResponse

	//statsKeyNobidErrorRequests stats Key for counting No Bid cases from respective partner
	statsKeyNobidErrorRequests

	//statsKeyNobidderStatusErrorRequests stats Key for counting No Bidders Status present in Prebid  Server response
	statsKeyNobidderStatusErrorRequests

	//statsKeyLoggerErrorRequests stats Key for counting number of Wrapper logger failures for a given publisher,profile  and version
	statsKeyLoggerErrorRequests

	//statsKey24PublisherRequests stats key to count no of 2.4 requests for a publisher
	statsKey24PublisherRequests

	//statsKey25BadRequests stats key to count no of bad requests at 2.5 endpoint
	statsKey25BadRequests

	//statsKey25PublisherRequests stats key to count no of 2.5 requests for a publisher
	statsKey25PublisherRequests

	//statsKeyAMPBadRequests stats Key for counting number of AMP bad Requests
	statsKeyAMPBadRequests

	//statsKeyAMPPublisherRequests stats Key for counting number of AMP Request for a publisher
	statsKeyAMPPublisherRequests

	//statsKeyAMPCacheError stats Key for counting cache error for given pub and profile
	statsKeyAMPCacheError

	//statsKeyPublisherInvProfileAMPRequests stats Key for Counting AMP requests with Invalid Profile Id for respective publisher
	statsKeyPublisherInvProfileAMPRequests

	//statsKeyVideoBadRequests stats Key for counting number of Video Request
	statsKeyVideoBadRequests

	//statsKeyVideoPublisherRequests stats Key for counting number of Video Request for a publisher
	statsKeyVideoPublisherRequests

	//statsKeyVideoCacheError stats Key for counting cache error
	statsKeyVideoCacheError

	//statsKeyPublisherInvProfileVideoRequests stats Key for Counting Video requests with Invalid Profile Id for respective publisher
	statsKeyPublisherInvProfileVideoRequests

	//statsKeyInvalidCreatives stats Key for counting invalid creatives for Publisher Partner
	statsKeyInvalidCreatives

	//statsKeyPlatformPublisherPartnerRequests stats Key for counting Platform Publisher Partner level Requests
	statsKeyPlatformPublisherPartnerRequests

	//statsKeyPlatformPublisherPartnerResponses stats Key for counting Platform Publisher Partner level Responses
	statsKeyPlatformPublisherPartnerResponses

	//statsKeyPublisherResponseEncodingErrors stats Key to count errors during response encoding at Publisher level
	statsKeyPublisherResponseEncodingErrors

	//Bucketwise latency related Stats Keys for publisher partner level for response time

	//statsKeyA2000 response time above 2000ms
	statsKeyA2000
	//statsKeyA1500 response time between 1500ms and 2000ms
	statsKeyA1500
	//statsKeyA1000 response time between 1000ms and 1500ms
	statsKeyA1000
	//statsKeyA900 response time between 900ms and 1000ms
	statsKeyA900
	//statsKeyA800 response time between 800ms and 900ms
	statsKeyA800
	//statsKeyA700 response time between 700ms and 800ms
	statsKeyA700
	//statsKeyA600 response time between 600ms and 700ms
	statsKeyA600
	//statsKeyA500 response time between 500ms and 600ms
	statsKeyA500
	//statsKeyA400 response time between 400ms and 500ms
	statsKeyA400
	//statsKeyA300 response time between 300ms and 400ms
	statsKeyA300
	//statsKeyA200 response time between 200ms and 300ms
	statsKeyA200
	//statsKeyA100 response time between 100ms and 200ms
	statsKeyA100
	//statsKeyA50 response time between 50ms and 100ms
	statsKeyA50
	//statsKeyL50 response time less than 50ms
	statsKeyL50

	//Bucketwise latency related Stats Keys for a publisher for pre-processing time

	//statsKeyPrTimeAbv100 bucket for pre processing time above 100ms
	statsKeyPrTimeAbv100
	//statsKeyPrTimeAbv50 bucket for pre processing time bw 50ms anb 100ms
	statsKeyPrTimeAbv50
	//statsKeyPrTimeAbv10 bucket for pre processing time bw 10ms anb 50ms
	statsKeyPrTimeAbv10
	//statsKeyPrTimeAbv1 bucket for pre processing time bw 1ms anb 10ms
	statsKeyPrTimeAbv1
	//statsKeyPrTimeBlw1 bucket for pre processing time below 1ms
	statsKeyPrTimeBlw1

	//statsKeyBannerImpDisabledViaConfig stats Key for Counting banner impressions that are disabled via config for a publisher/profile
	statsKeyBannerImpDisabledViaConfig

	// ********************* CTV Stats *********************

	//statsKeyCTVPrebidFailedImpression for counting number of CTV prebid side failed impressions
	statsKeyCTVPrebidFailedImpression
	//statsKeyCTVRequests for counting number of CTV  Requests
	statsKeyCTVRequests
	//statsKeyCTVBadRequests for counting number of CTV  Bad Requests
	statsKeyCTVBadRequests
	//statsKeyCTVPublisherRequests for counting number of CTV  Publisher Requests
	statsKeyCTVPublisherRequests
	//statsKeyCTVHTTPMethodRequests for counting number of CTV  Publisher GET/POST Requests
	statsKeyCTVHTTPMethodRequests
	//statsKeyCTVValidationDetail for tracking error with granularity
	statsKeyCTVValidationErr
	//statsKeyIncompleteAdPods for tracking incomplete AdPods because of any reason
	statsKeyIncompleteAdPods
	//statsKeyCTVReqImpstWithConfig for tracking requests that had config and were not overwritten by database config
	statsKeyCTVReqImpstWithConfig
	//statsKeyTotalAdPodImpression for tracking no of AdPod impressions
	statsKeyTotalAdPodImpression
	//statsKeyAdPodSecondsMissed for tracking no pf seconds that were missed because of our algos
	statsKeyReqTotalAdPodImpression
	//statsKeyReqAdPodSecondsMissed for tracking no pf seconds that were missed because of our algos
	statsKeyAdPodSecondsMissed
	//statsKeyReqImpDurationYield is for tracking the number on adpod impressions generated for give min and max request imp durations
	statsKeyReqImpDurationYield
	//statsKeyReqWithAdPodCount if for counting requests with AdPods
	statsKeyReqWithAdPodCount
	//statsKeyBidDuration for counting number of bids of video duration
	statsKeyBidDuration

	//statsKeyPBSAuctionRequests stats Key for counting PBS Auction endpoint Requests
	statsKeyPBSAuctionRequests

	//statsKeyInjectTrackerErrorCount stats key for counting error during injecting tracker in Creative
	statsKeyInjectTrackerErrorCount

	//statsBidResponsesByDealUsingPBS stats key for counting number of bids received which for given deal id, profile id, publisherid
	statsBidResponsesByDealUsingPBS

	//statsBidResponsesByDealUsingHB stats key for counting number of bids received which for given deal id, profile id, publisherid
	statsBidResponsesByDealUsingHB

	// statsPartnerTimeoutInPBS stats key for countiing number of timeouts occured for given publisher and profile
	statsPartnerTimeoutInPBS

	// This is to declare the array of stats, add new stats above this
	maxNumOfStats
	// NOTE - DON'T ADD NEW STATS KEY BELOW THIS. NEW STATS SHOULD BE ADDED ABOVE maxNumOfStats
)

// constant to defines status-code used while sending stats to server
const (
	statusSetupFail = iota
	statusPublishSuccess
	statusPublishFail
)
