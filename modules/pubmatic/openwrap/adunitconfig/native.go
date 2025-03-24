package adunitconfig

import (
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func UpdateNativeObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, div string) (adUnitCtx models.AdUnitCtx) {
	defer func() {
		if r := recover(); r != nil {
			rCtx.MetricsEngine.RecordOpenWrapServerPanicStats(rCtx.HostName, "UpdateNativeObjectWithAdunitConfig")
			glog.Error(string(debug.Stack()))
		}
	}()

	nativeAdUnitConfigEnabled := true
	defer func() {
		if imp.Native != nil && !nativeAdUnitConfigEnabled {
			rCtx.MetricsEngine.RecordImpDisabledViaConfigStats(models.ImpTypeNative, rCtx.PubIDStr, rCtx.ProfileIDStr)
		}
	}()

	if rCtx.AdUnitConfig == nil || len(rCtx.AdUnitConfig.Config) == 0 {
		return
	}

	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil {
		adUnitCtx.UsingDefaultConfig = true

		if defaultAdUnitConfig.Native != nil && defaultAdUnitConfig.Native.Enabled != nil && !*defaultAdUnitConfig.Native.Enabled {
			nativeAdUnitConfigEnabled = false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Native: &adunitconfig.Native{Enabled: &nativeAdUnitConfigEnabled}}
			return
		}
	}

	adUnitCtx.SelectedSlotAdUnitConfig, adUnitCtx.MatchedSlot, adUnitCtx.IsRegex, adUnitCtx.MatchedRegex = selectSlot(rCtx, 1, 1, imp.TagID, div, rCtx.Source)
	if adUnitCtx.SelectedSlotAdUnitConfig != nil && adUnitCtx.SelectedSlotAdUnitConfig.Native != nil {
		adUnitCtx.UsingDefaultConfig = false

		if adUnitCtx.SelectedSlotAdUnitConfig.Native.Enabled != nil && !*adUnitCtx.SelectedSlotAdUnitConfig.Native.Enabled {
			nativeAdUnitConfigEnabled = false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Native: &adunitconfig.Native{Enabled: &nativeAdUnitConfigEnabled}}
			return
		}
	}

	adUnitCtx.AppliedSlotAdUnitConfig = getFinalSlotAdUnitConfig(adUnitCtx.SelectedSlotAdUnitConfig, defaultAdUnitConfig)

	return
}
