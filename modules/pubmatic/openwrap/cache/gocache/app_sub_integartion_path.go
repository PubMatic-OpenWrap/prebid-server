package gocache

import (
	"fmt"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorAppSubIntegrationPathUpdate = "[ErrorAppSubIntegrationPathUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetAppSubIntegrationPath() (map[string]int, error) {
	AppSubIntegrationPathMap, err := c.db.GetAppSubIntegrationPath()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AppSubIntegrationPathQuery, "", "")
		return AppSubIntegrationPathMap, fmt.Errorf(errorAppSubIntegrationPathUpdate, err)
	}
	return AppSubIntegrationPathMap, nil
}
