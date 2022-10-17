package floors

import (
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// PriceFloorRules defines the contract for bidrequest.ext.prebid.floors
type fetchReult struct {
	priceFloors openrtb_ext.PriceFloorRules `json:"pricefloors,omitempty"`
	fetchStatus int                         `json:"fetchstatus,omitempty"`
}

func fetchFloors(Account config.Account) *fetchReult {
	var fetchedResults fetchReult
	fetchedResults.fetchStatus = -1
	return &fetchedResults
}
