package resolver

import "github.com/prebid/openrtb/v20/openrtb2"

// currencyResolver resolves the currency of the adapter response
type currencyResolver struct {
	valueResolver
}

func (r *currencyResolver) getFromORTBObject(ortbResponse map[string]any) (any, bool) {
	if curr, ok := ortbResponse[curKey]; ok && curr != "" {
		return curr, true
	}
	return nil, false
}

func (r *currencyResolver) autoDetect(request *openrtb2.BidRequest, node map[string]any) (any, bool) {
	return nil, false
}

func (r *currencyResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid[currencyKey] = value
}
