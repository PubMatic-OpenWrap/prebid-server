package adunitconfig

import (
	"runtime/debug"
	"strconv"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func UpdateBannerObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, div string, metricEngine metrics.MetricsEngine) (adUnitCtx models.AdUnitCtx) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	if rCtx.AdUnitConfig == nil || len(rCtx.AdUnitConfig.Config) == 0 {
		return
	}

	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil {
		if defaultAdUnitConfig.Banner != nil && defaultAdUnitConfig.Banner.Enabled != nil && !*defaultAdUnitConfig.Banner.Enabled {
			f := false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Banner: &adunitconfig.Banner{Enabled: &f}}
			metricEngine.RecordImpDisabledViaConfigStats(models.ImpTypeBanner, strconv.Itoa(rCtx.PubID), strconv.Itoa(rCtx.ProfileID))
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
		if adUnitCtx.SelectedSlotAdUnitConfig.Banner.Enabled != nil && !*adUnitCtx.SelectedSlotAdUnitConfig.Banner.Enabled {
			f := false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Banner: &adunitconfig.Banner{Enabled: &f}}
			metricEngine.RecordImpDisabledViaConfigStats(models.ImpTypeBanner, strconv.Itoa(rCtx.PubID), strconv.Itoa(rCtx.ProfileID))
			return
		}
	}

	adUnitCtx.AppliedSlotAdUnitConfig = getFinalSlotAdUnitConfig(adUnitCtx.SelectedSlotAdUnitConfig, defaultAdUnitConfig)

	return
}
