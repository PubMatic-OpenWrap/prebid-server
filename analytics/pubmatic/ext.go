package pubmatic

import (
	"encoding/json"

	"github.com/prebid/prebid-server/openrtb_ext"
)

// ImpExtension - Impression Extension
type ImpExtension struct {
	Wrapper     *ExtImpWrapper              `json:"wrapper,omitempty"`
	Bidder      map[string]*BidderExtension `json:"bidder,omitempty"`
	SKAdnetwork json.RawMessage             `json:"skadn,omitempty"`
	Reward      *int                        `json:"reward,omitempty"`
	Data        json.RawMessage             `json:"data,omitempty"`
	Prebid      *openrtb_ext.ExtImpPrebid   `json:"prebid,omitempty"`
}

// ExtImpWrapper - Impression wrapper Extension
type ExtImpWrapper struct {
	Div *string `json:"div,omitempty"`
}

// BidderExtension - Bidder specific items
type BidderExtension struct {
	KeyWords []KeyVal  `json:"keywords,omitempty"`
	DealTier *DealTier `json:"dealtier,omitempty"`
}

// KeyVal structure to store bidder related custom key-values
type KeyVal struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"value,omitempty"`
}

// DealTier - Deal information for individual bidders
type DealTier struct {
	Prefix      string `json:"prefix,omitempty"`
	MinDealTier int    `json:"mindealtier,omitempty"`
}
