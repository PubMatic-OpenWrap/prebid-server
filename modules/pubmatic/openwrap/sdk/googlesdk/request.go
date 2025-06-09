package googlesdk

import (
	"encoding/base64"
	"errors"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	jsoniter "github.com/json-iterator/go"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/feature"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	sdkparser "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/parser"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/sdkutils"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

const androidAppId = "com.google.ads.mediation.pubmatic.PubMaticMediationAdapter"
const iOSAppId = "GADMediationAdapterPubMatic"

var jsoniterator = jsoniter.ConfigCompatibleWithStandardLibrary

type GoogleSDK struct {
	metricsEngine metrics.MetricsEngine
	config        config.Config
	features      feature.Features
	wrapper       wrapperData
}

type wrapperData struct {
	PublisherId string
	ProfileId   string
	TagId       string
}

func (wd *wrapperData) setProfileID(request *openrtb2.BidRequest) {
	if len(wd.ProfileId) == 0 {
		return
	}

	if request.Ext == nil {
		request.Ext = []byte(`{}`)
	}

	request.Ext, _ = jsonparser.Set(request.Ext, []byte(wd.ProfileId), "prebid", "bidderparams", "pubmatic", "wrapper", "profileid")
}

func (wd *wrapperData) setPublisherId(request *openrtb2.BidRequest) {
	if len(wd.PublisherId) == 0 {
		return
	}

	if request.App == nil {
		request.App = &openrtb2.App{}
	}

	if request.App.Publisher == nil {
		request.App.Publisher = &openrtb2.Publisher{}
	}

	request.App.Publisher.ID = wd.PublisherId
}

func (wd *wrapperData) setTagId(request *openrtb2.BidRequest) {
	if len(wd.TagId) == 0 || len(request.Imp) == 0 {
		return
	}

	request.Imp[0].TagID = wd.TagId
}

// NewGoogleSDK creates a new instance of GoogleSDK
func NewGoogleSDK(metricsEngine metrics.MetricsEngine, cfg config.Config, features feature.Features) *GoogleSDK {
	gsdk := GoogleSDK{
		metricsEngine: metricsEngine,
		features:      features,
		config:        cfg,
	}

	return &gsdk
}

func (gs *GoogleSDK) preProcessRequest(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return nil, errors.New("empty request body")
	}

	var request *openrtb2.BidRequest
	if err := jsoniterator.Unmarshal(body, &request); err != nil {
		return body, err
	}

	// Fetch wrapper data from the request body
	gs.getWrapperData(request)

	// If wrapper data is not found, return error
	// This is a critical check, as wrapper data is essential for request processing
	if len(gs.wrapper.PublisherId) == 0 || len(gs.wrapper.ProfileId) == 0 || len(gs.wrapper.TagId) == 0 {
		return body, errors.New("missing wrapper data: publisherId, profileId or tagId")
	}

	// Modify request with static data
	modifyRequestWithStaticData(request)

	// Set wrapper data in the request
	gs.setWrapperData(request)

	// Google SDK specific modifications
	gs.modifyRequestWithGoogleFeature(request)

	// Marshal the modified request
	modifiedBody, err := jsoniterator.Marshal(request)
	if err != nil {
		return body, err
	}

	return modifiedBody, nil
}

func (gs *GoogleSDK) getWrapperData(request *openrtb2.BidRequest) {
	if request == nil || len(request.Imp) == 0 {
		return
	}

	adunitMappingByte, datatype, _, err := jsonparser.Get(request.Imp[0].Ext, "ad_unit_mapping")
	if adunitMappingByte == nil || err != nil || datatype != jsonparser.Array {
		glog.Errorf("[GoogleSDK] [Error]: failed to get ad unit mapping %v", err)
		return
	}

	var adunitMapping []map[string]interface{}
	if err := jsoniterator.Unmarshal(adunitMappingByte, &adunitMapping); err != nil {
		glog.Errorf("[GoogleSDK] [Error]: failed to unmarshal ad unit mapping %v", err)
		return
	}

	for _, mapping := range adunitMapping {
		keyvals, ok := mapping["keyvals"].([]interface{})
		if !ok {
			continue
		}

		for _, kv := range keyvals {
			kvMap, ok := kv.(map[string]any)
			if !ok {
				continue
			}

			key, ok := kvMap["key"].(string)
			if !ok {
				continue
			}
			value, ok := kvMap["value"].(string)
			if !ok {
				continue
			}

			switch key {
			case "publisher_id":
				gs.wrapper.PublisherId = value
			case "profile_id":
				gs.wrapper.ProfileId = value
			case "ad_unit_id":
				gs.wrapper.TagId = value
			}
		}

		// Check if all values are found
		if len(gs.wrapper.PublisherId) > 0 && len(gs.wrapper.ProfileId) > 0 && len(gs.wrapper.TagId) > 0 {
			break
		}
	}
}

func (gs *GoogleSDK) setWrapperData(request *openrtb2.BidRequest) {
	// Set Publisher Id
	gs.wrapper.setPublisherId(request)

	// Set profile Id at ext.prebid.bidderparams.pubmatic.wrapper.profileid
	gs.wrapper.setProfileID(request)

	// Set Tag Id
	gs.wrapper.setTagId(request)
}

func getSignalData(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return nil, errors.New("empty request body")
	}

	data, dataType, _, err := jsonparser.Get(body, "imp", "[0]", "ext", "buyer_generated_request_data")
	if err != nil || dataType != jsonparser.Array {
		return nil, errors.New("failed to get buyer generated request data: " + err.Error())
	}

	var signalData []byte
	_, err = jsonparser.ArrayEach(data, func(sdkData []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || dataType != jsonparser.Object {
			return
		}

		id, err := jsonparser.GetString(sdkData, "source_app", "id")
		if err != nil || (id != androidAppId && id != iOSAppId) {
			return
		}

		signal, err := jsonparser.GetString(sdkData, "data")
		if err != nil || len(signal) == 0 {
			return
		}

		// decode base64 signal
		signalData, err = base64.StdEncoding.DecodeString(signal)
		if err != nil {
			return
		}
	})
	if err != nil {
		return nil, err
	}

	return signalData, nil
}

func (gs *GoogleSDK) ModifyRequestWithGoogleSDKParams(requestBody []byte) []byte {
	if len(requestBody) == 0 {
		return requestBody
	}

	// Pre-process request
	requestBody, err := gs.preProcessRequest(requestBody)
	if err != nil {
		glog.Errorf("[GoogleSDK] [Error]: failed to pre-process request for publisher %v and profile %v: %v", gs.wrapper.PublisherId, gs.wrapper.ProfileId, err)
		return requestBody
	}

	//Get Signal data and if signal data is not found, process request without signal data
	signalData, err := getSignalData(requestBody)
	if err != nil || len(signalData) == 0 {
		gs.metricsEngine.RecordSignalDataStatus(gs.wrapper.PublisherId, gs.wrapper.ProfileId, models.MissingSignal)
		glog.Errorf("[GoogleSDK] [Error]: failed to get signal data: %v", err)
	}

	requestBody, err = gs.patchSignalDataToRequest(requestBody, signalData)
	if err != nil {
		glog.Errorf("[GoogleSDK] [Error]: failed to patch signal data to request: %v", err)
	}

	return requestBody
}

func (gs *GoogleSDK) patchSignalDataToRequest(requestBody []byte, signalData []byte) ([]byte, error) {
	if gs.config.Template.GoogleSDK.Enable && len(gs.config.Template.GoogleSDK.DeserializedData) > 0 {
		var request, signal map[string]any
		if err := jsoniterator.Unmarshal(requestBody, &request); err != nil {
			glog.Errorf("[GoogleSDK] [Error]: failed to unmarshal request body for publisher %v and profile %v: %v", gs.wrapper.PublisherId, gs.wrapper.ProfileId, err)
			return requestBody, err
		}

		if err := jsoniterator.Unmarshal(signalData, &signal); err != nil {
			glog.Errorf("[GoogleSDK] [Error]: failed to unmarshal signal data for publisher %v and profile %v: %v", gs.wrapper.PublisherId, gs.wrapper.ProfileId, err)
			gs.metricsEngine.RecordSignalDataStatus(gs.wrapper.PublisherId, gs.wrapper.ProfileId, models.InvalidSignal)
			return requestBody, err
		}

		sdkparser.ParseTemplateAndSetValues(gs.config.Template.GoogleSDK.DeserializedData, signal, request)

		// Marshal the modified request back to JSON
		modifiedRequest, err := jsoniterator.Marshal(request)
		if err != nil {
			glog.Errorf("[GoogleSDK] [Error]: failed to marshal modified request: %v", err)
			return requestBody, err
		}
		return modifiedRequest, nil
	}

	sdkRequest := &openrtb2.BidRequest{}
	if err := jsoniterator.Unmarshal(requestBody, sdkRequest); err != nil {
		return requestBody, err
	}

	if len(signalData) > 0 {
		var signal *openrtb2.BidRequest
		if err := jsoniterator.Unmarshal(signalData, &signal); err != nil {
			glog.Errorf("[GoogleSDK] [Error]: failed to unmarshal signal data: %v", err)
			gs.metricsEngine.RecordSignalDataStatus(gs.wrapper.PublisherId, gs.wrapper.ProfileId, models.InvalidSignal)
			return requestBody, err
		}

		// Modify request with signal data
		modifyRequestWithSignalData(sdkRequest, signal)
	}

	// Marshal the modified request back to JSON
	modifiedRequest, err := jsoniterator.Marshal(sdkRequest)
	if err != nil {
		glog.Errorf("[GoogleSDK] [Error]: failed to marshal modified request: %v", err)
		return requestBody, err
	}

	return modifiedRequest, nil
}

func (gs *GoogleSDK) modifyRequestWithGoogleFeature(request *openrtb2.BidRequest) {
	if request == nil || len(request.Imp) == 0 || gs.features == nil {
		return
	}

	for i := range request.Imp {
		bannerSizes := GetFlexSlotSizes(request.Imp[i].Banner, gs.features)
		SetFlexSlotSizes(request.Imp[i].Banner, bannerSizes)
	}
}

func modifyRequestWithStaticData(request *openrtb2.BidRequest) {
	if len(request.Imp) == 0 {
		return
	}

	// Always set secure to 1
	request.Imp[0].Secure = ptrutil.ToPtr(int8(1))

	//Set gpid
	if len(request.Imp[0].TagID) > 0 {
		request.Imp[0].Ext, _ = jsonparser.Set(request.Imp[0].Ext, []byte(strconv.Quote(request.Imp[0].TagID)), "gpid")
	}

	// Remove metric
	request.Imp[0].Metric = nil

	// Remove banner if impression is rewarded and banner and video both are present
	if request.Imp[0].Rwdd == 1 && request.Imp[0].Banner != nil && request.Imp[0].Video != nil {
		request.Imp[0].Banner = nil
	}

	// Remove native from request
	request.Imp[0].Native = nil

}

func modifyRequestWithSignalData(request *openrtb2.BidRequest, signalData *openrtb2.BidRequest) {
	if request == nil || signalData == nil {
		return
	}

	modifyImpression(request, signalData.Imp)
	modifyApp(request, signalData.App)
	modifyDevice(request, signalData.Device)
	modifyRegs(request, signalData.Regs)
	modifySource(request, signalData.Source)
	modifyUser(request, signalData.User)

	// Request Ext
	request.Ext, _ = sdkutils.CopyPath(signalData.Ext, request.Ext, "wrapper", "clientconfig")
}

func modifyUser(request *openrtb2.BidRequest, signalUser *openrtb2.User) {
	if signalUser == nil {
		return
	}

	if request.User == nil {
		request.User = &openrtb2.User{}
	}

	if request.User.Ext == nil {
		request.User.Ext = []byte(`{}`)
	}

	request.User.Ext, _ = sdkutils.CopyPath(signalUser.Ext, request.User.Ext, "sessionduration")
	request.User.Ext, _ = sdkutils.CopyPath(signalUser.Ext, request.User.Ext, "impdepth")
}

func modifySource(request *openrtb2.BidRequest, signalSource *openrtb2.Source) {
	if signalSource == nil {
		return
	}

	if request.Source == nil {
		request.Source = &openrtb2.Source{}
	}

	request.Source.Ext, _ = sdkutils.CopyPath(signalSource.Ext, request.Source.Ext, "omidpn")
	request.Source.Ext, _ = sdkutils.CopyPath(signalSource.Ext, request.Source.Ext, "omidpv")
}

func modifyRegs(request *openrtb2.BidRequest, signalRegs *openrtb2.Regs) {
	if signalRegs == nil {
		return
	}

	if request.Regs == nil {
		request.Regs = &openrtb2.Regs{}
	}

	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "dsa", "dsarequired")
	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "dsa", "pubrender")
	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "dsa", "datatopub")
	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "gpp")
	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "gpp_sid")
}

func modifyDevice(request *openrtb2.BidRequest, signalDevice *openrtb2.Device) {
	if signalDevice == nil {
		return
	}

	if request.Device == nil {
		request.Device = &openrtb2.Device{}
	}

	if len(signalDevice.UA) > 0 {
		request.Device.UA = signalDevice.UA
	}

	if len(signalDevice.Make) > 0 {
		request.Device.Make = signalDevice.Make
	}

	if len(signalDevice.Model) > 0 {
		request.Device.Model = signalDevice.Model
	}

	if signalDevice.JS != nil {
		request.Device.JS = signalDevice.JS
	}

	if signalDevice.IP != "" {
		request.Device.IP = signalDevice.IP
	}

	if signalDevice.Geo != nil {
		request.Device.Geo = signalDevice.Geo
	}

	if signalDevice.HWV != "" {
		request.Device.HWV = signalDevice.HWV
	}
}

func modifyApp(request *openrtb2.BidRequest, signalApp *openrtb2.App) {
	if signalApp == nil {
		return
	}

	if request.App == nil {
		request.App = &openrtb2.App{}
	}

	if len(signalApp.Domain) > 0 {
		request.App.Domain = signalApp.Domain
	}

	if signalApp.Paid != nil {
		request.App.Paid = signalApp.Paid
	}

	if len(signalApp.Keywords) > 0 {
		request.App.Keywords = signalApp.Keywords
	}
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
}

func modifyImpExtension(requestImpExt, signalImpExt []byte) []byte {
	if signalImpExt == nil {
		return requestImpExt
	}

	if len(requestImpExt) == 0 {
		requestImpExt = []byte(`{}`)
	}

	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "skadn", "versions")
	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "skadn", "version")
	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "skadn", "skoverlay")
	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "skadn", "productpage")
	return requestImpExt
}

func modifyImpression(request *openrtb2.BidRequest, signalImps []openrtb2.Imp) {
	if len(request.Imp) == 0 || len(signalImps) == 0 {
		return
	}

	if signalImps[0].DisplayManager != "" {
		request.Imp[0].DisplayManager = signalImps[0].DisplayManager
	}

	if signalImps[0].DisplayManagerVer != "" {
		request.Imp[0].DisplayManagerVer = signalImps[0].DisplayManagerVer
	}

	// Update clickbrowser
	// TODO: This is shallow copy, check if we need deep copy
	if signalImps[0].ClickBrowser != nil {
		request.Imp[0].ClickBrowser = signalImps[0].ClickBrowser
	}

	// Update banner
	modifyBanner(request.Imp[0].Banner, signalImps[0].Banner)

	// Update video (replace entire video object from signal except battr)
	var battrVideo []adcom1.CreativeAttribute
	if request.Imp[0].Video != nil && len(request.Imp[0].Video.BAttr) > 0 {
		battrVideo = make([]adcom1.CreativeAttribute, len(request.Imp[0].Video.BAttr))
		copy(battrVideo, request.Imp[0].Video.BAttr)
	}

	if signalImps[0].Video != nil {
		request.Imp[0].Video = signalImps[0].Video
		if len(battrVideo) > 0 {
			request.Imp[0].Video.BAttr = battrVideo
		}
	}

	// Update native
	request.Imp[0].Native = signalImps[0].Native
	if request.Imp[0].Native != nil {
		request.Imp[0].Native.Request = string(jsonparser.Delete([]byte(request.Imp[0].Native.Request), "privacy"))
	}

	// Update imp extension
	request.Imp[0].Ext = modifyImpExtension(request.Imp[0].Ext, signalImps[0].Ext)
}
