package resolver

import "github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"

type currencyResolver struct{}

func (r *currencyResolver) getFromORTBObject(ortbResponse map[string]any) (any, bool) {
	return ortbResponse["cur"], true
}

func (r *currencyResolver) getUsingBidderParamLocation(ortbResponse map[string]any, path string) (any, bool) {
	return util.GetValueFromLocation(ortbResponse, path)
}

func (r *currencyResolver) autoDetect(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *currencyResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid["Currency"] = value
}
