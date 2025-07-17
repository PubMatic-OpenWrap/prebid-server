package gocache

import (
	"errors"
)

func (c *cache) GetThrottlePartnersWithCriteria(country string, criteria string, criteriaValue int) ([]string, error) {
	if c.db == nil {
		return nil, errors.New("DB not initialized")
	}

	partnerthrottleMap := c.db.GetLatestCountryPartnerFilter()
	if partnerthrottleMap == nil {
		return nil, errors.New("partner filter cache empty")
	}

	var throttledPartners []string
	if countryRecords, exists := partnerthrottleMap[country]; exists {
		for _, record := range countryRecords {
			if record.Criteria == criteria && record.CriteriaThreshold == criteriaValue {
				throttledPartners = append(throttledPartners, record.FeatureValue)
			}
		}
	}

	return throttledPartners, nil
}
