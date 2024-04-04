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

		bannertype, _ := jsonparser.GetString(maxImpression.Banner.Ext, "bannertype")
		if bannertype == "rewarded" {
			maxImpression.Banner = nil
		}
	}

	var sdkImpExt map[string]any
	if err := json.Unmarshal(sdkImpression.Ext, &sdkImpExt); err != nil {
		return
	}

	var maxImpExt map[string]any
	json.Unmarshal(maxImpression.Ext, &maxImpExt)

	if reward, ok := sdkImpExt["reward"]; ok {
		maxImpExt["reward"] = reward
	}

	if skadn, ok := sdkImpExt["skadn"]; ok {
		maxImpExt["skadn"] = skadn
	}

	maxImpression.Ext, _ = json.Marshal(maxImpExt)
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

	sdkAtts, _, _, err := jsonparser.Get(sdkDevice.Ext, "atts")
	if err == nil {
		if len(maxRequest.Device.Ext) == 0 {
			maxRequest.Device.Ext = json.RawMessage(`{}`)
		}
		maxRequest.Device.Ext, _ = jsonparser.Set(maxRequest.Device.Ext, sdkAtts, "atts")
	}

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

	var sdkRegsExt map[string]any
	if err := json.Unmarshal(sdkRegs.Ext, &sdkRegsExt); err != nil {
		return
	}

	if maxRequest.Regs == nil {
		maxRequest.Regs = &openrtb2.Regs{}
	}

	var maxRegsExt map[string]any
	if err := json.Unmarshal(maxRequest.Regs.Ext, &maxRegsExt); err != nil {
		maxRegsExt = make(map[string]any)
	}

	if gdpr, ok := sdkRegsExt["gdpr"]; ok {
		maxRegsExt["gdpr"] = gdpr
	}

	if gpp, ok := sdkRegsExt["gpp"]; ok {
		maxRegsExt["gpp"] = gpp
	}

	if gpp_sid, ok := sdkRegsExt["gpp_sid"]; ok {
		maxRegsExt["gpp_sid"] = gpp_sid
	}

	if us_privacy, ok := sdkRegsExt["us_privacy"]; ok {
		maxRegsExt["us_privacy"] = us_privacy
	}

	maxRequest.Regs.Ext, _ = json.Marshal(maxRegsExt)
}

func updateSource(sdkSource *openrtb2.Source, maxRequest *openrtb2.BidRequest) {
	if sdkSource == nil || len(sdkSource.Ext) == 0 {
		return
	}

	var sdkSourceExt map[string]any
	if err := json.Unmarshal(sdkSource.Ext, &sdkSourceExt); err != nil {
		return
	}

	if maxRequest.Source == nil {
		maxRequest.Source = &openrtb2.Source{}
	}

	var maxSourceExt map[string]any
	if err := json.Unmarshal(maxRequest.Source.Ext, &maxSourceExt); err != nil {
		maxSourceExt = make(map[string]any)
	}

	if omidpn, ok := sdkSourceExt["omidpn"]; ok {
		maxSourceExt["omidpn"] = omidpn
	}

	if omidpv, ok := sdkSourceExt["omidpv"]; ok {
		maxSourceExt["omidpv"] = omidpv
	}

	maxRequest.Source.Ext, _ = json.Marshal(maxSourceExt)
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

	var sdkUserExt map[string]any
	if err := json.Unmarshal(sdkUser.Ext, &sdkUserExt); err != nil {
		return
	}

	var maxUserExt map[string]any
	if err := json.Unmarshal(maxRequest.User.Ext, &maxUserExt); err != nil {
		maxUserExt = make(map[string]any)
	}

	if consent, ok := sdkUserExt["consent"]; ok {
		maxUserExt["consent"] = consent
	}

	if eids, ok := sdkUserExt["eids"]; ok {
		maxUserExt["eids"] = eids
	}

	maxRequest.User.Ext, _ = json.Marshal(maxUserExt)
}
