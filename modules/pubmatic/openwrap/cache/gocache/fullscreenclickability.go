package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

var (
	errorFscDspMsg = "[ErrorFscDspUpdate]:%w"
)

// Populates cache with Fsc-Dsp Threshold Percentages
func (c *cache) GetFSCThresholdPerDSP() (map[int]int, error) {
	fscThreshold, err := c.db.GetFSCThresholdPerDSP()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllDspFscPcntQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.AllDspFscPcntQuery, "", "", err)
		return fscThreshold, fmt.Errorf(errorFscDspMsg, err)
	}
	return fscThreshold, nil
}
