package gocache

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// GetProfileAdUnitMultiFloors returns profileadunitmultifloors.
func (c *cache) GetProfileAdUnitMultiFloors() (models.ProfileAdUnitMultiFloors, error) {
	profileAdUnitMultiFloors, err := c.db.GetProfileAdUnitMultiFloors()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.ProfileAdUnitMultiFloorsQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.ProfileAdUnitMultiFloorsQuery, "", "", err)
		return models.ProfileAdUnitMultiFloors{}, err
	}
	return profileAdUnitMultiFloors, nil
}
