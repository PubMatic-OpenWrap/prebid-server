package publisherfeature

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type gdprCountryCodes struct {
	codes [2]map[string]struct{}
	index int
}

func newGDPRCountryCodes() gdprCountryCodes {
	return gdprCountryCodes{
		codes: [2]map[string]struct{}{
			make(map[string]struct{}),
			make(map[string]struct{}),
		},
		index: 0,
	}
}

func (fe *feature) updateGDPRCountryCodes() {
	var err error
	//fetch gdpr countrycodes
	gdprCountryCodes, errorGDPRCountryUpdate := fe.cache.GetGDPRCountryCodes()
	if errorGDPRCountryUpdate != nil {
		err = models.ErrorWrap(err, errorGDPRCountryUpdate)
	}
	// assign fetched codes to the inactive map
	if gdprCountryCodes != nil {
		fe.gdprCountryCodes.codes[fe.gdprCountryCodes.index^1] = gdprCountryCodes
	}

	// toggle the index to make the updated map active
	fe.gdprCountryCodes.index ^= 1
	if err != nil {
		glog.Error(err.Error())
	}
}

// IsCountryGDPREnabled returns true if country is gdpr enabled
func (fe *feature) IsCountryGDPREnabled(countryCode string) bool {
	codes := fe.gdprCountryCodes.codes[fe.gdprCountryCodes.index]
	_, enabled := codes[countryCode]
	return enabled
}
