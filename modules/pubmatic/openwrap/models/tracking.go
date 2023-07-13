package models

// impression tracker url parameters
const (
	// constants for query parameter names for tracker call
	TRKPubID             = "pubid"
	TRKPageURL           = "purl"
	TRKTimestamp         = "tst"
	TRKIID               = "iid"
	TRKProfileID         = "pid"
	TRKVersionID         = "pdvid"
	TRKIP                = "ip"
	TRKUserAgent         = "ua"
	TRKSlotID            = "slot"
	TRKAdunit            = "au"
	TRKRewardedInventory = "rwrd"
	TRKPartnerID         = "pn"
	TRKBidderCode        = "bc"
	TRKKGPV              = "kgpv"
	TRKGrossECPM         = "eg"
	TRKNetECPM           = "en"
	TRKBidID             = "bidid"
	TRKOrigBidID         = "origbidid"
	TRKQMARK             = "?"
	TRKAmpersand         = "&"
	TRKSSAI              = "ssai"
)

// video error tracker url parameters
const (
	ERROperIDValue    = "8"
	ERROperID         = "operId"
	ERROperIDParam    = ERROperID + "=" + ERROperIDValue
	ERRPubID          = "p"
	ERRProfileID      = "pid"
	ERRVersionID      = "v"
	ERRTimestamp      = "ts"
	ERRPartnerID      = "pn"
	ERRBidderCode     = "bc"
	ERRAdunit         = "au"
	ERRCreativeID     = "crId"
	ERRErrorCode      = "ier"
	ERRErrorCodeMacro = "[ERRORCODE]"
	ERRErrorCodeParam = ERRErrorCode + "=" + ERRErrorCodeMacro
	ERRSUrl           = "sURL" // key represents either domain or bundle from request
	ERRPlatform       = "pfi"
	ERRAdvertiser     = "adv"
	ERRSSAI           = "ssai"
)

// EventTrackingMacros Video Event Tracker's custom macros
type EventTrackingMacros string

const (
	MacroProfileID           EventTrackingMacros = "[PROFILE_ID]"            // Pass Profile ID using this macro
	MacroProfileVersionID    EventTrackingMacros = "[PROFILE_VERSION]"       // Pass Profile's version ID using this macro
	MacroUnixTimeStamp       EventTrackingMacros = "[UNIX_TIMESTAMP]"        // Pass Current Unix Time when Event Tracking URL is generated using this macro
	MacroPlatform            EventTrackingMacros = "[PLATFORM]"              // Pass PubMatic's Platform using this macro
	MacroWrapperImpressionID EventTrackingMacros = "[WRAPPER_IMPRESSION_ID]" // Pass Wrapper Impression ID using this macro
	MacroSSAI                EventTrackingMacros = "[SSAI]"                  // Pass SSAI vendor name using this macro
)

// DspId for Pixel Based Open Measurement
const (
	DspId_DV360 = 80
)
