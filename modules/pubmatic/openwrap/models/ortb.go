package models

// OpenRTB Constants
const (
	//constant for adformat
	Banner = "banner"
	Video  = "video"
	Native = "native"

	//constants for reading video keys from adunit Config
	VideoMinDuration    = "minduration"
	VideoMaxDuration    = "maxduration"
	VideoSkip           = "skip"
	VideoSkipMin        = "skipmin"
	VideoSkipAfter      = "skipafter"
	VideoBattr          = "battr"
	VideoConnectionType = "connectiontype"
	VideoMinBitRate     = "minbitrate"
	VideoMaxBitRate     = "maxbitrate"
	VideoMaxExtended    = "maxextended"
	VideoStartDelay     = "startdelay"
	VideoPlacement      = "placement"
	VideoLinearity      = "linearity"
	VideoMimes          = "mimes"
	VideoProtocol       = "protocol"
	VideoProtocols      = "protocols"
	VideoW              = "w"
	VideoH              = "h"
	VideoSequence       = "sequence"
	VideoBoxingAllowed  = "boxingallowed"
	VideoPlaybackMethod = "playbackmethod"
	VidepPlaybackEnd    = "playbackend"
	VideoDelivery       = "delivery"
	VideoPos            = "pos"
	VideoAPI            = "api"
	VideoCompanionType  = "companiontype"
	VideoComapanionAd   = "companionad"

	//banner obj
	BannerFormat   = "format"
	BannerW        = "w"
	BannerH        = "h"
	BannerWMax     = "wmax"
	BannerHMax     = "hmax"
	BannerWMin     = "wmin"
	BannerHMin     = "hmin"
	BannerBType    = "btype"
	BannerBAttr    = "battr"
	BannerPos      = "pos"
	BannerMimes    = "mimes"
	BannerTopFrame = "topframe"
	BannerExpdir   = "expdir"
	BannerAPI      = "api"
	BannerID       = "id"
	BannerVcm      = "vcm"

	//format object
	FormatW      = "w"
	FormatH      = "h"
	FormatWRatio = "wratio"
	FormatHRatio = "hratio"
	FormatWmin   = "wmin"
)

type ConsentType int

const (
	Unknown ConsentType = iota
	TCF_V1
	TCF_V2
	CCPA
)
