package exchange

import "github.com/prebid/openrtb/v20/openrtb3"

// SeatNonBid list the reasons why bid was not resulted in positive bid
// reason could be either No bid, Error, Request rejection or Response rejection
// Reference:  https://github.com/InteractiveAdvertisingBureau/openrtb/blob/master/extensions/community_extensions/seat-non-bid.md
const (
	NoBidUnknownError                      openrtb3.NoBidReason = 0   // No Bid - General
	ResponseRejectedCategoryMappingInvalid openrtb3.NoBidReason = 303 // Response Rejected - Category Mapping Invalid
	ErrorGeneral                           openrtb3.NoBidReason = 100 // Error - General
	ErrorTimeout                           openrtb3.NoBidReason = 101 // Error - Timeout
	ErrorInvalidBidResponse                openrtb3.NoBidReason = 102 // Error - Invalid Bid Response
	ErrorBidderUnreachable                 openrtb3.NoBidReason = 103 // Error - Bidder Unreachable
)

const (
	RequestBlockedGeneral              openrtb3.NoBidReason = 200 // Request Blocked - General
	RequestBlockedUnsupportedChannel   openrtb3.NoBidReason = 201 // Request Blocked - Unsupported Channel (app/site/dooh)
	RequestBlockedUnsupportedMediaType openrtb3.NoBidReason = 202 // Request Blocked - Unsupported Media Type (banner/video/native/audio)
	RequestBlockedOptimized            openrtb3.NoBidReason = 203 // Request Blocked - Optimized
	RequestBlockedPrivacy              openrtb3.NoBidReason = 204 // Request Blocked - Privacy
)

const (
	ResponseRejectedGeneral                      openrtb3.NoBidReason = 300 // Response Rejected - General
	ResponseRejectedBelowFloor                   openrtb3.NoBidReason = 301 // Response Rejected - Below Floor
	ResponseRejectedDuplicateBid                 openrtb3.NoBidReason = 302 // Response Rejected - Duplicate
	ResponseRejectedInvalidCategoryMapping       openrtb3.NoBidReason = 303 // Response Rejected - Category Mapping Invalid
	ResponseRejectedBelowDealFloor               openrtb3.NoBidReason = 304 // Bid was Below Deal Floor
	ResponseRejectedInvalidCreative              openrtb3.NoBidReason = 350 // Response Rejected - Invalid Creative
	ResponseRejectedCreativeSizeNotAllowed       openrtb3.NoBidReason = 351 // Response Rejected - Invalid Creative (Size Not Allowed)
	ResponseRejectedCreativeNotSecure            openrtb3.NoBidReason = 352 // Response Rejected - Invalid Creative (Not Secure)
	ResponseRejectedCreativeIncorrectFormat      openrtb3.NoBidReason = 353 // Response Rejected - Invalid Creative (Incorrect Format)
	ResponseRejectedCreativeMalware              openrtb3.NoBidReason = 354 // Response Rejected - Invalid Creative (Malware)
	ResponseRejectedCreativeAdvertiserExclusions openrtb3.NoBidReason = 355 // Creative Filtered - Advertiser Exclusions
	ResponseRejectedCreativeAdvertiserBlocking   openrtb3.NoBidReason = 356 // Creative Filtered - Advertiser Blocking
	ResponseRejectedCreativeCategoryExclusions   openrtb3.NoBidReason = 357 // Creative Filtered - Category Exclusions
)
