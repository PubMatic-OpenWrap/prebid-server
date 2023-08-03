package gocache

// Populates Cache with Fsc-Disabled Publishers
func (dbcache *cache) GetFSCDisabledPublishers() (map[int]struct{}, error) {
	return dbcache.db.GetFSCDisabledPublishers()
}

// Populates cache with Fsc-Dsp Threshold Percentages
func (dbcache *cache) GetFSCThresholdPerDSP() (map[int]int, error) {
	return dbcache.db.GetFSCThresholdPerDSP()
}
