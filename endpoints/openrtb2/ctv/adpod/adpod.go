package adpod

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/exchange/entities"
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
	GetSeatNonBid() openrtb_ext.NonBidCollection
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

// GetNonBidParamsFromTypesBid function returns NonBidParams from types Bid
func GetNonBidParamsFromTypesBid(bid *types.Bid, seat string) openrtb_ext.NonBidParams {
	if bid.ExtBid.Prebid == nil {
		bid.ExtBid.Prebid = &openrtb_ext.ExtBidPrebid{}
	}
	pbsOrtbBid := entities.PbsOrtbBid{
		Bid:               bid.Bid,
		BidMeta:           bid.ExtBid.Prebid.Meta,
		BidType:           bid.ExtBid.Prebid.Type,
		BidTargets:        bid.ExtBid.Prebid.Targeting,
		BidVideo:          bid.ExtBid.Prebid.Video,
		BidEvents:         bid.ExtBid.Prebid.Events,
		BidFloors:         bid.ExtBid.Prebid.Floors,
		DealPriority:      bid.ExtBid.Prebid.DealPriority,
		DealTierSatisfied: bid.DealTierSatisfied,
		GeneratedBidID:    bid.ExtBid.Prebid.BidId,
		OriginalBidCPM:    bid.OriginalBidCPM,
		OriginalBidCur:    bid.OriginalBidCur,
		TargetBidderCode:  bid.ExtBid.Prebid.TargetBidderCode,
		OriginalBidCPMUSD: bid.OriginalBidCPMUSD,
	}
	return entities.GetNonBidParamsFromPbsOrtbBid(&pbsOrtbBid, seat)
}

func addSeatNonBids(bids []*types.Bid) openrtb_ext.NonBidCollection {
	var snb openrtb_ext.NonBidCollection
	for _, bid := range bids {
		if bid.Nbr != nil {
			nonBidParams := GetNonBidParamsFromTypesBid(bid, bid.Seat)
			nonBidParams.NonBidReason = int(*bid.Nbr)
			snb.AddBid(openrtb_ext.NewNonBid(nonBidParams), bid.Seat)
		}
	}
	return snb
}
