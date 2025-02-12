package publisherfeature

import (
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type gdprCountryCodes struct {
	codes [2]models.HashSet
	index int
}

func newGDPRCountryCodes() gdprCountryCodes {
	return gdprCountryCodes{
		codes: [2]models.HashSet{
			make(models.HashSet),
			make(models.HashSet),
		},
		index: 0,
	}
}

// updateGDPRCountryCodes updates gdprCountryCodes fetched from DB to pubFeatureMap
func (fe *feature) updateGDPRCountryCodes() {
	gdprCountryCodes, err := fe.cache.GetGDPRCountryCodes()
	if err != nil || gdprCountryCodes == nil {
		return
	}
	// assign fetched codes to the inactive map
	fe.gdprCountryCodes.codes[fe.gdprCountryCodes.index^1] = gdprCountryCodes
	// toggle the index to make the updated map active
	fe.gdprCountryCodes.index ^= 1
}

// IsCountryGDPREnabled returns true if country is gdpr enabled
func (fe *feature) IsCountryGDPREnabled(countryCode string) bool {
	codes := fe.gdprCountryCodes.codes[fe.gdprCountryCodes.index]
	_, enabled := codes[countryCode]
	return enabled
}
