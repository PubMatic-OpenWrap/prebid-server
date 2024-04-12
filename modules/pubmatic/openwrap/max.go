package openwrap

import (
	"encoding/json"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
)

func getSignalData(requestBody []byte) string {
	var signal string
	signal, err := jsonparser.GetString(requestBody, "user", "data", "[0]", "segment", "[0]", "signal")
	if err != nil {
		signal = ""
	}
	return signal
}

func addSignalDataInRequest(signal string, maxRequest *openrtb2.BidRequest, clientconfigflag int) {
	if len(signal) == 0 {
		return
	}

	var sdkRequest openrtb2.BidRequest
	if err := json.Unmarshal([]byte(signal), &sdkRequest); err != nil {
		return
	}

	flg := []byte(`0`)
	if clientconfigflag == 1 {
		flg = []byte(`1`)
	}
	if maxReqExt, err := jsonparser.Set(maxRequest.Ext, flg, "wrapper", "clientconfig"); err == nil {
		maxRequest.Ext = maxReqExt
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

	if sdkImpression.Video != nil {
		maxImpression.Video = sdkImpression.Video
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
	if sdkRegs == nil {
		return
	}

	if maxRequest.Regs == nil {
		maxRequest.Regs = &openrtb2.Regs{}
	}

	maxRequest.Regs.COPPA = sdkRegs.COPPA
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
	maxRequest.User.Data = sdkUser.Data

	maxRequest.User.Ext = setIfKeysExists(sdkUser.Ext, maxRequest.User.Ext, "consent", "eids")
}

func setIfKeysExists(source []byte, target []byte, keys ...string) []byte {
	oldTarget := target
	for _, key := range keys {
		field, dataType, _, err := jsonparser.Get(source, key)
		if err != nil {
			continue
		}

		if len(target) == 0 {
			target = []byte(`{}`)
		}

		if dataType == jsonparser.String {
			quotedStr := strconv.Quote(string(field))
			field = []byte(quotedStr)
		}

		target, err = jsonparser.Set(target, field, key)
		if err != nil {
			return oldTarget
		}
	}
	return target
}
