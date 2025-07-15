package openwrap

import (
	"math/rand"
	"strings"
	"time"

	"github.com/golang/glog"
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

func (m *OpenWrap) applyPartnerThrottling(rCtx models.RequestCtx, partnerConfigMap map[int]map[string]string) bool {
	throttlePartners, err := m.cache.GetThrottlePartnersWithCriteria(rCtx.DeviceCtx.DerivedCountryCode, "gecpm", 0)
	if err != nil {
		glog.Errorf("Error getting throttled partners for country %s: %v", rCtx.DeviceCtx.DerivedCountryCode, err)
		return false
	}

	throttleMap := make(map[string]struct{}, len(throttlePartners))
	for _, bidder := range throttlePartners {
		throttleMap[bidder] = struct{}{}
	}

	// Create a single random generator instance seeded once per request
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	allPartnersThrottledFlag := true
	for _, cfg := range partnerConfigMap {
		bidderCode, ok := cfg[models.BidderCode]
		if !ok || bidderCode == "" {
			continue
		}

		if _, exists := throttleMap[bidderCode]; exists {
			// 5% fallback traffic logic
			if r.Float64() < 0.05 {
				glog.Infof("Allowing 5%% fallback traffic for throttled bidder: %s", bidderCode)
				continue
			}
			rCtx.AdapterThrottleMap[bidderCode] = struct{}{}
			m.metricEngine.RecordPartnerThrottledRequests(rCtx.PubIDStr, bidderCode)
		} else if allPartnersThrottledFlag {
			allPartnersThrottledFlag = false
		}
	}
	return allPartnersThrottledFlag
}
