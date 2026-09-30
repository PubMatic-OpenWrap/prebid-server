package publisherfeature

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type maxFloors struct {
	enabledPublishers map[int]struct{}
}

// updateMaxFloorsEnabledPublishers updates the maxFloors enabled publishers
func (fe *feature) updateMaxFloorsEnabledPublishers() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublishers := make(map[int]struct{})
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMaxFloors]; ok && val.Enabled == 1 {
			enabledPublishers[pubID] = struct{}{}
		}
	}

	fe.Lock()
	fe.maxFloors.enabledPublishers = enabledPublishers
	fe.Unlock()
}

func (fe *feature) IsMaxFloorsEnabled(pubid int) bool {
	fe.RLock()
	defer fe.RUnlock()

	_, isPresent := fe.maxFloors.enabledPublishers[pubid]
	return isPresent

}
