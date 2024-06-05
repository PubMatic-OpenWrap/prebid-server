package models

import (
	"encoding/json"
	"net/http"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/wakanda"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/usersync"
)

type RequestCtx struct {
	// PubID is the publisher id retrieved from request
	PubID int
	// ProfileID is the value received in profileid field in wrapper object.
	ProfileID int
	// DisplayID is the value received in versionid field in wrapper object.
	DisplayID int
	// VersionID is the unique id from DB associated with the incoming DisplayID
	VersionID int
	// DisplayVersionID is the DisplayID of the profile selected by OpenWrap incase DisplayID/versionid is 0
	DisplayVersionID int

	SSAuction          int
	SummaryDisable     int
	LogInfoFlag        int
	SSAI               string
	PartnerConfigMap   map[int]map[string]string
	SupportDeals       bool
	Platform           string
	LoggerImpressionID string
	ClientConfigFlag   int
	Country            string
	IP                 string
	TMax               int64

	//NYC_TODO: use enum?
	IsTestRequest                     int8
	ABTestConfig, ABTestConfigApplied int
	IsCTVRequest                      bool

	TrackerEndpoint, VideoErrorTrackerEndpoint string

	UA              string
	Cookies         string
	UidCookie       *http.Cookie
	KADUSERCookie   *http.Cookie
	ParsedUidCookie *usersync.Cookie
	OriginCookie    string

	Debug bool
	Trace bool

	//tracker
	PageURL   string
	StartTime int64
	DeviceCtx DeviceCtx

	//trackers per bid
	Trackers map[string]OWTracker

	//prebid-biddercode to seat/alias mapping
	PrebidBidderCode map[string]string

	// imp-bid ctx to avoid computing same thing for bidder params, logger and tracker
	ImpBidCtx          map[string]ImpCtx
	Aliases            map[string]string
	NewReqExt          *RequestExt
	ResponseExt        openrtb_ext.ExtBidResponse
	MarketPlaceBidders map[string]struct{}

	AdapterThrottleMap map[string]struct{}
	AdapterFilteredMap map[string]struct{}

	AdUnitConfig *adunitconfig.AdUnitConfig

	Source, Origin string

	SendAllBids bool
	WinningBids map[string]OwBid
	DroppedBids map[string][]openrtb2.Bid
	DefaultBids map[string]map[string][]openrtb2.Bid
	SeatNonBids map[string][]openrtb_ext.NonBid // map of bidder to list of nonbids

	BidderResponseTimeMillis map[string]int

	Endpoint               string
	PubIDStr, ProfileIDStr string // TODO: remove this once we completely move away from header-bidding
	MetricsEngine          metrics.MetricsEngine
	ReturnAllBidStatus     bool   // ReturnAllBidStatus stores the value of request.ext.prebid.returnallbidstatus
	Sshb                   string //Sshb query param to identify that the request executed heder-bidding or not, sshb=1(executed HB(8001)), sshb=2(reverse proxy set from HB(8001->8000)), sshb=""(direct request(8000)).

	DCName                 string
	CachePutMiss           int                                                          // to be used in case of CTV JSON endpoint/amp/inapp-ott-video endpoint
	CurrencyConversion     func(from string, to string, value float64) (float64, error) `json:"-"`
	MatchedImpression      map[string]int
	CustomDimensions       map[string]CustomDimension
	AmpVideoEnabled        bool //AmpVideoEnabled indicates whether to include a Video object in an AMP request.
	IsTBFFeatureEnabled    bool
	VastUnwrapEnabled      bool
	VastUnwrapStatsEnabled bool
	AppLovinMax            AppLovinMax
	LoggerDisabled         bool
	TrackerDisabled        bool
	ProfileType            int
	ProfileTypePlatform    int
	AppPlatform            int
	AppIntegrationPath     *int
	AppSubIntegrationPath  *int
	WakandaDebug           wakanda.DebugInterface
}

type OwBid struct {
	ID                   string
	NetEcpm              float64
	BidDealTierSatisfied bool
	Nbr                  *openrtb3.NoBidReason
}

func (r RequestCtx) GetVersionLevelKey(key string) string {
	if len(r.PartnerConfigMap) == 0 || len(r.PartnerConfigMap[VersionLevelConfigID]) == 0 {
		return ""
	}
	v := r.PartnerConfigMap[VersionLevelConfigID][key]
	return v
}

// DeviceCtx to cache device specific parameters
type DeviceCtx struct {
	DeviceIFA string
	IFATypeID *DeviceIFAType
	Platform  DevicePlatform
	Ext       *ExtDevice
}

type ImpCtx struct {
	ImpID             string
	TagID             string
	Div               string
	SlotName          string
	AdUnitName        string
	Secure            int
	BidFloor          float64
	BidFloorCur       string
	IsRewardInventory *int8
	Banner            bool
	Video             *openrtb2.Video
	Native            *openrtb2.Native
	IncomingSlots     []string
	Type              string // banner, video, native, etc
	Bidders           map[string]PartnerData
	NonMapped         map[string]struct{}

	NewExt json.RawMessage
	BidCtx map[string]BidCtx

	BannerAdUnitCtx AdUnitCtx
	VideoAdUnitCtx  AdUnitCtx

	//temp
	BidderError    string
	IsAdPodRequest bool
}

type PartnerData struct {
	PartnerID        int
	PrebidBidderCode string
	MatchedSlot      string
	KGP              string
	KGPV             string
	IsRegex          bool
	Params           json.RawMessage
	VASTTagFlag      bool
	VASTTagFlags     map[string]bool
}

type BidCtx struct {
	BidExt

	// EG gross net in USD for tracker and logger
	EG float64
	// EN gross net in USD for tracker and logger
	EN float64
}

type AdUnitCtx struct {
	MatchedSlot              string
	IsRegex                  bool
	MatchedRegex             string
	SelectedSlotAdUnitConfig *adunitconfig.AdConfig
	AppliedSlotAdUnitConfig  *adunitconfig.AdConfig
	UsingDefaultConfig       bool
	AllowedConnectionTypes   []int
}

type CustomDimension struct {
	Value     string `json:"value,omitempty"`
	SendToGAM *bool  `json:"sendtoGAM,omitempty"`
}

// FeatureData struct to hold feature data from cache
type FeatureData struct {
	Enabled int    // feature enabled/disabled
	Value   string // feature value if any
}

type AppLovinMax struct {
	Reject bool
}
