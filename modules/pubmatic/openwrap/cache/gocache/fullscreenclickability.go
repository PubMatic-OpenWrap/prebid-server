package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var (
	errorFscDspMsg = "[ErrorFscDspUpdate]:%w"
)

// Populates cache with Fsc-Dsp Threshold Percentages
func (c *cache) GetFSCThresholdPerDSP() (map[int]int, error) {
	fscThreshold, err := c.db.GetFSCThresholdPerDSP()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllDspFscPcntQuery, "", "")
		glog.Errorf("[PartialQueryFailure] for AllDspFscPcntQuery with err: %v", err)
		return fscThreshold, fmt.Errorf(errorFscDspMsg, err)
	}
	return fscThreshold, nil
}
