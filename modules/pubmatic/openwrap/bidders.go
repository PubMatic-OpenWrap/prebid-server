package openwrap

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

var alias = map[string]string{
	models.BidderAdGenerationAlias:      string(openrtb_ext.BidderAdgeneration),
	models.BidderDistrictmDMXAlias:      string(openrtb_ext.BidderDmx),
	models.BidderPubMaticSecondaryAlias: string(openrtb_ext.BidderPubmatic),
	models.BidderDistrictmAlias:         string(openrtb_ext.BidderAppnexus),
	models.BidderAndBeyondAlias:         string(openrtb_ext.BidderAdkernel),
	models.BidderMediaFuseAlias:         string(openrtb_ext.BidderAppnexus),
}

// IsAlias will return copy of exisiting alias
func IsAlias(bidder string) (string, bool) {
	v, ok := alias[bidder]
	return v, ok
}
