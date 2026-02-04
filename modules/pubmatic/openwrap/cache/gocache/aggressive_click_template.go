package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

var (
	errorActDspMsg = "[ErrorActDspUpdate]:%w"
)

// Populates cache with Act-Dsp Threshold Percentages
func (c *cache) GetACTThresholdPerDSP() (map[int]int, error) {
	actThreshold, err := c.db.GetACTThresholdPerDSP()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllDspActPcntQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.AllDspActPcntQuery, "", "", err)
		return actThreshold, fmt.Errorf(errorActDspMsg, err)
	}
	return actThreshold, nil
}
