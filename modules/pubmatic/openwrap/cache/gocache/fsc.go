package gocache

// Populates Cache with Fsc-Disabled Publishers
func (dbcache *cache) GetFSCDisabledPublishers() (map[int]struct{}, error) {
	disabledPublishersMap, err := dbcache.db.GetFSCDisabledPublishers()
	if err != nil {
		return disabledPublishersMap, err
	}
	// Not setting into cache as fsc maintains it own map
	// mcache.Set(constant.FscPublisher, disabledPublishersMap)
	return disabledPublishersMap, nil
}

// Populates cache with Fsc-Dsp Threshold Percentages
func (dbcache *cache) GetFSCThresholdPerDSP() (map[int]int, error) {
	dspThresholdsMap, err := dbcache.db.GetFSCThresholdPerDSP()
	if err != nil {
		return dspThresholdsMap, err
	}
	// Not setting into cache as fsc maintains it own map
	// mcache.Set(constant.FscPublisher, dspThresholdsMap)
	return dspThresholdsMap, nil
}
