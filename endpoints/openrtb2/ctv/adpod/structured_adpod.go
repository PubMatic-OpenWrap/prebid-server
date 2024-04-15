package adpod

import (
	"encoding/json"

	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type StructuredAdpod struct {
	AdpodCtx
	ImpBidMap         map[string][]*types.Bid
	WinningBid        map[string]types.Bid
	CategoryExclusion bool
}

type Slot struct {
	ImpId string
	Index int
}

func (da *StructuredAdpod) GetPodType() PodType {
	return da.Type
}

func (sa *StructuredAdpod) AddImpressions(imp openrtb2.Imp) {
	sa.Imps = append(sa.Imps, imp)
}

func (sa *StructuredAdpod) GenerateImpressions() {
	// We do not generate impressions in case of structured adpod
}

func (sa *StructuredAdpod) GetImpressions() []openrtb2.Imp {
	return sa.Imps
}

func (sa *StructuredAdpod) CollectBid(bid openrtb2.Bid, seat string) {
	ext := openrtb_ext.ExtBid{}
	if bid.Ext != nil {
		json.Unmarshal(bid.Ext, &ext)
	}

	adpodBid := types.Bid{
		Bid:               &bid,
		ExtBid:            ext,
		DealTierSatisfied: util.GetDealTierSatisfied(&ext),
		Seat:              seat,
	}
	bids := sa.ImpBidMap[bid.ImpID]

	bids = append(bids, &adpodBid)
	sa.ImpBidMap[bid.ImpID] = bids
}

func (sa *StructuredAdpod) PerformAuctionAndExclusion() {
	if len(sa.ImpBidMap) == 0 {
		return
	}

	// Sort Bids impression wise
	for impId, bids := range sa.ImpBidMap {
		util.SortBids(bids)
		sa.ImpBidMap[impId] = bids
	}

	// Fill up exclusion conditions
	sa.Exclusion.setExclusionConditions(len(sa.ImpBidMap))

	// Create Slots
	slots := make([]Slot, 0)

	for impId, bids := range sa.ImpBidMap {
		if len(bids) == 0 {
			continue
		}

		slot := Slot{
			ImpId: impId,
			Index: 0,
		}
		slots = append(slots, slot)
	}

	sa.selectBidForSlot(slots)

	// Select Winning bids
	for i := range slots {
		bids := sa.ImpBidMap[slots[i].ImpId]
		// Add validations on len of array and index chosen
		sa.WinningBid[slots[i].ImpId] = *bids[slots[i].Index]
	}

}

func (sa *StructuredAdpod) Validate() []error {
	return nil
}

func (sa *StructuredAdpod) GetAdpodSeatBids() []openrtb2.SeatBid {
	if len(sa.WinningBid) == 0 {
		return nil
	}

	var seatBid []openrtb2.SeatBid
	for _, bid := range sa.WinningBid {
		adpodSeat := openrtb2.SeatBid{
			Bid:  []openrtb2.Bid{*bid.Bid},
			Seat: bid.Seat,
		}
		seatBid = append(seatBid, adpodSeat)
	}

	return seatBid
}

func (sa *StructuredAdpod) GetAdpodExtension(blockedVastTagID map[string]map[string][]string) *types.ImpData {
	return nil
}

/************Structured Adpod Auction Methods***********************/

func (sa *StructuredAdpod) selectBidForSlot(slots []Slot) {
	if len(slots) == 0 {
		return
	}

	slotIndex := sa.getSlotIndexWithHighestBid(slots)

	// Get current bid for selected slot
	selectedSlot := slots[slotIndex]
	slotBids := sa.ImpBidMap[selectedSlot.ImpId]
	selectedBid := slotBids[selectedSlot.Index]

	if sa.Exclusion.shouldApplyExclusion() {
		if bidIndex, ok := sa.isBetterBidThanDeal(slots, slotIndex, selectedSlot, selectedBid); ok {
			selectedSlot.Index = bidIndex
			slots[slotIndex] = selectedSlot
		} else if sa.Exclusion.eitherExclusionNotMetForBid(selectedBid) {
			// Get bid for current slot for which category is not overlapping
			for i := selectedSlot.Index + 1; i < len(slotBids); i++ {
				if sa.Exclusion.allExclusionConditionsSatified(slotBids[i]) {
					selectedSlot.Index = i
					sa.Exclusion.updateExclusionConditions()
					break
				}
			}

			// Update selected Slot in slots array
			slots[slotIndex] = selectedSlot
		}
	} else {
		sa.Exclusion.updateExclusionConditions()
	}

	sa.Exclusion.addExclusionParameters(slotBids[selectedSlot.Index])

	// Swap selected slot at initial position
	slots[0], slots[slotIndex] = slots[slotIndex], slots[0]

	sa.selectBidForSlot(slots[1:])
}

func (sa *StructuredAdpod) getSlotIndexWithHighestBid(slots []Slot) int {
	var index int
	maxBid := &types.Bid{
		Bid: &openrtb2.Bid{},
	}

	for i := range slots {
		impBids := sa.ImpBidMap[slots[i].ImpId]
		bid := impBids[slots[i].Index]

		if bid.DealTierSatisfied == maxBid.DealTierSatisfied {
			if bid.Price > maxBid.Price {
				maxBid = bid
				index = i
			}
		} else if bid.DealTierSatisfied {
			maxBid = bid
			index = i
		}
	}

	return index
}

func isDealBid(bid *types.Bid) bool {
	return bid.DealTierSatisfied
}

func (sa *StructuredAdpod) isDealBidCatOverlapWithAnotherDealBid(slots []Slot, selectedSlotIndex int, selectedBid *types.Bid) bool {
	if exclusionParamsNotExists(selectedBid) {
		return false
	}

	catMap := make(map[string]bool)
	for _, cat := range selectedBid.Cat {
		catMap[cat] = true
	}

	domainMap := make(map[string]bool)
	for _, domain := range selectedBid.ADomain {
		domainMap[domain] = true
	}

	var isCatOverlap bool
	for i := range slots {
		if selectedSlotIndex == i {
			continue
		}
		slotBids := sa.ImpBidMap[slots[i].ImpId]
		bid := slotBids[slots[i].Index]

		for _, cat := range bid.Cat {
			if _, ok := catMap[cat]; ok {
				isCatOverlap = true
				break
			}
		}
		if isCatOverlap {
			break
		}
	}

	return isCatOverlap

}

func isBetterBidAvailable(slotBids []*types.Bid, selectedBid *types.Bid, selectedBidtIndex int) (int, bool) {
	var isBetterBidAvailable bool
	var betterBidIndex int

	catMap := make(map[string]bool)
	for _, cat := range selectedBid.Cat {
		catMap[cat] = true
	}

	for i := selectedBidtIndex + 1; i < len(slotBids); i++ {
		bid := slotBids[i]

		// Next bid should not be deal bid
		if bid.DealTierSatisfied {
			continue
		}

		// Category should not be overlaped
		var isCatOverlap bool
		for _, cat := range bid.Cat {
			if _, ok := catMap[cat]; ok {
				isCatOverlap = true
				break
			}
		}
		if isCatOverlap {
			continue
		}

		// Check for bid price is greater than deal price
		if bid.Price > selectedBid.Price {
			isBetterBidAvailable = true
			betterBidIndex = i
			break
		}

	}

	return betterBidIndex, isBetterBidAvailable
}

func (sa *StructuredAdpod) isBetterBidThanDeal(slots []Slot, selectedSlotIndx int, selectedSlot Slot, selectedBid *types.Bid) (int, bool) {
	selectedBidIndex := selectedSlot.Index

	if !isDealBid(selectedBid) {
		return selectedBidIndex, false
	}

	if !sa.isDealBidCatOverlapWithAnotherDealBid(slots, selectedSlotIndx, selectedBid) {
		return selectedBidIndex, false
	}

	var isBetterBid bool
	selectedBidIndex, isBetterBid = isBetterBidAvailable(sa.ImpBidMap[selectedSlot.ImpId], selectedBid, selectedBidIndex)

	return selectedBidIndex, isBetterBid
}
