package adpod

import (
	"encoding/json"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

type StructuredAdpod struct {
	models.AdpodCtx
	ImpBidMap          map[string][]*types.Bid
	WinningBid         map[string]types.Bid
	SelectedCategories map[string]bool
	SelectedDomains    map[string]bool
}

type Slot struct {
	ImpId     string
	Index     int
	TotalBids int
	NoBid     bool
}

func NewStructuredAdpod(podId string, reqAdPodExt *models.ExtRequestAdPod) *StructuredAdpod {
	adpod := StructuredAdpod{
		AdpodCtx: models.AdpodCtx{
			//PubId: pubId,
			Type: models.Structured,
			//ReqAdpodExt:   reqAdpodExt,
			Exclusion: getExclusionConfigs(podId, reqAdPodExt),
		},
		ImpBidMap:  make(map[string][]*types.Bid),
		WinningBid: make(map[string]types.Bid),
	}

	return &adpod
}

// func (da *StructuredAdpod) GetPodType() PodType {
// 	return da.Type
// }

func (sa *StructuredAdpod) AddImpressions(imp openrtb2.Imp) {
	sa.Imps = append(sa.Imps, imp)
}

func (sa *StructuredAdpod) GetImpressions() []*openrtb_ext.ImpWrapper {

	var imps []*openrtb_ext.ImpWrapper
	for _, imp := range sa.Imps {
		imps = append(imps, &openrtb_ext.ImpWrapper{Imp: &imp})
	}
	return imps
}

func (sa *StructuredAdpod) CollectBid(bid *openrtb2.Bid, seat string) {
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

func (sa *StructuredAdpod) HoldAuction() {
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
		if len(bids) == 0 {
			continue
		}

		slot := slots[i]
		if slot.NoBid {
			continue
		}

		sa.WinningBid[slot.ImpId] = *bids[slot.Index]
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

/****************************Exclusion*******************************/

func (sa *StructuredAdpod) addCategories(categories []string) {
	if sa.SelectedCategories == nil {
		sa.SelectedCategories = make(map[string]bool)
	}

	for _, cat := range categories {
		sa.SelectedCategories[cat] = true
	}
}

func (sa *StructuredAdpod) addDomains(domains []string) {
	if sa.SelectedDomains == nil {
		sa.SelectedDomains = make(map[string]bool)
	}

	for _, domain := range domains {
		sa.SelectedDomains[domain] = true
	}
}

func (sa *StructuredAdpod) isCategoryAlreadySelected(bid *types.Bid) bool {
	if !sa.Exclusion.IABCategoryExclusion {
		return false
	}

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

func (sa *StructuredAdpod) isDomainAlreadySelected(bid *types.Bid) bool {
	if !sa.Exclusion.AdvertiserDomainExclusion {
		return false
	}

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

func (sa *StructuredAdpod) isCatOverlap(cats []string, catMap map[string]bool) bool {
	if !sa.Exclusion.IABCategoryExclusion {
		return false
	}

	return isAtrributesOverlap(cats, catMap)
}

func (sa *StructuredAdpod) isDomainOverlap(domains []string, domainMap map[string]bool) bool {
	if !sa.Exclusion.AdvertiserDomainExclusion {
		return false
	}

	return isAtrributesOverlap(domains, domainMap)
}

func isAtrributesOverlap(attributes []string, checkMap map[string]bool) bool {
	for _, item := range attributes {
		if _, ok := checkMap[item]; ok {
			return true
		}
	}
	return false
}

/*******************Structured Adpod Auction Methods***********************/

func isDealBid(bid *types.Bid) bool {
	return bid.DealTierSatisfied
}

func (sa *StructuredAdpod) isOverlap(bid *types.Bid, catMap map[string]bool, domainMap map[string]bool) bool {
	return sa.isCatOverlap(bid.Cat, catMap) || sa.isDomainOverlap(bid.ADomain, domainMap)
}

func (sa *StructuredAdpod) selectBidForSlot(slots []Slot) {
	if len(slots) == 0 {
		return
	}

	slotIndex := sa.getSlotIndexWithHighestBid(slots)

	// Get current bid for selected slot
	selectedSlot := slots[slotIndex]
	slotBids := sa.ImpBidMap[selectedSlot.ImpId]
	selectedBid := slotBids[selectedSlot.Index]

	if sa.Exclusion.ShouldApplyExclusion() {
		if bidIndex, ok := sa.isBetterBidThanDeal(slots, slotIndex, selectedSlot, selectedBid); ok {
			selectedSlot.Index = bidIndex
			slots[slotIndex] = selectedSlot
		} else if sa.isCategoryAlreadySelected(selectedBid) || sa.isDomainAlreadySelected(selectedBid) {
			noBidSlot := true
			// Get bid for current slot for which category is not overlapping
			for i := selectedSlot.Index + 1; i < len(slotBids); i++ {
				if !sa.isCategoryAlreadySelected(slotBids[i]) && !sa.isDomainAlreadySelected(slotBids[i]) {
					selectedSlot.Index = i
					noBidSlot = false
					break
				}
			}

			// Update no bid status
			selectedSlot.NoBid = noBidSlot

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

// isBetterBidAvailable checks if a better bid is available for the selected slot.
// It returns true if
func (sa *StructuredAdpod) isBetterBidAvailable(slots []Slot, selectedSlotIndex int, selectedBid *types.Bid) bool {
	if len(selectedBid.Cat) == 0 && len(selectedBid.ADomain) == 0 {
		return false
	}

	catMap := createMapFromSlice(selectedBid.Cat)
	domainMap := createMapFromSlice(selectedBid.ADomain)

	return sa.shouldUpdateSelectedBid(slots, selectedSlotIndex, catMap, domainMap)
}

// shouldUpdateSelectedBid checks if a bid should be updated for a selected slot.
func (sa *StructuredAdpod) shouldUpdateSelectedBid(slots []Slot, selectedSlotIndex int, catMap map[string]bool, domainMap map[string]bool) bool {
	for i := range slots {
		if selectedSlotIndex == i {
			continue
		}
		slotBids := sa.ImpBidMap[slots[i].ImpId]
		slotIndex := slots[i].Index

		// Get bid for current slot
		bid := slotBids[slotIndex]

		if bid.DealTierSatisfied && sa.isOverlap(bid, catMap, domainMap) {
			return sa.shouldUpdateBid(slotBids, slotIndex, catMap, domainMap)
		}
	}
	return false
}

// shouldUpdateBid checks if a bid should be updated for a selected slot.
// It iterates through the remaining slot bids of overlapped slot starting from the given slot index,
// and checks exclusions conditions for only deal bids.
// It will ensure more deal bids in final adpod.
func (sa *StructuredAdpod) shouldUpdateBid(slotBids []*types.Bid, slotIndex int, catMap map[string]bool, domainMap map[string]bool) bool {
	for i := slotIndex + 1; i < len(slotBids); i++ {
		bid := slotBids[i]

		if !bid.DealTierSatisfied {
			break
		}

		if !sa.isOverlap(bid, catMap, domainMap) {
			return false
		}
	}
	return true
}

func (sa *StructuredAdpod) getBetterBid(slotBids []*types.Bid, selectedBid *types.Bid, selectedBidtIndex int) (int, bool) {
	catMap := createMapFromSlice(selectedBid.Cat)
	domainMap := createMapFromSlice(selectedBid.ADomain)

	for i := selectedBidtIndex + 1; i < len(slotBids); i++ {
		bid := slotBids[i]

		// Check for deal bid and select if exclusion conditions are satisfied
		if isDealBid(bid) {
			if !sa.isOverlap(bid, catMap, domainMap) {
				return i, true
			}
			continue
		}

		// New selected bid exclusion parameters should not be overlaped
		if sa.isOverlap(bid, catMap, domainMap) {
			continue
		}

		// Check for bid price is greater than deal price
		if bid.Price > selectedBid.Price {
			return i, true
		}
	}

	return selectedBidtIndex, false
}

func (sa *StructuredAdpod) isBetterBidThanDeal(slots []Slot, selectedSlotIndx int, selectedSlot Slot, selectedBid *types.Bid) (int, bool) {
	selectedBidIndex := selectedSlot.Index

	if !isDealBid(selectedBid) {
		return selectedBidIndex, false
	}

	if !sa.isBetterBidAvailable(slots, selectedSlotIndx, selectedBid) {
		return selectedBidIndex, false
	}

	return sa.getBetterBid(sa.ImpBidMap[selectedSlot.ImpId], selectedBid, selectedBidIndex)
}

func createMapFromSlice(slice []string) map[string]bool {
	resultMap := make(map[string]bool)
	for _, item := range slice {
		resultMap[item] = true
	}
	return resultMap
}

func (sa *StructuredAdpod) CollectAPRC(rctx models.RequestCtx) {

}

func (sa *StructuredAdpod) GetWinningBidsIds(rctx models.RequestCtx, winningBidIds map[string][]string) {
	if len(sa.WinningBid) == 0 {
		return
	}

	for _, bid := range sa.WinningBid {
		if len(bid.AdM) == 0 {
			continue
		}
		impCtx, ok := rctx.ImpBidCtx[bid.ImpID]
		if !ok {
			return
		}
		winningBidIds[bid.ImpID] = append(winningBidIds[bid.ImpID], bid.ID)
		impCtx.BidIDToAPRC[bid.ID] = models.StatusWinningBid
		rctx.ImpBidCtx[bid.ImpID] = impCtx
	}
}
