package openrtb_ext

import (
	"encoding/json"
)

// ExtImpPrebid defines the contract for bidrequest.imp[i].ext.prebid
type ExtImpPrebid struct {
	// StoredRequest specifies which stored impression to use, if any.
	StoredRequest *ExtStoredRequest `json:"storedrequest"`

	// IsRewardedInventory is a signal intended for video impressions. Must be 0 or 1.
	IsRewardedInventory int8 `json:"is_rewarded_inventory"`

	// Bidder is the preferred approach for providing paramters to be interepreted by the bidder's adapter.
	Bidder map[string]json.RawMessage `json:"bidder"`

	SKAdnetwork json.RawMessage `json:"skadn,omitempty"`
}

// ExtStoredRequest defines the contract for bidrequest.imp[i].ext.prebid.storedrequest
type ExtStoredRequest struct {
	ID string `json:"id"`
}

type ExtImpContext struct {
	Data *ExtData `json:"data"`
}

type ExtData struct {
	AdServer *ExtAdServer `json:"adserver"`
}

type ExtAdServer struct {
	Name   string `json:"name"`
	AdSlot string `json:"adslot"`
}
