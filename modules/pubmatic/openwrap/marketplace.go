package openwrap

import (
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// TODO: Make this generic implementation
func getMarketplaceBidders(reqABC *openrtb_ext.ExtAlternateBidderCodes, partnerConfigMap map[int]map[string]string) (*openrtb_ext.ExtAlternateBidderCodes, map[string]struct{}) {
	if reqABC != nil {
		return reqABC, nil
	}

	// string validations, etc will be done by api-wrapper-tag. Not need to repetitively do the typical string validations
	marketplaceBiddersDB := partnerConfigMap[models.VersionLevelConfigID][models.MarketplaceBidders]
	if len(marketplaceBiddersDB) == 0 {
		return nil, nil
	}
	marketplaceBidders := strings.Split(marketplaceBiddersDB, ",")

	bidderMap := make(map[string]struct{})
	for _, bidder := range marketplaceBidders {
		bidderMap[bidder] = struct{}{}
	}

	return &openrtb_ext.ExtAlternateBidderCodes{
		Enabled: true,
		Bidders: map[string]openrtb_ext.ExtAdapterAlternateBidderCodes{
			string(openrtb_ext.BidderPubmatic): { // How do we get non-pubmatic bidders?. Does core-platform even have it?
				Enabled:            true,
				AllowedBidderCodes: marketplaceBidders,
			},
		},
	}, bidderMap
}
