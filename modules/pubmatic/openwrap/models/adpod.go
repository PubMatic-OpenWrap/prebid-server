package models

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

type PodType int8

const (
	//BidderOWPrebidCTV for prebid adpod response
	BidderOWPrebidCTV = "prebid_ctv"
	AdpodKey          = "adpod"

	// Defaults
	DefaultMinAds                      = 1
	DefaultMaxAds                      = 3
	DefaultAdvertiserExclusionPercent  = 100
	DefaultIABCategoryExclusionPercent = 100

	// MinDuration represents index value where we can get minimum duration of given impression object
	MinDuration = 0
	// MaxDuration represents index value where we can get maximum duration of given impression object
	MaxDuration = 1

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

	//Pod type
	Structured PodType = 0
	Dynamic    PodType = 1
	Hybrid     PodType = 2
	NotAdpod   PodType = -1
)

// ImpAdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
	ImpID          string `json:"id,omitempty"`
	SequenceNumber int8   `json:"seq,omitempty"`
	MinDuration    int64  `json:"minduration,omitempty"`
	MaxDuration    int64  `json:"maxduration,omitempty"`
}

type Adpod interface {
	GetPodType() PodType
	AddImpressions(imp openrtb2.Imp)
	Validate() []error
	GetImpressions() []openrtb2.Imp
	CollectBid(bid *openrtb2.Bid, seat string)
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
	AdvertiserDomainExclusion bool
	IABCategoryExclusion      bool
}

func (ex *Exclusion) shouldApplyExclusion() bool {
	return ex.AdvertiserDomainExclusion || ex.IABCategoryExclusion
}
