package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorProfileTypePlatformMapUpdate = "[ErrorProfileTypePlatformMapUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetProfileTypePlatforms() (map[string]int, error) {
	profileTypePlatformMap, err := c.db.GetProfileTypePlatforms()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.ProfileTypePlatformMapQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.ProfileTypePlatformMapQuery, "", "", err)
		return profileTypePlatformMap, fmt.Errorf(errorProfileTypePlatformMapUpdate, err)
	}
	return profileTypePlatformMap, nil
}
