package models

import "github.com/prebid/openrtb/v20/openrtb2"

type Adpod interface{}

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
