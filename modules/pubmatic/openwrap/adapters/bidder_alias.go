package adapters

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

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
	case models.BidderDistrictmAlias:
		coreBidderName = string(openrtb_ext.BidderAppnexus)
	case models.BidderAndBeyondAlias:
		coreBidderName = string(openrtb_ext.BidderAdkernel)
	default:
		coreBidderName = bidderName

	}

	return coreBidderName
}
