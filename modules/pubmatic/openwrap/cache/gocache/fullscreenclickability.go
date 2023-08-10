package gocache

import "github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

// Populates Cache with Fsc-Disabled Publishers
func (c *cache) GetFSCDisabledPublishers() (map[int]struct{}, error) {
	fscDisabledPublishers, err := c.db.GetFSCDisabledPublishers()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllFscDisabledPublishersQuery, "", "")
	}
	return fscDisabledPublishers, err
}

// Populates cache with Fsc-Dsp Threshold Percentages
func (c *cache) GetFSCThresholdPerDSP() (map[int]int, error) {
	fscThreshold, err := c.db.GetFSCThresholdPerDSP()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllDspFscPcntQuery, "", "")
	}
	return fscThreshold, err
}
