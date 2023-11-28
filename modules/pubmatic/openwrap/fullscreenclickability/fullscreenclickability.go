package fullscreenclickability

import (
	"math/rand"

	"sync"

	"github.com/golang/glog"
	cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
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
	fscConfigs = fsc{
		cache:              c,
		disabledPublishers: make(map[int]struct{}),
		thresholdsPerDsp:   make(map[int]int),
		serviceStop:        make(chan struct{}),
	}

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

// fetch and update fsc config maps from DB
func updateFscConfigMapsFromCache(c cache.Cache) {
	var err error
	disabledPublishers, errPubFsc := c.GetFSCDisabledPublishers()
	if errPubFsc != nil {
		err = models.ErrorWrap(err, errPubFsc)
	}
	thresholdsPerDsp, errDspFsc := c.GetFSCThresholdPerDSP()
	if errDspFsc != nil {
		err = models.ErrorWrap(err, errDspFsc)
	}
	if err != nil {
		glog.Error(err.Error())
		return
	}
	fscConfigs.Lock()
	fscConfigs.disabledPublishers = disabledPublishers
	fscConfigs.thresholdsPerDsp = thresholdsPerDsp
	fscConfigs.Unlock()
}

// IsFscApplicable returns true if fsc can be applied (fsc=1)
func IsFscApplicable(pubId int, seat string, dspId int) bool {
	return models.IsPubmaticCorePartner(seat) && (fscConfigs.IsUnderFSCThreshold(pubId, dspId) != 0)
}

// Exposed for test cases
func SetAndResetFscWithMockCache(mockDb cache.Cache, dspThresholdMap map[int]int) func() {
	fscConfigs.cache = mockDb
	//mockDspID entry for testing fsc=1
	fscConfigs.thresholdsPerDsp = dspThresholdMap
	return func() {
		fscConfigs.cache = nil
		fscConfigs.thresholdsPerDsp = make(map[int]int)
		fscConfigs.disabledPublishers = make(map[int]struct{})
	}
}
