package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

// bidDealPriorityResolver retrieves the priority of the deal bid using the bidder param location.
// The determined dealPriority is subsequently assigned to adapterresponse.typedbid.dealPriority
type bidDealPriorityResolver struct {
	paramResolver
}

func (b *bidDealPriorityResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.dealpriority]", path)
	}
	val, ok := validateNumber[int](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.dealpriority]", path)
	}
	return val, nil
}

func (b *bidDealPriorityResolver) setValue(adapterBid map[string]any, value any) (err error) {
	adapterBid[bidDealPriorityKey] = value
	return
}
