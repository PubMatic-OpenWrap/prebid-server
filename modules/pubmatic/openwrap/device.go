package openwrap

import (
	"encoding/json"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func validateDevice(device *openrtb2.Device) {
	//unmarshal device ext
	var deviceExt models.ExtDevice
	err := json.Unmarshal(device.Ext, &deviceExt)
	if err != nil {
		return
	}

	if deviceExt.ExtDevice != nil {
		deviceExt.IFAType = strings.TrimSpace(deviceExt.IFAType)
		deviceExt.SessionID = strings.TrimSpace(deviceExt.SessionID)

		//refactor below condition
		if deviceExt.IFAType != "" {
			if device.IFA != "" {
				if _, ok := models.DeviceIFATypeID[strings.ToLower(deviceExt.IFAType)]; !ok {
					deviceExt.IFAType = ""
				}
			} else if deviceExt.SessionID != "" {
				device.IFA = deviceExt.SessionID
				deviceExt.IFAType = models.DeviceIFATypeSESSIONID
			} else {
				deviceExt.IFAType = ""
			}
		} else if deviceExt.SessionID != "" {
			device.IFA = deviceExt.SessionID
			deviceExt.IFAType = models.DeviceIFATypeSESSIONID
		}
	} else if deviceExt.SessionID != "" {
		deviceExt.ExtDevice = &openrtb_ext.ExtDevice{
			IFAType: models.DeviceIFATypeSESSIONID,
		}
		device.IFA = deviceExt.SessionID
	}

	device.Ext, _ = json.Marshal(deviceExt)
}
