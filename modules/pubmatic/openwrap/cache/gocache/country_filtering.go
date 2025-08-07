package gocache

import (
	"errors"
	"reflect"
)

func (c *cache) GetThrottlePartnersWithCriteria(country string) (map[string]struct{}, error) {

	if c.db == nil || reflect.ValueOf(c.db).IsNil() {
		return nil, errors.New("DB not initialized")
	}
	countryPartnerThrottleMap := c.db.GetLatestCountryPartnerFilter()
	if countryPartnerThrottleMap == nil {
		return nil, errors.New("partner filter cache is empty")
	}
	return countryPartnerThrottleMap[country], nil
}
