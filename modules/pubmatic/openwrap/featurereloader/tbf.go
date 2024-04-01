// Package tbf provides functionalities related to the Tracking-Beacon-First (TBF) feature.
// The package manages the configuration of the TBF feature, which includes publisher-profile-level
// traffic data, caching, and service reloader functionality.
package featurereloader

import (
	"math/rand"
	"sync"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
)

// tbf structure holds the configuration of Tracking-Beacon-First feature
type tbf struct {
	pubProfileTraffic map[int]map[int]int
	*sync.RWMutex
}

var tbfConfigs tbf

// limitTBFTrafficValues validates the traffic values from the given map of pub-prof-traffic
// to ensure they are constrained between 0 and 100 (inclusive).
// If a value is below 0 or above 100, it is set to 0. The original map is modified in place.
func limitTBFTrafficValues(pubProfTraffic map[int]map[int]int) {
	for _, profTraffic := range pubProfTraffic {
		for profID, traffic := range profTraffic {
			if traffic < 0 || traffic > 100 {
				profTraffic[profID] = 0
			}
		}
	}
}

// updateTBFConfigMapsFromCache loads the TBF traffic data from cache/database and updates the configuration map.
// If execution of db-query-fails then this function will not update the old config-values.
// This function is safe for concurrent access.
func updateTBFConfigMapsFromCache() error {

	pubProfileTrafficRate, err := reloaderConfig.cache.GetTBFTrafficForPublishers()
	if err != nil {
		return err
	}
	limitTBFTrafficValues(pubProfileTrafficRate)

	reloaderConfig.Lock()
	reloaderConfig.tbf.pubProfileTraffic = pubProfileTrafficRate
	reloaderConfig.Unlock()

	return nil
}

// IsEnabledTBFFeature returns false if TBF feature is disabled for pub-profile combination
// It makes use of predictTBFValue function to predict whether the request is eligible
// to track beacon first before adm based on the provided traffic percentage.
// This function is safe for concurrent access.
func IsEnabledTBFFeature(pubid int, profid int) bool {

	var trafficRate int
	var present bool

	reloaderConfig.RLock()
	if tbfConfigs.pubProfileTraffic != nil {
		trafficRate, present = tbfConfigs.pubProfileTraffic[pubid][profid]
	}
	reloaderConfig.RUnlock()

	if !present {
		return false
	}

	return predictTBFValue(trafficRate)
}

// predictTBFValue predicts whether a request is eligible for TBF feature
// based on the provided trafficRate value.“
func predictTBFValue(trafficRate int) bool {
	return rand.Intn(100) < trafficRate
}

// SetAndResetTBFConfig is exposed for test cases
func SetAndResetTBFConfig(mockDb cache.Cache, pubProfileTraffic map[int]map[int]int) func() {
	reloaderConfig.tbf.RWMutex = &sync.RWMutex{}
	reloaderConfig.cache = mockDb
	tbfConfigs.pubProfileTraffic = pubProfileTraffic
	return func() {
		reloaderConfig.cache = nil
		tbfConfigs.pubProfileTraffic = make(map[int]map[int]int)
	}
}
