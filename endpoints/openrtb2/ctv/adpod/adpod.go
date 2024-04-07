package adpod

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type PodType int8

const (
	PodTypeStructured PodType = 0
	PodTypeDynamic    PodType = 1
	PodTypeHybrid     PodType = 2
)

type Adpod interface {
	GetPodType() PodType
	AddImpressions(imp openrtb2.Imp)
	Validate() []error
	GenerateImpressions()
	GetImpressions() []openrtb2.Imp
	CollectBid(bid openrtb2.Bid, seat string)
	PerformAuctionAndExclusion()
	GetAdpodSeatBids() []openrtb2.SeatBid
	GetAdpodExtension(blockedVastTagID map[string]map[string][]string) *types.ImpData
}

type AdpodCtx struct {
	PubId         string
	Type          PodType
	Imps          []openrtb2.Imp
	ReqExt        *openrtb_ext.ExtRequestAdPod
	MetricsEngine metrics.MetricsEngine
}
