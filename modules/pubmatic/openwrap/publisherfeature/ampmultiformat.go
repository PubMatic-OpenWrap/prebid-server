package publisherfeature

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type ampMultiformat struct {
	enabledPublishers map[int]struct{}
}

// fetch and update fsc config maps from DB
func (fe *feature) updateAmpMutiformatConfigFromCache() {

	enabledPublishers := make(map[int]struct{})
	for pubID, featureID := range fe.publisherFeature {
		if featureID == models.FeatureAMPMultiFormat {
			enabledPublishers[pubID] = struct{}{}
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
