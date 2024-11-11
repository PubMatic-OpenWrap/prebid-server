package models

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

const (
	//BidderOWPrebidCTV for prebid adpod response
	BidderOWPrebidCTV string = "prebid_ctv"
)

const (
	DefaultMinAds                      = 1
	DefaultMaxAds                      = 3
	DefaultAdvertiserExclusionPercent  = 100
	DefaultIABCategoryExclusionPercent = 100
)

const (
	ADPOD = "adpod"
)

const (
	// MinDuration represents index value where we can get minimum duration of given impression object
	MinDuration = iota
	// MaxDuration represents index value where we can get maximum duration of given impression object
	MaxDuration
)

const (
	//StatusOK ...
	StatusOK int64 = 0
	//StatusWinningBid ...
	StatusWinningBid int64 = 1
	//StatusCategoryExclusion ...
	StatusCategoryExclusion int64 = 2
	//StatusDomainExclusion ...
	StatusDomainExclusion int64 = 3
	//StatusDurationMismatch ...
	StatusDurationMismatch int64 = 4
)

// ImpAdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
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

type PodType uint

const (
	NotAdpod   PodType = 0
	Dynamic    PodType = 1
	Structured PodType = 2
)

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

func (ex *Exclusion) ShouldApplyExclusion() bool {
	return ex.AdvertiserDomainExclusion || ex.IABCategoryExclusion
}

type Adpod interface {
	GetImpressions() []*openrtb_ext.ImpWrapper
	CollectBid(bid *openrtb2.Bid, seat string)
	HoldAuction()
	GetWinningBidsIds(rctx RequestCtx, winningBidIds map[string][]string)
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
type GeneratedSlotConfig struct {
	ImpID          string `json:"id,omitempty"`
	SequenceNumber int8   `json:"seq,omitempty"`
	MinDuration    int64  `json:"minduration,omitempty"`
	MaxDuration    int64  `json:"maxduration,omitempty"`
}
