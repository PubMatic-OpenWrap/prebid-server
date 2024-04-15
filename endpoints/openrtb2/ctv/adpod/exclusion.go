package adpod

import (
	"math"

	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
)

type Exclusion struct {
	Config ExclusionConfig

	// IAB Category
	SelectedCategories   map[string]bool
	SameCategoryBidCount int

	// Advertiser Domain
	SelectedDomains    map[string]bool
	SameDomainBidCount int
}

type ExclusionConfig struct {
	AdvertiserExclusionPercent  int // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent int // Percent value 0 means all ads should be of different IAB categories.
}

func (ex *Exclusion) shouldApplyExclusion() bool {
	return ex.SameCategoryBidCount == 0 || ex.SameDomainBidCount == 0
}

func (ex *Exclusion) setExclusionConditions(totalSlots int) {
	// IAB Category
	var sameCategoryBidCount int
	if ex.Config.IABCategoryExclusionPercent == 0 {
		sameCategoryBidCount = 0
	} else {
		sameCategoryBidCount = int(math.Floor(float64(totalSlots) * (float64(ex.Config.IABCategoryExclusionPercent) / 100)))
	}
	ex.SameCategoryBidCount = sameCategoryBidCount

	// Advertiser Domain
	var sameDomainBidCount int
	if ex.Config.AdvertiserExclusionPercent == 0 {
		sameDomainBidCount = 0
	} else {
		sameDomainBidCount = int(math.Floor(float64(totalSlots) * (float64(ex.Config.AdvertiserExclusionPercent) / 100)))
	}
	ex.SameDomainBidCount = sameDomainBidCount
}

func (ex *Exclusion) updateExclusionConditions() {
	// IAB Category
	if ex.SameCategoryBidCount != 0 {
		ex.SameCategoryBidCount--
	}

	// Advertiser Domain
	if ex.SameDomainBidCount != 0 {
		ex.SameDomainBidCount--
	}
}

func (ex *Exclusion) addExclusionParameters(bid *types.Bid) {
	if bid == nil {
		return
	}

	// Add bid categories to selected categories
	ex.addCategories(bid.Cat)

	// Add bid Domains to selected Domains
	ex.addDomains(bid.ADomain)
}

func (ex *Exclusion) eitherExclusionNotMetForBid(bid *types.Bid) bool {
	return ex.isCategoryAlreadySelected(bid) || ex.isDomainAlreadySelected(bid)
}

func (ex *Exclusion) allExclusionConditionsSatified(bid *types.Bid) bool {
	return !ex.isCategoryAlreadySelected(bid) && !ex.isDomainAlreadySelected(bid)
}

func exclusionParamsNotExists(bid *types.Bid) bool {
	if bid == nil {
		return true
	}

	return len(bid.Cat) == 0 && len(bid.ADomain) == 0
}

/**************************IAB Categories**********************/

// addCategories will collect categories for adpod
func (ex *Exclusion) addCategories(categories []string) {
	if ex.SelectedCategories == nil {
		ex.SelectedCategories = make(map[string]bool)
	}

	for _, cat := range categories {
		ex.SelectedCategories[cat] = true
	}
}

func (ex *Exclusion) isCategoryAlreadySelected(bid *types.Bid) bool {
	if ex.SameCategoryBidCount > 0 {
		return false
	}

	if bid == nil || bid.Cat == nil {
		return false
	}

	if ex.SelectedCategories == nil {
		return false
	}

	var alreadySelected bool
	for i := range bid.Cat {
		_, ok := ex.SelectedCategories[bid.Cat[i]]
		if ok {
			alreadySelected = true
			break
		}
	}

	return alreadySelected
}

/**************************Advertiser Domains**********************/

// addDomains will collect domains for adpod
func (ex *Exclusion) addDomains(domains []string) {
	if ex.SelectedDomains == nil {
		ex.SelectedDomains = make(map[string]bool)
	}

	for _, domain := range domains {
		ex.SelectedDomains[domain] = true
	}
}

func (ex *Exclusion) isDomainAlreadySelected(bid *types.Bid) bool {
	if ex.SameDomainBidCount > 0 {
		return false
	}

	if bid == nil || bid.ADomain == nil {
		return false
	}

	if ex.SelectedDomains == nil {
		return false
	}

	var alreadySelected bool
	for i := range bid.ADomain {
		_, ok := ex.SelectedDomains[bid.ADomain[i]]
		if ok {
			alreadySelected = true
			break
		}
	}

	return alreadySelected
}
