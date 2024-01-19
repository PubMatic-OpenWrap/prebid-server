package openwrap

import (
	"encoding/json"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func populateDeviceContext(dvc *models.DeviceCtx, device *openrtb2.Device) {
	if device == nil {
		return
	}
	//this is needed in determine ifa_type parameter
	dvc.DeviceIFA = device.IFA

	if device.Ext == nil {
		return
	}

	//unmarshal device ext
	var deviceExt models.ExtDevice
	if err := json.Unmarshal(device.Ext, &deviceExt); err != nil {
		return
	}
	dvc.Ext = &deviceExt

	//update device IFA Details
	updateDeviceIFADetails(dvc)
}

func updateDeviceIFADetails(dvc *models.DeviceCtx) {
	if dvc == nil || dvc.Ext == nil {
		return
	}

	deviceExt := dvc.Ext
	deviceExt.IFAType = strings.TrimSpace(deviceExt.IFAType)
	deviceExt.SessionID = strings.TrimSpace(deviceExt.SessionID)

	//refactor below condition
	if deviceExt.IFAType != "" {
		if dvc.DeviceIFA != "" {
			if _, ok := models.DeviceIFATypeID[strings.ToLower(deviceExt.IFAType)]; !ok {
				deviceExt.IFAType = ""
			}
		} else if deviceExt.SessionID != "" {
			dvc.DeviceIFA = deviceExt.SessionID
			deviceExt.IFAType = models.DeviceIFATypeSESSIONID
		} else {
			deviceExt.IFAType = ""
		}
	} else if deviceExt.SessionID != "" {
		dvc.DeviceIFA = deviceExt.SessionID
		deviceExt.IFAType = models.DeviceIFATypeSESSIONID
	}

	if ifaTypeID, ok := models.DeviceIFATypeID[strings.ToLower(deviceExt.IFAType)]; ok {
		dvc.IFATypeID = &ifaTypeID
	}
}

func amendDeviceObject(device *openrtb2.Device, dvc *models.DeviceCtx) {
	if device == nil || dvc == nil {
		return
	}

	//update device IFA
	if len(dvc.DeviceIFA) > 0 {
		device.IFA = dvc.DeviceIFA
	}

	//update device extension
	if dvc.Ext != nil {
		device.Ext, _ = json.Marshal(dvc.Ext)
	}
}
