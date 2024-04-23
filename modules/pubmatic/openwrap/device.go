package openwrap

import (
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func populateDeviceContext(dvc *models.DeviceCtx, device *openrtb2.Device, signalData *openrtb2.BidRequest) {
	if device == nil {
		return
	}
	//this is needed in determine ifa_type parameter
	dvc.DeviceIFA = device.IFA

	if signalData != nil && signalData.Device != nil {
		device.Ext = setIfKeysExists(signalData.Device.Ext, device.Ext, "atts")
	}

	if device.Ext == nil {
		return
	}

	//unmarshal device ext
	var deviceExt models.ExtDevice
	if err := deviceExt.UnmarshalJSON(device.Ext); err != nil {
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
	extIFATypeStr, _ := deviceExt.GetIFAType()
	extSessionIDStr, _ := deviceExt.GetSessionID()

	if extIFATypeStr == "" {
		if extSessionIDStr == "" {
			deviceExt.DeleteIFAType()
			deviceExt.DeleteSessionID()
			return
		}
		dvc.DeviceIFA = extSessionIDStr
		extIFATypeStr = models.DeviceIFATypeSESSIONID
	}
	if dvc.DeviceIFA != "" {
		if _, ok := models.DeviceIFATypeID[strings.ToLower(extIFATypeStr)]; !ok {
			extIFATypeStr = ""
		}
	} else if extSessionIDStr != "" {
		dvc.DeviceIFA = extSessionIDStr
		extIFATypeStr = models.DeviceIFATypeSESSIONID

	} else {
		extIFATypeStr = ""
	}

	if ifaTypeID, ok := models.DeviceIFATypeID[strings.ToLower(extIFATypeStr)]; ok {
		dvc.IFATypeID = &ifaTypeID
	}

	if extIFATypeStr == "" {
		deviceExt.DeleteIFAType()
	} else {
		deviceExt.SetIFAType(extIFATypeStr)
	}

	if extSessionIDStr == "" {
		deviceExt.DeleteSessionID()
	} else {
		deviceExt.SetSessionID(extSessionIDStr)
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
		device.Ext, _ = dvc.Ext.MarshalJSON()
	}
}
