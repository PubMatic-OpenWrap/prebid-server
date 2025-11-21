package ctv

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
)

// KeyParserMap map type which contains standard key parser functions
type KeyParserMap map[string]func(Parser) error

// ExtParserMap map type which contains extension parameter parser functions
type ExtParserMap map[string]func(Parser, string, *string) error

// IgnoreList map type which contains list of keys to ignore
type IgnoreList map[string]struct{}

// ParserMap contains standard, extensions, ignorelist parmeters parser functions
type ParserMap struct {
	KeyMapping KeyParserMap
	ExtMapping ExtParserMap
	IgnoreList IgnoreList
}

// ortbMapper is ParserMap for ortb parameters
var ortbMapper = &ParserMap{
	KeyMapping: KeyParserMap{
		//BidRequest
		ORTBBidRequestID:      Parser.ORTBBidRequestID,
		ORTBBidRequestTest:    Parser.ORTBBidRequestTest,
		ORTBBidRequestAt:      Parser.ORTBBidRequestAt,
		ORTBBidRequestTmax:    Parser.ORTBBidRequestTmax,
		ORTBBidRequestWseat:   Parser.ORTBBidRequestWseat,
		ORTBBidRequestWlang:   Parser.ORTBBidRequestWlang,
		ORTBBidRequestBseat:   Parser.ORTBBidRequestBseat,
		ORTBBidRequestAllImps: Parser.ORTBBidRequestAllImps,
		ORTBBidRequestCur:     Parser.ORTBBidRequestCur,
		ORTBBidRequestBcat:    Parser.ORTBBidRequestBcat,
		ORTBBidRequestBadv:    Parser.ORTBBidRequestBadv,
		ORTBBidRequestBapp:    Parser.ORTBBidRequestBapp,

		//Source
		ORTBSourceFD:     Parser.ORTBSourceFD,
		ORTBSourceTID:    Parser.ORTBSourceTID,
		ORTBSourcePChain: Parser.ORTBSourcePChain,
		ORTBSourceSChain: Parser.ORTBSourceSChain,

		//Site
		ORTBSiteID:            Parser.ORTBSiteID,
		ORTBSiteName:          Parser.ORTBSiteName,
		ORTBSiteDomain:        Parser.ORTBSiteDomain,
		ORTBSitePage:          Parser.ORTBSitePage,
		ORTBSiteRef:           Parser.ORTBSiteRef,
		ORTBSiteSearch:        Parser.ORTBSiteSearch,
		ORTBSiteMobile:        Parser.ORTBSiteMobile,
		ORTBSiteCat:           Parser.ORTBSiteCat,
		ORTBSiteSectionCat:    Parser.ORTBSiteSectionCat,
		ORTBSitePageCat:       Parser.ORTBSitePageCat,
		ORTBSitePrivacyPolicy: Parser.ORTBSitePrivacyPolicy,
		ORTBSiteKeywords:      Parser.ORTBSiteKeywords,

		//Site.Publisher
		ORTBSitePublisherID:     Parser.ORTBSitePublisherID,
		ORTBSitePublisherName:   Parser.ORTBSitePublisherName,
		ORTBSitePublisherCat:    Parser.ORTBSitePublisherCat,
		ORTBSitePublisherDomain: Parser.ORTBSitePublisherDomain,

		//Site.Content
		ORTBSiteContentID:                 Parser.ORTBSiteContentID,
		ORTBSiteContentEpisode:            Parser.ORTBSiteContentEpisode,
		ORTBSiteContentTitle:              Parser.ORTBSiteContentTitle,
		ORTBSiteContentSeries:             Parser.ORTBSiteContentSeries,
		ORTBSiteContentSeason:             Parser.ORTBSiteContentSeason,
		ORTBSiteContentArtist:             Parser.ORTBSiteContentArtist,
		ORTBSiteContentGenre:              Parser.ORTBSiteContentGenre,
		ORTBSiteContentAlbum:              Parser.ORTBSiteContentAlbum,
		ORTBSiteContentIsRc:               Parser.ORTBSiteContentIsRc,
		ORTBSiteContentURL:                Parser.ORTBSiteContentURL,
		ORTBSiteContentCat:                Parser.ORTBSiteContentCat,
		ORTBSiteContentProdQ:              Parser.ORTBSiteContentProdQ,
		ORTBSiteContentVideoQuality:       Parser.ORTBSiteContentVideoQuality,
		ORTBSiteContentContext:            Parser.ORTBSiteContentContext,
		ORTBSiteContentContentRating:      Parser.ORTBSiteContentContentRating,
		ORTBSiteContentUserRating:         Parser.ORTBSiteContentUserRating,
		ORTBSiteContentQaGmeDiarating:     Parser.ORTBSiteContentQaGmeDiarating,
		ORTBSiteContentKeywords:           Parser.ORTBSiteContentKeywords,
		ORTBSiteContentLiveStream:         Parser.ORTBSiteContentLiveStream,
		ORTBSiteContentSourceRelationship: Parser.ORTBSiteContentSourceRelationship,
		ORTBSiteContentLen:                Parser.ORTBSiteContentLen,
		ORTBSiteContentLanguage:           Parser.ORTBSiteContentLanguage,
		ORTBSiteContentEmbeddable:         Parser.ORTBSiteContentEmbeddable,

		//Site.Content.Network
		ORTBSiteContentNetworkID:     Parser.ORTBSiteContentNetworkID,
		ORTBSiteContentNetworkName:   Parser.ORTBSiteContentNetworkName,
		ORTBSiteContentNetworkDomain: Parser.ORTBSiteContentNetworkDomain,

		//Site.Content.Channel
		ORTBSiteContentChannelID:     Parser.ORTBSiteContentChannelID,
		ORTBSiteContentChannelName:   Parser.ORTBSiteContentChannelName,
		ORTBSiteContentChannelDomain: Parser.ORTBSiteContentChannelDomain,

		//Site.Content.Producer
		ORTBSiteContentProducerID:     Parser.ORTBSiteContentProducerID,
		ORTBSiteContentProducerName:   Parser.ORTBSiteContentProducerName,
		ORTBSiteContentProducerCat:    Parser.ORTBSiteContentProducerCat,
		ORTBSiteContentProducerDomain: Parser.ORTBSiteContentProducerDomain,

		//App
		ORTBAppID:            Parser.ORTBAppID,
		ORTBAppName:          Parser.ORTBAppName,
		ORTBAppBundle:        Parser.ORTBAppBundle,
		ORTBAppDomain:        Parser.ORTBAppDomain,
		ORTBAppStoreURL:      Parser.ORTBAppStoreURL,
		ORTBAppVer:           Parser.ORTBAppVer,
		ORTBAppPaid:          Parser.ORTBAppPaid,
		ORTBAppCat:           Parser.ORTBAppCat,
		ORTBAppSectionCat:    Parser.ORTBAppSectionCat,
		ORTBAppPageCat:       Parser.ORTBAppPageCat,
		ORTBAppPrivacyPolicy: Parser.ORTBAppPrivacyPolicy,
		ORTBAppKeywords:      Parser.ORTBAppKeywords,

		//App.Publisher
		ORTBAppPublisherID:     Parser.ORTBAppPublisherID,
		ORTBAppPublisherName:   Parser.ORTBAppPublisherName,
		ORTBAppPublisherCat:    Parser.ORTBAppPublisherCat,
		ORTBAppPublisherDomain: Parser.ORTBAppPublisherDomain,

		//App.Content
		ORTBAppContentID:                 Parser.ORTBAppContentID,
		ORTBAppContentEpisode:            Parser.ORTBAppContentEpisode,
		ORTBAppContentTitle:              Parser.ORTBAppContentTitle,
		ORTBAppContentSeries:             Parser.ORTBAppContentSeries,
		ORTBAppContentSeason:             Parser.ORTBAppContentSeason,
		ORTBAppContentArtist:             Parser.ORTBAppContentArtist,
		ORTBAppContentGenre:              Parser.ORTBAppContentGenre,
		ORTBAppContentAlbum:              Parser.ORTBAppContentAlbum,
		ORTBAppContentIsRc:               Parser.ORTBAppContentIsRc,
		ORTBAppContentURL:                Parser.ORTBAppContentURL,
		ORTBAppContentCat:                Parser.ORTBAppContentCat,
		ORTBAppContentProdQ:              Parser.ORTBAppContentProdQ,
		ORTBAppContentVideoQuality:       Parser.ORTBAppContentVideoQuality,
		ORTBAppContentContext:            Parser.ORTBAppContentContext,
		ORTBAppContentContentRating:      Parser.ORTBAppContentContentRating,
		ORTBAppContentUserRating:         Parser.ORTBAppContentUserRating,
		ORTBAppContentQaGmeDiarating:     Parser.ORTBAppContentQaGmeDiarating,
		ORTBAppContentKeywords:           Parser.ORTBAppContentKeywords,
		ORTBAppContentLiveStream:         Parser.ORTBAppContentLiveStream,
		ORTBAppContentSourceRelationship: Parser.ORTBAppContentSourceRelationship,
		ORTBAppContentLen:                Parser.ORTBAppContentLen,
		ORTBAppContentLanguage:           Parser.ORTBAppContentLanguage,
		ORTBAppContentEmbeddable:         Parser.ORTBAppContentEmbeddable,

		//App.Content.Network
		ORTBAppContentNetworkID:     Parser.ORTBAppContentNetworkID,
		ORTBAppContentNetworkName:   Parser.ORTBAppContentNetworkName,
		ORTBAppContentNetworkDomain: Parser.ORTBAppContentNetworkDomain,

		//App.Content.Channel
		ORTBAppContentChannelID:     Parser.ORTBAppContentChannelID,
		ORTBAppContentChannelName:   Parser.ORTBAppContentChannelName,
		ORTBAppContentChannelDomain: Parser.ORTBAppContentChannelDomain,

		//App.Content.Producer
		ORTBAppContentProducerID:     Parser.ORTBAppContentProducerID,
		ORTBAppContentProducerName:   Parser.ORTBAppContentProducerName,
		ORTBAppContentProducerCat:    Parser.ORTBAppContentProducerCat,
		ORTBAppContentProducerDomain: Parser.ORTBAppContentProducerDomain,

		//Video
		ORTBImpVideoMimes:          Parser.ORTBImpVideoMimes,
		ORTBImpVideoMinDuration:    Parser.ORTBImpVideoMinDuration,
		ORTBImpVideoMaxDuration:    Parser.ORTBImpVideoMaxDuration,
		ORTBImpVideoProtocols:      Parser.ORTBImpVideoProtocols,
		ORTBImpVideoPlayerWidth:    Parser.ORTBImpVideoPlayerWidth,
		ORTBImpVideoPlayerHeight:   Parser.ORTBImpVideoPlayerHeight,
		ORTBImpVideoStartDelay:     Parser.ORTBImpVideoStartDelay,
		ORTBImpVideoPlacement:      Parser.ORTBImpVideoPlacement,
		ORTBImpVideoPlcmt:          Parser.ORTBImpVideoPlcmt,
		ORTBImpVideoLinearity:      Parser.ORTBImpVideoLinearity,
		ORTBImpVideoSkip:           Parser.ORTBImpVideoSkip,
		ORTBImpVideoSkipMin:        Parser.ORTBImpVideoSkipMin,
		ORTBImpVideoSkipAfter:      Parser.ORTBImpVideoSkipAfter,
		ORTBImpVideoSequence:       Parser.ORTBImpVideoSequence,
		ORTBImpVideoBAttr:          Parser.ORTBImpVideoBAttr,
		ORTBImpVideoMaxExtended:    Parser.ORTBImpVideoMaxExtended,
		ORTBImpVideoMinBitrate:     Parser.ORTBImpVideoMinBitrate,
		ORTBImpVideoMaxBitrate:     Parser.ORTBImpVideoMaxBitrate,
		ORTBImpVideoBoxingAllowed:  Parser.ORTBImpVideoBoxingAllowed,
		ORTBImpVideoPlaybackMethod: Parser.ORTBImpVideoPlaybackMethod,
		ORTBImpVideoDelivery:       Parser.ORTBImpVideoDelivery,
		ORTBImpVideoPos:            Parser.ORTBImpVideoPos,
		ORTBImpVideoAPI:            Parser.ORTBImpVideoAPI,
		ORTBImpVideoCompanionType:  Parser.ORTBImpVideoCompanionType,

		//Regs
		ORTBRegsCoppa:        Parser.ORTBRegsCoppa,
		ORTBRegsGpp:          Parser.ORTBRegsGpp,
		ORTBRegsGppSid:       Parser.ORTBRegsGppSid,
		ORTBRegsExtGdpr:      Parser.ORTBRegsExtGdpr,
		ORTBRegsExtUSPrivacy: Parser.ORTBRegsExtUSPrivacy,

		//Imp
		ORTBImpID:                Parser.ORTBImpID,
		ORTBImpDisplayManager:    Parser.ORTBImpDisplayManager,
		ORTBImpDisplayManagerVer: Parser.ORTBImpDisplayManagerVer,
		ORTBImpInstl:             Parser.ORTBImpInstl,
		ORTBImpTagID:             Parser.ORTBImpTagID,
		ORTBImpBidFloor:          Parser.ORTBImpBidFloor,
		ORTBImpBidFloorCur:       Parser.ORTBImpBidFloorCur,
		ORTBImpClickBrowser:      Parser.ORTBImpClickBrowser,
		ORTBImpSecure:            Parser.ORTBImpSecure,
		ORTBImpIframeBuster:      Parser.ORTBImpIframeBuster,
		ORTBImpExp:               Parser.ORTBImpExp,
		ORTBImpPmp:               Parser.ORTBImpPmp,
		ORTBImpExtBidder:         Parser.ORTBImpExtBidder,
		ORTBImpExtPrebid:         Parser.ORTBImpExtPrebid,

		//Device Functions
		ORTBDeviceUserAgent:      Parser.ORTBDeviceUserAgent,
		ORTBDeviceIP:             Parser.ORTBDeviceIP,
		ORTBDeviceIpv6:           Parser.ORTBDeviceIpv6,
		ORTBDeviceDnt:            Parser.ORTBDeviceDnt,
		ORTBDeviceLmt:            Parser.ORTBDeviceLmt,
		ORTBDeviceDeviceType:     Parser.ORTBDeviceDeviceType,
		ORTBDeviceMake:           Parser.ORTBDeviceMake,
		ORTBDeviceModel:          Parser.ORTBDeviceModel,
		ORTBDeviceOs:             Parser.ORTBDeviceOs,
		ORTBDeviceOsv:            Parser.ORTBDeviceOsv,
		ORTBDeviceHwv:            Parser.ORTBDeviceHwv,
		ORTBDeviceWidth:          Parser.ORTBDeviceWidth,
		ORTBDeviceHeight:         Parser.ORTBDeviceHeight,
		ORTBDevicePpi:            Parser.ORTBDevicePpi,
		ORTBDevicePxRatio:        Parser.ORTBDevicePxRatio,
		ORTBDeviceJS:             Parser.ORTBDeviceJS,
		ORTBDeviceGeoFetch:       Parser.ORTBDeviceGeoFetch,
		ORTBDeviceFlashVer:       Parser.ORTBDeviceFlashVer,
		ORTBDeviceLanguage:       Parser.ORTBDeviceLanguage,
		ORTBDeviceCarrier:        Parser.ORTBDeviceCarrier,
		ORTBDeviceMccmnc:         Parser.ORTBDeviceMccmnc,
		ORTBDeviceConnectionType: Parser.ORTBDeviceConnectionType,
		ORTBDeviceIfa:            Parser.ORTBDeviceIfa,
		ORTBDeviceDidSha1:        Parser.ORTBDeviceDidSha1,
		ORTBDeviceDidMd5:         Parser.ORTBDeviceDidMd5,
		ORTBDeviceDpidSha1:       Parser.ORTBDeviceDpidSha1,
		ORTBDeviceDpidMd5:        Parser.ORTBDeviceDpidMd5,
		ORTBDeviceMacSha1:        Parser.ORTBDeviceMacSha1,
		ORTBDeviceMacMd5:         Parser.ORTBDeviceMacMd5,

		//Device.Geo
		ORTBDeviceGeoLat:           Parser.ORTBDeviceGeoLat,
		ORTBDeviceGeoLon:           Parser.ORTBDeviceGeoLon,
		ORTBDeviceGeoType:          Parser.ORTBDeviceGeoType,
		ORTBDeviceGeoAccuracy:      Parser.ORTBDeviceGeoAccuracy,
		ORTBDeviceGeoLastFix:       Parser.ORTBDeviceGeoLastFix,
		ORTBDeviceGeoIPService:     Parser.ORTBDeviceGeoIPService,
		ORTBDeviceGeoCountry:       Parser.ORTBDeviceGeoCountry,
		ORTBDeviceGeoRegion:        Parser.ORTBDeviceGeoRegion,
		ORTBDeviceGeoRegionFips104: Parser.ORTBDeviceGeoRegionFips104,
		ORTBDeviceGeoMetro:         Parser.ORTBDeviceGeoMetro,
		ORTBDeviceGeoCity:          Parser.ORTBDeviceGeoCity,
		ORTBDeviceGeoZip:           Parser.ORTBDeviceGeoZip,
		ORTBDeviceGeoUtcOffset:     Parser.ORTBDeviceGeoUtcOffset,

		//Device.Ext.IfaType
		ORTBDeviceExtIfaType:   Parser.ORTBDeviceExtIfaType,
		ORTBDeviceExtSessionID: Parser.ORTBDeviceExtSessionID,
		ORTBDeviceExtATTS:      Parser.ORTBDeviceExtATTS,

		//User
		ORTBUserID:         Parser.ORTBUserID,
		ORTBUserBuyerUID:   Parser.ORTBUserBuyerUID,
		ORTBUserYob:        Parser.ORTBUserYob,
		ORTBUserGender:     Parser.ORTBUserGender,
		ORTBUserKeywords:   Parser.ORTBUserKeywords,
		ORTBUserCustomData: Parser.ORTBUserCustomData,

		//User.Ext.Consent
		ORTBUserExtConsent: Parser.ORTBUserExtConsent,
		//User.Ext.EIDS
		ORTBUserExtEIDS: Parser.ORTBUserExtEIDS,
		//User.Ext.SessionDuration
		ORTBUserExtSessionDuration: Parser.ORTBUserExtSessionDuration,
		//User.Ext.ImpDepth
		ORTBUserExtImpDepth: Parser.ORTBUserExtImpDepth,
		//User.Data
		ORTBUserData: Parser.ORTBUserData,

		//User.Geo
		ORTBUserGeoLat:           Parser.ORTBUserGeoLat,
		ORTBUserGeoLon:           Parser.ORTBUserGeoLon,
		ORTBUserGeoType:          Parser.ORTBUserGeoType,
		ORTBUserGeoAccuracy:      Parser.ORTBUserGeoAccuracy,
		ORTBUserGeoLastFix:       Parser.ORTBUserGeoLastFix,
		ORTBUserGeoIPService:     Parser.ORTBUserGeoIPService,
		ORTBUserGeoCountry:       Parser.ORTBUserGeoCountry,
		ORTBUserGeoRegion:        Parser.ORTBUserGeoRegion,
		ORTBUserGeoRegionFips104: Parser.ORTBUserGeoRegionFips104,
		ORTBUserGeoMetro:         Parser.ORTBUserGeoMetro,
		ORTBUserGeoCity:          Parser.ORTBUserGeoCity,
		ORTBUserGeoZip:           Parser.ORTBUserGeoZip,
		ORTBUserGeoUtcOffset:     Parser.ORTBUserGeoUtcOffset,

		//ReqWrapperExtension
		ORTBProfileID:            Parser.ORTBProfileID,
		ORTBVersionID:            Parser.ORTBVersionID,
		ORTBSSAuctionFlag:        Parser.ORTBSSAuctionFlag,
		ORTBSumryDisableFlag:     Parser.ORTBSumryDisableFlag,
		ORTBClientConfigFlag:     Parser.ORTBClientConfigFlag,
		ORTBSupportDeals:         Parser.ORTBSupportDeals,
		ORTBIncludeBrandCategory: Parser.ORTBIncludeBrandCategory,
		ORTBSSAI:                 Parser.ORTBSSAI,
		ORTBKeyValues:            Parser.ORTBKeyValues,
		ORTBKeyValuesMap:         Parser.ORTBKeyValuesMap,

		//VideoExtension
		ORTBImpVideoExtOffset:                           Parser.ORTBImpVideoExtOffset,
		ORTBImpVideoExtAdPodMinAds:                      Parser.ORTBImpVideoExtAdPodMinAds,
		ORTBImpVideoExtAdPodMaxAds:                      Parser.ORTBImpVideoExtAdPodMaxAds,
		ORTBImpVideoExtAdPodMinDuration:                 Parser.ORTBImpVideoExtAdPodMinDuration,
		ORTBImpVideoExtAdPodMaxDuration:                 Parser.ORTBImpVideoExtAdPodMaxDuration,
		ORTBImpVideoExtAdPodAdvertiserExclusionPercent:  Parser.ORTBImpVideoExtAdPodAdvertiserExclusionPercent,
		ORTBImpVideoExtAdPodIABCategoryExclusionPercent: Parser.ORTBImpVideoExtAdPodIABCategoryExclusionPercent,

		//ReqAdPodExt
		ORTBRequestExtAdPodMinAds:                              Parser.ORTBRequestExtAdPodMinAds,
		ORTBRequestExtAdPodMaxAds:                              Parser.ORTBRequestExtAdPodMaxAds,
		ORTBRequestExtAdPodMinDuration:                         Parser.ORTBRequestExtAdPodMinDuration,
		ORTBRequestExtAdPodMaxDuration:                         Parser.ORTBRequestExtAdPodMaxDuration,
		ORTBRequestExtAdPodAdvertiserExclusionPercent:          Parser.ORTBRequestExtAdPodAdvertiserExclusionPercent,
		ORTBRequestExtAdPodIABCategoryExclusionPercent:         Parser.ORTBRequestExtAdPodIABCategoryExclusionPercent,
		ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent:  Parser.ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent,
		ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent: Parser.ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent,
		ORTBRequestExtAdPodIABCategoryExclusionWindow:          Parser.ORTBRequestExtAdPodIABCategoryExclusionWindow,
		ORTBRequestExtAdPodAdvertiserExclusionWindow:           Parser.ORTBRequestExtAdPodAdvertiserExclusionWindow,

		//ReqPrebidExt
		ORTBRequestExtPrebidTransparencyContent: Parser.ORTBRequestExtPrebidTransparencyContent,
		ORTBExtPrebidFloorsEnforcement:          Parser.ORTBExtPrebidFloorsEnforceFloorDeals,
		ORTBExtPrebidReturnAllBidStatus:         Parser.ORTBExtPrebidReturnAllBidStatus,
		ORTBExtPrebidBidderParamsPubmaticCDS:    Parser.ORTBExtPrebidBidderParamsPubmaticCDS,
	},
	ExtMapping: ExtParserMap{
		//Extensions
		ORTBBidRequestExt:          Parser.ORTBBidRequestExt,
		ORTBSourceExt:              Parser.ORTBSourceExt,
		ORTBRegsExt:                Parser.ORTBRegsExt,
		ORTBImpExt:                 Parser.ORTBImpExt,
		ORTBImpVideoExt:            Parser.ORTBImpVideoExt,
		ORTBSiteExt:                Parser.ORTBSiteExt,
		ORTBSiteContentNetworkExt:  Parser.ORTBSiteContentNetworkExt,
		ORTBSiteContentChannelExt:  Parser.ORTBSiteContentChannelExt,
		ORTBAppExt:                 Parser.ORTBAppExt,
		ORTBAppContentNetworkExt:   Parser.ORTBAppContentNetworkExt,
		ORTBAppContentChannelExt:   Parser.ORTBAppContentChannelExt,
		ORTBSitePublisherExt:       Parser.ORTBSitePublisherExt,
		ORTBSiteContentExt:         Parser.ORTBSiteContentExt,
		ORTBSiteContentProducerExt: Parser.ORTBSiteContentProducerExt,
		ORTBAppPublisherExt:        Parser.ORTBAppPublisherExt,
		ORTBAppContentExt:          Parser.ORTBAppContentExt,
		ORTBAppContentProducerExt:  Parser.ORTBAppContentProducerExt,
		ORTBDeviceExt:              Parser.ORTBDeviceExt,
		ORTBDeviceGeoExt:           Parser.ORTBDeviceGeoExt,
		ORTBUserExt:                Parser.ORTBUserExt,
		ORTBUserGeoExt:             Parser.ORTBUserGeoExt,
	},
	IgnoreList: IgnoreList{
		Debug: struct{}{},
	},
}

// GetORTBParserMap TODO
func GetORTBParserMap() *ParserMap {
	return ortbMapper
}

// ORTBParser interface which will be used to generate ortb request from API request
type ORTBParser interface {
	ParseORTBRequest(*ParserMap) (*openrtb2.BidRequest, error)
}

type ParseError struct {
	Code    *openrtb3.NoBidReason
	Message string
}

func NewParseError(code openrtb3.NoBidReason, message string) *ParseError {
	return &ParseError{
		Code:    &code,
		Message: message,
	}
}

func (e *ParseError) Error() string {
	return e.Message
}

func (e *ParseError) NBR() *openrtb3.NoBidReason {
	return e.Code
}

type Parser interface {
	ORTBParser

	//BidRequest
	ORTBBidRequestID() error
	ORTBBidRequestTest() error
	ORTBBidRequestAt() error
	ORTBBidRequestTmax() error
	ORTBBidRequestWseat() error
	ORTBBidRequestAllImps() error
	ORTBBidRequestCur() error
	ORTBBidRequestBcat() error
	ORTBBidRequestBadv() error
	ORTBBidRequestBapp() error
	ORTBBidRequestWlang() error
	ORTBBidRequestBseat() error

	//Source
	ORTBSourceFD() error
	ORTBSourceTID() error
	ORTBSourcePChain() error
	ORTBSourceSChain() error

	//Site
	ORTBSiteID() error
	ORTBSiteName() error
	ORTBSiteDomain() error
	ORTBSitePage() error
	ORTBSiteRef() error
	ORTBSiteSearch() error
	ORTBSiteMobile() error
	ORTBSiteCat() error
	ORTBSiteSectionCat() error
	ORTBSitePageCat() error
	ORTBSitePrivacyPolicy() error
	ORTBSiteKeywords() error

	//Site.Publisher
	ORTBSitePublisherID() error
	ORTBSitePublisherName() error
	ORTBSitePublisherCat() error
	ORTBSitePublisherDomain() error

	//Site.Content
	ORTBSiteContentID() error
	ORTBSiteContentEpisode() error
	ORTBSiteContentTitle() error
	ORTBSiteContentSeries() error
	ORTBSiteContentSeason() error
	ORTBSiteContentArtist() error
	ORTBSiteContentGenre() error
	ORTBSiteContentAlbum() error
	ORTBSiteContentIsRc() error
	ORTBSiteContentURL() error
	ORTBSiteContentCat() error
	ORTBSiteContentProdQ() error
	ORTBSiteContentVideoQuality() error
	ORTBSiteContentContext() error
	ORTBSiteContentContentRating() error
	ORTBSiteContentUserRating() error
	ORTBSiteContentQaGmeDiarating() error
	ORTBSiteContentKeywords() error
	ORTBSiteContentLiveStream() error
	ORTBSiteContentSourceRelationship() error
	ORTBSiteContentLen() error
	ORTBSiteContentLanguage() error
	ORTBSiteContentEmbeddable() error

	//Site.Content.Network
	ORTBSiteContentNetworkID() error
	ORTBSiteContentNetworkName() error
	ORTBSiteContentNetworkDomain() error

	//Site.Content.Channel
	ORTBSiteContentChannelID() error
	ORTBSiteContentChannelName() error
	ORTBSiteContentChannelDomain() error

	//Site.Content.Producer
	ORTBSiteContentProducerID() error
	ORTBSiteContentProducerName() error
	ORTBSiteContentProducerCat() error
	ORTBSiteContentProducerDomain() error

	//App
	ORTBAppID() error
	ORTBAppName() error
	ORTBAppBundle() error
	ORTBAppDomain() error
	ORTBAppStoreURL() error
	ORTBAppVer() error
	ORTBAppPaid() error
	ORTBAppCat() error
	ORTBAppSectionCat() error
	ORTBAppPageCat() error
	ORTBAppPrivacyPolicy() error
	ORTBAppKeywords() error

	//App.Publisher
	ORTBAppPublisherID() error
	ORTBAppPublisherName() error
	ORTBAppPublisherCat() error
	ORTBAppPublisherDomain() error

	//App.Content
	ORTBAppContentID() error
	ORTBAppContentEpisode() error
	ORTBAppContentTitle() error
	ORTBAppContentSeries() error
	ORTBAppContentSeason() error
	ORTBAppContentArtist() error
	ORTBAppContentGenre() error
	ORTBAppContentAlbum() error
	ORTBAppContentIsRc() error
	ORTBAppContentURL() error
	ORTBAppContentCat() error
	ORTBAppContentProdQ() error
	ORTBAppContentVideoQuality() error
	ORTBAppContentContext() error
	ORTBAppContentContentRating() error
	ORTBAppContentUserRating() error
	ORTBAppContentQaGmeDiarating() error
	ORTBAppContentKeywords() error
	ORTBAppContentLiveStream() error
	ORTBAppContentSourceRelationship() error
	ORTBAppContentLen() error
	ORTBAppContentLanguage() error
	ORTBAppContentEmbeddable() error

	//App.Content.Network
	ORTBAppContentNetworkID() error
	ORTBAppContentNetworkName() error
	ORTBAppContentNetworkDomain() error

	//App.Content.Channel
	ORTBAppContentChannelID() error
	ORTBAppContentChannelName() error
	ORTBAppContentChannelDomain() error

	//App.Content.Producer
	ORTBAppContentProducerID() error
	ORTBAppContentProducerName() error
	ORTBAppContentProducerCat() error
	ORTBAppContentProducerDomain() error

	//Video
	ORTBImpVideoMimes() error
	ORTBImpVideoMinDuration() error
	ORTBImpVideoMaxDuration() error
	ORTBImpVideoProtocols() error
	ORTBImpVideoPlayerWidth() error
	ORTBImpVideoPlayerHeight() error
	ORTBImpVideoStartDelay() error
	ORTBImpVideoPlacement() error
	ORTBImpVideoPlcmt() error
	ORTBImpVideoLinearity() error
	ORTBImpVideoSkip() error
	ORTBImpVideoSkipMin() error
	ORTBImpVideoSkipAfter() error
	ORTBImpVideoSequence() error
	ORTBImpVideoBAttr() error
	ORTBImpVideoMaxExtended() error
	ORTBImpVideoMinBitrate() error
	ORTBImpVideoMaxBitrate() error
	ORTBImpVideoBoxingAllowed() error
	ORTBImpVideoPlaybackMethod() error
	ORTBImpVideoDelivery() error
	ORTBImpVideoPos() error
	ORTBImpVideoAPI() error
	ORTBImpVideoCompanionType() error

	//Regs
	ORTBRegsCoppa() error

	//Imp
	ORTBImpID() error
	ORTBImpDisplayManager() error
	ORTBImpDisplayManagerVer() error
	ORTBImpInstl() error
	ORTBImpTagID() error
	ORTBImpBidFloor() error
	ORTBImpBidFloorCur() error
	ORTBImpClickBrowser() error
	ORTBImpSecure() error
	ORTBImpIframeBuster() error
	ORTBImpExp() error
	ORTBImpPmp() error
	ORTBImpExtBidder() error
	ORTBImpExtPrebid() error

	//Device Functions
	ORTBDeviceUserAgent() error
	ORTBDeviceDnt() error
	ORTBDeviceLmt() error
	ORTBDeviceIP() error
	ORTBDeviceIpv6() error
	ORTBDeviceDeviceType() error
	ORTBDeviceMake() error
	ORTBDeviceModel() error
	ORTBDeviceOs() error
	ORTBDeviceOsv() error
	ORTBDeviceHwv() error
	ORTBDeviceWidth() error
	ORTBDeviceHeight() error
	ORTBDevicePpi() error
	ORTBDevicePxRatio() error
	ORTBDeviceJS() error
	ORTBDeviceGeoFetch() error
	ORTBDeviceFlashVer() error
	ORTBDeviceLanguage() error
	ORTBDeviceCarrier() error
	ORTBDeviceMccmnc() error
	ORTBDeviceConnectionType() error
	ORTBDeviceIfa() error
	ORTBDeviceDidSha1() error
	ORTBDeviceDidMd5() error
	ORTBDeviceDpidSha1() error
	ORTBDeviceDpidMd5() error
	ORTBDeviceMacSha1() error
	ORTBDeviceMacMd5() error

	//DeviceExtIfaType
	ORTBDeviceExtIfaType() error
	ORTBDeviceExtSessionID() error
	ORTBDeviceExtATTS() error

	//Device.Geo
	ORTBDeviceGeoLat() error
	ORTBDeviceGeoLon() error
	ORTBDeviceGeoType() error
	ORTBDeviceGeoAccuracy() error
	ORTBDeviceGeoLastFix() error
	ORTBDeviceGeoIPService() error
	ORTBDeviceGeoCountry() error
	ORTBDeviceGeoRegion() error
	ORTBDeviceGeoRegionFips104() error
	ORTBDeviceGeoMetro() error
	ORTBDeviceGeoCity() error
	ORTBDeviceGeoZip() error
	ORTBDeviceGeoUtcOffset() error

	//User
	ORTBUserID() error
	ORTBUserBuyerUID() error
	ORTBUserYob() error
	ORTBUserGender() error
	ORTBUserKeywords() error
	ORTBUserCustomData() error

	//User.Geo
	ORTBUserGeoLat() error
	ORTBUserGeoLon() error
	ORTBUserGeoType() error
	ORTBUserGeoAccuracy() error
	ORTBUserGeoLastFix() error
	ORTBUserGeoIPService() error
	ORTBUserGeoCountry() error
	ORTBUserGeoRegion() error
	ORTBUserGeoRegionFips104() error
	ORTBUserGeoMetro() error
	ORTBUserGeoCity() error
	ORTBUserGeoZip() error
	ORTBUserGeoUtcOffset() error

	//Regs.Ext.Gdpr
	ORTBRegsExtGdpr() error
	ORTBRegsExtUSPrivacy() error
	//Regs.Gpp
	ORTBRegsGpp() error
	ORTBRegsGppSid() error

	//User.Ext.Consent
	ORTBUserExtConsent() error

	//User.Ext.EIDS
	ORTBUserExtEIDS() error

	//User.Ext.SessionDuration
	ORTBUserExtSessionDuration() error

	//User.Ext.ImpDepth
	ORTBUserExtImpDepth() error

	//User.Data
	ORTBUserData() error

	//Req.Ext.Parameters
	ORTBProfileID() error
	ORTBVersionID() error
	ORTBSSAuctionFlag() error
	ORTBSumryDisableFlag() error
	ORTBClientConfigFlag() error
	ORTBSupportDeals() error
	ORTBIncludeBrandCategory() error
	ORTBSSAI() error
	ORTBKeyValues() error
	ORTBKeyValuesMap() error

	//VideoExtension
	ORTBImpVideoExtOffset() error
	ORTBImpVideoExtAdPodMinAds() error
	ORTBImpVideoExtAdPodMaxAds() error
	ORTBImpVideoExtAdPodMinDuration() error
	ORTBImpVideoExtAdPodMaxDuration() error
	ORTBImpVideoExtAdPodAdvertiserExclusionPercent() error
	ORTBImpVideoExtAdPodIABCategoryExclusionPercent() error

	//ReqAdPodExt
	ORTBRequestExtAdPodMinAds() error
	ORTBRequestExtAdPodMaxAds() error
	ORTBRequestExtAdPodMinDuration() error
	ORTBRequestExtAdPodMaxDuration() error
	ORTBRequestExtAdPodAdvertiserExclusionPercent() error
	ORTBRequestExtAdPodIABCategoryExclusionPercent() error
	ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent() error
	ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent() error
	ORTBRequestExtAdPodIABCategoryExclusionWindow() error
	ORTBRequestExtAdPodAdvertiserExclusionWindow() error

	//ReqPrebidExt
	ORTBRequestExtPrebidTransparencyContent() error
	ORTBExtPrebidFloorsEnforceFloorDeals() error
	ORTBExtPrebidReturnAllBidStatus() error
	ORTBExtPrebidBidderParamsPubmaticCDS() error

	//ORTB Extensions
	ORTBBidRequestExt(string, *string) error
	ORTBSourceExt(string, *string) error
	ORTBRegsExt(string, *string) error
	ORTBImpExt(string, *string) error
	ORTBImpVideoExt(string, *string) error
	ORTBSiteExt(string, *string) error
	ORTBSiteContentNetworkExt(string, *string) error
	ORTBSiteContentChannelExt(string, *string) error
	ORTBAppExt(string, *string) error
	ORTBAppContentNetworkExt(string, *string) error
	ORTBAppContentChannelExt(string, *string) error
	ORTBSitePublisherExt(string, *string) error
	ORTBSiteContentExt(string, *string) error
	ORTBSiteContentProducerExt(string, *string) error
	ORTBAppPublisherExt(string, *string) error
	ORTBAppContentExt(string, *string) error
	ORTBAppContentProducerExt(string, *string) error
	ORTBDeviceExt(string, *string) error
	ORTBDeviceGeoExt(string, *string) error
	ORTBUserExt(string, *string) error
	ORTBUserGeoExt(string, *string) error
}
