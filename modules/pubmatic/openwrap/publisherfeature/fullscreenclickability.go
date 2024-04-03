package publisherfeature

import (
	"math/rand"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type fsc struct {
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
}

// updateFscConfigMapsFromCache update the fsc disabled publishers and thresholds per dsp
func (fe *feature) updateFscConfigMapsFromCache() error {
	if fe.publisherFeature == nil {
		return nil
	}

	thresholdsPerDsp, err := fe.cache.GetFSCThresholdPerDSP()
	if err != nil {
		return err
	}

	disabledPublishers := make(map[int]struct{})
	for pubID, feature := range fe.publisherFeature {
		for featureID, featureDetails := range feature {
			if featureID == models.FeatureFSC && featureDetails.Enabled == 0 {
				disabledPublishers[pubID] = struct{}{}
			}
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
func (fe *feature) isUnderFSCThreshold(pubid int, dspid int) int {
	fe.RLock()
	defer fe.RUnlock()

	if _, isPresent := fe.fsc.disabledPublishers[pubid]; isPresent {
		return 0
	}

	if dspThreshold, isPresent := fe.fsc.thresholdsPerDsp[dspid]; isPresent && predictFscValue(dspThreshold) {
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
