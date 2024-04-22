package openwrap

import (
	"encoding/json"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func getSignalData(requestBody []byte) *openrtb2.BidRequest {
	signal, err := jsonparser.GetString(requestBody, "user", "data", "[0]", "segment", "[0]", "signal")
	if err != nil {
		return nil
	}

	signalData := &openrtb2.BidRequest{
		Regs: &openrtb2.Regs{
			COPPA: -1,
		},
	}
	if err := json.Unmarshal([]byte(signal), signalData); err != nil {
		return nil
	}
	return signalData
}

func addSignalDataInRequest(signalData *openrtb2.BidRequest, maxRequest *openrtb2.BidRequest, clientconfigflag int) {
	flg := []byte(`0`)
	if clientconfigflag == 1 {
		flg = []byte(`1`)
	}
	if maxReqExt, err := jsonparser.Set(maxRequest.Ext, flg, "prebid", "bidderparams", "pubmatic", "wrapper", "clientconfig"); err == nil {
		maxRequest.Ext = maxReqExt
	}

	if len(signalData.Imp) > 0 {
		updateImpression(signalData.Imp[0], &maxRequest.Imp[0])
	}
	updateDevice(signalData.Device, maxRequest)
	updateApp(signalData.App, maxRequest)
	updateRegs(signalData.Regs, maxRequest)
	updateSource(signalData.Source, maxRequest)
	updateUser(signalData.User, maxRequest)
}

func updateImpression(sdkImpression openrtb2.Imp, maxImpression *openrtb2.Imp) {
	if maxImpression == nil {
		return
	}

	if sdkImpression.DisplayManager != "" {
		maxImpression.DisplayManager = sdkImpression.DisplayManager
	}

	if sdkImpression.DisplayManagerVer != "" {
		maxImpression.DisplayManagerVer = sdkImpression.DisplayManagerVer
	}

	if sdkImpression.ClickBrowser != nil {
		maxImpression.ClickBrowser = sdkImpression.ClickBrowser
	}

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
}

func updateDevice(sdkDevice *openrtb2.Device, maxRequest *openrtb2.BidRequest) {
	if sdkDevice == nil {
		return
	}

	if maxRequest.Device == nil {
		maxRequest.Device = &openrtb2.Device{}
	}

	if sdkDevice.MCCMNC != "" {
		maxRequest.Device.MCCMNC = sdkDevice.MCCMNC
	}

	if sdkDevice.ConnectionType != nil {
		maxRequest.Device.ConnectionType = sdkDevice.ConnectionType
	}

	if sdkDevice.Geo == nil {
		return
	}

	if maxRequest.Device.Geo == nil {
		maxRequest.Device.Geo = &openrtb2.Geo{}
	}

	if sdkDevice.Geo.City != "" {
		maxRequest.Device.Geo.City = sdkDevice.Geo.City
	}

	if sdkDevice.Geo.UTCOffset != 0 {
		maxRequest.Device.Geo.UTCOffset = sdkDevice.Geo.UTCOffset
	}
}

func updateApp(sdkApp *openrtb2.App, maxRequest *openrtb2.BidRequest) {
	if sdkApp == nil {
		return
	}

	if maxRequest.App == nil {
		maxRequest.App = &openrtb2.App{}
	}

	if sdkApp.Paid != nil {
		maxRequest.App.Paid = sdkApp.Paid
	}

	if sdkApp.Keywords != "" {
		maxRequest.App.Keywords = sdkApp.Keywords
	}

	if sdkApp.Domain != "" {
		maxRequest.App.Domain = sdkApp.Domain
	}
}

func updateRegs(sdkRegs *openrtb2.Regs, maxRequest *openrtb2.BidRequest) {
	if sdkRegs == nil {
		return
	}

	if maxRequest.Regs == nil {
		maxRequest.Regs = &openrtb2.Regs{}
	}

	if sdkRegs.COPPA != -1 {
		maxRequest.Regs.COPPA = sdkRegs.COPPA
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

	if sdkUser.Yob != 0 {
		maxRequest.User.Yob = sdkUser.Yob
	}

	if sdkUser.Gender != "" {
		maxRequest.User.Gender = sdkUser.Gender
	}

	if sdkUser.Keywords != "" {
		maxRequest.User.Keywords = sdkUser.Keywords
	}

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

func updateImpressionExt(signalDataImpExt json.RawMessage, impExt *models.ImpExtension) {
	if skdan, _, _, err := jsonparser.Get(signalDataImpExt, "skadn"); err == nil {
		impExt.SKAdnetwork = skdan
	}
	if reward, err := jsonparser.GetInt(signalDataImpExt, "reward"); err == nil {
		impExt.Reward = openrtb2.Int8Ptr(int8(reward))
	}
}
