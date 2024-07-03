package adpod

import (
	"encoding/json"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/constant"
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
		// Validate the array length and index chosen
		if len(bids) > slots[i].Index {
			selectedBid := bids[slots[i].Index]
			selectedBid.Status = constant.StatusWinningBid
			sa.WinningBid[slots[i].ImpId] = *selectedBid
		}

	}
}

func (sa *structuredAdpod) Validate() []error {
	return nil
}

func (sa *structuredAdpod) GetWinningBids() []openrtb2.SeatBid {
	return sa.GetAdpodSeatBids()
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

func (sa *structuredAdpod) GetSeatNonBid(snb *openrtb_ext.NonBidCollection) {
	for _, bids := range sa.ImpBidMap {
		for _, bid := range bids {
			if bid.Status != constant.StatusWinningBid {
				nonBidParams := GetNonBidParamsFromPbsOrtbBid(bid, bid.Seat)
				convertedReason := ConvertAPRCToNBRC(bid.Status)
				if convertedReason != nil {
					nonBidParams.NonBidReason = int(*convertedReason)
				}
				snb.AddBid(openrtb_ext.NewNonBid(nonBidParams), bid.Seat)
			}
		}
	}
	return
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

func (sa *structuredAdpod) isCatOverlap(cats []string, catMap map[string]bool) bool {
	if !sa.Exclusion.IABCategoryExclusion {
		return false
	}

	return isAtrributesOverlap(cats, catMap)
}

func (sa *structuredAdpod) isDomainOverlap(domains []string, domainMap map[string]bool) bool {
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

func (sa *structuredAdpod) isOverlap(bid *types.Bid, catMap map[string]bool, domainMap map[string]bool) bool {
	return sa.isCatOverlap(bid.Cat, catMap) || sa.isDomainOverlap(bid.ADomain, domainMap)
}

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
			slotBids[selectedSlot.Index].Status = constant.StatusCategoryExclusion
			selectedSlot.Index = bidIndex
			slots[slotIndex] = selectedSlot
		} else if sa.isCategoryAlreadySelected(selectedBid) || sa.isDomainAlreadySelected(selectedBid) {
			slotBids[selectedSlot.Index].Status = constant.StatusCategoryExclusion
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

// isBetterBidAvailable checks if a better bid is available for the selected slot.
// It returns true if
func (sa *structuredAdpod) isBetterBidAvailable(slots []Slot, selectedSlotIndex int, selectedBid *types.Bid) bool {
	if len(selectedBid.Cat) == 0 && len(selectedBid.ADomain) == 0 {
		return false
	}

	catMap := createMapFromSlice(selectedBid.Cat)
	domainMap := createMapFromSlice(selectedBid.ADomain)

	return sa.shouldUpdateSelectedBid(slots, selectedSlotIndex, catMap, domainMap)
}

// shouldUpdateSelectedBid checks if a bid should be updated for a selected slot.
func (sa *structuredAdpod) shouldUpdateSelectedBid(slots []Slot, selectedSlotIndex int, catMap map[string]bool, domainMap map[string]bool) bool {
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
func (sa *structuredAdpod) shouldUpdateBid(slotBids []*types.Bid, slotIndex int, catMap map[string]bool, domainMap map[string]bool) bool {
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

func (sa *structuredAdpod) getBetterBid(slotBids []*types.Bid, selectedBid *types.Bid, selectedBidtIndex int) (int, bool) {
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

func (sa *structuredAdpod) isBetterBidThanDeal(slots []Slot, selectedSlotIndx int, selectedSlot Slot, selectedBid *types.Bid) (int, bool) {
	selectedBidIndex := selectedSlot.Index

	if !isDealBid(selectedBid) {
		return selectedBidIndex, false
	}

	if !sa.isBetterBidAvailable(slots, selectedSlotIndx, selectedBid) {
		return selectedBidIndex, false
	}

	return sa.getBetterBid(sa.ImpBidMap[selectedSlot.ImpId], selectedBid, selectedBidIndex)
}
