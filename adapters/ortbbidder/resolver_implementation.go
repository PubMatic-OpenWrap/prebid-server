package ortbbidder

import (
	"github.com/prebid/openrtb/v20/openrtb2"
)

type mtypeResolver struct{}

func (r *mtypeResolver) fromOriginalObject(bid map[string]any) (any, bool) {
	mType, ok := bid["mtype"].(float64)
	if !ok {
		return nil, false
	}
	return getMediaTypeForBidFromMType(openrtb2.MarkupType(mType)), true
}

func (r *mtypeResolver) fromParamLocation(ortbResponse map[string]any, path string) (any, bool) {
	return getValueFromLocation(ortbResponse, path)
}

func (r *mtypeResolver) autoDetect(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *mtypeResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid["mtype"] = value
}
