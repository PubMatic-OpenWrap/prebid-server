package fullscreenclickability

import (
	"math/rand"

	"sync"

	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	// "git.pubmatic.com/PubMatic/go-common/logger"
)

type fsc struct {
	cache              cache.Cache
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
	sync.RWMutex
	serviceStop chan (bool)
}

var fscConfigs fsc

// These reloaders will be called only to Forced-Write into Cache post timer based call.

// Initializing reloader with cache-refresh default-expiry + 30 mins (to avoid DB load post cache refresh)
func Init(c cache.Cache, defaultExpiry int) {
	//init fsc configs
	fscConfigs.cache = c
	fscConfigs.disabledPublishers = make(map[int]struct{})
	fscConfigs.thresholdsPerDsp = make(map[int]int)
	fscConfigs.serviceStop = make(chan bool)

	go initiateReloader(c, defaultExpiry)
	// logger.Info("Initialized FSC cache update reloaders for publisher and dsp fsc configuraitons")

}

// Exposed to access fsc object
func GetFscInstance() *fsc {
	return &fscConfigs
}

/*
	IsUnderFSCThreshold:- returns fsc 1/0 based on

1. When publisher has disabled FSC in DB, return 0
2. If FSC is enabled for publisher(default), consider DSP-threshold , and predict value of fsc 0 or 1.
3. If dspId is not present return 0
*/

func (f *fsc) IsUnderFSCThreshold(pubid int, dspid int) int {
	f.RLock()
	defer f.RUnlock()

	if _, isPresent := f.disabledPublishers[pubid]; isPresent {
		return 0
	}

	if dspThreshold, isPresent := f.thresholdsPerDsp[dspid]; isPresent && predictFscValue(dspThreshold) {
		return 1
	}
	return 0
}

func predictFscValue(threshold int) bool {
	return (rand.Intn(100)) < threshold
}

func StopFscReloaderService() {
	//updating serviceStop flag to true
	fscConfigs.serviceStop <- true
}

func updateFscConfigMapsFromCache(c cache.Cache) {
	fscConfigs.Lock()
	defer fscConfigs.Unlock()

	var err error
	if fscConfigs.disabledPublishers, err = c.GetFSCDisabledPublishers(); err != nil {
		// logger.Error(err.Error())

	}
	if fscConfigs.thresholdsPerDsp, err = c.GetFSCThresholdPerDSP(); err != nil {
		// logger.Error(err.Error())
	}

}

// IsFscApplicable returns true if fsc can be applied (fsc=1)
func IsFscApplicable(pubId int, seat string, dspId int) bool {
	if models.IsPubmaticCorePartner(seat) && (fscConfigs.IsUnderFSCThreshold(pubId, dspId) != 0) {
		return true
	}
	return false
}

// // Exposed for test cases
// func SetAndResetFscWithMockCache(mockDb dbcache.Cache, dspThresholdMap map[int]int) func() {
// 	fscConfigs.cache = mockDb
// 	//mockDspID entry for testing fsc=1
// 	fscConfigs.thresholdsPerDsp = dspThresholdMap
// 	return func() {
// 		fscConfigs.cache = nil
// 		fscConfigs.thresholdsPerDsp = make(map[int]int)
// 	}
// }

// func ResetInitFscReloaderTest() {
// 	//setting empty to mock routine
// 	initiateFscReloader = func(c dbcache.Cache, expiryTime int) {}
// }
