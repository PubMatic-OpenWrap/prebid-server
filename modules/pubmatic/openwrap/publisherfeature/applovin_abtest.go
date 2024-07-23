package publisherfeature

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type applovinABTest struct {
	enabledPublisherProfile map[int]map[string]models.ApplovinAdUnitFloors
}

func (fe *feature) updateAdunitConfigFeature() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublisherProfile := make(map[int]map[string]models.ApplovinAdUnitFloors)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureAdUnitConfig]; ok && val.Enabled == 1 && len(val.Value) > 0 {
			var profileAdUnitConfig map[string]models.ApplovinAdUnitFloors
			if err := json.Unmarshal([]byte(val.Value), &profileAdUnitConfig); err != nil {
				glog.Errorf("ErrJSONUnmarshalFailed Applovin ABTest Feature: pubid: %d profileadunitconfig: %s err: %s", pubID, val.Value, err.Error())
				continue
			}

			for profileID, adUnitConfig := range profileAdUnitConfig {
				enabledPublisherProfile[pubID][profileID] = adUnitConfig
			}
		}
	}
	fe.Lock()
	fe.applovinABTest.enabledPublisherProfile = enabledPublisherProfile
	fe.Unlock()
}

func (fe *feature) IsApplovinABTestEnabled(pubID int, profileID string) bool {
	fe.RLock()
	defer fe.RUnlock()
	_, isPresent := fe.applovinABTest.enabledPublisherProfile[pubID][profileID]
	return isPresent
}

func (fe *feature) GetApplovinMaxFloors(pubID int, profileID string) models.ApplovinAdUnitFloors {
	fe.RLock()
	defer fe.RUnlock()
	if adunitfloors, isPresent := fe.applovinABTest.enabledPublisherProfile[pubID][profileID]; isPresent {
		return adunitfloors
	}
	return models.ApplovinAdUnitFloors{}
}
