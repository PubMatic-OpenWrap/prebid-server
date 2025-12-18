package publisherfeature

type performanceDSPs struct {
	dsps  [2]map[int]struct{}
	index int
}

func newPerformanceDSPs() performanceDSPs {
	return performanceDSPs{
		dsps: [2]map[int]struct{}{
			make(map[int]struct{}),
			make(map[int]struct{}),
		},
		index: 0,
	}
}

// updatePerformanceDSPs updates performanceDSPs fetched from DB to pubFeatureMap
func (fe *feature) updatePerformanceDSPs() {
	performanceDSPs, err := fe.cache.GetPerformanceDSPs()
	if err != nil || performanceDSPs == nil {
		return
	}
	// assign fetched dsps to the inactive map
	fe.performanceDSPs.dsps[fe.performanceDSPs.index^1] = performanceDSPs
	// toggle the index to make the updated map active
	fe.performanceDSPs.index ^= 1
}

// GetEnabledPerformanceDSPs returns enabled performance dsp ids
func (fe *feature) GetEnabledPerformanceDSPs() map[int]struct{} {
	return fe.performanceDSPs.dsps[fe.performanceDSPs.index]
}
