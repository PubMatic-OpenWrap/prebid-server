package gocache

import (
	"fmt"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

var errorPubFeatureUpdate = "[ErrorPubFeatureUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetPublisherFeatureMap() (map[int]map[int]models.FeatureData, error) {
	publisherFeatureMap, err := c.db.GetPublisherFeatureMap()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.PublisherFeatureMapQuery, "", "")
		return publisherFeatureMap, fmt.Errorf(errorPubFeatureUpdate, err)
	}
	return publisherFeatureMap, nil
}
