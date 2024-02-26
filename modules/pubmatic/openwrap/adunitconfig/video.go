package adunitconfig

import (
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func UpdateVideoObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, div string, connectionType *adcom1.ConnectionType) (adUnitCtx models.AdUnitCtx) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	videoAdUnitConfigEnabled := true
	defer func() {
		if imp.Video != nil && !videoAdUnitConfigEnabled {
			rCtx.MetricsEngine.RecordImpDisabledViaConfigStats(models.ImpTypeVideo, rCtx.PubIDStr, rCtx.ProfileIDStr)
		}
	}()

	if rCtx.AdUnitConfig == nil || len(rCtx.AdUnitConfig.Config) == 0 {
		return
	}

	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil {
		adUnitCtx.UsingDefaultConfig = true

		if defaultAdUnitConfig.Video != nil && defaultAdUnitConfig.Video.Enabled != nil && !*defaultAdUnitConfig.Video.Enabled {
			videoAdUnitConfigEnabled = false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Video: &adunitconfig.Video{Enabled: &videoAdUnitConfigEnabled}}
			return
		}
	}

	var height, width int64
	if imp.Video != nil  {
		if imp.Video.H != nil {
			height = *imp.Video.H
			}
		if imp.Video.W != nil {
			width = *imp.Video.W
		}
	}

	adUnitCtx.SelectedSlotAdUnitConfig, adUnitCtx.MatchedSlot, adUnitCtx.IsRegex, adUnitCtx.MatchedRegex = selectSlot(rCtx, height, width, imp.TagID, div, rCtx.Source)
	if adUnitCtx.SelectedSlotAdUnitConfig != nil && adUnitCtx.SelectedSlotAdUnitConfig.Video != nil {
		adUnitCtx.UsingDefaultConfig = false
		if adUnitCtx.SelectedSlotAdUnitConfig.Video.Enabled != nil && !*adUnitCtx.SelectedSlotAdUnitConfig.Video.Enabled {
			videoAdUnitConfigEnabled = false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Video: &adunitconfig.Video{Enabled: &videoAdUnitConfigEnabled}}
			return
		}
	}

	adUnitCtx.AppliedSlotAdUnitConfig = getFinalSlotAdUnitConfig(adUnitCtx.SelectedSlotAdUnitConfig, defaultAdUnitConfig)
	if adUnitCtx.AppliedSlotAdUnitConfig == nil {
		return
	}

	adUnitCtx.AllowedConnectionTypes = getDefaultAllowedConnectionTypes(rCtx.AdUnitConfig)

	// updateAllowedConnectionTypes := !adUnitCtx.UsingDefaultConfig
	// if adUnitCtx.AppliedSlotAdUnitConfig != nil && adUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
	// 	adUnitCtx.AppliedSlotAdUnitConfig.Video.Config != nil && len(adUnitCtx.AppliedSlotAdUnitConfig.Video.Config.ConnectionType) != 0 {
	// 	updateAllowedConnectionTypes = updateAllowedConnectionTypes && true
	// }

	// // disable video if connection type is not present in allowed connection types from config
	// if connectionType != nil {
	// 	//check connection type in slot config
	// 	if updateAllowedConnectionTypes {
	// 		adUnitCtx.AllowedConnectionTypes = configObjInVideoConfig.ConnectionType
	// 	}

	// 	if allowedConnectionTypes != nil && !checkValuePresentInArray(allowedConnectionTypes, int(*connectionType)) {
	// 		f := false
	// 		adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Video: &adunitconfig.Video{Enabled: &f}}
	//      rCtx.MetricsEngine.RecordVideoImpDisabledViaConnTypeStats(rCtx.PubIDStr,rCtx.ProfIDStr)
	// 		return
	// 	}
	// }

	return
}
