package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorAppIntegrationPathMapUpdate = "[ErrorAppIntegrationPathMapUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetAppIntegrationPaths() (map[string]int, error) {
	appIntegrationPathMap, err := c.db.GetAppIntegrationPaths()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AppIntegrationPathMapQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.AppIntegrationPathMapQuery, "", "", err)
		return appIntegrationPathMap, fmt.Errorf(errorAppIntegrationPathMapUpdate, err)
	}
	return appIntegrationPathMap, nil
}
