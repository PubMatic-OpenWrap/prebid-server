package models

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

type Adpod interface {
	CollectBid(bid *openrtb2.Bid, seat string)
	HoldAuction()
	CollectAPRC(impCtxMap map[string]ImpCtx)
	GetWinningBidsIds(impCtxMap map[string]ImpCtx, ImpToWinningBids map[string]map[string]bool)
}

// AdpodCtx context for adpod
type AdpodCtx struct {
	PodId          string
	Type           PodType
	Imps           []openrtb2.Imp
	Exclusion      Exclusion
	ProfileConfigs *AdpodProfileConfig
}

// Exclusion config for adpod
type Exclusion struct {
	AdvertiserDomainExclusion bool
	IABCategoryExclusion      bool
}

type GeneratedSlotConfig struct {
	ImpID          string `json:"id,omitempty"`
	SequenceNumber int8   `json:"seq,omitempty"`
	MinDuration    int64  `json:"minduration,omitempty"`
	MaxDuration    int64  `json:"maxduration,omitempty"`
}

type PodConfig struct {
	PodID       string
	PodDur      int64
	MaxSeq      int64
	MinDuration int64
	MaxDuration int64
	RqdDurs     []int64
}

type Bid struct {
	*openrtb2.Bid
	openrtb_ext.ExtBid
	Duration          int
	Status            int64
	DealTierSatisfied bool
	Seat              string
}

type AdPodBid struct {
	Bids          []*Bid
	Price         float64
	Cat           []string
	ADomain       []string
	OriginalImpID string
	SeatName      string
}
