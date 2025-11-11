package publisherfeature

type performantDSPs struct {
	dsps  [2]map[int]struct{}
	index int
}

func newPerformantDSPs() performantDSPs {
	return performantDSPs{
		dsps: [2]map[int]struct{}{
			make(map[int]struct{}),
			make(map[int]struct{}),
		},
		index: 0,
	}
}

// updatePerformantDSPs updates performantDSPs fetched from DB to pubFeatureMap
func (fe *feature) updatePerformantDSPs() {
	performantDSPs, err := fe.cache.GetPerformantDSPs()
	if err != nil || performantDSPs == nil {
		return
	}
	// assign fetched dsps to the inactive map
	fe.performantDSPs.dsps[fe.performantDSPs.index^1] = performantDSPs
	// toggle the index to make the updated map active
	fe.performantDSPs.index ^= 1
}

// GetEnabledPerformantDSPs returns enabled performant dsp ids
func (fe *feature) GetEnabledPerformantDSPs() map[int]struct{} {
	return fe.performantDSPs.dsps[fe.performantDSPs.index]
}
