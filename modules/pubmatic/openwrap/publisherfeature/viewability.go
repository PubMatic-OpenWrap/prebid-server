package publisherfeature

type inViewEnabledPublishers struct {
	pubs  [2]map[int]struct{}
	index int
}

func newInViewEnabledPublishers() inViewEnabledPublishers {
	return inViewEnabledPublishers{
		pubs: [2]map[int]struct{}{
			make(map[int]struct{}),
			make(map[int]struct{}),
		},
		index: 0,
	}
}

// updateInViewEnabledPublishers updates inViewEnabledPublishers fetched from DB to pubFeatureMap
func (fe *feature) updateInViewEnabledPublishers() {
	inViewEnabledPublishers, err := fe.cache.GetInViewEnabledPublishers()
	if err != nil || inViewEnabledPublishers == nil {
		return
	}
	// assign fetched inViewEnabledPublishers to the inactive map
	fe.inViewEnabledPublishers.pubs[fe.inViewEnabledPublishers.index^1] = inViewEnabledPublishers
	// toggle the index to make the updated map active
	fe.inViewEnabledPublishers.index ^= 1
}

// GetInViewEnabledPublishers returns enabled inViewEnabledPublishers
func (fe *feature) GetInViewEnabledPublishers() map[int]struct{} {
	return fe.inViewEnabledPublishers.pubs[fe.inViewEnabledPublishers.index]
}
