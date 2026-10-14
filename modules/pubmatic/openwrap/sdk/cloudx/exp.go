package cloudx

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
)

const cloudXDefaultExp = 3000

// applyCloudXBidExpIfMissing sets bid.exp from outbound imp.exp when the partner omits it.
// When imp.exp is missing, bid.exp defaults to 3000.
func applyCloudXBidExpIfMissing(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) {
	if bidResponse == nil {
		return
	}

	for i := range bidResponse.SeatBid {
		for j := range bidResponse.SeatBid[i].Bid {
			bid := &bidResponse.SeatBid[i].Bid[j]
			if bid.Exp > 0 {
				continue
			}
			impID, _ := models.GetImpressionID(bid.ImpID)
			if impCtx, ok := rctx.ImpBidCtx[impID]; ok && impCtx.Exp > 0 {
				bid.Exp = impCtx.Exp
				continue
			}
			bid.Exp = cloudXDefaultExp
		}
	}
}
