package adpod

import (
	"encoding/json"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/v2/metrics"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

type structuredAdpod struct {
	AdpodCtx
	ImpBidMap          map[string][]*types.Bid
	WinningBid         map[string]types.Bid
	SelectedCategories map[string]bool
	SelectedDomains    map[string]bool
}

type Slot struct {
	ImpId     string
	Index     int
	TotalBids int
}

func NewStructuredAdpod(pubId string, metricsEngine metrics.MetricsEngine, reqAdpodExt *openrtb_ext.ExtRequestAdPod) *structuredAdpod {
	adpod := structuredAdpod{
		AdpodCtx: AdpodCtx{
			PubId:         pubId,
			Type:          Structured,
			ReqAdpodExt:   reqAdpodExt,
			MetricsEngine: metricsEngine,
		},
		ImpBidMap:  make(map[string][]*types.Bid),
		WinningBid: make(map[string]types.Bid),
	}

	return &adpod
}

func (da *structuredAdpod) GetPodType() PodType {
	return da.Type
}

func (sa *structuredAdpod) AddImpressions(imp openrtb2.Imp) {
	sa.Imps = append(sa.Imps, imp)
}

func (sa *structuredAdpod) GetImpressions() []openrtb2.Imp {
	return sa.Imps
}

func (sa *structuredAdpod) CollectBid(bid *openrtb2.Bid, seat string) {
	ext := openrtb_ext.ExtBid{}
	if bid.Ext != nil {
		json.Unmarshal(bid.Ext, &ext)
	}

	adpodBid := types.Bid{
		Bid:               bid,
		ExtBid:            ext,
		DealTierSatisfied: util.GetDealTierSatisfied(&ext),
		Seat:              seat,
	}
	bids := sa.ImpBidMap[bid.ImpID]

	bids = append(bids, &adpodBid)
	sa.ImpBidMap[bid.ImpID] = bids
}

func (sa *structuredAdpod) HoldAuction() {
	if len(sa.ImpBidMap) == 0 {
		return
	}

	// Sort Bids impression wise
	for _, bids := range sa.ImpBidMap {
		util.SortBids(bids)
	}

	// Create Slots
	slots := make([]Slot, 0)

	for impId, bids := range sa.ImpBidMap {
		slot := Slot{
			ImpId:     impId,
			Index:     0,
			TotalBids: len(bids),
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

func (sa *structuredAdpod) Validate() []error {
	return nil
}

func (sa *structuredAdpod) GetAdpodSeatBids() []openrtb2.SeatBid {
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

func (sa *structuredAdpod) GetAdpodExtension(blockedVastTagID map[string]map[string][]string) *types.ImpData {
	return nil
}

/****************************Exclusion*******************************/

func (sa *structuredAdpod) addCategories(categories []string) {
	if sa.SelectedCategories == nil {
		sa.SelectedCategories = make(map[string]bool)
	}

	for _, cat := range categories {
		sa.SelectedCategories[cat] = true
	}
}

func (sa *structuredAdpod) addDomains(domains []string) {
	if sa.SelectedDomains == nil {
		sa.SelectedDomains = make(map[string]bool)
	}

	for _, domain := range domains {
		sa.SelectedDomains[domain] = true
	}
}

func (sa *structuredAdpod) isCategoryAlreadySelected(bid *types.Bid) bool {
	if bid == nil || bid.Cat == nil {
		return false
	}

	if sa.SelectedCategories == nil {
		return false
	}

	for i := range bid.Cat {
		if _, ok := sa.SelectedCategories[bid.Cat[i]]; ok {
			return true
		}
	}

	return false
}

func (sa *structuredAdpod) isDomainAlreadySelected(bid *types.Bid) bool {
	if bid == nil || bid.ADomain == nil {
		return false
	}

	if sa.SelectedDomains == nil {
		return false
	}

	for i := range bid.ADomain {
		if _, ok := sa.SelectedDomains[bid.ADomain[i]]; ok {
			return true
		}
	}

	return false
}

/************Structured Adpod Auction Methods***********************/

func (sa *structuredAdpod) selectBidForSlot(slots []Slot) {
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
		} else if sa.isCategoryAlreadySelected(selectedBid) || sa.isDomainAlreadySelected(selectedBid) {
			// Get bid for current slot for which category is not overlapping
			for i := selectedSlot.Index + 1; i < len(slotBids); i++ {
				if !sa.isCategoryAlreadySelected(slotBids[i]) && !sa.isDomainAlreadySelected(slotBids[i]) {
					selectedSlot.Index = i
					break
				}
			}

			// Update selected Slot in slots array
			slots[slotIndex] = selectedSlot
		}
	}

	// Add bid categories to selected categories
	sa.addCategories(slotBids[selectedSlot.Index].Cat)
	sa.addDomains(slotBids[selectedSlot.Index].ADomain)

	// Swap selected slot at initial position
	slots[0], slots[slotIndex] = slots[slotIndex], slots[0]

	sa.selectBidForSlot(slots[1:])
}

func (sa *structuredAdpod) getSlotIndexWithHighestBid(slots []Slot) int {
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

func (sa *structuredAdpod) isDealBidCatOverlapWithAnotherDealBid(slots []Slot, selectedSlotIndex int, selectedBid *types.Bid) bool {
	if len(selectedBid.Cat) == 0 {
		return false
	}

	catMap := make(map[string]bool)
	for _, cat := range selectedBid.Cat {
		catMap[cat] = true
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

func (sa *structuredAdpod) isBetterBidThanDeal(slots []Slot, selectedSlotIndx int, selectedSlot Slot, selectedBid *types.Bid) (int, bool) {
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
