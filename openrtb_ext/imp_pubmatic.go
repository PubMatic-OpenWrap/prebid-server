package openrtb_ext

import "encoding/json"

// ExtImpPubmatic defines the contract for bidrequest.imp[i].ext.pubmatic
// PublisherId is mandatory parameters, others are optional parameters
// AdSlot is identifier for specific ad placement or ad tag
// Keywords is bid specific parameter,
// WrapExt needs to be sent once per bid request

type ExtImpPubmatic struct {
	PublisherId string                  `json:"publisherId"`
	AdSlot      string                  `json:"adSlot"`
	Dctr        string                  `json:"dctr,omitempty"`
	PmZoneID    string                  `json:"pmzoneid,omitempty"`
	WrapExt     json.RawMessage         `json:"wrapper,omitempty"`
	Keywords    []*ExtImpPubmaticKeyVal `json:"keywords,omitempty"`
	Kadfloor    string                  `json:"kadfloor,omitempty"`
	BidViewabilityScore *ExtBidViewabilityScore `json:"bidViewability,omitempty"`
}

// ExtImpPubmaticKeyVal defines the contract for bidrequest.imp[i].ext.pubmatic.keywords[i]
type ExtImpPubmaticKeyVal struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"value,omitempty"`
}

//ExtBidViewabilityScore defines the contract for bidrequest.imp[i].ext.pubmatic.bidViewability
type ExtBidViewabilityScore struct {
	Rendered      int     `json:"rendered,omitempty"`
	Viewed        int     `json:"viewed,omitempty"`
	CreatedAt     int     `json:"createdAt,omitempty"`
	UpdatedAt     int     `json:"updatedAt,omitempty"`
	LastViewed    float64 `json:"lastViewed,omitempty"`
	TotalViewTime float64 `json:"totalViewTime,omitempty"`
}