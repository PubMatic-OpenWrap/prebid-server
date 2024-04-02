package openwrap

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
)

func addSignalDataInRequest(requestBody []byte) {
	var signal string
	jsonparser.ArrayEach(requestBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		name, err := jsonparser.GetString(value, "name")
		if err != nil {
			return
		}

		if name == "Publisher Passed" {
			signal, err = jsonparser.GetString(value, "segment", "[0]", "signal")
			if err != nil {
				return
			}
		}
	}, "user", "data")

	if len(signal) == 0 {
		return
	}

	var maxRequest, sdkRequest openrtb2.BidRequest
	if err := json.Unmarshal([]byte(signal), &sdkRequest); err != nil {
		return
	}
	if err := json.Unmarshal(requestBody, &maxRequest); err != nil {
		return
	}
	updateRegs(sdkRequest.Regs, maxRequest.Regs)
	updateSource(sdkRequest.Source, maxRequest.Source)
	updateUser(sdkRequest.User, maxRequest.User)
	updateApp(sdkRequest.App, maxRequest.App)
	updateDevice(sdkRequest.Device, maxRequest.Device)
	updateImpression(sdkRequest.Imp, maxRequest.Imp)
}

func updateImpression(sdkImpression []openrtb2.Imp, maxImpression []openrtb2.Imp) {
	if sdkImpression == nil || maxImpression == nil {
		return
	}

	sdkImp := sdkImpression[0]
	maxImp := maxImpression[0]

	maxImp.DisplayManager = sdkImp.DisplayManager
	maxImp.DisplayManagerVer = sdkImp.DisplayManagerVer
	maxImp.ClickBrowser = sdkImp.ClickBrowser

	var blockedAttributes []adcom1.CreativeAttribute
	if maxImp.Video != nil {
		blockedAttributes = maxImp.Video.BAttr
	}

	maxImp.Video = sdkImp.Video
	if maxImp.Video != nil {
		maxImp.Video.BAttr = blockedAttributes
	}

	if maxImp.Banner != nil {
		if sdkImp.Banner != nil {
			maxImp.Banner.API = sdkImp.Banner.API
		}
		bannertype, _ := jsonparser.GetString(maxImp.Banner.Ext, "bannertype")
		if bannertype == "rewarded" {
			maxImp.Banner = nil
		}
	}

	var sdkImpExt map[string]interface{}
	if err := json.Unmarshal(sdkImp.Ext, &sdkImpExt); err != nil {
		return
	}

	var maxImpExt map[string]interface{}
	if err := json.Unmarshal(maxImp.Ext, &maxImpExt); err != nil {
		return
	}

	if reward, ok := sdkImpExt["reward"]; ok {
		maxImpExt["reward"] = reward
	}

	skadn, ok := sdkImpExt["skadn"]
	if !ok {
		return
	}

	if _, ok := maxImpExt["skadn"]; !ok {
		maxImpExt["skadn"] = map[string]interface{}{}
	}

	maxImpExt["skadn"] = skadn
	maxImp.Ext, _ = json.Marshal(maxImpExt)
}

func updateDevice(sdkDevice *openrtb2.Device, maxDevice *openrtb2.Device) {
	if sdkDevice == nil {
		return
	}

	if maxDevice == nil {
		maxDevice = &openrtb2.Device{}
	}

	maxDevice.MCCMNC = sdkDevice.MCCMNC
	maxDevice.ConnectionType = sdkDevice.ConnectionType

	if sdkDevice.Geo == nil {
		return
	}

	if maxDevice.Geo == nil {
		maxDevice.Geo = &openrtb2.Geo{}
	}

	maxDevice.Geo.City = sdkDevice.Geo.City
	maxDevice.Geo.UTCOffset = sdkDevice.Geo.UTCOffset

	// for geo.dma which is non-ortb parameter add it to prebid-openrtb fork

	sdkAtts, _, _, err := jsonparser.Get(sdkDevice.Ext, "atts")
	if err != nil {
		return
	}
	jsonparser.Set(maxDevice.Ext, sdkAtts, "atts")
}

func updateApp(sdkApp *openrtb2.App, maxApp *openrtb2.App) {
	if sdkApp == nil {
		return
	}

	if maxApp == nil {
		maxApp = &openrtb2.App{}
	}

	maxApp.Paid = sdkApp.Paid
	maxApp.Keywords = sdkApp.Keywords
	maxApp.Domain = sdkApp.Domain
}

func updateRegs(sdkRegs *openrtb2.Regs, maxRegs *openrtb2.Regs) {
	if sdkRegs == nil || len(sdkRegs.Ext) == 0 {
		return
	}

	var sdkRegsExt map[string]interface{}
	if err := json.Unmarshal(sdkRegs.Ext, &sdkRegsExt); err != nil {
		return
	}

	if maxRegs == nil {
		maxRegs = &openrtb2.Regs{}
	}

	var maxRegsExt map[string]interface{}
	if err := json.Unmarshal(maxRegs.Ext, &maxRegsExt); err != nil {
		return
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
		maxRegsExt["gdpr"] = us_privacy
	}

	maxRegs.Ext, _ = json.Marshal(maxRegsExt)
}

func updateSource(sdkSource *openrtb2.Source, maxSource *openrtb2.Source) {
	if sdkSource == nil || len(sdkSource.Ext) == 0 {
		return
	}

	var sdkSourceExt map[string]interface{}
	if err := json.Unmarshal(sdkSource.Ext, &sdkSourceExt); err != nil {
		return
	}

	if maxSource == nil {
		maxSource = &openrtb2.Source{}
	}

	var maxSourceExt map[string]interface{}
	if err := json.Unmarshal(maxSource.Ext, &maxSourceExt); err != nil {
		return
	}

	if omidpn, ok := sdkSourceExt["omidpn"]; ok {
		maxSourceExt["omidpn"] = omidpn
	}

	if omidpv, ok := sdkSourceExt["omidpv"]; ok {
		maxSourceExt["omidpv"] = omidpv
	}

	maxSource.Ext, _ = json.Marshal(maxSourceExt)
}

func updateUser(sdkUser *openrtb2.User, maxUser *openrtb2.User) {
	if sdkUser == nil {
		return
	}

	if maxUser == nil {
		maxUser = &openrtb2.User{}
	}

	maxUser.Yob = sdkUser.Yob
	maxUser.Gender = sdkUser.Gender
	maxUser.Keywords = sdkUser.Keywords

	//Is this correct? Why doc says to set data.id, data.name separately
	maxUser.Data = append(maxUser.Data, sdkUser.Data...)

	var sdkUserExt map[string]interface{}
	if err := json.Unmarshal(sdkUser.Ext, &sdkUserExt); err != nil {
		return
	}

	var maxUserExt map[string]interface{}
	if err := json.Unmarshal(maxUser.Ext, &maxUserExt); err != nil {
		return
	}

	if consent, ok := sdkUserExt["consent"]; ok {
		maxUserExt["consent"] = consent
	}

	if eids, ok := sdkUserExt["eids"]; ok {
		maxUserExt["eids"] = eids
	}

	maxUser.Ext, _ = json.Marshal(maxUserExt)
}
