package feature

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/parser"
)

func EnrichVASTForSSUFeature(bidResponse *openrtb2.BidResponse, vastXMLHandler parser.VASTXMLHandler) {
	for _, seatBid := range bidResponse.SeatBid {
		for i := range seatBid.Bid {
			bid := &seatBid.Bid[i]
			if bid.AdM == "" {
				continue
			}
			bid.AdM = UpdateADMWithAdvCat(vastXMLHandler, bid.AdM, bid.Cat, bid.ADomain)
		}
	}
}

func UpdateADMWithAdvCat(vastXMLHandler parser.VASTXMLHandler, adm string, categories []string, advertiser []string) string {
	if err := vastXMLHandler.Parse(adm); err != nil {
		return adm
	}
	if len(categories) > 0 {
		updatedAdm, err := vastXMLHandler.AddCategoryTag(categories)
		if err == nil {
			adm = updatedAdm
		}
	}

	if len(advertiser) > 0 {
		updatedAdm, err := vastXMLHandler.AddAdvertiserTag(advertiser[0])
		if err == nil {
			adm = updatedAdm
		}
	}
	return adm
}
