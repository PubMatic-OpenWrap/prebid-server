package gocache

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// GetGDPRCountryCodes returns gdprcountrycodes fetched from DB which will be saved in publisherFeatureMap
func (c *cache) GetGDPRCountryCodes() (models.HashSet, error) {
	gdprCountryCodes, err := c.db.GetGDPRCountryCodes()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.GDPRCountryCodesQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.GDPRCountryCodesQuery, "", "", err)
		return gdprCountryCodes, err
	}
	return gdprCountryCodes, nil
}
