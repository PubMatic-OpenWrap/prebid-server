package openwrap

import (
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

func populateDeviceContext(dvc *models.DeviceCtx, device *openrtb2.Device) {
	if device == nil {
		return
	}
	//this is needed in determine ifa_type parameter
	dvc.DeviceIFA = strings.TrimSpace(device.IFA)

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
	extIFAType, ifaTypeFound := deviceExt.GetIFAType()
	extSessionID, _ := deviceExt.GetSessionID()

	if ifaTypeFound {
		if dvc.DeviceIFA != "" {
			if ifaTypeID, ok := models.DeviceIFATypeID[strings.ToLower(extIFAType)]; !ok {
				deviceExt.DeleteIFAType()
			} else {
				dvc.IFATypeID = &ifaTypeID
				deviceExt.SetIFAType(extIFAType)
			}
		} else if extSessionID != "" {
			dvc.DeviceIFA = extSessionID
			dvc.IFATypeID = ptrutil.ToPtr(models.DeviceIfaTypeIdSessionId)
			deviceExt.SetIFAType(models.DeviceIFATypeSESSIONID)
		} else {
			deviceExt.DeleteIFAType()
		}
	} else if extSessionID != "" && dvc.DeviceIFA == "" {
		dvc.DeviceIFA = extSessionID
		dvc.IFATypeID = ptrutil.ToPtr(models.DeviceIfaTypeIdSessionId)
		deviceExt.SetIFAType(models.DeviceIFATypeSESSIONID)
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
