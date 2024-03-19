package adapters

import (
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// Alias will return copy of exisiting alias
func Alias() map[string]string {

	return map[string]string{
		models.BidderAdGenerationAlias:      string(openrtb_ext.BidderAdgeneration),
		models.BidderDistrictmDMXAlias:      string(openrtb_ext.BidderDmx),
		models.BidderPubMaticSecondaryAlias: string(openrtb_ext.BidderPubmatic),
		models.BidderDistrictmAlias:         string(openrtb_ext.BidderAppnexus),
		models.BidderAndBeyondAlias:         string(openrtb_ext.BidderAdkernel),
		models.BidderMediaFuseAlias:         string(openrtb_ext.BidderAppnexus),
	}
}

//ResolveOWBidder it resolves hardcoded bidder alias names

func ResolveOWBidder(bidderName string) string {
	var coreBidderName string

	switch bidderName {
	case models.BidderAdGenerationAlias:
		coreBidderName = string(openrtb_ext.BidderAdgeneration)
	case models.BidderDistrictmDMXAlias:
		coreBidderName = string(openrtb_ext.BidderDmx)
	case models.BidderPubMaticSecondaryAlias:
		coreBidderName = string(openrtb_ext.BidderPubmatic)
	case models.BidderDistrictmAlias, models.BidderMediaFuseAlias:
		coreBidderName = string(openrtb_ext.BidderAppnexus)
	case models.BidderAndBeyondAlias:
		coreBidderName = string(openrtb_ext.BidderAdkernel)
	case models.BidderAdformAdfAlias:
		coreBidderName = string(openrtb_ext.BidderAdf)
	default:
		coreBidderName = bidderName
	}
	return coreBidderName
}
