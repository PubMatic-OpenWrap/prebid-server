package publisherfeature

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type ampMultiformat struct {
	enabledPublishers map[int]struct{}
}

// updateAmpMutiformatEnabledPublishers updates the ampMultiformat enabled publishers
func (fe *feature) updateAmpMutiformatEnabledPublishers() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublishers := make(map[int]struct{})
	for pubID, feature := range fe.publisherFeature {
		for featureID, featureDetails := range feature {
			if featureID == models.FeatureAMPMultiFormat && featureDetails.Enabled == 1 {
				enabledPublishers[pubID] = struct{}{}
			}
		}
	}

	fe.Lock()
	fe.ampMultiformat.enabledPublishers = enabledPublishers
	fe.Unlock()
}

func (fe *feature) IsAmpMultformatEnabled(pubid int) bool {
	fe.RLock()
	defer fe.RUnlock()

	if _, isPresent := fe.ampMultiformat.enabledPublishers[pubid]; isPresent {
		return true
	}
	return false
}
