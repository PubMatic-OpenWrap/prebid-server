package models

import (
	"encoding/json"
	"net/http"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/usersync"
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
	NewReqExt          *RequestExt
	ResponseExt        openrtb_ext.ExtBidResponse
	MarketPlaceBidders map[string]struct{}

	AdapterThrottleMap map[string]struct{}

	AdUnitConfig *adunitconfig.AdUnitConfig

	Source, Origin string

	SendAllBids bool
	WinningBids WinningBids
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
	CachePutMiss           int // to be used in case of CTV JSON endpoint/amp/inapp-ott-video endpoint
	CurrencyConversion     func(from string, to string, value float64) (float64, error)
	MatchedImpression      map[string]int

	Errors []error
}

type OwBid struct {
	ID                   string
	NetEcpm              float64
	BidDealTierSatisfied bool
	Nbr                  *openrtb3.NonBidStatusCode
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
	NewExt            json.RawMessage
	BidCtx            map[string]BidCtx
	BannerAdUnitCtx   AdUnitCtx
	VideoAdUnitCtx    AdUnitCtx
	AdpodConfig       *AdPod
	ImpAdPodCfg       []*ImpAdPodConfig

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

type WinningBids map[string][]*OwBid

func (w WinningBids) IsWinningBid(impId, bidId string) bool {
	var isWinningBid bool

	wbids, ok := w[impId]
	if !ok {
		return isWinningBid
	}

	for i := range wbids {
		if bidId == wbids[i].ID {
			isWinningBid = true
			break
		}
	}

	return isWinningBid
}

func (w WinningBids) AppendBid(impId string, bid *OwBid) {
	wbid, ok := w[impId]
	if !ok {
		wbid = make([]*OwBid, 0)
	}

	wbid = append(wbid, bid)
	w[impId] = wbid
}

// isNewWinningBid calculates if the new bid (nbid) will win against the current winning bid (wbid) given preferDeals.
func IsNewWinningBid(bid, wbid *OwBid, preferDeals bool) bool {
	if preferDeals {
		//only wbid has deal
		if wbid.BidDealTierSatisfied && !bid.BidDealTierSatisfied {
			bid.Nbr = GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid)
			return false
		}
		//only bid has deal
		if !wbid.BidDealTierSatisfied && bid.BidDealTierSatisfied {
			wbid.Nbr = GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid)
			return true
		}
	}
	//both have deal or both do not have deal
	if bid.NetEcpm > wbid.NetEcpm {
		wbid.Nbr = GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid)
		return true
	}
	bid.Nbr = GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid)
	return false
}

func GetNonBidStatusCodePtr(nbr openrtb3.NonBidStatusCode) *openrtb3.NonBidStatusCode {
	return &nbr
}
