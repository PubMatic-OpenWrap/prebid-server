package auction

import (
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

type structuredPodSelection struct {
	bids      []*podBid
	totalCPM  float64
	dealCount int
}

func validateAndCollectBids(rCtx *models.RequestCtx, podCfg models.AdpodConfig, bidresponse *openrtb2.BidResponse) [][]*podBid {
	// map impID -> slot index
	slotIdxByImp := make(map[string]int, len(podCfg.Slots))
	for i := range podCfg.Slots {
		slotIdxByImp[podCfg.Slots[i].Id] = i
	}

	bidsPerSlot := make([][]*podBid, len(podCfg.Slots))
	for _, seatBid := range bidresponse.SeatBid {
		for _, bid := range seatBid.Bid {
			idx, ok := slotIdxByImp[bid.ImpID]
			if !ok {
				continue
			}

			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			if bid.Price <= 0 || bid.Dur <= 0 {
				bidCtx.Nbr = ptrutil.ToPtr(exchange.ResponseRejectedInvalidCreative)
				impCtx.BidCtx[bid.ID] = bidCtx
				rCtx.ImpBidCtx[bid.ImpID] = impCtx
				continue
			}

			if !durationOK(bid.Dur, podCfg.Slots[idx]) {
				bidCtx.Nbr = ptrutil.ToPtr(exchange.ResponseRejectedInvalidCreative)
				impCtx.BidCtx[bid.ID] = bidCtx
				rCtx.ImpBidCtx[bid.ImpID] = impCtx
				continue
			}

			var dealtierSatisfied bool
			if bidCtx.Prebid != nil {
				dealtierSatisfied = bidCtx.Prebid.DealTierSatisfied
			}

			bidsPerSlot[idx] = append(bidsPerSlot[idx], newPodBid(bid, dealtierSatisfied))
		}
	}

	return bidsPerSlot
}

func getBestAdpodCombination(podCfg models.AdpodConfig, candsPerSlot [][]*podBid, supportDeals bool) *structuredPodSelection {
	best := &structuredPodSelection{dealCount: -1}

	var dfs func(
		slotIdx int,
		usedDom, usedCat map[string]struct{},
		current []*podBid,
		curCPM float64,
		curDeals int,
	)

	dfs = func(
		slotIdx int,
		usedDom, usedCat map[string]struct{},
		current []*podBid,
		curCPM float64,
		curDeals int,
	) {
		if slotIdx == len(podCfg.Slots) {
			if !supportDeals {
				// supportdeals=false → pure price-based
				if curCPM > best.totalCPM {
					best.totalCPM = curCPM
					best.bids = append([]*podBid(nil), current...)
				}
				return
			}

			// supportdeals=true → prefer more DealTierSatisfied bids, then price
			if curDeals > best.dealCount ||
				(curDeals == best.dealCount && curCPM > best.totalCPM) {
				best.dealCount = curDeals
				best.totalCPM = curCPM
				best.bids = append([]*podBid(nil), current...)
			}
			return
		}

		// Option 1: skip this slot
		dfs(slotIdx+1, usedDom, usedCat, current, curCPM, curDeals)

		// Option 2: try each candidate for this slot
		for _, c := range candsPerSlot[slotIdx] {
			if !exclusionSatisfied(podCfg.Exclusion, c, usedDom, usedCat) {
				continue
			}

			nd := deepCloneMap(usedDom)
			nc := deepCloneMap(usedCat)
			if podCfg.Exclusion.AdvertiserDomainExclusion {
				for _, d := range c.Bid.ADomain {
					nd[d] = struct{}{}
				}
			}
			if podCfg.Exclusion.IABCategoryExclusion {
				for _, cat := range c.Bid.Cat {
					nc[cat] = struct{}{}
				}
			}

			nextDeals := curDeals
			// Count only tier‑satisfied bids when supportdeals is enabled
			if supportDeals && c.DealTierSatisfied {
				nextDeals++
			}

			dfs(
				slotIdx+1,
				nd,
				nc,
				append(current, c),
				curCPM+c.Bid.Price,
				nextDeals,
			)
		}
	}

	dfs(0, map[string]struct{}{}, map[string]struct{}{}, nil, 0, 0)

	if len(best.bids) == 0 {
		return nil
	}
	return best
}

func StructuredAdpodAuction(rCtx *models.RequestCtx, bidresponse *openrtb2.BidResponse, podConfig models.AdpodConfig) []error {
	if bidresponse == nil || len(bidresponse.SeatBid) == 0 || len(podConfig.Slots) == 0 {
		return nil
	}

	// filter and collect bids
	bidsPerSlot := validateAndCollectBids(rCtx, podConfig, bidresponse)
	if len(bidsPerSlot) == 0 {
		return nil
	}

	var supportDeals bool
	if rCtx.NewReqExt != nil {
		supportDeals = rCtx.NewReqExt.Prebid.SupportDeals
	}

	best := getBestAdpodCombination(podConfig, bidsPerSlot, supportDeals)
	if best == nil {
		return []error{fmt.Errorf("no valid adpod combination found")}
	}

	// 1) Map winners per imp
	winnersByImp := make(map[string]*podBid, len(best.bids))
	for _, wb := range best.bids {
		winnersByImp[wb.ImpID] = wb
	}

	// Form OW winning bids
	for impId, bid := range winnersByImp {
		impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
		if !ok {
			continue
		}
		bidCtx, ok := impCtx.BidCtx[bid.ID]
		if !ok {
			continue
		}
		owBid := &models.OwBid{
			ID:      bid.ID,
			NetEcpm: bidCtx.EN,
		}
		if bidCtx.BidExt.Prebid != nil {
			owBid.BidDealTierSatisfied = bidCtx.BidExt.Prebid.DealTierSatisfied
		}
		rCtx.WinningBids.AppendBid(impId, owBid)
	}

	// 2) For every candidate, if it’s not a winner and has no NBR yet,
	//    mark it as lost to deal or lost to higher bid.
	for _, slotCands := range bidsPerSlot {
		for _, c := range slotCands {
			// skip if this bid already has an NBR (e.g., exclusion/invalid)
			if c.Nbr != nil {
				continue
			}
			winner, ok := winnersByImp[c.ImpID]
			if !ok || winner.ID == c.ID {
				// no winner for this imp, or this bid IS the winner
				continue
			}

			// If supportDeals is false, always treat as price-based loss.
			if !rCtx.SupportDeals {
				c.Nbr = ptrutil.ToPtr(nbr.LossBidLostToHigherBid)
				continue
			}

			// supportDeals == true:
			// Winner satisfied deal tier and loser did not → lost to deal bid.
			// Otherwise → lost to higher bid.
			if winner.DealTierSatisfied && !c.DealTierSatisfied {
				c.Nbr = ptrutil.ToPtr(nbr.LossBidLostToDealBid)
			} else {
				c.Nbr = ptrutil.ToPtr(nbr.LossBidLostToHigherBid)
			}
		}
	}

	// Update the NBRS in the bid Ctx
	for _, slotCands := range bidsPerSlot {
		for _, c := range slotCands {
			impCtx, ok := rCtx.ImpBidCtx[c.ImpID]
			if !ok {
				continue
			}
			bidCtx, ok := impCtx.BidCtx[c.ID]
			if !ok {
				continue
			}
			bidCtx.Nbr = c.Nbr
			impCtx.BidCtx[c.ID] = bidCtx
			rCtx.ImpBidCtx[c.ImpID] = impCtx
		}
	}

	return nil
}
