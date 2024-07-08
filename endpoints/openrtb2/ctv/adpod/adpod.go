package adpod

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/constant"
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
	CollectBid(bid *openrtb2.Bid, seat string)
	HoldAuction()
	GetAdpodSeatBids() []openrtb2.SeatBid
	GetWinningBids() []openrtb2.SeatBid
	GetAdpodExtension(blockedVastTagID map[string]map[string][]string) *types.ImpData
	GetSeatNonBid(snb *openrtb_ext.NonBidCollection)
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
	AdvertiserDomainExclusion bool
	IABCategoryExclusion      bool
}

func (ex *Exclusion) shouldApplyExclusion() bool {
	return ex.AdvertiserDomainExclusion || ex.IABCategoryExclusion
}

// GetNonBidParamsFromPbsOrtbBid function returns NonBidParams from PbsOrtbBid
func GetNonBidParamsFromPbsOrtbBid(bid *types.Bid, seat string) openrtb_ext.NonBidParams {
	adapterCode := seat
	if bid.Prebid.Meta == nil {
		bid.Prebid.Meta = &openrtb_ext.ExtBidPrebidMeta{}
	}
	bid.Prebid.Meta.AdapterCode = adapterCode
	return openrtb_ext.NonBidParams{
		Bid:               bid.Bid,
		OriginalBidCPM:    bid.OriginalBidCPM,
		OriginalBidCur:    bid.OriginalBidCur,
		DealPriority:      bid.Prebid.DealPriority,
		DealTierSatisfied: bid.Prebid.DealTierSatisfied,
		GeneratedBidID:    bid.Prebid.BidId,
		TargetBidderCode:  bid.Prebid.TargetBidderCode,
		OriginalBidCPMUSD: bid.OriginalBidCPMUSD,
		BidMeta:           bid.Prebid.Meta,
		BidType:           bid.Prebid.Type,
		BidTargets:        bid.Prebid.Targeting,
		BidVideo:          bid.Prebid.Video,
		BidEvents:         bid.Prebid.Events,
		BidFloors:         bid.Prebid.Floors,
	}
}

func addSeatNonBids(snb *openrtb_ext.NonBidCollection, bids []*types.Bid) {
	for _, bid := range bids {
		if bid.Status != constant.StatusWinningBid {
			nonBidParams := GetNonBidParamsFromPbsOrtbBid(bid, bid.Seat)
			convertedReason := ConvertAPRCToNBRC(bid.Status)
			if convertedReason != nil {
				nonBidParams.NonBidReason = int(*convertedReason)
			}
			snb.AddBid(openrtb_ext.NewNonBid(nonBidParams), bid.Seat)
		}
	}
}
