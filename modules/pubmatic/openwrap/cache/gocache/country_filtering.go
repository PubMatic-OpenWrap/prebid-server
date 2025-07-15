package gocache

import (
	"errors"
)

const CountryPartnerFiltering = "country_partner_filtering"

func (c *cache) GetThrottlePartnersWithCriteria(country string, criteria string, criteriaValue int) ([]string, error) {
	if c.db == nil {
		return nil, errors.New("DB not initialized")
	}

	partnerMap := c.db.GetLatestCountryPartnerFilter()
	if partnerMap == nil {
		return nil, errors.New("partner filter cache empty")
	}

	var throttledPartners []string
	if countryRecords, exists := partnerMap[country]; exists {
		for _, record := range countryRecords {
			if record.Criteria == criteria && record.CriteriaThreshold == criteriaValue {
				throttledPartners = append(throttledPartners, record.FeatureValue)
			}
		}
	}

	return throttledPartners, nil
}
