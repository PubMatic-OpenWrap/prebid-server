package gocache

import (
	"fmt"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

var (
	errorFscPubMsg = "[ErrorFscPubUpdate]:%w"
	errorFscDspMsg = "[ErrorFscDspUpdate]:%w"
)

// Populates Cache with Fsc-Disabled Publishers
func (c *cache) GetFSCDisabledPublishers() (map[int]struct{}, error) {
	fscDisabledPublishers, err := c.db.GetFSCDisabledPublishers()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllFscDisabledPublishersQuery, "", "")
		return fscDisabledPublishers, fmt.Errorf(errorFscPubMsg, err)
	}
	return fscDisabledPublishers, nil
}

// Populates cache with Fsc-Dsp Threshold Percentages
func (c *cache) GetFSCThresholdPerDSP() (map[int]int, error) {
	fscThreshold, err := c.db.GetFSCThresholdPerDSP()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllDspFscPcntQuery, "", "")
		return fscThreshold, fmt.Errorf(errorFscDspMsg, err)
	}
	return fscThreshold, nil
}
