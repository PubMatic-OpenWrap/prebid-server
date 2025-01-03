package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var errorGDPRCountryUpdate = "[ErrorGDPRCountryUpdate]:%w"

// We are not saving data in cache here
func (c *cache) GetGDPRCountryCodes() (map[string]struct{}, error) {
	gdprCountryCodes, err := c.db.GetGDPRCountryCodes()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.GDPRCountryCodesQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.GDPRCountryCodesQuery, "", "", err)
		return gdprCountryCodes, fmt.Errorf(errorGDPRCountryUpdate, err)
	}
	return gdprCountryCodes, nil
}
