package openwrap

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
)

func getSignalData(requestBody []byte) string {
	var signal string
	var signalReceived bool
	jsonparser.ArrayEach(requestBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if signalReceived {
			return
		}

		name, err := jsonparser.GetString(value, "name")
		if err != nil {
			return
		}

		if name == "Publisher Passed" {
			signal, err = jsonparser.GetString(value, "segment", "[0]", "signal")
			if err != nil {
				return
			}
			signalReceived = true
		}
	}, "user", "data")

	return signal
}

func addSignalDataInRequest(signal string, maxRequest *openrtb2.BidRequest) {
	if len(signal) == 0 {
		return
	}

	var sdkRequest openrtb2.BidRequest
	if err := json.Unmarshal([]byte(signal), &sdkRequest); err != nil {
		return
	}

	if len(sdkRequest.Imp) > 0 {
		updateImpression(sdkRequest.Imp[0], &maxRequest.Imp[0])
	}
	updateDevice(sdkRequest.Device, maxRequest)
	updateApp(sdkRequest.App, maxRequest)
	updateRegs(sdkRequest.Regs, maxRequest)
	updateSource(sdkRequest.Source, maxRequest)
	updateUser(sdkRequest.User, maxRequest)
}

func updateImpression(sdkImpression openrtb2.Imp, maxImpression *openrtb2.Imp) {
	if maxImpression == nil {
		return
	}

	maxImpression.DisplayManager = sdkImpression.DisplayManager
	maxImpression.DisplayManagerVer = sdkImpression.DisplayManagerVer
	maxImpression.ClickBrowser = sdkImpression.ClickBrowser

	var blockedAttributes []adcom1.CreativeAttribute
	if maxImpression.Video != nil && sdkImpression.Video != nil {
		blockedAttributes = maxImpression.Video.BAttr
		maxImpression.Video = sdkImpression.Video
		maxImpression.Video.BAttr = blockedAttributes
	}

	if maxImpression.Banner != nil {
		if sdkImpression.Banner != nil {
			maxImpression.Banner.API = sdkImpression.Banner.API
		}

		bannertype, err := jsonparser.GetString(maxImpression.Banner.Ext, "bannertype")
		if err == nil && bannertype == "rewarded" {
			maxImpression.Banner = nil
		}
	}

	maxImpression.Ext = setIfKeysExists(sdkImpression.Ext, maxImpression.Ext, "reward", "skadn")
}

func updateDevice(sdkDevice *openrtb2.Device, maxRequest *openrtb2.BidRequest) {
	if sdkDevice == nil {
		return
	}

	if maxRequest.Device == nil {
		maxRequest.Device = &openrtb2.Device{}
	}

	maxRequest.Device.MCCMNC = sdkDevice.MCCMNC
	maxRequest.Device.ConnectionType = sdkDevice.ConnectionType

	maxRequest.Device.Ext = setIfKeysExists(sdkDevice.Ext, maxRequest.Device.Ext, "atts")

	if sdkDevice.Geo == nil {
		return
	}

	if maxRequest.Device.Geo == nil {
		maxRequest.Device.Geo = &openrtb2.Geo{}
	}

	maxRequest.Device.Geo.City = sdkDevice.Geo.City
	maxRequest.Device.Geo.UTCOffset = sdkDevice.Geo.UTCOffset

	// for geo.dma which is non-ortb parameter add it to prebid-openrtb fork
}

func updateApp(sdkApp *openrtb2.App, maxRequest *openrtb2.BidRequest) {
	if sdkApp == nil {
		return
	}

	if maxRequest.App == nil {
		maxRequest.App = &openrtb2.App{}
	}

	maxRequest.App.Paid = sdkApp.Paid
	maxRequest.App.Keywords = sdkApp.Keywords
	maxRequest.App.Domain = sdkApp.Domain
}

func updateRegs(sdkRegs *openrtb2.Regs, maxRequest *openrtb2.BidRequest) {
	if sdkRegs == nil || len(sdkRegs.Ext) == 0 {
		return
	}

	if maxRequest.Regs == nil {
		maxRequest.Regs = &openrtb2.Regs{}
	}

	maxRequest.Regs.Ext = setIfKeysExists(sdkRegs.Ext, maxRequest.Regs.Ext, "gdpr", "gpp", "gpp_sid", "us_privacy")
}

func updateSource(sdkSource *openrtb2.Source, maxRequest *openrtb2.BidRequest) {
	if sdkSource == nil || len(sdkSource.Ext) == 0 {
		return
	}

	if maxRequest.Source == nil {
		maxRequest.Source = &openrtb2.Source{}
	}

	maxRequest.Source.Ext = setIfKeysExists(sdkSource.Ext, maxRequest.Source.Ext, "omidpn", "omidpv")
}

func updateUser(sdkUser *openrtb2.User, maxRequest *openrtb2.BidRequest) {
	if sdkUser == nil {
		return
	}

	if maxRequest.User == nil {
		maxRequest.User = &openrtb2.User{}
	}

	maxRequest.User.Yob = sdkUser.Yob
	maxRequest.User.Gender = sdkUser.Gender
	maxRequest.User.Keywords = sdkUser.Keywords

	//Is this correct? Why doc says to set data.id, data.name separately
	maxRequest.User.Data = append(maxRequest.User.Data, sdkUser.Data...)

	maxRequest.User.Ext = setIfKeysExists(sdkUser.Ext, maxRequest.User.Ext, "consent", "eids")
}

func setIfKeysExists(source []byte, target []byte, keys ...string) []byte {
	oldTarget := target
	for _, key := range keys {
		field, _, _, err := jsonparser.Get(source, key)
		if err != nil {
			continue
		}

		if len(target) == 0 {
			target = []byte(`{}`)
		}

		target, err = jsonparser.Set(target, field, key)
		if err != nil {
			return oldTarget
		}
	}
	return target
}
