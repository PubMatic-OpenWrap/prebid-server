package resolver

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

type mtypeResolver struct{}

func (r *mtypeResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	mtype, ok := bid["mtype"].(float64)
	if !ok {
		return nil, false
	}
	return util.GetMediaType(openrtb2.MarkupType(mtype)), true
}

func (r *mtypeResolver) getUsingBidderParamLocation(ortbResponse map[string]any, path string) (any, bool) {
	return util.GetValueFromLocation(ortbResponse, path)
}

func (r *mtypeResolver) autoDetect(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *mtypeResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid["BidType"] = value
}
