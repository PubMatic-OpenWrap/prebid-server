package floors

import (
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// fetchReult defines the contract for fetched floors results
type fetchReult struct {
	priceFloors openrtb_ext.PriceFloorRules `json:"pricefloors,omitempty"`
	fetchStatus int                         `json:"fetchstatus,omitempty"`
}

// fetchAccountFloors this function fetch floors JSON for given account
var fetchAccountFloors = func(account config.Account) *fetchReult {
	var fetchedResults fetchReult
	fetchedResults.fetchStatus = -1
	return &fetchedResults
}
