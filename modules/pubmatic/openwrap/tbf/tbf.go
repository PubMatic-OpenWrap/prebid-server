// Package tbf provides functionalities related to the Tracking-Beacon-First (TBF) feature.
// The package manages the configuration of the TBF feature, which includes publisher-profile-level
// traffic data, caching, and service reloader functionality.
package tbf

import (
	"math/rand"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
)

// tbf structure holds the configuration of Tracking-Beacon-First feature
type tbf struct {
	pubProfileTraffic map[int]map[int]int

	cache cache.Cache
	*sync.RWMutex
	serviceStop chan (struct{})
}

var tbfConfigs tbf

// initiateTBFReloader periodically update the TBF configuration from database
var initiateTBFReloader = func(c cache.Cache, expiryTime int) {
	glog.Info("TBF Reloader start")
	ticker := time.NewTicker(time.Duration(expiryTime) * time.Second)

	for {
		updateTBFConfigMapsFromCache()
		select {
		case _ = <-tbfConfigs.serviceStop:
			return
		case t := <-ticker.C:
			glog.Infof("TBF Reloader loads cache @%v", t)
		}
	}
}

// Init function initializes parameters of the tbfConfigs
// It starts the TBF reloader service in background
func Init(defaultExpiry int, cache cache.Cache) {

	tbfConfigs.cache = cache
	tbfConfigs.pubProfileTraffic = make(map[int]map[int]int)
	tbfConfigs.serviceStop = make(chan struct{})
	tbfConfigs.RWMutex = &sync.RWMutex{}

	go initiateTBFReloader(cache, defaultExpiry)
	glog.Info("Initialized TBF cache reloaders to update publishers TBF configurations")
}

// StopTBFReloaderService sends signal to stop the reloader service
func StopTBFReloaderService() {
	tbfConfigs.serviceStop <- struct{}{}
}

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

	pubProfileTrafficRate, err := tbfConfigs.cache.GetTBFTrafficForPublishers()
	if err != nil {
		return err
	}
	limitTBFTrafficValues(pubProfileTrafficRate)

	tbfConfigs.Lock()
	tbfConfigs.pubProfileTraffic = pubProfileTrafficRate
	tbfConfigs.Unlock()

	return nil
}

// IsEnabledTBFFeature returns false if TBF feature is disabled for pub-profile combination
// It makes use of predictTBFValue function to predict whether the request is eligible
// to track beacon first before adm based on the provided traffic percentage.
// This function is safe for concurrent access.
func IsEnabledTBFFeature(pubid int, profid int) bool {

	var trafficRate int
	var present bool

	tbfConfigs.RLock()
	if tbfConfigs.pubProfileTraffic != nil {
		trafficRate, present = tbfConfigs.pubProfileTraffic[pubid][profid]
	}
	tbfConfigs.RUnlock()

	if !present {
		return false
	}

	return predictTBFValue(trafficRate)
}

// predictTBFValue predicts whether a request is eligible for TBF feature
// based on the provided trafficRate value.
func predictTBFValue(trafficRate int) bool {
	return rand.Intn(100) < trafficRate
}

// SetAndResetTBFConfig is exposed for test cases
func SetAndResetTBFConfig(mockDb cache.Cache, pubProfileTraffic map[int]map[int]int) func() {
	tbfConfigs.RWMutex = &sync.RWMutex{}
	tbfConfigs.cache = mockDb
	tbfConfigs.pubProfileTraffic = pubProfileTraffic
	return func() {
		tbfConfigs.cache = nil
		tbfConfigs.pubProfileTraffic = make(map[int]map[int]int)
	}
}

// ResetTBFReloader is exposed for test cases
func ResetTBFReloader() {
	initiateTBFReloader = func(c cache.Cache, expiryTime int) {}
}
