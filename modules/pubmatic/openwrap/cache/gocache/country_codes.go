package gocache

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// GetCountryCodesMapping returns countrycodesmapping fetched from DB which will be saved in publisherFeatureMap
func (c *cache) GetCountryCodesMapping() (models.CountryCodesMapping, error) {
	countryCodesMapping, err := c.db.GetCountryCodesMapping()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.CountryCodesMappingQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.CountryCodesMappingQuery, "", "", err)
		return countryCodesMapping, err
	}
	return countryCodesMapping, nil
}
