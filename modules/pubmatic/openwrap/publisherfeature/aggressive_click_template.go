package publisherfeature

import (
	"math/rand"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type act struct {
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
}

// updateActConfigMapsFromCache update the act disabled publishers and thresholds per dsp
func (fe *feature) updateActConfigMapsFromCache() error {
	thresholdsPerDsp, err := fe.cache.GetACTThresholdPerDSP()
	if err != nil {
		return err
	}

	disabledPublishers := make(map[int]struct{})
	if fe.publisherFeature != nil {
		for pubID, feature := range fe.publisherFeature {
			if val, ok := feature[models.FeatureACT]; ok && val.Enabled == 0 {
				disabledPublishers[pubID] = struct{}{}
			}
		}
	}

	fe.Lock()
	fe.act.disabledPublishers = disabledPublishers
	if thresholdsPerDsp != nil {
		fe.act.thresholdsPerDsp = thresholdsPerDsp
	}
	fe.Unlock()
	return nil
}

/*
isUnderACTThreshold:- returns act 1/0 based on:
1. When publisher has disabled ACT in DB, return 0
2. If ACT is enabled for publisher(default), consider DSP-threshold , and predict value of act 0 or 1.
3. If dspId is not present return 0
*/
func (fe *feature) isUnderACTThreshold(pubid int, dspid int) int {
	fe.RLock()
	defer fe.RUnlock()

	if _, isPresent := fe.act.disabledPublishers[pubid]; isPresent {
		return 0
	}

	if dspThreshold, isPresent := fe.act.thresholdsPerDsp[dspid]; isPresent && predictActValue(dspThreshold) {
		return 1
	}
	return 0
}

func predictActValue(threshold int) bool {
	return (rand.Intn(100)) < threshold
}

// IsActApplicable returns true if act can be applied (act=1)
func (fe *feature) IsActApplicable(pubId int, seat string, dspId int) bool {
	return models.IsPubmaticCorePartner(seat) && (fe.isUnderACTThreshold(pubId, dspId) != 0)
}
