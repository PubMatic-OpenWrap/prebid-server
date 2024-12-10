package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorAppSubIntegrationPathMapUpdate = "[ErrorAppSubIntegrationPathMapUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetAppSubIntegrationPaths() (map[string]int, error) {
	appSubIntegrationPathMap, err := c.db.GetAppSubIntegrationPaths()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AppSubIntegrationPathMapQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.AppSubIntegrationPathMapQuery, "", "", err)
		return appSubIntegrationPathMap, fmt.Errorf(errorAppSubIntegrationPathMapUpdate, err)
	}
	return appSubIntegrationPathMap, nil
}
