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

// GetMBMFPhase1PubId returns phase1pubidmultifloors fetched from DB which will be saved in publisherFeatureMap
func (c *cache) GetMBMFPhase1PubId() (map[int]struct{}, error) {
	pubIdMultiFloors, err := c.db.GetMBMFPhase1PubId()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.MBMFPhase1PubIdQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.MBMFPhase1PubIdQuery, "", "", err)
		return pubIdMultiFloors, err
	}
	return pubIdMultiFloors, nil
}
