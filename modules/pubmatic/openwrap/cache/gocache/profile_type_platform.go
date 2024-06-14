package gocache

import (
	"fmt"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorProfileTypePlatformUpdate = "[ErrorProfileTypePlatformUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetProfileTypePlatform() (map[string]int, error) {
	profileTypePlatformMap, err := c.db.GetProfileTypePlatform()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.ProfileTypePlatformQuery, "", "")
		return profileTypePlatformMap, fmt.Errorf(errorProfileTypePlatformUpdate, err)
	}
	return profileTypePlatformMap, nil
}
