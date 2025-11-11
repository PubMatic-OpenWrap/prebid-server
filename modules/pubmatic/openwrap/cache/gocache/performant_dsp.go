package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

var (
	errorPerformantDSPMsg = "[ErrorPerformantDSPUpdate]:%w"
)

// Populates cache with Performant DSPs
func (c *cache) GetPerformantDSPs() (map[int]struct{}, error) {
	performantDSPs, err := c.db.GetPerformantDSPs()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.PerformantDSPsQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.PerformantDSPsQuery, "", "", err)
		return performantDSPs, fmt.Errorf(errorPerformantDSPMsg, err)
	}
	return performantDSPs, nil
}
