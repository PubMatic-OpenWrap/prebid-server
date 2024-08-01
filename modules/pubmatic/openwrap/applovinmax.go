package openwrap

import (
	"encoding/json"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func getSignalData(requestBody []byte, rctx models.RequestCtx) *openrtb2.BidRequest {
	signal, err := jsonparser.GetString(requestBody, "user", "data", "[0]", "segment", "[0]", "signal")
	if err != nil {
		signalType := models.InvalidSignal
		if err == jsonparser.KeyPathNotFoundError {
			signalType = models.MissingSignal
		}
		rctx.MetricsEngine.RecordSignalDataStatus(getAppPublisherID(requestBody), getProfileID(requestBody), signalType)
		return nil
	}

	signalData := &openrtb2.BidRequest{}
	if err := json.Unmarshal([]byte(signal), signalData); err != nil {
		rctx.MetricsEngine.RecordSignalDataStatus(getAppPublisherID(requestBody), getProfileID(requestBody), models.InvalidSignal)
		return nil
	}
	return signalData
}

func addSignalDataInRequest(signalData *openrtb2.BidRequest, maxRequest *openrtb2.BidRequest) {
	updateRequestWrapper(signalData.Ext, maxRequest)
	updateImpression(signalData.Imp, maxRequest.Imp)
	updateDevice(signalData.Device, maxRequest)
	updateApp(signalData.App, maxRequest)
	updateRegs(signalData.Regs, maxRequest)
	updateSource(signalData.Source, maxRequest)
	updateUser(signalData.User, maxRequest)
}

func updateImpression(signalImps []openrtb2.Imp, maxImps []openrtb2.Imp) {
	if len(maxImps) == 0 || len(signalImps) == 0 {
		return
	}

	signalImp := signalImps[0]
	if signalImp.DisplayManager != "" {
		maxImps[0].DisplayManager = signalImp.DisplayManager
	}

	if signalImp.DisplayManagerVer != "" {
		maxImps[0].DisplayManagerVer = signalImp.DisplayManagerVer
	}

	if signalImp.ClickBrowser != nil {
		maxImps[0].ClickBrowser = signalImp.ClickBrowser
	}

	if signalImp.Video != nil {
		maxImps[0].Video = signalImp.Video
	}

	if maxImps[0].Banner != nil {
		if signalImp.Banner != nil && len(signalImp.Banner.API) > 0 {
			maxImps[0].Banner.API = signalImp.Banner.API
		}

		bannertype, err := jsonparser.GetString(maxImps[0].Banner.Ext, "bannertype")
		if err == nil && bannertype == models.TypeRewarded {
			maxImps[0].Banner = nil
		}
	}

	maxImps[0].Ext = setIfKeysExists(signalImp.Ext, maxImps[0].Ext, "reward", "skadn")
}

func updateDevice(signalDevice *openrtb2.Device, maxRequest *openrtb2.BidRequest) {
	if signalDevice == nil {
		return
	}

	if maxRequest.Device == nil {
		maxRequest.Device = &openrtb2.Device{}
	}

	if signalDevice.MCCMNC != "" {
		maxRequest.Device.MCCMNC = signalDevice.MCCMNC
	}

	if signalDevice.ConnectionType != nil {
		maxRequest.Device.ConnectionType = signalDevice.ConnectionType
	}

	maxRequest.Device.Ext = setIfKeysExists(signalDevice.Ext, maxRequest.Device.Ext, "atts")

	if signalDevice.Geo == nil {
		return
	}

	if maxRequest.Device.Geo == nil {
		maxRequest.Device.Geo = &openrtb2.Geo{}
	}

	if signalDevice.Geo.City != "" {
		maxRequest.Device.Geo.City = signalDevice.Geo.City
	}

	if signalDevice.Geo.UTCOffset != 0 {
		maxRequest.Device.Geo.UTCOffset = signalDevice.Geo.UTCOffset
	}
}

func updateApp(signalApp *openrtb2.App, maxRequest *openrtb2.BidRequest) {
	if signalApp == nil {
		return
	}

	if maxRequest.App == nil {
		maxRequest.App = &openrtb2.App{}
	}

	if signalApp.Paid != nil {
		maxRequest.App.Paid = signalApp.Paid
	}

	if signalApp.Keywords != "" {
		maxRequest.App.Keywords = signalApp.Keywords
	}

	if signalApp.Domain != "" {
		maxRequest.App.Domain = signalApp.Domain
	}
}

func updateRegs(signalRegs *openrtb2.Regs, maxRequest *openrtb2.BidRequest) {
	if signalRegs == nil {
		return
	}

	if maxRequest.Regs == nil {
		maxRequest.Regs = &openrtb2.Regs{}
	}

	if signalRegs.COPPA != 0 {
		maxRequest.Regs.COPPA = signalRegs.COPPA
	}
	maxRequest.Regs.Ext = setIfKeysExists(signalRegs.Ext, maxRequest.Regs.Ext, "gdpr", "gpp", "gpp_sid", "us_privacy")
}

func updateSource(signalSource *openrtb2.Source, maxRequest *openrtb2.BidRequest) {
	if signalSource == nil || len(signalSource.Ext) == 0 {
		return
	}

	if maxRequest.Source == nil {
		maxRequest.Source = &openrtb2.Source{}
	}

	maxRequest.Source.Ext = setIfKeysExists(signalSource.Ext, maxRequest.Source.Ext, "omidpn", "omidpv")
}

func updateUser(signalUser *openrtb2.User, maxRequest *openrtb2.BidRequest) {
	if signalUser == nil {
		return
	}

	if maxRequest.User == nil {
		maxRequest.User = &openrtb2.User{}
	}

	if signalUser.Yob != 0 {
		maxRequest.User.Yob = signalUser.Yob
	}

	if signalUser.Gender != "" {
		maxRequest.User.Gender = signalUser.Gender
	}

	if signalUser.Keywords != "" {
		maxRequest.User.Keywords = signalUser.Keywords
	}

	maxRequest.User.Data = signalUser.Data
	maxRequest.User.Ext = setIfKeysExists(signalUser.Ext, maxRequest.User.Ext, "consent", "eids")
}

func setIfKeysExists(source []byte, target []byte, keys ...string) []byte {
	newTarget := target
	if len(keys) > 0 && len(newTarget) == 0 {
		newTarget = []byte(`{}`)
	}

	for _, key := range keys {
		field, dataType, _, err := jsonparser.Get(source, key)
		if err != nil {
			continue
		}

		if dataType == jsonparser.String {
			quotedStr := strconv.Quote(string(field))
			field = []byte(quotedStr)
		}

		newTarget, err = jsonparser.Set(newTarget, field, key)
		if err != nil {
			return target
		}
	}

	if len(newTarget) == 2 {
		return target
	}
	return newTarget
}

func updateRequestWrapper(signalExt json.RawMessage, maxRequest *openrtb2.BidRequest) {
	clientConfigFlag, err := jsonparser.GetInt(signalExt, "wrapper", "clientconfig")
	if err != nil || clientConfigFlag != 1 {
		return
	}

	if maxReqExt, err := jsonparser.Set(maxRequest.Ext, []byte(`1`), "prebid", "bidderparams", "pubmatic", "wrapper", "clientconfig"); err == nil {
		maxRequest.Ext = maxReqExt
	}
}

func updateAppLovinMaxRequest(requestBody []byte, rctx models.RequestCtx) []byte {
	signalData := getSignalData(requestBody, rctx)
	if signalData == nil {
		return modifyRequestBody(requestBody)
	}

	maxRequest := &openrtb2.BidRequest{}
	if err := json.Unmarshal(requestBody, maxRequest); err != nil {
		return requestBody
	}

	addSignalDataInRequest(signalData, maxRequest)
	if maxRequestbytes, err := json.Marshal(maxRequest); err == nil {
		return maxRequestbytes
	}
	return requestBody
}

func updateAppLovinMaxResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) models.AppLovinMax {
	rctx.AppLovinMax.Reject = false

	if bidResponse.NBR != nil {
		if !rctx.Debug {
			rctx.AppLovinMax.Reject = true
		}
	} else if len(bidResponse.SeatBid) == 0 || len(bidResponse.SeatBid[0].Bid) == 0 {
		rctx.AppLovinMax.Reject = true
	}
	return rctx.AppLovinMax
}

func applyAppLovinMaxResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if rctx.AppLovinMax.Reject {
		return bidResponse
	}

	//This condition is applied only in case if debug=1 refer func updateMaxAppLovinResponse
	if bidResponse.NBR != nil {
		return bidResponse
	}

	resp, err := json.Marshal(bidResponse)
	if err != nil {
		return bidResponse
	}

	signaldata := `{"` + models.SignalData + `":` + strconv.Quote(string(resp)) + `}`
	*bidResponse = openrtb2.BidResponse{
		ID:    bidResponse.ID,
		BidID: bidResponse.SeatBid[0].Bid[0].ID,
		Cur:   bidResponse.Cur,
		SeatBid: []openrtb2.SeatBid{
			{
				Bid: []openrtb2.Bid{
					{
						ID:    bidResponse.SeatBid[0].Bid[0].ID,
						ImpID: bidResponse.SeatBid[0].Bid[0].ImpID,
						Price: bidResponse.SeatBid[0].Bid[0].Price,
						BURL:  bidResponse.SeatBid[0].Bid[0].BURL,
						Ext:   json.RawMessage(signaldata),
					},
				},
			},
		},
	}
	return bidResponse
}

func getAppPublisherID(requestBody []byte) string {
	if pubId, err := jsonparser.GetString(requestBody, "app", "publisher", "id"); err == nil && len(pubId) > 0 {
		return pubId
	}
	return ""
}

func getProfileID(requestBody []byte) string {
	if profileId, err := jsonparser.GetInt(requestBody, "ext", "prebid", "bidderparams", "pubmatic", "wrapper", "profileid"); err == nil {
		a := strconv.Itoa(int(profileId))
		return a
	}
	return ""
}

// modifyRequestBody modifies displaymanger and banner object in req if signal is missing/invalid
func modifyRequestBody(requestBody []byte) []byte {
	if body, err := jsonparser.Set(requestBody, []byte(strconv.Quote("PubMatic_OpenWrap_SDK")), "imp", "[0]", "displaymanager"); err == nil {
		requestBody = jsonparser.Delete(body, "imp", "[0]", "displaymanagerver")
	}

	if bannertype, err := jsonparser.GetString(requestBody, "imp", "[0]", "banner", "ext", "bannertype"); err == nil && bannertype == models.TypeRewarded {
		requestBody = jsonparser.Delete(requestBody, "imp", "[0]", "banner")
	}
	return requestBody
}

// getApplovinABTestFloors fetches adunitwise floors for pub-profile
func (m OpenWrap) getApplovinABTestFloors(rctx models.RequestCtx) models.ABTestConfig {
	if rctx.Endpoint == models.EndpointAppLovinMax && m.pubFeatures.IsApplovinMultiFloorsEnabled(rctx.PubID, rctx.ProfileIDStr) {
		return models.ABTestConfig{
			Enabled: true,
			Config:  m.pubFeatures.GetApplovinABTestFloors(rctx.PubID, rctx.ProfileIDStr),
		}
	}
	return models.ABTestConfig{Enabled: false}
}
