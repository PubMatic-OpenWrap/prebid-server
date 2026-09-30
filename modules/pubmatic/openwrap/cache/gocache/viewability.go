package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

var errorInViewEnabledPublishersUpdate = "[ErrorInViewEnabledPublishersUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetInViewEnabledPublishers() (map[int]struct{}, error) {
	inViewEnabledPublishers, err := c.db.GetInViewEnabledPublishers()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.InViewEnabledPublishersQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.InViewEnabledPublishersQuery, "", "", err)
		return inViewEnabledPublishers, fmt.Errorf(errorInViewEnabledPublishersUpdate, err)
	}
	return inViewEnabledPublishers, nil
}
