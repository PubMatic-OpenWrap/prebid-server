package gocache

import (
	"errors"
	"reflect"
)

func (c *cache) GetThrottlePartnersWithCriteria(country string) ([]string, error) {

	if c.db == nil || reflect.ValueOf(c.db).IsNil() {
		return nil, errors.New("DB not initialized")
	}
	partnerthrottleMap := c.db.GetLatestCountryPartnerFilter()
	if partnerthrottleMap == nil {
		return nil, errors.New("partner filter cache is empty")
	}
	return partnerthrottleMap[country], nil
}
