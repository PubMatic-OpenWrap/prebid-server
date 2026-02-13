package auction

import (
	"fmt"
	"sort"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

// dynamicPodSelection holds the best combination of bids for a dynamic pod
type dynamicPodSelection struct {
	bids      []*podBid
	totalCPM  float64
	totalDur  int64
	dealCount int
}

// DynamicAdpodAuction performs auction for dynamic adpod where a single impression
// represents the entire pod and multiple bids compete for slots within pod constraints.
// It maximizes pod value while respecting maxseq (max ads) and poddur (total duration) constraints.
func DynamicAdpodAuction(rCtx *models.RequestCtx, podConfig models.AdpodConfig, bidresponse *openrtb2.BidResponse) []error {
	if bidresponse == nil || len(bidresponse.SeatBid) == 0 || len(podConfig.Slots) == 0 {
		return nil
	}

	// For dynamic pod, we expect exactly one flexible slot
	var dynamicSlot *models.SlotConfig
	for i := range podConfig.Slots {
		if podConfig.Slots[i].Flexible {
			dynamicSlot = &podConfig.Slots[i]
			break
		}
	}

	if dynamicSlot == nil {
		return nil
	}

	// Collect and validate all bids for the dynamic slot
	candidates := collectDynamicBids(rCtx, dynamicSlot, bidresponse)
	if len(candidates) == 0 {
		return nil
	}

	var supportDeals bool
	if rCtx.NewReqExt != nil {
		supportDeals = rCtx.NewReqExt.Prebid.SupportDeals
	}

	// Sort bids: by deal tier (if supportDeals), then by price descending
	sortDynamicBids(candidates, supportDeals)

	// Find best combination using knapsack-like approach
	best := findBestDynamicCombination(
		candidates,
		dynamicSlot.MaxSeq,
		dynamicSlot.PodDur,
		podConfig.Exclusion,
		supportDeals,
	)

	if best == nil || len(best.bids) == 0 {
		return []error{fmt.Errorf("no valid bids found for dynamic adpod")}
	}

	// Build winner set for quick lookup
	winnerSet := make(map[string]struct{}, len(best.bids))
	for _, wb := range best.bids {
		winnerSet[wb.ID] = struct{}{}
	}

	// Record winning bids
	for _, bid := range best.bids {
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
		rCtx.WinningBids.AppendBid(bid.ImpID, owBid)
	}

	// Mark losing bids with appropriate NBR
	markDynamicLosingBids(rCtx, candidates, winnerSet, best, supportDeals)

	// Update NBRs in bid context
	updateDynamicBidCtx(rCtx, candidates)

	return nil
}

// collectDynamicBids collects and validates bids for a dynamic slot
func collectDynamicBids(rCtx *models.RequestCtx, slot *models.SlotConfig, bidresponse *openrtb2.BidResponse) []*podBid {
	var candidates []*podBid

	for _, seatBid := range bidresponse.SeatBid {
		for _, bid := range seatBid.Bid {
			// Only consider bids for this slot's impression
			if bid.ImpID != slot.Id {
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

			// Validate price and duration
			if bid.Price <= 0 || bid.Dur <= 0 {
				bidCtx.Nbr = ptrutil.ToPtr(exchange.ResponseRejectedInvalidCreative)
				impCtx.BidCtx[bid.ID] = bidCtx
				rCtx.ImpBidCtx[bid.ImpID] = impCtx
				continue
			}

			// Validate duration against slot constraints
			if !durationOK(bid.Dur, *slot) {
				bidCtx.Nbr = ptrutil.ToPtr(exchange.ResponseRejectedInvalidCreative)
				impCtx.BidCtx[bid.ID] = bidCtx
				rCtx.ImpBidCtx[bid.ImpID] = impCtx
				continue
			}

			var dealTierSatisfied bool
			if bidCtx.Prebid != nil {
				dealTierSatisfied = bidCtx.Prebid.DealTierSatisfied
			}

			candidates = append(candidates, newPodBid(bid, dealTierSatisfied))
		}
	}

	return candidates
}

// sortDynamicBids sorts bids by deal tier (if enabled), then by CPM/duration ratio descending
func sortDynamicBids(bids []*podBid, supportDeals bool) {
	sort.Slice(bids, func(i, j int) bool {
		if supportDeals {
			// Deal-satisfied bids come first
			if bids[i].DealTierSatisfied != bids[j].DealTierSatisfied {
				return bids[i].DealTierSatisfied
			}
		}
		// Then sort by CPM per second (value density) for better knapsack packing
		cpmPerSecI := bids[i].Price / float64(bids[i].Dur)
		cpmPerSecJ := bids[j].Price / float64(bids[j].Dur)
		if cpmPerSecI != cpmPerSecJ {
			return cpmPerSecI > cpmPerSecJ
		}
		// Tie-breaker: prefer higher absolute price
		return bids[i].Price > bids[j].Price
	})
}

// findBestDynamicCombination finds the best combination of bids that maximizes pod value
// while respecting maxseq and poddur constraints using branch-and-bound approach
func findBestDynamicCombination(
	candidates []*podBid,
	maxSeq int64,
	podDur int64,
	exclusion models.ExclusionConfig,
	supportDeals bool,
) *dynamicPodSelection {
	best := &dynamicPodSelection{dealCount: -1}

	// Use branch-and-bound with pruning for efficiency
	var search func(
		idx int,
		current []*podBid,
		curCPM float64,
		curDur int64,
		curDeals int,
		usedDom, usedCat map[string]struct{},
	)

	search = func(
		idx int,
		current []*podBid,
		curCPM float64,
		curDur int64,
		curDeals int,
		usedDom, usedCat map[string]struct{},
	) {
		// Check if current selection is better than best
		if len(current) > 0 {
			if isBetterDynamicSelection(curCPM, curDeals, best.totalCPM, best.dealCount, supportDeals) {
				best.totalCPM = curCPM
				best.totalDur = curDur
				best.dealCount = curDeals
				best.bids = append([]*podBid(nil), current...)
			}
		}

		// Pruning: if we've reached max ads or processed all candidates, stop
		if int64(len(current)) >= maxSeq || idx >= len(candidates) {
			return
		}

		// Try each remaining candidate
		for i := idx; i < len(candidates); i++ {
			c := candidates[i]

			// Check duration constraint
			if curDur+c.Dur > podDur {
				continue
			}

			// Check exclusion constraints
			if !exclusionSatisfied(exclusion, c, usedDom, usedCat) {
				continue
			}

			// Clone exclusion sets
			nd := deepCloneMap(usedDom)
			nc := deepCloneMap(usedCat)
			if exclusion.AdvertiserDomainExclusion {
				for _, d := range c.ADomain {
					nd[d] = struct{}{}
				}
			}
			if exclusion.IABCategoryExclusion {
				for _, cat := range c.Cat {
					nc[cat] = struct{}{}
				}
			}

			nextDeals := curDeals
			if supportDeals && c.DealTierSatisfied {
				nextDeals++
			}

			search(
				i+1,
				append(current, c),
				curCPM+c.Price,
				curDur+c.Dur,
				nextDeals,
				nd,
				nc,
			)
		}
	}

	search(0, nil, 0, 0, 0, map[string]struct{}{}, map[string]struct{}{})

	if len(best.bids) == 0 {
		return nil
	}
	return best
}

// isBetterDynamicSelection compares two selections and returns true if new is better
func isBetterDynamicSelection(newCPM float64, newDeals int, bestCPM float64, bestDeals int, supportDeals bool) bool {
	if !supportDeals {
		// Pure price-based comparison
		return newCPM > bestCPM
	}
	// Deal-based: prefer more deals, then higher CPM
	if newDeals > bestDeals {
		return true
	}
	if newDeals == bestDeals && newCPM > bestCPM {
		return true
	}
	return false
}

// markDynamicLosingBids marks all non-winning bids with appropriate NBR
func markDynamicLosingBids(
	rCtx *models.RequestCtx,
	candidates []*podBid,
	winnerSet map[string]struct{},
	best *dynamicPodSelection,
	supportDeals bool,
) {
	// Find the highest deal-satisfied winner for comparison
	var hasWinnerWithDeal bool
	for _, wb := range best.bids {
		if wb.DealTierSatisfied {
			hasWinnerWithDeal = true
			break
		}
	}

	for _, c := range candidates {
		if c.Nbr != nil {
			// Already has NBR (e.g., invalid creative)
			continue
		}
		if _, isWinner := winnerSet[c.ID]; isWinner {
			continue
		}

		// Determine loss reason
		if !supportDeals || !hasWinnerWithDeal {
			c.Nbr = ptrutil.ToPtr(nbr.LossBidLostToHigherBid)
		} else if hasWinnerWithDeal && !c.DealTierSatisfied {
			c.Nbr = ptrutil.ToPtr(nbr.LossBidLostToDealBid)
		} else {
			c.Nbr = ptrutil.ToPtr(nbr.LossBidLostToHigherBid)
		}
	}
}

// updateDynamicBidCtx updates the bid context with NBRs for all candidates
func updateDynamicBidCtx(rCtx *models.RequestCtx, candidates []*podBid) {
	for _, c := range candidates {
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
