package openwrap

import (
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func setFloorsExt(requestExt *models.RequestExt, configMap map[int]map[string]string, setMaxFloor bool, isDynamicFloorEnabledPub bool) {
	versionConfigMap := configMap[models.VersionLevelConfigID]
	if versionConfigMap == nil {
		return
	}

	if requestExt != nil && requestExt.Prebid.Floors != nil && requestExt.Prebid.Floors.Enabled != nil && !(*requestExt.Prebid.Floors.Enabled) {
		return
	}

	if requestExt.Prebid.Floors == nil {
		requestExt.Prebid.Floors = new(openrtb_ext.PriceFloorRules)
	}
	if requestExt.Prebid.Floors.Enabled == nil {
		requestExt.Prebid.Floors.Enabled = ptrutil.ToPtr(true)
	}

	if requestExt.Prebid.Floors.Enforcement == nil {
		requestExt.Prebid.Floors.Enforcement = new(openrtb_ext.PriceFloorEnforcement)
	}

	if versionConfigMap[models.PLATFORM_KEY] == models.PLATFORM_APP {
		if !isDynamicFloorEnabledPub || versionConfigMap[models.FloorModuleEnabled] != "1" {
			setFloorsDefaultsForApp(requestExt, setMaxFloor)
			return
		}
		setFloorsData(requestExt, versionConfigMap)
		requestExt.Prebid.Floors.SetMaxFloor = true
		return
	}
	setFloorsData(requestExt, versionConfigMap)
	requestExt.Prebid.Floors.SetMaxFloor = setMaxFloor
}

func setFloorsData(requestExt *models.RequestExt, versionConfigMap map[string]string) {
	if requestExt.Prebid.Floors.FloorMin == 0 {
		floorMin, ok := versionConfigMap[models.FloorMin]
		if ok && floorMin != "" {
			floorMinValue, err := strconv.ParseFloat(floorMin, 64)
			if err != nil {
				glog.V(models.LogLevelDebug).Info("Failed to parse floorMin: %s", floorMin)
			}
			requestExt.Prebid.Floors.FloorMin = floorMinValue
		}
	}

	if requestExt.Prebid.Floors.Enforcement.EnforcePBS == nil {
		requestExt.Prebid.Floors.Enforcement.EnforcePBS = ptrutil.ToPtr(true)
		floorType, typeExists := versionConfigMap[models.FloorType]
		if typeExists && strings.ToLower(floorType) == models.SoftFloorType {
			*requestExt.Prebid.Floors.Enforcement.EnforcePBS = false
		}
	}

	if versionConfigMap[models.PLATFORM_KEY] == models.PLATFORM_APP && requestExt.Prebid.Floors.Enforcement.FloorDeals == nil {
		requestExt.Prebid.Floors.Enforcement.FloorDeals = ptrutil.ToPtr(true)
		dealsEnforcement, ok := versionConfigMap[models.FloorDealEnforcement]
		if ok && dealsEnforcement == "0" {
			*requestExt.Prebid.Floors.Enforcement.FloorDeals = false
		}
	}

	//Based on floorPriceModuleEnabled(appLevelDynamicFloorFlag) flag, dynamic fetch would be enabled/disabled
	if versionConfigMap[models.FloorModuleEnabled] == "1" {
		url, urlExists := versionConfigMap[models.PriceFloorURL]
		if urlExists {
			if requestExt.Prebid.Floors.Location == nil {
				requestExt.Prebid.Floors.Location = new(openrtb_ext.PriceFloorEndpoint)
			}
			requestExt.Prebid.Floors.Location.URL = url
		}
	}
}

func setFloorsDefaultsForApp(requestExt *models.RequestExt, setMaxFloor bool) {
	if requestExt.Prebid.Floors.Enforcement.FloorDeals == nil {
		requestExt.Prebid.Floors.Enforcement.FloorDeals = ptrutil.ToPtr(true)
	}
	if requestExt.Prebid.Floors.Enforcement.EnforcePBS == nil {
		requestExt.Prebid.Floors.Enforcement.EnforcePBS = ptrutil.ToPtr(true)
	}
	requestExt.Prebid.Floors.SetMaxFloor = setMaxFloor
}
