package publisherfeature

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type appLovinMultiFloors struct {
	enabledPublisherProfile map[int]map[string]models.ApplovinAdUnitFloors
}

func (fe *feature) updateApplovinMultiFloorsFeature() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublisherProfile := make(map[int]map[string]models.ApplovinAdUnitFloors)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureApplovinMultiFloors]; ok && val.Enabled == 1 && len(val.Value) > 0 {
			var profileAdUnitFloors map[string]models.ApplovinAdUnitFloors
			if err := json.Unmarshal([]byte(val.Value), &profileAdUnitFloors); err != nil {
				glog.Errorf("ErrJSONUnmarshalFailed Applovin ABTest Feature: pubid: %d profileAdUnitFloors: %s err: %s", pubID, val.Value, err.Error())
				continue
			}

			if _, pubIDPresent := enabledPublisherProfile[pubID]; !pubIDPresent {
				enabledPublisherProfile[pubID] = make(map[string]models.ApplovinAdUnitFloors)
			}
			for profileID, adUnitFloors := range profileAdUnitFloors {
				enabledPublisherProfile[pubID][profileID] = adUnitFloors
			}
		}
	}
	fe.Lock()
	fe.appLovinMultiFloors.enabledPublisherProfile = enabledPublisherProfile
	fe.Unlock()
}

func (fe *feature) IsApplovinMultiFloorsEnabled(pubID int, profileID string) bool {
	fe.RLock()
	defer fe.RUnlock()
	_, isPresent := fe.appLovinMultiFloors.enabledPublisherProfile[pubID][profileID]
	return isPresent
}

func (fe *feature) GetApplovinMultiFloors(pubID int, profileID string) models.ApplovinAdUnitFloors {
	fe.RLock()
	defer fe.RUnlock()
	if adunitfloors, isPresent := fe.appLovinMultiFloors.enabledPublisherProfile[pubID][profileID]; isPresent {
		return adunitfloors
	}
	return models.ApplovinAdUnitFloors{}
}
