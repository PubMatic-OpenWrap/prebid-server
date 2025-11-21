package models

import (
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
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
	Adpod = "adpod"
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

type PodType int8

const (
	PodTypeStructured PodType = iota + 1
	PodTypeDynamic
	PodTypeHybrid
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

// Adpod Context
type AdpodCtx map[string]AdpodConfig

type SlotConfig struct {
	// Imp ID of the slot
	Id string `json:"id,omitempty"`

	// slot position indicator (spec: video.slotinpod)
	SlotInPod adcom1.SlotPositionInPod `json:"slotinpod,omitempty"`

	// slot-level duration constraints (spec: video.minduration / video.maxduration)
	MinDuration int64 `json:"minduration,omitempty"`
	MaxDuration int64 `json:"maxduration,omitempty"`

	// exact allowed durations (mutually exclusive with minduration/maxduration)
	RqdDurs []int64 `json:"rqddurs,omitempty"`

	// dynamic/hybrid related (spec: video.poddur = total dynamic portion length)
	// For normalized hybrid modelling we allow one slot to carry pod-level dynamic info.
	PodDur int64 `json:"poddur,omitempty"` // total dynamic portion seconds (if present)
	MaxSeq int64 `json:"maxseq,omitempty"` // spec: maximum # ads in dynamic portion

	// CTV 2.5 dynamic adpod config
	MinAds                      int64 `json:"minads,omitempty"`
	MaxAds                      int64 `json:"maxads,omitempty"`
	MinPodDuration              int64 `json:"minpodduration,omitempty"`
	MaxPodDuration              int64 `json:"maxpodduration,omitempty"`
	AdvertiserExclusionPercent  *int  `json:"excladv,omitempty"`    // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int  `json:"excliabcat,omitempty"` // Percent value 0 means all ads should be of different IAB categories.

	// helper flag: true when this slot is flexible/dynamic (poddur/maxseq/mincpmpersec present)
	Flexible bool `json:"flexible,omitempty"`
}

type ExclusionConfig struct {
	AdvertiserDomainExclusion bool `json:"advertiserdomainexclusion,omitempty"`
	IABCategoryExclusion      bool `json:"iabcategoryexclusion,omitempty"`
}

// AdpodConfig configuration for adpod
type AdpodConfig struct {
	PodID     string             `json:"podid,omitempty"`  // spec: podid (string recommended)
	PodSeq    adcom1.PodSequence `json:"podseq,omitempty"` // spec: podseq (0 any, -1 last, 1 first)
	PodType   PodType            `json:"podtype,omitempty"`
	Exclusion ExclusionConfig    `json:"exclusion,omitempty"`
	Slots     []SlotConfig       `json:"slots,omitempty"`
}

func (a AdpodCtx) AddAdpodConfig(imp *openrtb2.Imp) {
	if imp == nil || imp.Video == nil || imp.Video.PodID == "" {
		return
	}

	// Get existing config or create a new one
	config, exists := a[imp.Video.PodID]
	if !exists {
		config = AdpodConfig{
			PodID:  imp.Video.PodID,
			PodSeq: imp.Video.PodSeq,
		}
	}

	// Add the slot to the config
	config.AddSlot(imp)

	// Determine pod type based on slots
	config.DeterminePodType()

	// Update the context with the modified config
	a[imp.Video.PodID] = config
}

func (a *AdpodConfig) AddSlot(imp *openrtb2.Imp) {
	slot := SlotConfig{
		Id:          imp.ID,
		SlotInPod:   imp.Video.SlotInPod,
		MinDuration: imp.Video.MinDuration,
		MaxDuration: imp.Video.MaxDuration,
		RqdDurs:     imp.Video.RqdDurs,
		PodDur:      imp.Video.PodDur,
		MaxSeq:      imp.Video.MaxSeq,
	}

	if slot.PodDur > 0 {
		slot.Flexible = true
	}

	a.Slots = append(a.Slots, slot)
}

func (a *AdpodConfig) DeterminePodType() {
	if len(a.Slots) == 0 {
		return
	}

	flexibleCount := 0
	nonFlexibleCount := 0
	for _, slot := range a.Slots {
		if slot.Flexible {
			flexibleCount++
		} else {
			nonFlexibleCount++
		}
	}

	// Determine pod type based on the counts
	switch {
	case flexibleCount == 1 && nonFlexibleCount == 0:
		a.PodType = PodTypeDynamic
	case flexibleCount == 0 && nonFlexibleCount > 0:
		a.PodType = PodTypeStructured
	case flexibleCount > 0 && nonFlexibleCount > 0:
		a.PodType = PodTypeHybrid
	}
}
