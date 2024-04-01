package publisherfeature

import (
	"math/rand"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type fsc struct {
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
}

// fetch and update fsc config maps from DB
func (fe *feature) updateFscConfigMapsFromCache() error {
	var err error
	thresholdsPerDsp, errDspFsc := fe.cache.GetFSCThresholdPerDSP()
	if errDspFsc != nil {
		err = models.ErrorWrap(err, errDspFsc)
	}
	if err != nil {
		return err
	}

	disabledPublishers := make(map[int]struct{})
	for pubID, featureID := range fe.publisherFeature {
		if featureID == models.FeatureFSC {
			disabledPublishers[pubID] = struct{}{}
		}
	}

	fe.Lock()
	fe.fsc.disabledPublishers = disabledPublishers
	fe.fsc.thresholdsPerDsp = thresholdsPerDsp
	fe.Unlock()
	return nil
}

/*
IsUnderFSCThreshold:- returns fsc 1/0 based on:
1. When publisher has disabled FSC in DB, return 0
2. If FSC is enabled for publisher(default), consider DSP-threshold , and predict value of fsc 0 or 1.
3. If dspId is not present return 0
*/
func (re *feature) isUnderFSCThreshold(pubid int, dspid int) int {
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

// IsFscApplicable returns true if fsc can be applied (fsc=1)
func (fe *feature) IsFscApplicable(pubId int, seat string, dspId int) bool {
	return models.IsPubmaticCorePartner(seat) && (fe.isUnderFSCThreshold(pubId, dspId) != 0)
}
