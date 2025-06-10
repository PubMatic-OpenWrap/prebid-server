package publisherfeature

import (
	"encoding/json"
	"strings"
	"sync/atomic"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// mbmfData holds all the fields that need double-buffering
type mbmfData struct {
	enabledCountries         map[int]models.HashSet
	enabledPublishers        map[int]bool
	profileAdUnitLevelFloors models.ProfileAdUnitMultiFloors
	instlFloors              map[int]*models.MultiFloors
	rwddFloors               map[int]*models.MultiFloors
}

// mbmf represents Multi-Bid Multi-Floor settings using double-buffering
type mbmf struct {
	data  [2]mbmfData
	index atomic.Int32
}

func newMBMF() *mbmf {
	m := mbmf{
		data: [2]mbmfData{
			{
				enabledCountries:         make(map[int]models.HashSet),
				enabledPublishers:        make(map[int]bool),
				profileAdUnitLevelFloors: make(models.ProfileAdUnitMultiFloors),
				instlFloors:              make(map[int]*models.MultiFloors),
				rwddFloors:               make(map[int]*models.MultiFloors),
			},
			{
				enabledCountries:         make(map[int]models.HashSet),
				enabledPublishers:        make(map[int]bool),
				profileAdUnitLevelFloors: make(models.ProfileAdUnitMultiFloors),
				instlFloors:              make(map[int]*models.MultiFloors),
				rwddFloors:               make(map[int]*models.MultiFloors),
			},
		},
	}
	m.index.Store(0)
	return &m
}

func (fe *feature) updateMBMF() {
	if fe.publisherFeature == nil {
		return
	}
	nextIdx := fe.mbmf.index.Load() ^ 1
	fe.updateMBMFCountries(nextIdx)
	fe.updateMBMFPublishers(nextIdx)
	fe.updateProfileAdUnitLevelFloors(nextIdx)
	fe.updateMBMFInstlFloors(nextIdx)
	fe.updateMBMFRwddFloors(nextIdx)
	fe.mbmf.index.Store(nextIdx)
}

func (fe *feature) updateMBMFCountries(nextIdx int32) {
	enabledCountries := make(map[int]models.HashSet)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFCountry]; ok && val.Enabled == 1 {
			countries := strings.Split(val.Value, ",")
			if _, exists := enabledCountries[pubID]; !exists {
				enabledCountries[pubID] = models.HashSet{}
			}
			for _, country := range countries {
				country = strings.TrimSpace(country)
				if country != "" {
					enabledCountries[pubID][country] = struct{}{}
				}
			}
		}
	}
	fe.mbmf.data[nextIdx].enabledCountries = enabledCountries
}

func (fe *feature) updateMBMFPublishers(nextIdx int32) {
	enabledPublishers := make(map[int]bool)
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureMBMFPublisher]; ok {
			enabledPublishers[pubID] = val.Enabled == 1
		}
	}
	fe.mbmf.data[nextIdx].enabledPublishers = enabledPublishers
}

// updateProfileAdUnitLevelFloors updates profileAdUnitLevelFloors fetched from DB to pubFeatureMap
func (fe *feature) updateProfileAdUnitLevelFloors(nextIdx int32) {
	floors, err := fe.cache.GetProfileAdUnitMultiFloors()
	if err != nil || floors == nil {
		return
	}
	fe.mbmf.data[nextIdx].profileAdUnitLevelFloors = floors
}

// updateMBMFInstlFloors updates mbmfInstlFloors fetched from DB to pubFeatureMap
func (fe *feature) updateMBMFInstlFloors(nextIdx int32) {
	instlFloors := make(map[int]*models.MultiFloors)
	for pubID, feature := range fe.publisherFeature {
		if floors := extractMultiFloors(feature, models.FeatureMBMFInstlFloors, pubID); floors != nil {
			instlFloors[pubID] = floors
		}
	}
	fe.mbmf.data[nextIdx].instlFloors = instlFloors
}

// updateMBMFRwddFloors updates mbmfRwddFloors fetched from DB to pubFeatureMap
func (fe *feature) updateMBMFRwddFloors(nextIdx int32) {
	rwddFloors := make(map[int]*models.MultiFloors)
	for pubID, feature := range fe.publisherFeature {
		if floors := extractMultiFloors(feature, models.FeatureMBMFRwddFloors, pubID); floors != nil {
			rwddFloors[pubID] = floors
		}
	}
	fe.mbmf.data[nextIdx].rwddFloors = rwddFloors
}

// IsMBMFCountryForPublisher returns true if country specified for MBMF in DB
func (fe *feature) IsMBMFCountryForPublisher(countryCode string, pubID int) bool {
	idx := fe.mbmf.index.Load()
	publisherCountries := fe.mbmf.data[idx].enabledCountries
	if pubCountryCodes, isPresent := publisherCountries[pubID]; isPresent {
		// check for countryCode in pubID specific countries
		_, isPresent := pubCountryCodes[countryCode]
		return isPresent
	}

	// check for pubID=0 countries
	_, isPresent := publisherCountries[0][countryCode]
	return isPresent
}

// IsMBMFPublisherEnabled returns true if publisher not present in DB or it is present in is_enabled=1
func (fe *feature) IsMBMFPublisherEnabled(pubID int) bool {
	idx := fe.mbmf.index.Load()
	publishers := fe.mbmf.data[idx].enabledPublishers
	isPublisherEnabled, isPresent := publishers[pubID]
	if !isPresent {
		return true
	}
	return isPublisherEnabled
}

// IsMBMFEnabledForAdUnitFormat returns true if publisher entry is not present
// OR it is present and is_enabled=1 for the given adunit format.
func (fe *feature) IsMBMFEnabledForAdUnitFormat(pubID int, adunitFormat string) bool {
	var floors map[int]*models.MultiFloors
	idx := fe.mbmf.index.Load()

	switch adunitFormat {
	case models.AdUnitFormatInstl:
		floors = fe.mbmf.data[idx].instlFloors
	case models.AdUnitFormatRwddVideo:
		floors = fe.mbmf.data[idx].rwddFloors
	default:
		return false
	}

	multiFloors, isPresent := floors[pubID]
	// Return true if no entry is present or if present and active
	return !isPresent || multiFloors.IsActive
}

// GetMBMFFloorsForAdUnitFormat returns floors for publisher specified for MBMF in DB
func (fe *feature) GetMBMFFloorsForAdUnitFormat(pubID int, adunitFormat string) *models.MultiFloors {
	var floors map[int]*models.MultiFloors
	idx := fe.mbmf.index.Load()

	switch adunitFormat {
	case models.AdUnitFormatInstl:
		floors = fe.mbmf.data[idx].instlFloors
	case models.AdUnitFormatRwddVideo:
		floors = fe.mbmf.data[idx].rwddFloors
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
	idx := fe.mbmf.index.Load()
	profileAdUnitfloors := fe.mbmf.data[idx].profileAdUnitLevelFloors
	adunitFloors, ok := profileAdUnitfloors[profileID]
	if !ok {
		return nil
	}
	return adunitFloors
}

// extractMultiFloors handles extraction of MultiFloors for a given feature key and publisher
func extractMultiFloors(featureMap map[int]models.FeatureData, featureKey int, pubID int) *models.MultiFloors {
	if val, ok := featureMap[featureKey]; ok {
		// If is_enabled is 0, set multifloors IsActive false for publisher
		if val.Enabled == 0 {
			return &models.MultiFloors{IsActive: false}
		}

		floors := models.MultiFloors{IsActive: true}
		if err := json.Unmarshal([]byte(val.Value), &floors); err != nil {
			glog.Errorf(models.ErrMBMFFloorsUnmarshal, pubID, "", err.Error())
			return nil
		}
		return &floors
	}
	return nil
}
