package publisherfeature

import (
	"encoding/json"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type mbmf struct {
	enabledCountries         [2]models.HashSet
	enabledPublishers        [2]map[int]bool
	profileAdUnitLevelFloors [2]models.ProfileAdUnitMultiFloors
	instlFloors              [2]map[int]models.MultiFloors
	rwddFloors               [2]map[int]models.MultiFloors
	index                    int32
}

func newMBMF() mbmf {
	return mbmf{
		enabledCountries: [2]models.HashSet{
			make(models.HashSet),
			make(models.HashSet),
		},
		enabledPublishers: [2]map[int]bool{
			make(map[int]bool),
			make(map[int]bool),
		},
		profileAdUnitLevelFloors: [2]models.ProfileAdUnitMultiFloors{
			make(models.ProfileAdUnitMultiFloors),
			make(models.ProfileAdUnitMultiFloors),
		},
		instlFloors: [2]map[int]models.MultiFloors{
			make(map[int]models.MultiFloors),
			make(map[int]models.MultiFloors),
		},
		rwddFloors: [2]map[int]models.MultiFloors{
			make(map[int]models.MultiFloors),
			make(map[int]models.MultiFloors),
		},
		index: 0,
	}
}

func (fe *feature) updateMBMF() {
	if fe.publisherFeature == nil {
		return
	}
	fe.updateMBMFCountries()
	fe.updateMBMFPublishers()
	fe.updateMbmfInstlFloors()
	fe.updateMbmfRwddFloors()
	fe.updateProfileAdUnitLevelFloors()
	fe.mbmf.index ^= 1
}

// updateGDPRCountryCodes updates gdprCountryCodes fetched from DB to pubFeatureMap
func (fe *feature) updateProfileAdUnitLevelFloors() {
	floors, err := fe.cache.GetProfileAdUnitMultiFloors()
	if err != nil || floors == nil {
		return
	}
	fe.mbmf.profileAdUnitLevelFloors[fe.mbmf.index^1] = floors
}

func (fe *feature) updateMBMFCountries() {
	enabledCountries := make(models.HashSet)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFCountry]; ok && pubID == 0 && val.Enabled == 1 {
			countries := strings.Split(val.Value, ",")
			for _, country := range countries {
				country = strings.TrimSpace(country)
				if country != "" {
					enabledCountries[country] = struct{}{}
				}
			}
		}
	}
	fe.mbmf.enabledCountries[fe.mbmf.index^1] = enabledCountries
}

func (fe *feature) updateMBMFPublishers() {
	enabledPublishers := make(map[int]bool)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFPublisher]; ok {
			enabledPublishers[pubID] = val.Enabled == 1
		}
	}
	fe.mbmf.enabledPublishers[fe.mbmf.index^1] = enabledPublishers
}

func (fe *feature) updateMbmfInstlFloors() {
	instlFloors := make(map[int]models.MultiFloors)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFInstlFloors]; ok && val.Enabled == 1 {
			var floors models.MultiFloors
			if err := json.Unmarshal([]byte(val.Value), &floors); err != nil {
				glog.Errorf(models.ErrMBMFFloorsUnmarshal, pubID, "", err.Error())
				continue
			}
			instlFloors[pubID] = floors
		}
	}
	fe.mbmf.instlFloors[fe.mbmf.index^1] = instlFloors
}

func (fe *feature) updateMbmfRwddFloors() {
	rwddFloors := make(map[int]models.MultiFloors)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFRwddFloors]; ok && val.Enabled == 1 {
			var floors models.MultiFloors
			if err := json.Unmarshal([]byte(val.Value), &floors); err != nil {
				glog.Errorf(models.ErrMBMFFloorsUnmarshal, pubID, "", err.Error())
				continue
			}
			rwddFloors[pubID] = floors
		}
	}
	fe.mbmf.rwddFloors[fe.mbmf.index^1] = rwddFloors
}

// func (fe *feature) GetMbmfEnabledCountries() models.HashSet {
// 	return fe.mbmf.enabledCountries[fe.mbmf.index]
// }

// IsMBMFCountry returns true if country specified for MBMF in DB
func (fe *feature) IsMBMFCountry(countryCode string) bool {
	countries := fe.mbmf.enabledCountries[fe.mbmf.index]
	_, isPresent := countries[countryCode]
	return isPresent
}

// IsMBMFPublisherInDB returns true if publisher specified for MBMF in DB
func (fe *feature) IsMBMFPublisherInDB(pubID int) bool {
	publishers := fe.mbmf.enabledPublishers[fe.mbmf.index]
	_, isPresent := publishers[pubID]
	return isPresent
}

// IsMBMFPublisherEnabled returns true if publisher specified for MBMF in DB
func (fe *feature) IsMBMFPublisherEnabled(pubID int) bool {
	publishers := fe.mbmf.enabledPublishers[fe.mbmf.index]
	isPublisherEnabled, isPresent := publishers[pubID]
	return isPresent && isPublisherEnabled
}

// IsMBMFEnabledForAdUnitFormat returns true if publisher specified for MBMF in DB
func (fe *feature) IsMBMFEnabledForAdUnitFormat(pubID int, adUnitFormat string) bool {
	if adUnitFormat == models.AdUnitFormatInstl {
		_, present := fe.mbmf.instlFloors[fe.mbmf.index][pubID]
		return present
	}
	if adUnitFormat == models.AdUnitFormatRwddVideo {
		_, present := fe.mbmf.rwddFloors[fe.mbmf.index][pubID]
		return present
	}
	return false
}

// GetMBMFFloorsForAdUnitFormat returns floors for publisher specified for MBMF in DB
func (fe *feature) GetMBMFFloorsForAdUnitFormat(pubID int, adUnitFormat string) models.MultiFloors {
	var floors map[int]models.MultiFloors

	switch adUnitFormat {
	case models.AdUnitFormatInstl:
		floors = fe.mbmf.instlFloors[fe.mbmf.index]
	case models.AdUnitFormatRwddVideo:
		floors = fe.mbmf.rwddFloors[fe.mbmf.index]
	default:
		return models.MultiFloors{}
	}

	adFormatFloors, ok := floors[pubID]
	if !ok {
		return floors[models.DefaultAdUnitFormatFloors]
	}
	return adFormatFloors
}

// GetProfileAdUnitMultiFloors returns adunitlevel floors for publisher specified for MBMF in DB
func (fe *feature) GetProfileAdUnitMultiFloors(profileID int) map[string]models.MultiFloors {
	profileAdUnitfloors := fe.mbmf.profileAdUnitLevelFloors[fe.mbmf.index]
	adunitFloors, ok := profileAdUnitfloors[profileID]
	if !ok {
		return nil
	}
	return adunitFloors
}
