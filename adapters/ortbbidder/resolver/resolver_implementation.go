package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

type mtypeResolver struct{}

func (r *mtypeResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	mtype, ok := bid["mtype"].(float64)
	if !ok {
		return nil, false
	}
	return util.GetMType(mtype), true
}

func (r *mtypeResolver) getUsingBidderParam(ortbResponse map[string]any, path string) (any, bool) {
	return util.GetValueFromLocation(ortbResponse, path)
}

func (r *mtypeResolver) autoDetect(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *mtypeResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid["BidType"] = value
}

type currencyResolver struct{}

func (r *currencyResolver) getFromORTBObject(ortbResponse map[string]any) (any, bool) {
	return ortbResponse["cur"], true
}

func (r *currencyResolver) getUsingBidderParam(ortbResponse map[string]any, path string) (any, bool) {
	return util.GetValueFromLocation(ortbResponse, path)
}

func (r *currencyResolver) autoDetect(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *currencyResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid["Currency"] = value
}
