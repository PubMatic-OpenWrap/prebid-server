package models

import (
	"encoding/json"
	"net/http"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
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

	UA            string
	Cookies       string
	UidCookie     *http.Cookie
	KADUSERCookie *http.Cookie
	OriginCookie  string

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
	WinningBids WinningBids
	DroppedBids map[string][]openrtb2.Bid
	NoSeatBids  map[string]map[string][]openrtb2.Bid

	BidderResponseTimeMillis map[string]int

	Endpoint               string
	PubIDStr, ProfileIDStr string // TODO: remove this once we completely move away from header-bidding
	MetricsEngine          metrics.MetricsEngine
}

type OwBid struct {
	ID                   string
	NetEcpm              float64
	BidDealTierSatisfied bool
}

func (r RequestCtx) GetVersionLevelKey(key string) (string, bool) {
	if len(r.PartnerConfigMap) == 0 || len(r.PartnerConfigMap[VersionLevelConfigID]) == 0 {
		return "", false
	}
	v, ok := r.PartnerConfigMap[VersionLevelConfigID][key]
	return v, ok
}

type ImpCtx struct {
	ImpID             string
	TagID             string
	Div               string
	Secure            int
	IsRewardInventory *int8
	Banner            bool
	Video             *openrtb2.Video
	IncomingSlots     []string
	Type              string // banner, video, native, etc
	Bidders           map[string]PartnerData
	NonMapped         map[string]struct{}
	NewExt            json.RawMessage
	BidCtx            map[string]BidCtx
	BannerAdUnitCtx   AdUnitCtx
	VideoAdUnitCtx    AdUnitCtx
	AdpodConfig       *AdPod
	ImpAdPodCfg       []*ImpAdPodConfig
}

type PartnerData struct {
	PartnerID        int
	PrebidBidderCode string
	MatchedSlot      string
	KGP              string
	KGPV             string
	IsRegex          bool
	Params           json.RawMessage
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
