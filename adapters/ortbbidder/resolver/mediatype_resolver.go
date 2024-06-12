package resolver

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

// mtypeResolver resolves the media type of the type bid
type mtypeResolver struct {
	valueResolver
}

func (r *mtypeResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	mtype, ok := bid[mtypeKey].(float64)
	if !ok && mtype == 0 {
		return nil, false
	}
	return util.GetMediaType(openrtb2.MarkupType(mtype)), true
}

func (r *mtypeResolver) autoDetect(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *mtypeResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid[bidTypeKey] = value
}
