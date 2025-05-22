package openwrap

import (
	"strings"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func getCountryFilterConfig(partnerConfigMap map[int]map[string]string) (mode string, countryCodes string) {
	mode = models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.CountryFilterModeKey)
	if mode == "" {
		return "", ""
	}

	countryCodes = models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.CountryCodesKey)
	return mode, countryCodes

}

func isCountryAllowed(country string, mode string, countryCodes string) bool {
	if mode == "" || countryCodes == "" {
		return true
	}

	found := strings.Contains(countryCodes, country)

	// For allowlist (mode "1"), return true if country is found
	// For blocklist (mode "0"), return true if country is not found
	return (mode == "1" && found) || (mode == "0" && !found)
}

func shouldApplyCountryFilter(endpoint string) bool {
	return endpoint == models.EndpointAppLovinMax || endpoint == models.EndpointGoogleSDK
}
