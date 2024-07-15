package models

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

// ImpAdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
	ImpID          string `json:"id,omitempty"`
	SequenceNumber int8   `json:"seq,omitempty"`
	MinDuration    int64  `json:"minduration,omitempty"`
	MaxDuration    int64  `json:"maxduration,omitempty"`
}
