package openwrap

import (
	"math/rand"
	"strconv"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// GetAdapterThrottleMap creates a map of adapters that should be throttled and returns whether all partners are throttled.
func GetAdapterThrottleMap(partnerConfigMap map[int]map[string]string, adapterThrottleMap map[string]struct{}) (map[string]struct{}, bool) {
	totalValidPartners := 0
	if adapterThrottleMap == nil {
		adapterThrottleMap = make(map[string]struct{})
	}

	for _, partnerConfig := range partnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		bidderCode := partnerConfig[models.BidderCode]
		if bidderCode == "" {
			continue
		}

		totalValidPartners++

		// If already marked as throttled, skip further check
		if _, alreadyThrottled := adapterThrottleMap[bidderCode]; alreadyThrottled {
			continue
		}

		if ThrottleAdapter(partnerConfig) {
			adapterThrottleMap[bidderCode] = struct{}{}
		}
	}

	allPartnersThrottled := totalValidPartners > 0 && len(adapterThrottleMap) == totalValidPartners
	return adapterThrottleMap, allPartnersThrottled
}

// ThrottleAdapter this function returns bool value for whether a adapter should be throttled or not
func ThrottleAdapter(partnerConfig map[string]string) bool {
	if partnerConfig[models.THROTTLE] == "100" || partnerConfig[models.THROTTLE] == "" {
		return false
	}

	if partnerConfig[models.THROTTLE] == "0" {
		return true
	}

	//else check throttle value based on random no
	throttle, _ := strconv.ParseFloat(partnerConfig[models.THROTTLE], 64)
	throttle = 100 - throttle

	randomNumberBelow100 := GetRandomNumberBelow100()
	return !(float64(randomNumberBelow100) >= throttle)
}

var GetRandomNumberBelow100 = func() int {
	return rand.Intn(99)
}

var GetRandomNumberIn1To100 = func() int {
	return rand.Intn(100) + 1
}
