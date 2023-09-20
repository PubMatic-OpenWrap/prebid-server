package models

import (
	"net/url"
)

type RequestAPI int

const (
	ADMIN_API RequestAPI = iota
	GADS_API
	OpenRTB_V23_API
	OpenRTB_V24_API
	OpenRTB_V241_API
	OpenRTB_V25_API
	OpenRTB_AMP_API
	OpenRTB_VIDEO_API
	OpenRTB_VIDEO_OPENRTB_API
	OpenRTB_VIDEO_VAST_API
	OpenRTB_VIDEO_JSON_API
)

const (
	SLOT_KEY                 = "slot"
	KEY_VALUE_KEY            = "keyValue"
	ID_KEY                   = "id"
	DIV_KEY                  = "div"
	SLOT_INDEX_KEY           = "slotIndex"
	PROFILE_KEY              = "profileid"
	GA_ID_KEY                = "gaId"
	SITE_ID                  = "siteId"
	ADTAG_ID                 = "adTagId"
	BID_REQUEST_ID           = "bidRequestId"
	IMPRESSION_ID            = "impId"
	DM_KEY                   = "dm"
	AS_KEY                   = "as"
	WRAPPER_KEY              = "wrapper"
	ADPOD_KEY                = "adpod"
	BIDDER_KEY               = "bidder"
	RESPONSE_TYPE_KEY        = "rs"
	PM_CB_KEY                = "pm_cb"
	SERVER_SIDE_AUCTION_FLAG = "ssauction"
	SUMMERY_DISABLE_FLAG     = "sumry_disable"
	KVAL_PARAM_KEY           = "kval_param"
	SA_VERSION_KEY           = "SAVersion"
	PAGE_URL_KEY             = "pageURL"
	REF_URL_KEY              = "refurl"
	IN_IFRAME_KEY            = "inIframe"
	KAD_PAGE_URL_KEY         = "kadpageurl"
	RAN_REQ_KEY              = "ranreq"
	KLT_STAMP_KEY            = "kltstamp"
	TIMEZONE_KEY             = "timezone"
	SCREEN_RESOLUTION_KEY    = "screenResolution"
	ADTYPE_KEY               = "adType"
	ADPOSITION_KEY           = "adPosition"
	ADVISIBILITY_KEY         = "adVisibility"
	IABCAT_KEY               = "iabcat"
	AWT_KEY                  = "awt"
	ZONEID_KEY               = "pmZoneId"
	SITECODE_KEY             = "sitecode"
	UDID_KEY                 = "udid"
	UDID_TYPE_KEY            = "udidtype"
	UDID_HASH_KEY            = "udidhash"
	ORMMA_KEY                = "ormma"
	AD_ORIENTATION_KEY       = "adOrientation"
	DEVICE_ORIENTATION_KEY   = "deviceOrientation"
	LOCCAT_KEY               = "loccat"
	LOCBRAND_KEY             = "locbrand"
	KADFLOOR_KEY             = "kadfloor"
	RID_KEY                  = "rid"
	LOC_SOURCE_KEY           = "loc_source"
	ETHN_KEY                 = "ethn"
	KEYWORDS_KEY             = "keywords"
	LOC_KEY                  = "loc"
	CAT_KEY                  = "cat"
	API_KEY                  = "api"
	NETTYPE_KEY              = "nettype"
	CONSENT                  = "consent"
	GET_METHOD_QUERY_PARAM   = "json"
	PAGE_URL_HEADER          = "Referer"
	SKAdnetworkKey           = "skadn"
	OmidpvKey                = "omidpv"
	OmidpnKey                = "omidpn"
	RewardKey                = "reward"
	DataKey                  = "data"
	DEAL_TIER_KEY            = "dealtier"
	FluidStr                 = "fluid"
	DeviceSessionID          = "session_id"
	DeviceIfaType            = "ifa_type"
)

type ResponseType int

const (
	ORTB_RESPONSE   ResponseType = 1 + iota //openRTB default response
	GADS_RESPONSE_1                         //gshow ad response
	GADS_RESPONSE_2                         //DM gpt generic response
)

const (
	// USD denotes currency USD
	USD = "USD"
)

// constants related to Video request
const (
	PlayerSizeKey         = "sz"
	SizeStringSeparator   = "x"
	DescriptionURLKey     = "description_url"
	URLKey                = "url"
	MimesSeparator        = ","
	MultipleSizeSeparator = "|"
	AppRequestURLKey      = "pwtapp"
	Comma                 = ","
)

// OpenWrap Video request params
const (
	OWMimeTypes      = "pwtmime"
	OWMinAdDuration  = "pwtvmnd"
	OWMaxAdDuration  = "pwtvmxd"
	OWStartDelay     = "pwtdly"
	OWPlaybackMethod = "pwtplbk"
	OWAPI            = "pwtvapi"
	OWProtocols      = "pwtprots"
	OWSize           = "pwtvsz"
	OWBAttr          = "pwtbatr"
	OWLinearity      = "pwtvlin"
	OWPlacement      = "pwtvplc"
	OWMaxBitrate     = "pwtmxbr"
	OWMinBitrate     = "pwtmnbr"
	OWSkippable      = "pwtskp"
	OWProtocol       = "pwtprot"
	OWSkipMin        = "pwtskmn"
	OWSkipAfter      = "pwtskat"
	OWSequence       = "pwtseq"
	OWMaxExtended    = "pwtmxex"
	OWDelivery       = "pwtdvry"
	OWPos            = "pwtvpos"
	OWBoxingAllowed  = "pwtbox"
	OWBidderParams   = "pwtbidrprm"
	OwAppKeywords    = "pwtappkw"
	OWUserEids       = "pwteids"
)

// OpenWrap Mobile params
const (
	OWAppId              = "pwtappid"
	OWAppName            = "pwtappname"
	OWAppDomain          = "pwtappdom"
	OWAppBundle          = "pwtappbdl"
	OWAppStoreURL        = "pwtappurl"
	OWAppCat             = "pwtappcat"
	OWAppPaid            = "pwtapppd"
	OWDeviceUA           = "pwtua"
	OWDeviceLMT          = "pwtlmt"
	OWDeviceDNT          = "pwtdnt"
	OWDeviceIP           = "pwtip"
	OWDeviceJS           = "pwtjs"
	OWDeviceIfa          = "pwtifa"
	OWDeviceDidsha1      = "pwtdidsha1"
	OWDeviceDidmd5       = "pwtdidmd5"
	OWDeviceDpidsha1     = "pwtdpidsha1"
	OWDeviceDpidmd5      = "pwtdpidmd5"
	OWDeviceMacsha1      = "pwtmacsha1"
	OWDeviceMacmd5       = "pwtmacmd5"
	OWUserID             = "pwtuid"
	OWGeoLat             = "pwtlat"
	OWGeoLon             = "pwtlon"
	OWGeoType            = "pwtgtype"
	OWGeoCountry         = "pwtcntr"
	OWGeoCity            = "pwtcity"
	OWGeoMetro           = "pwtmet"
	OWGeoZip             = "pwtzip"
	OWUTOffset           = "pwtuto"
	OWContentGenre       = "pwtgenre"
	OWContentTitle       = "pwttitle"
	OWUserYob            = "pwtyob"
	OWUserGender         = "pwtgender"
	OWSourceOmidPv       = "pwtomidpv"
	OWSourceOmidPn       = "pwtomidpn"
	OWDeviceExtIfaType   = "pwtifatype"
	OWDeviceExtSessionID = "pwtsessionid"
	OWImpPrebidExt       = "pwtimpprebidext"
)

// constants for DFP Video request parameters
const (
	DFPMinAdDuration = "min_ad_duration"
	DFPMaxAdDuration = "max_ad_duration"
	DFPSize          = PlayerSizeKey
	DFPVAdType       = "vad_type"
	DFPVPos          = "vpos"
	DFPVpmute        = "vpmute"
	DFPVpa           = "vpa"
)

// constants for oRTB Request Video parameters
const (
	MimeORTBParam           = "Mimes"
	MinDurationORTBParam    = "MinDuration"
	MaxDurationORTBParam    = "MaxDuration"
	ProtocolsORTBParam      = "Protocols"
	ProtocolORTBParam       = "Protocol"
	WORTBParam              = "W"
	HORTBParam              = "H"
	SizeORTBParam           = "sz"
	StartDelayORTBParam     = "StartDelay"
	PlacementORTBParam      = "Placement"
	LinearityORTBParam      = "Linearity"
	SkipORTBParam           = "Skip"
	SkipMinORTBParam        = "SkipMin"
	SkipAfterORTBParam      = "SkipAfter"
	SequenceORTBParam       = "Sequence"
	BAttrORTBParam          = "BAttr"
	MaxExtendedORTBParam    = "MaxExtended"
	MinBitrateORTBParam     = "MinBitrate"
	MaxBitrateORTBParam     = "MaxBitrate"
	BoxingAllowedORTBParam  = "BoxingAllowed"
	PlaybackMethodORTBParam = "PlaybackMethod"
	DeliveryORTBParam       = "Delivery"
	PosORTBParam            = "Pos"
	CompanionadORTBParam    = "Companionad"
	APIORTBParam            = "API"
	CompanionTypeORTBParam  = "CompanionType"
	AppIDORTBParam          = "AppID"
	AppNameORTBParam        = "AppName"
	AppBundleORTBParam      = "AppBundle"
	AppStoreURLORTBParam    = "AppStoreURL"
	AppDomainORTBParam      = "AppDomain"
	AppCatORTBParam         = "AppCat"
	AppPaidORTBParam        = "AppPaid"
	DeviceUAORTBParam       = "DeviceUA"
	DeviceDNTORTBParam      = "DeviceDNT"
	DeviceLMTORTBParam      = "DeviceLMT"
	DeviceJSORTBParam       = "DeviceJS"
	DeviceIPORTBParam       = "DeviceIP"
	DeviceIfaORTBParam      = "DeviceIfa"
	DeviceDidsha1ORTBParam  = "DeviceDidsha1"
	DeviceDidmd5ORTBParam   = "DeviceDidmd5"
	DeviceDpidsha1ORTBParam = "DeviceDpidsha1"
	DeviceDpidmd5ORTBParam  = "DeviceDpidmd5"
	DeviceMacsha1ORTBParam  = "DeviceMacsha1"
	DeviceMacmd5ORTBParam   = "DeviceMacmd5"
	GeoLatORTBParam         = "GeoLat"
	GeoLonORTBParam         = "GeoLon"
	GeoTypeORTBParam        = "GeoType"
	GeoCountryORTBParam     = "GeoCountry"
	GeoCityORTBParam        = "GeoCity"
	GeoMetroORTBParam       = "GeoMetro"
	GeoZipORTBParam         = "GeoZip"
	GeoUTOffsetORTBParam    = "GeoUTOffset"
	UserIDORTBParam         = "UserId"
	UserYobORTBParam        = "UserYob"
	UserGenderORTBParam     = "UserGender"
	SourceOmidpvORTBParam   = "SourceOmidpv"
	SourceOmidpnORTBParam   = "SourceOmidpn"
	ContentGenreORTBParam   = "Genre"
	ContentTitleORTBParam   = "Title"
	BidderParams            = "BidderParams"
	DeviceExtSessionID      = "DeviceExtSessionID"
	DeviceExtIfaType        = "DeviceExtIfaType"
	ImpPrebidExt            = "ImpPrebidExt"
)

// ORTBToDFPOWMap is Map of ORTB params to DFP and OW params. 0th position in map value denotes DFP param and 1st position in value denotes OW param. To populate a given ORTB parameter, preference would be given to DFP value and if its not present, OW value would be used
var ORTBToDFPOWMap = map[string][]string{
	MimeORTBParam:           {OWMimeTypes, ""},
	MinDurationORTBParam:    {OWMinAdDuration, DFPMinAdDuration},
	MaxDurationORTBParam:    {OWMaxAdDuration, DFPMaxAdDuration},
	StartDelayORTBParam:     {OWStartDelay, DFPVPos},
	PlaybackMethodORTBParam: {OWPlaybackMethod, ""},
	APIORTBParam:            {OWAPI, ""},
	ProtocolsORTBParam:      {OWProtocols, ""},
	SizeORTBParam:           {OWSize, DFPSize},
	BAttrORTBParam:          {OWBAttr, ""},
	LinearityORTBParam:      {OWLinearity, DFPVAdType},
	PlacementORTBParam:      {OWPlacement, ""},
	MaxBitrateORTBParam:     {OWMaxBitrate, ""},
	MinBitrateORTBParam:     {OWMinBitrate, ""},
	SkipORTBParam:           {OWSkippable, ""},
	SkipMinORTBParam:        {OWSkipMin, ""},
	SkipAfterORTBParam:      {OWSkipAfter, ""},
	ProtocolORTBParam:       {OWProtocol, ""},
	SequenceORTBParam:       {OWSequence, ""},
	MaxExtendedORTBParam:    {OWMaxExtended, ""},
	BoxingAllowedORTBParam:  {OWBoxingAllowed, ""},
	DeliveryORTBParam:       {OWDelivery, ""},
	PosORTBParam:            {OWPos, ""},
	AppIDORTBParam:          {OWAppId, ""},
	AppNameORTBParam:        {OWAppName, ""},
	AppBundleORTBParam:      {OWAppBundle, ""},
	AppStoreURLORTBParam:    {OWAppStoreURL, ""},
	AppCatORTBParam:         {OWAppCat, ""},
	AppPaidORTBParam:        {OWAppPaid, ""},
	AppDomainORTBParam:      {OWAppDomain, ""},
	DeviceUAORTBParam:       {OWDeviceUA, ""},
	DeviceDNTORTBParam:      {OWDeviceDNT, ""},
	DeviceLMTORTBParam:      {OWDeviceLMT, ""},
	DeviceJSORTBParam:       {OWDeviceJS, ""},
	DeviceIPORTBParam:       {OWDeviceIP, ""},
	DeviceIfaORTBParam:      {OWDeviceIfa, ""},
	DeviceDidsha1ORTBParam:  {OWDeviceDidsha1, ""},
	DeviceDidmd5ORTBParam:   {OWDeviceDidmd5, ""},
	DeviceDpidsha1ORTBParam: {OWDeviceDpidsha1, ""},
	DeviceDpidmd5ORTBParam:  {OWDeviceDpidmd5, ""},
	DeviceMacsha1ORTBParam:  {OWDeviceMacsha1, ""},
	DeviceMacmd5ORTBParam:   {OWDeviceMacmd5, ""},
	GeoLatORTBParam:         {OWGeoLat, ""},
	GeoLonORTBParam:         {OWGeoLon, ""},
	GeoTypeORTBParam:        {OWGeoType, ""},
	UserIDORTBParam:         {OWUserID, ""},
	GeoCountryORTBParam:     {OWGeoCountry, ""},
	GeoCityORTBParam:        {OWGeoCity, ""},
	GeoMetroORTBParam:       {OWGeoMetro, ""},
	GeoZipORTBParam:         {OWGeoZip, ""},
	GeoUTOffsetORTBParam:    {OWUTOffset, ""},
	ContentGenreORTBParam:   {OWContentGenre, ""},
	ContentTitleORTBParam:   {OWContentTitle, ""},
	UserYobORTBParam:        {OWUserYob, ""},
	UserGenderORTBParam:     {OWUserGender, ""},
	SourceOmidpvORTBParam:   {OWSourceOmidPv, ""},
	SourceOmidpnORTBParam:   {OWSourceOmidPn, ""},
	BidderParams:            {OWBidderParams, ""},
	DeviceExtSessionID:      {OWDeviceExtSessionID, ""},
	DeviceExtIfaType:        {OWDeviceExtIfaType, ""},
	ImpPrebidExt:            {OWImpPrebidExt, ""},
}

// DFP Video positions constants
const (
	Preroll  = "preroll"
	Midroll  = "midroll"
	Postroll = "postroll"
)

// VideoPositionToStartDelayMap is a map of DFP Video positions to Start Delay integer values in oRTB request
var VideoPositionToStartDelayMap = map[string]string{
	Preroll:  "0",
	Midroll:  "-1",
	Postroll: "-2",
}

// DFP Video linearity (vad_type) constants
const (
	Linear    = "linear"
	Nonlinear = "nonlinear"
)

// LinearityMap is a map of DFP Linearity values to oRTB values
var LinearityMap = map[string]string{
	Linear:    "1",
	Nonlinear: "2",
}

// Mime types
const (
	All        = "0"
	VideoMP4   = "1" // video/mp4
	VPAIDFlash = "2" // application/x-shockwave-flash (VPAID - FLASH)
	VideoWMV   = "3" // video/wmv
	VideoH264  = "4" // video/h264
	VideoWebm  = "5" // video/webm
	VPAIDJS    = "6" // application/javascript (VPAID - JS)
	VideoOGG   = "7" // video/ogg
	VideoFLV   = "8" // video/flv (Flash Video)
)

// MimeIDToValueMap is a map of Mime IDs to string values
var MimeIDToValueMap = map[string]string{
	All:        "All",
	VideoMP4:   "video/mp4",
	VPAIDFlash: "application/x-shockwave-flash",
	VideoWMV:   "video/wmv",
	VideoH264:  "video/h264",
	VideoWebm:  "video/webm",
	VPAIDJS:    "application/javascript",
	VideoOGG:   "video/ogg",
	VideoFLV:   "video/flv",
}

// CheckIfValidQueryParamFlag checks if given query parameter has a valid flag value(i.e. 0 or 1)
func CheckIfValidQueryParamFlag(values url.Values, key string) bool {
	validationFailed := false
	paramValue := values.Get(key)
	if paramValue == "" {
		return validationFailed
	}
	if paramValue != "0" && paramValue != "1" {
		validationFailed = true
	}
	return validationFailed
}
