package openwrap

import (
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type OpenWrapAdapter struct {
	endpoint  string
}

// Builder builds a new instance of the AdButler onsite adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &OpenWrapAdapter{
		endpoint:    config.Endpoint,
	}
	return bidder, nil
}

