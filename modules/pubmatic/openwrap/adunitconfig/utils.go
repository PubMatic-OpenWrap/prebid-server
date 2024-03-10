package adunitconfig

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

// TODO use this
func GetMatchedSlotName(rCtx models.RequestCtx, imp openrtb2.Imp, impExt models.ImpExtension) (slotAdUnitConfig *adunitconfig.AdConfig, isRegex bool) {
	div := ""
	var height, width int64
	if imp.Video != nil {
		if imp.Video.H != nil {
			height = *imp.Video.H
		}
		if imp.Video.W != nil {
			width = *imp.Video.W
		}
	}
	tagID := imp.TagID

	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	slotName := models.GenerateSlotName(height, width, rCtx.AdUnitConfig.ConfigPattern, tagID, div, rCtx.Source)

	var ok bool
	slotAdUnitConfig, ok = rCtx.AdUnitConfig.Config[slotName]
	if ok {
		return
	}

	// for slot, adUnitConfig := range rCtx.AdUnitConfig.Config {

	// }

	return
}

func getDefaultAllowedConnectionTypes(adUnitConfigMap *adunitconfig.AdUnitConfig) []int {
	if adUnitConfigMap == nil {
		return nil
	}

	if v, ok := adUnitConfigMap.Config[models.AdunitConfigDefaultKey]; ok && v.Video != nil && v.Video.Config != nil && len(v.Video.Config.CompanionType) != 0 {
		return v.Video.Config.ConnectionType
	}

	return nil
}

func checkValuePresentInArray(intArray []int, value int) bool {
	for _, eachVal := range intArray {
		if eachVal == value {
			return true
		}
	}
	return false
}

// update slotConfig with final AdUnit config to apply with
func getFinalSlotAdUnitConfig(slotConfig, defaultConfig *adunitconfig.AdConfig) *adunitconfig.AdConfig {
	// nothing available
	if slotConfig == nil && defaultConfig == nil {
		return nil
	}

	// only default available
	if slotConfig == nil {
		return defaultConfig
	}

	// only slot available
	if defaultConfig == nil {
		return slotConfig
	}

	// both available, merge both with priority to slot

	if (slotConfig.BidFloor == nil || *slotConfig.BidFloor == 0.0) && defaultConfig.BidFloor != nil {
		slotConfig.BidFloor = defaultConfig.BidFloor

		slotConfig.BidFloorCur = ptrutil.ToPtr(models.USD)
		if defaultConfig.BidFloorCur != nil {
			slotConfig.BidFloorCur = defaultConfig.BidFloorCur
		}
	}

	//slotConfig has bidfloor and not have BidFloorCur set by default USD
	if slotConfig.BidFloor != nil && *slotConfig.BidFloor > float64(0) && slotConfig.BidFloorCur == nil {
		slotConfig.BidFloorCur = ptrutil.ToPtr(models.USD)
	}

	if slotConfig.Banner == nil {
		slotConfig.Banner = defaultConfig.Banner
	}

	if slotConfig.Video == nil {
		slotConfig.Video = defaultConfig.Video
	}

	if slotConfig.Floors == nil {
		slotConfig.Floors = defaultConfig.Floors
	}

	return slotConfig
}
