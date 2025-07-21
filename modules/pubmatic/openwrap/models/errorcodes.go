package models

import (
	"fmt"
)

type ErrorCode = int

// IError Interface for Custom Errors
type IError interface {
	error
	Code() ErrorCode
}

// Error Structure for Custom Errors
type Error struct {
	IError
	code    ErrorCode
	message string
}

// NewError New Object
func NewError(code ErrorCode, message string) IError {
	return &Error{code: code, message: message}
}

// Code Returns Error Code
func (e *Error) Code() ErrorCode {
	return e.code
}

// Error Returns Error Message
func (e *Error) Error() string {
	return e.message
}

const (
	ErrJSONMarshalFailed      = `error:[json_marshal_failed] object:[%s] message:[%s]`
	ErrJSONUnmarshalFailed    = `error:[json_unmarshal_failed] object:[%s] message:[%s] payload:[%s]`
	ErrTypeCastFailed         = `error:[type_cast_failed] key:[%s] type:[%s] value:[%v]`
	ErrHTTPNewRequestFailed   = `error:[setup_new_request_failed] method:[%s] endpoint:[%s] message:[%s]`
	ErrRowScanFailed          = `error:[db_row_scan_failed] name:[%s] error:[%s]`
	ErrQueryParsingFailed     = `error:[query_parsing_failed object:[%s] message:[%s] payload:[%s]`
	ErrDeserializationFailed  = `error:[deserialization_failed] object:[%s] message:[%s]  pubid:[%s] payload:[%s]`
	ErrSchainValidationFailed = `error:[schain_validation_failed] object:[%s] message:[%s] pubid:[%s] profileid:[%d] payload:[%s]`
)

var (
	ErrSuccess                        = NewError(0, "")
	ErrGADSUnmappedSlot               = NewError(1, "Slot not mapped")
	ErrGADSMissingConfig              = NewError(2, "Missing Configuration")
	ErrTimeout                        = NewError(3, "Timeout Error")
	ErrNoBidPrebid                    = NewError(4, "No Bid")
	ErrPartnerTimeout                 = NewError(5, "Partner Timed out")
	ErrInvalidConfiguration           = NewError(6, "Invalid Configuration")
	ErrNoGDPRConsent                  = NewError(7, "No Consent Present")
	ErrInvalidCreative                = NewError(8, "Invalid Creative")
	ErrCachePutFailed                 = NewError(9, "Cache PUT Failed")
	ErrInvalidParameter               = NewError(10, "Invalid Parameters")
	ErrAllPartnerThrottled            = NewError(11, "All partners throttled")
	ErrPartnerThrottled               = NewError(12, "Partner throttled")
	ErrBannerVideoDisabled            = NewError(13, "Banner/Video disabled through config")
	ErrPartnerContextDeadlineExceeded = NewError(14, "context deadline exceeded")
	ErrPrebidDefaultTimeout           = NewError(15, "Timed out")
	ErrInvalidHTTPRequestMethod       = NewError(16, "Invalid Request Method")
	ErrInvalidImpression              = NewError(17, "No Valid Impression Found")
	ErrBadRequest                     = NewError(18, "BadRequest")
	ErrPrebidUnknownError             = NewError(19, "Unknown error received from Prebid") // PrebidUnknownError is the Error code for Unknown Error returned from prebid-server
	ErrInvalidAdUnitConfig            = NewError(20, "Invalid adUnit config uploaded")
	ErrInvalidDealTierExt             = NewError(21, "Invalid deal tier info")
	ErrInvalidEndpoint                = NewError(22, "Invalid OpenWrap Endpoint")
	ErrBidderParamsValidationError    = NewError(23, "Bidder params passed are not valid as per schema definition")
	ErrPrebidCTVPreFilteringError     = NewError(24, "Prebid ctv adpod request failed to call partners")
	ErrPrebidCTVPostFilteringError    = NewError(25, "Prebid ctv adpod post bidding filtering")
	ErrReverseProxy                   = NewError(26, "Error Reverse Proxy")

	// ErrPrebidInvalidCustomPriceGranularity returnw new error with prebid core error message
	ErrPrebidInvalidCustomPriceGranularity = func(err error) IError {
		return NewError(26, fmt.Sprintf("Invalid Custom Price Granularity Config : [%s]", err.Error()))
	}
	//ORTB Validation Error Codes
	ErrInternalApp                                = NewError(500, "API Error")
	ErrMissingRequestID                           = NewError(600, "Missing Request ID")                                                                                                              //Missing Request ID
	ErrMissingImpressions                         = NewError(601, "Missing Impressions")                                                                                                             //Missing Impressions
	ErrMissingSiteApp                             = NewError(602, "Missing Site/App Object")                                                                                                         //Missing Site/App Object
	ErrSiteAppBothPresent                         = NewError(603, "Site App Both Present")                                                                                                           //Site App Both Present
	ErrMissingPublisherID                         = NewError(604, "Missing Publisher ID")                                                                                                            //Missing Publisher ID
	ErrMissingTagID                               = NewError(605, "Missing Tag ID")                                                                                                                  //Missing Tag ID
	ErrMissingImpressionID                        = NewError(606, "Missing Impression ID")                                                                                                           //Missing Impression ID
	ErrMissingAdType                              = NewError(607, "Missing Ad Type(Banner/Video)")                                                                                                   //Missing Ad Type(Banner/Video)
	ErrMissingAdSize                              = NewError(608, "Missing Ad Sizes")                                                                                                                //Missing Ad Sizes
	ErrMissingMIME                                = NewError(609, "Missing MIMEs")                                                                                                                   //Missing MIMEs
	ErrMissingNativeRequest                       = NewError(610, "Missing Native Request Payload")                                                                                                  //Missing Native Request Payload
	ErrMissingVideoObject                         = NewError(611, "Missing Video Object")                                                                                                            //Missing Video Object
	ErrInvalidRequestExtension                    = NewError(612, "Invalid Request Extension")                                                                                                       //Invalid Request Extension
	ErrInvalidCrossPodAdvertiserExclusionPercent  = NewError(613, "request.ext.adpod.crosspodexcladv must be a number between 0 and 100")                                                            //Invalid Value crosspodexcladv
	ErrInvalidCrossPodIABCategoryExclusionPercent = NewError(614, "request.ext.adpod.crosspodexcliabcat must be a number between 0 and 100")                                                         //Invalid Value crosspodexcliabcat
	ErrInvalidIABCategoryExclusionWindow          = NewError(615, "request.ext.adpod.excliabcatwindow must be postive number")                                                                       //Negative excliabcatwindow Value
	ErrInvalidAdvertiserExclusionWindow           = NewError(616, "request.ext.adpod.excladvwindow must be postive number")                                                                          //Negative excladvwindow Value
	ErrInvalidVideoExtension                      = NewError(617, "Invalid Video Extensions")                                                                                                        //Invalid Video Extensions
	ErrInvalidAdPodOffset                         = NewError(618, "imp.video.ext.offset must be postive number")                                                                                     //Negative Video AdPod offset Value
	ErrInvalidMinAds                              = NewError(619, "%key%.adpod.minads must be positive number")                                                                                      //Negative minads Value
	ErrInvalidMaxAds                              = NewError(620, "%key%.adpod.maxads must be positive number")                                                                                      //Negative maxads Value
	ErrInvalidMinDuration                         = NewError(621, "%key%.adpod.adminduration must be positive number")                                                                               //Negative minduration Value
	ErrInvalidMaxDuration                         = NewError(622, "%key%.adpod.admaxduration must be positive number")                                                                               //Negative maxduration Value
	ErrInvalidAdvertiserExclusionPercent          = NewError(623, "%key%.adpod.excladv must be number between 0 and 100")                                                                            //Invalid Value excladv
	ErrInvalidIABCategoryExclusionPercent         = NewError(624, "%key%.adpod.excliabcat must be number between 0 and 100")                                                                         //Invalid Value excliabcat
	ErrInvalidMinMaxAds                           = NewError(625, "%key%.adpod.minads must be less than %key%.adpod.maxads")                                                                         //minads greater than maxads
	ErrInvalidMinMaxDuration                      = NewError(626, "%key%.adpod.adminduration must be less than %key%.adpod.admaxduration")                                                           //minduration greater than maxduration
	ErrInvalidVideoMinDuration                    = NewError(627, "imp.video.minduration must be positive number")                                                                                   //negative minduration
	ErrInvalidVideoMaxDuration                    = NewError(628, "imp.video.maxduration must be positive number")                                                                                   //negative maxduration
	ErrInvalidVideoDurations                      = NewError(629, "imp.video.minduration must be less than imp.video.maxduration")                                                                   //video minduration greater than video maxduration
	ErrInvalidMinMaxDurationRange                 = NewError(630, "adpod duration checks for adminduration,admaxduration,minads,maxads are not in video minduration and maxduration duration range") //minmaxduration range failed
	ErrInvalidPublisherID                         = NewError(631, "Invalid Publisher ID")                                                                                                            //Invalid Publisher ID
	ErrInvalidVastTag                             = NewError(632, "Invalid vast tag configuration")                                                                                                  //Invalid Vast tag
	ErrInvalidAdServerURL                         = NewError(633, "Invalid adserver URL")
	ErrInvalidResponseFormat                      = NewError(634, "Invalid response format, must be 'json' or 'redirect'")
	ErrOWRedirectURLMissing                       = NewError(635, "OWRedirect URL is missing")
	ErrFetchAdpodConfig                           = NewError(636, "Failed to fetch adpod config")
	ErrMissingProfileID                           = NewError(700, "Missing Profile ID") //Missing Profile ID
	ErrCacheWarmup                                = NewError(701, "Cache Warmup")       //Cache Warmup
	ErrInvalidAdformat                            = NewError(702, "Invalid AdFormat")
	ErrMissingAdRuleConfigurations                = NewError(703, "Missing Adrule Configuration")
	ErrInvalidAdrule                              = NewError(704, "Invalid Adrule")
)
