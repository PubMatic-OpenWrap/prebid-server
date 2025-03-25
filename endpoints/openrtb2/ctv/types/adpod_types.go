package types

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

// Bid openrtb bid object with extra parameters
type Bid struct {
	*openrtb2.Bid
	openrtb_ext.ExtBid
	Duration          int
	Status            constant.BidStatus
	DealTierSatisfied bool
	Seat              string
}

// ExtCTVBidResponse object for ctv bid resposne object
type ExtCTVBidResponse struct {
	openrtb_ext.ExtBidResponse
	AdPod *BidResponseAdPodExt `json:"adpod,omitempty"`
}

// BidResponseAdPodExt object for ctv bidresponse adpod object
type BidResponseAdPodExt struct {
	Response openrtb2.BidResponse `json:"bidresponse,omitempty"`
	Config   map[string]*ImpData  `json:"config,omitempty"`
}

// AdPodBid combination contains ImpBid
type AdPodBid struct {
	Bids          []*Bid
	Price         float64
	Cat           []string
	ADomain       []string
	OriginalImpID string
	SeatName      string
}

// AdPodBids combination contains ImpBid
type AdPodBids []*AdPodBid

// BidsBuckets bids bucket
type BidsBuckets map[int][]*Bid

// ImpAdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
	ImpID          string `json:"id,omitempty"`
	SequenceNumber int8   `json:"seq,omitempty"`
	MinDuration    int64  `json:"minduration,omitempty"`
	MaxDuration    int64  `json:"maxduration,omitempty"`
}

// ImpData example
type ImpData struct {
	//AdPodGenerator
	ImpID           string                        `json:"-"`
	Bid             *AdPodBid                     `json:"-"`
	VideoExt        *openrtb_ext.ExtVideoAdPod    `json:"vidext,omitempty"`
	Config          []*ImpAdPodConfig             `json:"imp,omitempty"`
	BlockedVASTTags map[string][]string           `json:"blockedtags,omitempty"`
	Error           *openrtb_ext.ExtBidderMessage `json:"ec,omitempty"`
}
