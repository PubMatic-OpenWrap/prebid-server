package gocache

import (
	"fmt"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorAppSubIntegrationPathUpdate = "[ErrorAppSubIntegrationPathUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetAppSubIntegrationPaths() (map[string]int, error) {
	appSubIntegrationPathMap, err := c.db.GetAppSubIntegrationPaths()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AppSubIntegrationPathQuery, "", "")
		return appSubIntegrationPathMap, fmt.Errorf(errorAppSubIntegrationPathUpdate, err)
	}
	return appSubIntegrationPathMap, nil
}
