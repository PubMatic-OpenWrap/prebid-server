package publisherfeature

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// updateFscAndActConfigMapsFromCache updates both FSC and ACT config maps from cache in one call
// using GetFSCAndACTThresholdsPerDSP to avoid duplicate DB/cache round-trips.
func (fe *feature) updateFscAndActConfigMapsFromCache() error {
	fscThresholds, actThresholds, err := fe.cache.GetFSCAndACTThresholdsPerDSP()
	if err != nil {
		return err
	}

	disabledFsc := make(map[int]struct{})
	disabledAct := make(map[int]struct{})
	if fe.publisherFeature != nil {
		for pubID, feature := range fe.publisherFeature {
			if val, ok := feature[models.FeatureFSC]; ok && val.Enabled == 0 {
				disabledFsc[pubID] = struct{}{}
			}
			if val, ok := feature[models.FeatureACT]; ok && val.Enabled == 0 {
				disabledAct[pubID] = struct{}{}
			}
		}
	}

	fe.Lock()
	fe.fsc.disabledPublishers = disabledFsc
	fe.act.disabledPublishers = disabledAct
	if fscThresholds != nil {
		fe.fsc.thresholdsPerDsp = fscThresholds
	}
	if actThresholds != nil {
		fe.act.thresholdsPerDsp = actThresholds
	}
	fe.Unlock()
	return nil
}
