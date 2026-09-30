package publisherfeature

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type dynamicFloor struct {
	enabledPublishers [2]map[int]struct{}
	index             int32
}

func newDynamicFloor() dynamicFloor {
	return dynamicFloor{
		enabledPublishers: [2]map[int]struct{}{
			make(map[int]struct{}),
			make(map[int]struct{}),
		},
		index: 0,
	}
}

func (fe *feature) updateDynamicFloorEnabledPublishers() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublishers := make(map[int]struct{})
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureDynamicFloor]; ok && val.Enabled == 1 {
			enabledPublishers[pubID] = struct{}{}
		}
	}

	fe.dynamicFloor.enabledPublishers[fe.dynamicFloor.index^1] = enabledPublishers
	fe.dynamicFloor.index ^= 1
}

func (fe *feature) IsDynamicFloorEnabledPublisher(pubID int) bool {
	if fe.dynamicFloor.enabledPublishers[fe.dynamicFloor.index] == nil {
		return false
	}
	_, isPresent := fe.dynamicFloor.enabledPublishers[fe.dynamicFloor.index][pubID]
	return isPresent
}
