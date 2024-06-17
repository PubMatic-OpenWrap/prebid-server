package gocache

import (
	"fmt"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorAppIntegrationPathUpdate = "[ErrorAppIntegrationPathUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetAppIntegrationPaths() (map[string]int, error) {
	AppIntegrationPathMap, err := c.db.GetAppIntegrationPaths()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AppIntegrationPathQuery, "", "")
		return AppIntegrationPathMap, fmt.Errorf(errorAppIntegrationPathUpdate, err)
	}
	return AppIntegrationPathMap, nil
}
