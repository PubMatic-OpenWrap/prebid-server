package adpod

import (
	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/adpod"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func FormAdpodBidsAndPerformExclusion(response *openrtb2.BidResponse, rctx models.RequestCtx) (map[string][]string, []error) {
	var (
		errs []error
	)

	if len(response.SeatBid) == 0 {
		return nil, errs
	}

	collectBids(response, rctx)
	return doAdpodAuction(rctx), nil
	// impAdpodBidsMap, _ := generateAdpodBids(response.SeatBid, rctx.ImpBidCtx, rctx.AdpodProfileConfig)
	// adpodBids, errs := doAdPodExclusions(impAdpodBidsMap, rctx.ImpBidCtx)
	// if len(errs) > 0 {
	// 	return nil, errs
	// }

	// // Record APRC for bids
	// collectAPRC(impAdpodBidsMap, rctx.ImpBidCtx)

	// winningBidIds, err := GetWinningBidsIds(adpodBids, rctx.ImpBidCtx)
	// if err != nil {
	// 	return nil, []error{err}
	// }
}

func collectBids(response *openrtb2.BidResponse, rctx models.RequestCtx) {

	for i := range response.SeatBid {
		seat := response.SeatBid[i]
		for j := range seat.Bid {
			bid := &seat.Bid[j]

			if bid.Price == 0 {
				continue
			}

			if len(bid.ID) == 0 {
				bidId, err := jsonparser.GetString(bid.Ext, "prebid", "bidid")
				if err == nil {
					bid.ID = bidId
				}
			}

			// originalImpID, _ := DecodeImpressionID(bid.ImpID) //TODO: check if we can reomove and maintain map

			value, err := util.GetTargeting(openrtb_ext.HbCategoryDurationKey, openrtb_ext.BidderName(seat.Seat), *bid)
			if nil == err {
				// ignore error
				adpod.AddTargetingKey(bid, openrtb_ext.HbCategoryDurationKey, value)
			}

			value, err = util.GetTargeting(openrtb_ext.HbpbConstantKey, openrtb_ext.BidderName(seat.Seat), *bid)
			if nil == err {
				// ignore error
				adpod.AddTargetingKey(bid, openrtb_ext.HbpbConstantKey, value)
			}

			podId, ok := rctx.ImpToPodId[bid.ImpID]
			if !ok {
				continue
			}

			adpodCtx, ok := rctx.AdpodCtx[podId]
			if !ok {
				continue
			}

			adpodCtx.CollectBid(bid, seat.Seat)
		}
	}
}

func doAdpodAuction(rCtx models.RequestCtx) map[string][]string {
	winningBidIds := map[string][]string{}
	for _, adpodCtx := range rCtx.AdpodCtx {
		adpodCtx.HoldAuction()
		adpodCtx.CollectAPRC(rCtx)
		adpodCtx.GetWinningBidsIds(rCtx, winningBidIds)
	}
	return winningBidIds
}
