package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

var (
	errorPerformanceDSPMsg = "[ErrorPerformanceDSPUpdate]:%w"
)

// Populates cache with Performance DSPs
func (c *cache) GetPerformanceDSPs() (map[int]struct{}, error) {
	performanceDSPs, err := c.db.GetPerformanceDSPs()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.PerformanceDSPsQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.PerformanceDSPsQuery, "", "", err)
		return performanceDSPs, fmt.Errorf(errorPerformanceDSPMsg, err)
	}
	return performanceDSPs, nil
}
