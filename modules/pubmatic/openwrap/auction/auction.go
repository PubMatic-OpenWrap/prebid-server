package auction

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
)

/*
At this point of time,
 1. For price-based auction (request with supportDeals = false),
    all rejected bids will have NonBR code as LossLostToHigherBid which is expected.
 2. For request with supportDeals = true :
    2.1) If all bids are non-deal-bids (bidExt.Prebid.DealTierSatisfied = false)
    then NonBR code for them will be LossLostToHigherBid which is expected.
    2.2) If one of the bid is deal-bid (bidExt.Prebid.DealTierSatisfied = true)
    expectation:
    all rejected non-deal bids should have NonBR code as LossLostToDealBid
    all rejected deal-bids should have NonBR code as LossLostToHigherBid
    addLostToDealBidNonBRCode function will make sure that above expectation are met.
*/
func Auction(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) {
	anyDealTierSatisfyingBid := false
	winningBids := rctx.WinningBids
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impId := bid.ImpID

			// skip adpod bids
			// TODO: Check in slot config for structured and hybrid adpod
			if _, ok := rctx.AdpodCtx[impId]; ok {
				continue
			}

			impCtx, ok := rctx.ImpBidCtx[impId]
			if !ok {
				continue
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			bidExt := bidCtx.BidExt

			bidDealTierSatisfied := false
			if bidExt.Prebid != nil {
				bidDealTierSatisfied = bidExt.Prebid.DealTierSatisfied
				if bidDealTierSatisfied {
					anyDealTierSatisfyingBid = true // found at least one bid which satisfies dealTier
				}
			}

			owbid := models.OwBid{
				ID:                   bid.ID,
				NetEcpm:              bidExt.NetECPM,
				BidDealTierSatisfied: bidDealTierSatisfied,
			}

			var wbid models.OwBid
			var wbids []*models.OwBid
			var oldWinBidFound bool

			wbids, oldWinBidFound = winningBids[bid.ImpID]
			if len(wbids) > 0 {
				wbid = *wbids[0]
			}
			if !oldWinBidFound {
				winningBids[bid.ImpID] = make([]*models.OwBid, 1)
				winningBids[bid.ImpID][0] = &owbid
			} else if models.IsNewWinningBid(&owbid, &wbid, rctx.SupportDeals) {
				winningBids[bid.ImpID][0] = &owbid
			}

			// update NonBr codes for current bid
			if owbid.Nbr != nil {
				bidExt.Nbr = owbid.Nbr
			}

			// if current bid is winner then update NonBr code for earlier winning bid
			if winningBids.IsWinningBid(impId, owbid.ID) && oldWinBidFound {
				winBidCtx := rctx.ImpBidCtx[impId].BidCtx[wbid.ID]
				winBidCtx.BidExt.Nbr = wbid.Nbr
				rctx.ImpBidCtx[impId].BidCtx[wbid.ID] = winBidCtx
			}

			bidCtx.BidExt = bidExt
			rctx.ImpBidCtx[impId].BidCtx[bid.ID] = bidCtx
		}
	}

	rctx.WinningBids = winningBids

	if anyDealTierSatisfyingBid {
		addLostToDealBidNonBRCode(&rctx)
	}

}

// addLostToDealBidNonBRCode function sets the NonBR code of all lost-bids not satisfying dealTier to LossBidLostToDealBid
func addLostToDealBidNonBRCode(rctx *models.RequestCtx) {
	if !rctx.SupportDeals {
		return
	}

	for impID, impCtx := range rctx.ImpBidCtx {
		// Do not update the nbr in case of adpod bids
		if impCtx.AdpodConfig != nil {
			continue
		}

		_, ok := rctx.WinningBids[impID]
		if !ok {
			continue
		}

		for bidID, bidCtx := range impCtx.BidCtx {
			// do not update NonBR for winning bid
			if rctx.WinningBids.IsWinningBid(impID, bidID) {
				continue
			}

			bidDealTierSatisfied := false
			if bidCtx.BidExt.Prebid != nil {
				bidDealTierSatisfied = bidCtx.BidExt.Prebid.DealTierSatisfied
			}
			// do not update NonBr if lost-bid satisfies dealTier
			// because it can have NonBr reason as LossBidLostToHigherBid
			if bidDealTierSatisfied {
				continue
			}
			bidCtx.BidExt.Nbr = nbr.LossBidLostToDealBid.Ptr()
			rctx.ImpBidCtx[impID].BidCtx[bidID] = bidCtx
		}
	}
}
