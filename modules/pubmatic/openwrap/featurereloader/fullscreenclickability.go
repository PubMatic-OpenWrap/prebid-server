package featurereloader

import (
	"math/rand"

	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type fsc struct {
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
}

// Exposed to access fsc object
// func GetFscInstance() *fsc {
// 	return &FscConfigs
// }

/*
IsUnderFSCThreshold:- returns fsc 1/0 based on:
1. When publisher has disabled FSC in DB, return 0
2. If FSC is enabled for publisher(default), consider DSP-threshold , and predict value of fsc 0 or 1.
3. If dspId is not present return 0
*/
func (re *reloader) IsUnderFSCThreshold(pubid int, dspid int) int {
	re.RLock()
	defer re.RUnlock()

	if _, isPresent := re.fsc.disabledPublishers[pubid]; isPresent {
		return 0
	}

	if dspThreshold, isPresent := re.fsc.thresholdsPerDsp[dspid]; isPresent && predictFscValue(dspThreshold) {
		return 1
	}
	return 0
}

func predictFscValue(threshold int) bool {
	return (rand.Intn(100)) < threshold
}

// fetch and update fsc config maps from DB
func updateFscConfigMapsFromCache(c cache.Cache, publisherFeatureMap map[int]int) error {
	var err error
	thresholdsPerDsp, errDspFsc := c.GetFSCThresholdPerDSP()
	if errDspFsc != nil {
		err = models.ErrorWrap(err, errDspFsc)
	}
	if err != nil {
		return err
	}

	disabledPublishers := make(map[int]struct{})
	for pubID, featureID := range publisherFeatureMap {
		if featureID == models.FeatureFSC {
			disabledPublishers[pubID] = struct{}{}
		}
	}

	reloaderConfig.Lock()
	reloaderConfig.fsc.disabledPublishers = disabledPublishers
	reloaderConfig.fsc.thresholdsPerDsp = thresholdsPerDsp
	reloaderConfig.Unlock()
	return nil
}

// IsFscApplicable returns true if fsc can be applied (fsc=1)
func IsFscApplicable(pubId int, seat string, dspId int) bool {
	return models.IsPubmaticCorePartner(seat) && (reloaderConfig.IsUnderFSCThreshold(pubId, dspId) != 0)
}

// Exposed for test cases
func SetAndResetFscWithMockCache(mockDb cache.Cache, dspThresholdMap map[int]int) func() {
	reloaderConfig.cache = mockDb
	//mockDspID entry for testing fsc=1
	reloaderConfig.fsc.thresholdsPerDsp = dspThresholdMap
	return func() {
		reloaderConfig.cache = nil
		reloaderConfig.fsc.thresholdsPerDsp = make(map[int]int)
		reloaderConfig.fsc.disabledPublishers = make(map[int]struct{})
	}
}
