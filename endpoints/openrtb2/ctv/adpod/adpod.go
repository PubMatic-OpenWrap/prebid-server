package adpod

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/metrics"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

type PodType int8

const (
	Structured PodType = 0
	Dynamic    PodType = 1
	Hybrid     PodType = 2
	NotAdpod   PodType = -1
)

type Adpod interface {
	GetPodType() PodType
	AddImpressions(imp openrtb2.Imp)
	Validate() []error
	GetImpressions() []openrtb2.Imp
	CollectBid(bid openrtb2.Bid, seat string)
	HoldAuction()
	GetAdpodSeatBids() []openrtb2.SeatBid
	GetAdpodExtension(blockedVastTagID map[string]map[string][]string) *types.ImpData
}

type AdpodCtx struct {
	PubId         string
	Type          PodType
	Imps          []openrtb2.Imp
	ReqAdpodExt   *openrtb_ext.ExtRequestAdPod
	Exclusion     Exclusion
	MetricsEngine metrics.MetricsEngine
}

type Exclusion struct {
	AdvertiserExclusionPercent  int // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent int // Percent value 0 means all ads should be of different IAB categories.
	SelectedCategories          map[string]bool
	SelectedDomains             map[string]bool
}
