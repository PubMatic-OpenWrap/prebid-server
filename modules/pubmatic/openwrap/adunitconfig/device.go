package adunitconfig

import (
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
)

func ReplaceDeviceTypeFromAdUnitConfig(rCtx models.RequestCtx, device **openrtb2.Device) {
	if *device == nil {
		*device = &openrtb2.Device{}
	} else if (*device).DeviceType != 0 {
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

	if adUnitCfg == nil || adUnitCfg.Device == nil {
		return
	}

	(*device).DeviceType = adcom1.DeviceType(adUnitCfg.Device.DeviceType)
}
