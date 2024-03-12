package adunitconfig

import (
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func UpdateFloorsExtObjectFromAdUnitConfig(rCtx models.RequestCtx, requestExt *models.RequestExt) {
	if requestExt.Prebid.Floors != nil {
		return
	}

	var adUnitCfg *adunitconfig.AdConfig
	for _, impCtx := range rCtx.ImpBidCtx {
		if impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig != nil {
			adUnitCfg = impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig
			break
		}
		if impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig != nil {
			adUnitCfg = impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig
			break
		}
	}

	if adUnitCfg == nil || adUnitCfg.Floors == nil {
		return
	}

	requestExt.Prebid.Floors = adUnitCfg.Floors
}
