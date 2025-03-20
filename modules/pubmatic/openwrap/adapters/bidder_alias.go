package adapters

import (
	"strings"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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
		models.BidderAppStockAlias:          string(openrtb_ext.BidderLimelightDigital),
		models.BidderAdsYieldAlias:          string(openrtb_ext.BidderLimelightDigital),
		models.BidderIionadsAlias:           string(openrtb_ext.BidderLimelightDigital),
	}
}

//ResolveOWBidder it resolves hardcoded bidder alias names

func ResolveOWBidder(bidderName string) string {
	var coreBidderName = bidderName
	bidderName = strings.TrimSuffix(bidderName, "_deprecated")

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
	case models.BidderTrustxAlias:
		coreBidderName = string(openrtb_ext.BidderGrid)
	case models.BidderSynacormediaAlias:
		coreBidderName = string(openrtb_ext.BidderImds)
	case models.BidderViewDeos:
		coreBidderName = string(openrtb_ext.BidderAdtelligent)
	case models.BidderAppStockAlias, models.BidderAdsYieldAlias, models.BidderIionadsAlias:
		coreBidderName = string(openrtb_ext.BidderLimelightDigital)
	}
	return coreBidderName
}
