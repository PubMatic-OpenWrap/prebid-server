package feature

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/parser"
)

func EnrichVASTWithSSUFeature(bidResponse *openrtb2.BidResponse, vastXMLHandler parser.VASTXMLHandler) {
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

func UpdateADMWithAdvCat(vastXMLHandler parser.VASTXMLHandler, AdM string, cat []string, adomain []string) string {
	if err := vastXMLHandler.Parse(AdM); err != nil {
		return AdM
	}
	if len(cat) > 0 {
		adm, err := vastXMLHandler.AddCategoryTag(cat)
		if err == nil {
			AdM = adm
		}
	}

	if len(adomain) > 0 {
		adm, err := vastXMLHandler.AddAdvertiserTag(adomain[0])
		if err == nil {
			AdM = adm
		}
	}
	return AdM
}
