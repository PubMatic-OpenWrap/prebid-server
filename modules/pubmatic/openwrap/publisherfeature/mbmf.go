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
	instlFloors              [2]map[int]*models.MultiFloors
	rwddFloors               [2]map[int]*models.MultiFloors
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
		instlFloors: [2]map[int]*models.MultiFloors{
			make(map[int]*models.MultiFloors),
			make(map[int]*models.MultiFloors),
		},
		rwddFloors: [2]map[int]*models.MultiFloors{
			make(map[int]*models.MultiFloors),
			make(map[int]*models.MultiFloors),
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
	fe.updateProfileAdUnitLevelFloors()
	fe.updateMBMFInstlFloors()
	fe.updateMBMFRwddFloors()
	fe.mbmf.index ^= 1
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

// updateProfileAdUnitLevelFloors updates profileAdUnitLevelFloors fetched from DB to pubFeatureMap
func (fe *feature) updateProfileAdUnitLevelFloors() {
	floors, err := fe.cache.GetProfileAdUnitMultiFloors()
	if err != nil || floors == nil {
		return
	}
	fe.mbmf.profileAdUnitLevelFloors[fe.mbmf.index^1] = floors
}

// updateMBMFInstlFloors updates mbmfInstlFloors fetched from DB to pubFeatureMap
func (fe *feature) updateMBMFInstlFloors() {
	instlFloors := make(map[int]*models.MultiFloors)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFInstlFloors]; ok {
			// If is_enabled is 0, set multifloors IsActive false for publisher
			if val.Enabled == 0 {
				instlFloors[pubID] = &models.MultiFloors{IsActive: false}
				continue
			}

			floors := models.MultiFloors{IsActive: true}
			if err := json.Unmarshal([]byte(val.Value), &floors); err != nil {
				glog.Errorf(models.ErrMBMFFloorsUnmarshal, pubID, "", err.Error())
				continue
			}
			instlFloors[pubID] = &floors
		}
	}
	fe.mbmf.instlFloors[fe.mbmf.index^1] = instlFloors
}

// updateMBMFRwddFloors updates mbmfRwddFloors fetched from DB to pubFeatureMap
func (fe *feature) updateMBMFRwddFloors() {
	rwddFloors := make(map[int]*models.MultiFloors)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFRwddFloors]; ok {
			// If is_enabled is 0, set multifloors IsActive false for publisher
			if val.Enabled == 0 {
				rwddFloors[pubID] = &models.MultiFloors{IsActive: false}
				continue
			}

			floors := models.MultiFloors{IsActive: true}
			if err := json.Unmarshal([]byte(val.Value), &floors); err != nil {
				glog.Errorf(models.ErrMBMFFloorsUnmarshal, pubID, "", err.Error())
				continue
			}
			rwddFloors[pubID] = &floors
		}
	}
	fe.mbmf.rwddFloors[fe.mbmf.index^1] = rwddFloors
}

// IsMBMFCountry returns true if country specified for MBMF in DB
func (fe *feature) IsMBMFCountry(countryCode string) bool {
	countries := fe.mbmf.enabledCountries[fe.mbmf.index]
	_, isPresent := countries[countryCode]
	return isPresent
}

// IsMBMFPublisherEnabled returns true if publisher not present in DB or it is present in is_enabled=1
func (fe *feature) IsMBMFPublisherEnabled(pubID int) bool {
	publishers := fe.mbmf.enabledPublishers[fe.mbmf.index]
	isPublisherEnabled, isPresent := publishers[pubID]
	if !isPresent {
		return true
	}
	return isPublisherEnabled
}

// IsMBMFEnabledForAdUnitFormat returns true if publisher entry no present OR it is present but is_enabled=1 for given adformat
func (fe *feature) IsMBMFEnabledForAdUnitFormat(pubID int, adUnitFormat string) bool {
	if adUnitFormat == models.AdUnitFormatInstl {
		instlFloors := fe.mbmf.instlFloors[fe.mbmf.index]
		multiFloors, isPresent := instlFloors[pubID]
		//(pub entry not present or value stored in DB is incorrect) OR (pub entry present and it is enabled)
		if !isPresent || multiFloors.IsActive {
			return true
		}
	}
	if adUnitFormat == models.AdUnitFormatRwddVideo {
		rwddFloors := fe.mbmf.rwddFloors[fe.mbmf.index]
		multiFloors, isPresent := rwddFloors[pubID]
		//(pub entry not present or value stored in DB is incorrect) OR (pub entry present and it is enabled)
		if !isPresent || multiFloors.IsActive {
			return true
		}
	}
	return false
}

// GetMBMFFloorsForAdUnitFormat returns floors for publisher specified for MBMF in DB
func (fe *feature) GetMBMFFloorsForAdUnitFormat(pubID int, adunitFormat string) *models.MultiFloors {
	var floors map[int]*models.MultiFloors

	switch adunitFormat {
	case models.AdUnitFormatInstl:
		floors = fe.mbmf.instlFloors[fe.mbmf.index]
	case models.AdUnitFormatRwddVideo:
		floors = fe.mbmf.rwddFloors[fe.mbmf.index]
	default:
		return nil
	}

	adFormatFloors, ok := floors[pubID]
	if ok && adFormatFloors != nil {
		return adFormatFloors
	}

	defaultFloors := floors[models.DefaultAdUnitFormatFloors]
	if defaultFloors != nil {
		return defaultFloors
	}
	glog.Errorf("MBMF default floors not found for pubID %d and adunitFormat %s", pubID, adunitFormat)
	return nil
}

// GetProfileAdUnitMultiFloors returns adunitlevel floors for publisher specified for MBMF in DB
func (fe *feature) GetProfileAdUnitMultiFloors(profileID int) map[string]*models.MultiFloors {
	profileAdUnitfloors := fe.mbmf.profileAdUnitLevelFloors[fe.mbmf.index]
	adunitFloors, ok := profileAdUnitfloors[profileID]
	if !ok {
		return nil
	}
	return adunitFloors
}
