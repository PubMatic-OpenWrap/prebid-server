package googlesdk

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/sdkutils"
)

const AppId = "com.google.ads.mediation.pubmatic.PubMaticMediationAdapter"

type wrapperData struct {
	PublisherId string
	ProfileId   string
	TagId       string
}

func getSignalData(body []byte) *openrtb2.BidRequest {
	if len(body) == 0 {
		return nil
	}

	data, dataType, _, err := jsonparser.Get(body, "imp", "[0]", "ext", "buyer_generated_request_data")
	if err != nil || dataType != jsonparser.Array {
		return nil
	}

	var signalData *openrtb2.BidRequest

	// Process each element in buyer_generated_request_data
	_, err = jsonparser.ArrayEach(data, func(sdkData []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || dataType != jsonparser.Object {
			return
		}

		id, err := jsonparser.GetString(sdkData, "source_app", "id")
		if err != nil || id != AppId {
			return
		}

		signal, err := jsonparser.GetString(sdkData, "data")
		if err != nil || len(signal) == 0 {
			return
		}

		signalData = &openrtb2.BidRequest{}
		if err := json.Unmarshal([]byte(signal), signalData); err != nil {
			signalData = nil
		}
	})
	if err != nil {
		return nil
	}

	return signalData
}

func getWrapperData(body []byte) (*wrapperData, error) {
	if len(body) == 0 {
		return nil, errors.New("empty request body")
	}

	keyVal, dataType, _, err := jsonparser.Get(body, "imp", "[0]", "ext", "ad_unit_mapping", "Keyval")
	if err != nil || dataType != jsonparser.Array {
		return nil, errors.New("failed to get Keyval object")
	}

	// Use a switch-based approach for better performance
	wprData := &wrapperData{}
	found := false

	// Process each Keyval object
	if _, err = jsonparser.ArrayEach(keyVal, func(values []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || dataType != jsonparser.Object {
			return
		}

		// Get key and value in a single block to reduce error checking
		key, err := jsonparser.GetString(values, "key")
		if err != nil || key == "" {
			return
		}
		value, err := jsonparser.GetString(values, "value")
		if err != nil || value == "" {
			return
		}

		// Use switch for better performance than if-else chain
		switch key {
		case "publisher_id":
			wprData.PublisherId = value
			found = true
		case "profile_id":
			wprData.ProfileId = value
			found = true
		case "ad_unit_id":
			wprData.TagId = value
			found = true
		}
	}); err != nil {
		return nil, errors.New("failed to process wrapper data")
	}

	if !found {
		return nil, nil
	}

	return wprData, nil
}

func setProfileID(requestBody []byte, wrapperData *wrapperData) []byte {
	if wrapperData == nil || len(wrapperData.ProfileId) == 0 {
		return requestBody
	}

	requestBody, _ = jsonparser.Set(requestBody, []byte(wrapperData.ProfileId), "ext", "prebid", "bidderparams", "pubmatic", "wrapper", "profileid")
	return requestBody
}

func ModifyRequestWithGoogleSDKParams(requestBody []byte) []byte {
	if len(requestBody) == 0 {
		return requestBody
	}

	// Get wrapper data
	wrapperData, err := getWrapperData(requestBody)
	if err != nil {
		return requestBody
	}

	// Set profile Id at ext.prebid.bidderparams.pubmatic.wrapper.profileid
	requestBody = setProfileID(requestBody, wrapperData)

	signalData := getSignalData(requestBody)
	// if signal data is not present, forward request without patching
	if signalData == nil {
		return requestBody
	}

	sdkRequest := &openrtb2.BidRequest{}
	if err := json.Unmarshal(requestBody, sdkRequest); err != nil {
		return requestBody
	}

	modifyRequestWithSignalData(sdkRequest, signalData, wrapperData)

	modifiedRequest, err := json.Marshal(sdkRequest)
	if err != nil {
		return requestBody
	}

	return modifiedRequest
}

func modifyRequestWithSignalData(request *openrtb2.BidRequest, signalData *openrtb2.BidRequest, wrapperData *wrapperData) {
	modifyImpression(request.Imp, signalData.Imp, wrapperData)
	modifyApp(request.App, signalData.App, wrapperData)
	modifyDevice(request.Device, signalData.Device)
	modifyRegs(request.Regs, signalData.Regs)
	modifySource(request.Source, signalData.Source)
}

func modifySource(requestSource *openrtb2.Source, signalSource *openrtb2.Source) {
	if signalSource == nil {
		return
	}

	if requestSource == nil {
		requestSource = &openrtb2.Source{}
	}

	requestSource.Ext, _ = sdkutils.CopyPath(signalSource.Ext, requestSource.Ext, "omidpn")
	requestSource.Ext, _ = sdkutils.CopyPath(signalSource.Ext, requestSource.Ext, "omidpv")
}

func modifyRegs(requestRegs *openrtb2.Regs, signalRegs *openrtb2.Regs) {
	if signalRegs == nil {
		return
	}

	if requestRegs == nil {
		requestRegs = &openrtb2.Regs{}
	}

	requestRegs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, requestRegs.Ext, "dsa", "dsarequired")
	requestRegs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, requestRegs.Ext, "dsa", "pubrender")
	requestRegs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, requestRegs.Ext, "dsa", "datatopub")

}

func modifyDevice(requestDevice *openrtb2.Device, signalDevice *openrtb2.Device) {
	if signalDevice == nil {
		return
	}

	if requestDevice == nil {
		requestDevice = &openrtb2.Device{}
	}

	if len(signalDevice.UA) > 0 {
		requestDevice.UA = signalDevice.UA
	}

	if len(signalDevice.Make) > 0 {
		requestDevice.Make = signalDevice.Make
	}

	if len(signalDevice.Model) > 0 {
		requestDevice.Model = signalDevice.Model
	}

	if signalDevice.JS != nil {
		requestDevice.JS = signalDevice.JS
	}

	if signalDevice.Geo != nil {
		requestDevice.Geo = signalDevice.Geo
	}
}

func modifyApp(requestApp *openrtb2.App, signalApp *openrtb2.App, wrapperData *wrapperData) {
	if signalApp == nil {
		return
	}

	if requestApp == nil {
		requestApp = &openrtb2.App{}
	}

	if len(signalApp.Domain) > 0 {
		requestApp.Domain = signalApp.Domain
	}

	if signalApp.Paid != nil {
		requestApp.Paid = signalApp.Paid
	}

	if len(signalApp.Keywords) > 0 {
		requestApp.Keywords = signalApp.Keywords
	}

	if len(signalApp.StoreURL) > 0 {
		requestApp.StoreURL = signalApp.StoreURL
	}

	if requestApp.Publisher == nil {
		requestApp.Publisher = &openrtb2.Publisher{}
	}

	requestApp.Publisher.ID = wrapperData.PublisherId
}

func modifyBanner(requestBanner *openrtb2.Banner, signalBanner *openrtb2.Banner) {
	if requestBanner == nil || signalBanner == nil {
		return
	}

	if len(signalBanner.MIMEs) > 0 {
		requestBanner.MIMEs = signalBanner.MIMEs
	}

	if len(signalBanner.API) > 0 {
		requestBanner.API = signalBanner.API
	}

	if signalBanner.Vcm != nil {
		requestBanner.Vcm = signalBanner.Vcm
	}
}

func modifyNative(requestNative *openrtb2.Native, signalNative *openrtb2.Native) {
	if requestNative == nil || signalNative == nil {
		return
	}

	if len(signalNative.Ver) > 0 {
		requestNative.Ver = signalNative.Ver
	}

	if len(signalNative.API) > 0 {
		requestNative.API = signalNative.API
	}
}

func modifyImpExtension(requestImpExt, signalImpExt []byte) []byte {
	if signalImpExt == nil {
		return requestImpExt
	}

	if len(requestImpExt) == 0 {
		requestImpExt = []byte(`{}`)
	}

	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "skadn", "versions")
	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "skadn", "skoverlay")
	return requestImpExt
}

func modifyImpression(requestImps []openrtb2.Imp, signalImps []openrtb2.Imp, wrapperData *wrapperData) {
	if len(requestImps) == 0 || len(signalImps) == 0 {
		return
	}

	signalImp := signalImps[0]
	if signalImp.DisplayManager != "" {
		requestImps[0].DisplayManager = signalImp.DisplayManager
	}

	if signalImp.DisplayManagerVer != "" {
		requestImps[0].DisplayManagerVer = signalImp.DisplayManagerVer
	}

	// Update clickbrowser
	// TODO: This is shallow copy, check if we need deep copy
	if signalImp.ClickBrowser != nil {
		requestImps[0].ClickBrowser = signalImp.ClickBrowser
	}

	// Update banner
	modifyBanner(requestImps[0].Banner, signalImp.Banner)

	// Update video (replace entire video object from signal except battr)
	var battrVideo []adcom1.CreativeAttribute
	if requestImps[0].Video != nil && len(requestImps[0].Video.BAttr) > 0 {
		battrVideo = make([]adcom1.CreativeAttribute, len(requestImps[0].Video.BAttr))
		copy(battrVideo, requestImps[0].Video.BAttr)
	}

	if signalImp.Video != nil {
		requestImps[0].Video = signalImp.Video
		if len(battrVideo) > 0 {
			requestImps[0].Video.BAttr = battrVideo
		}
	}

	// Update native
	modifyNative(requestImps[0].Native, signalImp.Native)

	// Update imp extension
	requestImps[0].Ext = modifyImpExtension(requestImps[0].Ext, signalImps[0].Ext)

	//Set gpid
	requestImps[0].Ext, _ = jsonparser.Set(requestImps[0].Ext, []byte(strconv.Quote(requestImps[0].TagID)), "gpid")

	// Update tagId from adunit mapping in request
	requestImps[0].TagID = wrapperData.TagId
}
