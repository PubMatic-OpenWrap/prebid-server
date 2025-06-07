// function to enrich vast xml with google ssu cat and advertiser and call from auctionresp hook and HB

package feature

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/parser"
)

func EnrichVASTWithSSUFeature(bidResponse *openrtb2.BidResponse, vastXMLHandler parser.VASTXMLHandler) {
	for _, seatBid := range bidResponse.SeatBid {
		for i := range seatBid.Bid {
			bid := &seatBid.Bid[i] // use pointer so changes reflect
			if bid.AdM == "" {
				continue
			}

			if err := vastXMLHandler.Parse(bid.AdM); err != nil {
				continue
			}
			if len(bid.Cat) > 0 {
				adm, err := vastXMLHandler.AddCategoryTag(bid.Cat)
				if err == nil {
					bid.AdM = adm
				}
			}

			if len(bid.ADomain) > 0 {
				adm, err := vastXMLHandler.AddAdvertiserTag(bid.ADomain[0])
				if err == nil {
					bid.AdM = adm
				}
			}
		}
	}

}
