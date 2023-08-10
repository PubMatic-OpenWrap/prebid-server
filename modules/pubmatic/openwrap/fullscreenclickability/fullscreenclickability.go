package fullscreenclickability

import (
	"math/rand"

	"sync"

	"github.com/golang/glog"
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type fsc struct {
	cache              cache.Cache
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
	sync.RWMutex
	serviceStop chan struct{}
}

var fscConfigs fsc

// Initializing reloader with cache-refresh default-expiry + 30 mins (to avoid DB load post cache refresh)
func Init(c cache.Cache, defaultExpiry int) {
	//init fsc configs
	fscConfigs.cache = c
	fscConfigs.disabledPublishers = make(map[int]struct{})
	fscConfigs.thresholdsPerDsp = make(map[int]int)
	fscConfigs.serviceStop = make(chan struct{})

	go initiateReloader(c, defaultExpiry+1800)
	glog.Info("Initialized FSC cache update reloaders for publisher and dsp fsc configuraitons")

}

// Exposed to access fsc object
func GetFscInstance() *fsc {
	return &fscConfigs
}

/*
IsUnderFSCThreshold:- returns fsc 1/0 based on:
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

func StopReloaderService() {
	//updating serviceStop flag to true
	close(fscConfigs.serviceStop)
}

// fetch and update fsc config maps from DB
func updateFscConfigMapsFromCache(c cache.Cache) {
	disabledPublishers, err := c.GetFSCDisabledPublishers()
	if err != nil {
		glog.Error("ErrUpdateFscCache:", err.Error())
	}
	thresholdsPerDsp, err := c.GetFSCThresholdPerDSP()
	if err != nil {
		glog.Error("ErrUpdateFscCache:", err.Error())
	}
	fscConfigs.Lock()
	fscConfigs.disabledPublishers = disabledPublishers
	fscConfigs.thresholdsPerDsp = thresholdsPerDsp
	fscConfigs.Unlock()
}

// IsFscApplicable returns true if fsc can be applied (fsc=1)
func IsFscApplicable(pubId int, seat string, dspId int) bool {
	if models.IsPubmaticCorePartner(seat) && (fscConfigs.IsUnderFSCThreshold(pubId, dspId) != 0) {
		return true
	}
	return false
}

// Exposed for test cases
func SetAndResetFscWithMockCache(mockDb cache.Cache, dspThresholdMap map[int]int) func() {
	fscConfigs.cache = mockDb
	//mockDspID entry for testing fsc=1
	fscConfigs.thresholdsPerDsp = dspThresholdMap
	return func() {
		fscConfigs.cache = nil
		fscConfigs.thresholdsPerDsp = make(map[int]int)
	}
}

// func ResetInitFscReloaderTest() {
// 	//setting empty to mock routine
// 	initiateReloader = func(c cache.Cache, expiryTime int) {}
// }
