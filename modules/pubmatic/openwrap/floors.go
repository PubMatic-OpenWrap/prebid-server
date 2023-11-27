package openwrap

import (
	"strings"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

func setFloorsExt(requestExt *models.RequestExt, configMap map[int]map[string]string) {
	if configMap == nil || configMap[models.VersionLevelConfigID] == nil {
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

	if requestExt.Prebid.Floors.Enforcement.EnforcePBS == nil {
		// By default enforcemnt will be true i.e hard floor
		requestExt.Prebid.Floors.Enforcement.EnforcePBS = ptrutil.ToPtr(true)

		floorType, typeExists := configMap[models.VersionLevelConfigID][models.FloorType]
		if typeExists && strings.ToLower(floorType) == models.SoftFloorType {
			*requestExt.Prebid.Floors.Enforcement.EnforcePBS = false
		}
	}

	// Based on floorPriceModuleEnabled flag, dynamic fetch would be enabled/disabled
	enableFlag, isFlagPresent := configMap[models.VersionLevelConfigID][models.FloorModuleEnabled]
	if isFlagPresent && enableFlag == "1" {
		url, urlExists := configMap[models.VersionLevelConfigID][models.PriceFloorURL]
		if urlExists {
			if requestExt.Prebid.Floors.Location == nil {
				requestExt.Prebid.Floors.Location = new(openrtb_ext.PriceFloorEndpoint)
			}
			requestExt.Prebid.Floors.Location.URL = url
		}
	}
}
