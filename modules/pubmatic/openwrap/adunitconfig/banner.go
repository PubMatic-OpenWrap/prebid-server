package adunitconfig

import (
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func UpdateBannerObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, div string) (adUnitCtx models.AdUnitCtx) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	bannerAdUnitConfigEnabled := true
	defer func() {
		if imp.Banner != nil && !bannerAdUnitConfigEnabled {
			rCtx.MetricsEngine.RecordImpDisabledViaConfigStats(models.ImpTypeBanner, rCtx.PubIDStr, rCtx.ProfileIDStr)
		}
	}()

	if rCtx.AdUnitConfig == nil || len(rCtx.AdUnitConfig.Config) == 0 {
		return
	}

	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil {
		adUnitCtx.UsingDefaultConfig = true

		if defaultAdUnitConfig.Banner != nil && defaultAdUnitConfig.Banner.Enabled != nil && !*defaultAdUnitConfig.Banner.Enabled {
			bannerAdUnitConfigEnabled = false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Banner: &adunitconfig.Banner{Enabled: &bannerAdUnitConfigEnabled}}
			return
		}
	}

	var height, width int64
	if imp.Banner != nil {
		if imp.Banner.H != nil {
			height = *imp.Banner.H
		}
		if imp.Banner.W != nil {
			width = *imp.Banner.W
		}
	}

	adUnitCtx.SelectedSlotAdUnitConfig, adUnitCtx.MatchedSlot, adUnitCtx.IsRegex, adUnitCtx.MatchedRegex = selectSlot(rCtx, height, width, imp.TagID, div, rCtx.Source)
	if adUnitCtx.SelectedSlotAdUnitConfig != nil && adUnitCtx.SelectedSlotAdUnitConfig.Banner != nil {
		adUnitCtx.UsingDefaultConfig = false

		if adUnitCtx.SelectedSlotAdUnitConfig.Banner.Enabled != nil && !*adUnitCtx.SelectedSlotAdUnitConfig.Banner.Enabled {
			bannerAdUnitConfigEnabled = false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Banner: &adunitconfig.Banner{Enabled: &bannerAdUnitConfigEnabled}}
			return
		}
	}

	adUnitCtx.AppliedSlotAdUnitConfig = getFinalSlotAdUnitConfig(adUnitCtx.SelectedSlotAdUnitConfig, defaultAdUnitConfig)

	return
}
