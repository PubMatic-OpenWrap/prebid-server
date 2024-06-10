package resolver

type currencyResolver struct {
	valueResolver
}

func (r *currencyResolver) getFromORTBObject(ortbResponse map[string]any) (any, bool) {
	return ortbResponse["cur"], true
}

func (r *currencyResolver) autoDetect(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *currencyResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid["Currency"] = value
}
