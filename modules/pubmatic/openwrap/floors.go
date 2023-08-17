package openwrap

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func setPriceFloorFetchURL(requestExt *models.RequestExt, configMap map[int]map[string]string) {

	if configMap == nil || configMap[models.VersionLevelConfigID] == nil {
		return
	}

	if requestExt != nil && requestExt.Prebid.Floors != nil && requestExt.Prebid.Floors.Enabled != nil && !(*requestExt.Prebid.Floors.Enabled) {
		return
	}

	url, urlExists := configMap[models.VersionLevelConfigID][models.PriceFloorURL]
	if urlExists {
		if requestExt.Prebid.Floors == nil {
			requestExt.Prebid.Floors = &openrtb_ext.PriceFloorRules{}
		}
		if requestExt.Prebid.Floors.Enabled == nil {
			requestExt.Prebid.Floors.Enabled = new(bool)
		}
		*requestExt.Prebid.Floors.Enabled = true

		enable, enabledExists := configMap[models.VersionLevelConfigID][models.FloorModuleEnabled]
		if enabledExists && enable != "1" {
			*requestExt.Prebid.Floors.Enabled = false
			return
		}

		if requestExt.Prebid.Floors.Location == nil {
			requestExt.Prebid.Floors.Location = new(openrtb_ext.PriceFloorEndpoint)
		}
		requestExt.Prebid.Floors.Location.URL = url
	}

}
