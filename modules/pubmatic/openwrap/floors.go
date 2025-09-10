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

	appLevelDynamicFloorFlag := versionConfigMap[models.FloorModuleEnabled]
	// By default enforcemnt will be true i.e hard floor
	requestExt.Prebid.Floors.Enforcement.EnforcePBS = ptrutil.ToPtr(true)
	if versionConfigMap[models.PLATFORM_KEY] == models.PLATFORM_APP {
		// floor deal will be true by default for platform in-app
		requestExt.Prebid.Floors.Enforcement.FloorDeals = ptrutil.ToPtr(true)
		if !isDynamicFloorEnabledPub || appLevelDynamicFloorFlag != "1" {
			requestExt.Prebid.Floors.SetMaxFloor = setMaxFloor
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

	if requestExt.Prebid.Floors.Enforcement.FloorDeals == nil {
		dealsEnforcement, ok := versionConfigMap[models.FloorDealEnforcement]
		if ok && dealsEnforcement == "0" {
			*requestExt.Prebid.Floors.Enforcement.FloorDeals = false
		}
	}

	if requestExt.Prebid.Floors.Enforcement.EnforcePBS == nil {
		floorType, typeExists := versionConfigMap[models.FloorType]
		if typeExists && strings.ToLower(floorType) == models.SoftFloorType {
			*requestExt.Prebid.Floors.Enforcement.EnforcePBS = false
		}
	}

	//Based on floorPriceModuleEnabled(appLevelDynamicFloorFlag) flag, dynamic fetch would be enabled/disabled
	url, urlExists := versionConfigMap[models.PriceFloorURL]
	if urlExists {
		if requestExt.Prebid.Floors.Location == nil {
			requestExt.Prebid.Floors.Location = new(openrtb_ext.PriceFloorEndpoint)
		}
		requestExt.Prebid.Floors.Location.URL = url
	}
}
