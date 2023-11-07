package models

import (
	"encoding/json"
	"net/http"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/usersync"
)

type RequestCtx struct {
	PubID, ProfileID, DisplayID, VersionID int
	SSAuction                              int
	SummaryDisable                         int
	LogInfoFlag                            int
	SSAI                                   string
	PartnerConfigMap                       map[int]map[string]string
	SupportDeals                           bool
	Platform                               string
	LoggerImpressionID                     string
	ClientConfigFlag                       int

	IP   string
	TMax int64

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
	PageURL        string
	StartTime      int64
	DevicePlatform DevicePlatform

	//trackers per bid
	Trackers map[string]OWTracker

	//prebid-biddercode to seat/alias mapping
	PrebidBidderCode map[string]string

	// imp-bid ctx to avoid computing same thing for bidder params, logger and tracker
	ImpBidCtx          map[string]ImpCtx
	Aliases            map[string]string
	NewReqExt          json.RawMessage
	ResponseExt        json.RawMessage
	MarketPlaceBidders map[string]struct{}

	AdapterThrottleMap map[string]struct{}

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
	ReturnAllBidStatus     bool // ReturnAllBidStatus stores the value of request.ext.prebid.returnallbidstatus
	CurrencyConversion     func(from string, to string, value float64) (float64, error)
}

type OwBid struct {
	ID                   string
	NetEcpm              float64
	BidDealTierSatisfied bool
}

func (r RequestCtx) GetVersionLevelKey(key string) string {
	if len(r.PartnerConfigMap) == 0 || len(r.PartnerConfigMap[VersionLevelConfigID]) == 0 {
		return ""
	}
	v := r.PartnerConfigMap[VersionLevelConfigID][key]
	return v
}

type ImpCtx struct {
	ImpID             string
	TagID             string
	Div               string
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
