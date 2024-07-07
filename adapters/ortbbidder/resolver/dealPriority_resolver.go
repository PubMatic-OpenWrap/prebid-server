package resolver

import "github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"

// bidDealPriorityResolver retrieves the priority of the deal bid using the bidder param location.
// The determined dealPriority is subsequently assigned to adapterresponse.typedbid.dealPriority
type bidDealPriorityResolver struct {
	defaultValueResolver
}

func (b *bidDealPriorityResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateNumber[int](value)
}

func (b *bidDealPriorityResolver) setValue(adapterBid map[string]any, value any) bool {
	adapterBid[bidDealPriorityKey] = value
	return true
}
