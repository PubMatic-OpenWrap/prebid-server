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

	signalData := &openrtb2.BidRequest{}
	if err := json.Unmarshal([]byte(signal), signalData); err != nil {
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
		if signalImp.Banner != nil {
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

func updateAppLovinMaxRequest(requestBody []byte) []byte {
	signalData := getSignalData(requestBody)
	if signalData == nil {
		return requestBody
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

func updateMaxAppLovinResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) models.MaxAppLovin {
	maxAppLovin := models.MaxAppLovin{Reject: false}

	if bidResponse.NBR != nil {
		if !rctx.Debug {
			maxAppLovin.Reject = true
		}
		return maxAppLovin
	}

	if len(bidResponse.SeatBid) == 0 || len(bidResponse.SeatBid[0].Bid) == 0 {
		maxAppLovin.Reject = true
		return maxAppLovin
	}
	return maxAppLovin
}

func applyMaxAppLovinResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if rctx.MaxAppLovin.Reject {
		*bidResponse = openrtb2.BidResponse{}
		return bidResponse
	}

	if bidResponse.NBR != nil {
		return bidResponse
	}

	resp, err := json.Marshal(bidResponse)
	if err != nil {
		*bidResponse = openrtb2.BidResponse{}
		return bidResponse
	}

	signaldata := `{"signaldata":` + strconv.Quote(string(resp)) + `}`
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
