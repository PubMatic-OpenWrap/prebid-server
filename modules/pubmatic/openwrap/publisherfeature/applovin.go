package publisherfeature

import (
	"encoding/json"
	"strconv"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type appLovinMultiFloors struct {
	enabledPublisherProfile map[int]map[string]models.ApplovinAdUnitFloors
}

type appLovinMaxSchain struct {
	enabledPublisherProfile map[int]int
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

func (fe *feature) updateApplovinMaxSchainFeature() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublisherProfile := make(map[int]int)

	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureAppLovinMaxSchain]; ok && val.Enabled == 1 && len(val.Value) > 0 {
			percentage, err := strconv.Atoi(val.Value)
			if err != nil {
				glog.Errorf("ErrInvalidPercentage ApplovinMaxSchain Feature: pubid: %d value: %s err: %s",
					pubID, val.Value, err.Error())
				continue
			}

			enabledPublisherProfile[pubID] = percentage
		}
	}

	fe.Lock()
	fe.appLovinMaxSchain.enabledPublisherProfile = enabledPublisherProfile
	fe.Unlock()
}

func (fe *feature) IsApplovinMaxSchainEnabled(pubID int) bool {
	fe.RLock()
	defer fe.RUnlock()
	_, isPresent := fe.appLovinMaxSchain.enabledPublisherProfile[pubID]
	return isPresent
}

func (fe *feature) GetApplovinMaxSchainPercentage(pubID int) int {
	fe.RLock()
	defer fe.RUnlock()
	if percentage, isPresent := fe.appLovinMaxSchain.enabledPublisherProfile[pubID]; isPresent {
		return percentage
	}
	return 0
}
