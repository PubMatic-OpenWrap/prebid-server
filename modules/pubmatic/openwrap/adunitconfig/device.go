package adunitconfig

import (
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func ReplaceDeviceTypeFromAdUnitConfig(rCtx models.RequestCtx, device *openrtb2.Device) *openrtb2.Device {
	if device != nil && device.DeviceType != 0 {
		return device
	}

	var adUnitCfg *adunitconfig.AdConfig
	for _, impCtx := range rCtx.ImpBidCtx {
		if impCtx.Type == models.ImpTypeBanner {
			adUnitCfg = impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig
			break
		}
		if impCtx.Type == models.ImpTypeVideo {
			adUnitCfg = impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig
			break
		}
	}

	if device == nil {
		device = new(openrtb2.Device)
	}

	if adUnitCfg == nil || adUnitCfg.Device == nil {
		return device
	}

	device.DeviceType = adcom1.DeviceType(adUnitCfg.Device.DeviceType)

	return device
}
