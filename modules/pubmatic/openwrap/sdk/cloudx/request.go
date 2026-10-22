package cloudx

import (
	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	jsoniter "github.com/json-iterator/go"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/sdk/sdkutils"
	"github.com/prebid/prebid-server/v4/util/ptrutil"
)

var jsoniterator = jsoniter.ConfigCompatibleWithStandardLibrary

type CloudX struct {
	metricsEngine metrics.MetricsEngine
	publisherId   string
	profileId     string
}

func NewCloudX(metricsEngine metrics.MetricsEngine) *CloudX {
	return &CloudX{
		metricsEngine: metricsEngine,
	}
}

// ModifyRequestWithCloudXParams merges CloudX signal into the request. The signal is stringified
// JSON at imp.ext.buyer_generated_request_data. rctx is a pointer so the parsed signal bid request
// can be stored on rCtx.SignalRequest for PubMatic-only EDS at before_validation.
func (l *CloudX) ModifyRequestWithCloudXParams(requestBody []byte, rctx *models.RequestCtx) []byte {
	if len(requestBody) == 0 {
		return nil
	}

	request := &openrtb2.BidRequest{}
	if err := jsoniterator.Unmarshal(requestBody, request); err != nil {
		glog.Errorf("[CloudX] [Error]: failed to unmarshal request body with error %v and request body %s", err, string(requestBody))
		return requestBody
	}

	// modify request with static data
	l.modifyRequestWithStaticData(request)

	// Set publisher id
	if request.App != nil && request.App.Publisher != nil {
		l.publisherId = request.App.Publisher.ID
	}

	// Set profile id
	if profileID, _, _, _ := jsonparser.Get(request.Ext, "prebid", "bidderparams", "pubmatic", "wrapper", "profileid"); profileID != nil {
		l.profileId = string(profileID)
	}

	// modify request with signal data
	l.modifyRequestWithSignalData(request, rctx)

	modifiedRequest, err := jsoniterator.Marshal(request)
	if err != nil {
		return requestBody
	}

	return modifiedRequest
}

func (l *CloudX) modifyRequestWithStaticData(request *openrtb2.BidRequest) {
	if request == nil {
		return
	}

	if len(request.Imp) > 0 {
		// Set banner to nil for rewarded inventorys
		if request.Imp[0].Video != nil && request.Imp[0].Rwdd == 1 {
			request.Imp[0].Banner = nil
		}

		// Set imp.secure as 1
		request.Imp[0].Secure = ptrutil.ToPtr(int8(1))

		// Remove metric
		request.Imp[0].Metric = nil

		// Remove native from request
		request.Imp[0].Native = nil

		// Remove video from request
		request.Imp[0].Video = nil
	}

	if request.App != nil {
		// delete app.ext.sessionDepth
		request.App.Ext = jsonparser.Delete(request.App.Ext, "sessionDepth")
	}
}

func (l *CloudX) modifyRequestWithSignalData(request *openrtb2.BidRequest, rctx *models.RequestCtx) {
	if request == nil || len(request.Imp) == 0 || len(request.Imp[0].Ext) == 0 {
		return
	}

	// Signal is stringified JSON, not base64.
	signalJSON, err := jsonparser.GetString(request.Imp[0].Ext, "buyer_generated_request_data")
	if signalJSON == "" || err != nil {
		l.metricsEngine.RecordSignalDataStatus(l.publisherId, l.profileId, models.MissingSignal)
		return
	}

	var signal *openrtb2.BidRequest
	if err := jsoniterator.Unmarshal([]byte(signalJSON), &signal); err != nil || signal == nil {
		l.metricsEngine.RecordSignalDataStatus(l.publisherId, l.profileId, models.InvalidSignal)
		return
	}

	// Keep parsed signal for EDS; signal ext.eds is not merged onto the shared request ext.
	if rctx != nil {
		rctx.SignalRequest = signal
	}

	modifyImpression(request, signal.Imp)
	modifyRegs(request, signal.Regs)
	modifyApp(request, signal.App)
	modifyDevice(request, signal.Device)
	modifyUser(request, signal.User)
	modifySource(request, signal.Source)

	// Request Ext
	request.Ext, _ = sdkutils.CopyPath(signal.Ext, request.Ext, "wrapper", "clientconfig")
}

func modifyBanner(requestBanner *openrtb2.Banner, signalBanner *openrtb2.Banner) {
	sdkutils.MergeBanner(requestBanner, signalBanner)
	// cloudx sends device size in the request, so we need to overwrite the request banner format with the signal banner format
	if requestBanner != nil && signalBanner != nil && len(signalBanner.Format) > 0 {
		requestBanner.Format = signalBanner.Format
	}
}

func modifyImpression(request *openrtb2.BidRequest, signalImps []openrtb2.Imp) {
	if len(request.Imp) == 0 || len(signalImps) == 0 {
		return
	}

	// read secure from signal
	if signalImps[0].Secure != nil {
		request.Imp[0].Secure = signalImps[0].Secure
	}

	if signalImps[0].Exp > 0 {
		request.Imp[0].Exp = signalImps[0].Exp
	}

	if signalImps[0].DisplayManager != "" {
		request.Imp[0].DisplayManager = signalImps[0].DisplayManager
	}

	if signalImps[0].DisplayManagerVer != "" {
		request.Imp[0].DisplayManagerVer = signalImps[0].DisplayManagerVer
	}

	if signalImps[0].ClickBrowser != nil {
		request.Imp[0].ClickBrowser = signalImps[0].ClickBrowser
	}

	// modify banner
	modifyBanner(request.Imp[0].Banner, signalImps[0].Banner)

	// modify video
	if signalImps[0].Video != nil {
		request.Imp[0].Video = signalImps[0].Video
	}

	// modify native
	request.Imp[0].Native = nil
	if signalImps[0].Native != nil {
		request.Imp[0].Native = signalImps[0].Native
	}

	// modify ext
	request.Imp[0].Ext = modifyImpExtension(request.Imp[0].Ext, signalImps[0].Ext)
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
	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "skadn", "skadnetids")
	requestImpExt, _ = sdkutils.CopyPath(signalImpExt, requestImpExt, "owsdk")

	return requestImpExt
}

func modifyRegs(request *openrtb2.BidRequest, signalRegs *openrtb2.Regs) {
	if signalRegs == nil {
		return
	}

	if request.Regs == nil {
		request.Regs = &openrtb2.Regs{}
	}

	if signalRegs.COPPA > 0 {
		request.Regs.COPPA = signalRegs.COPPA
	}

	// Read privacy fields from signal regs.ext and set them on the request regs object when present.
	if gpp, err := jsonparser.GetString(signalRegs.Ext, "gpp"); err == nil && gpp != "" {
		request.Regs.GPP = gpp
	}
	if _, dataType, _, err := jsonparser.Get(signalRegs.Ext, "gpp_sid"); err == nil && dataType == jsonparser.Array {
		var sids []int8
		jsonparser.ArrayEach(signalRegs.Ext, func(value []byte, _ jsonparser.ValueType, _ int, err error) {
			if err != nil {
				return
			}
			n, convErr := jsonparser.ParseInt(value)
			if convErr != nil {
				return
			}
			sids = append(sids, int8(n))
		}, "gpp_sid")
		if len(sids) > 0 {
			request.Regs.GPPSID = sids
		}
	}
	if gdpr, err := jsonparser.GetInt(signalRegs.Ext, "gdpr"); err == nil {
		request.Regs.GDPR = ptrutil.ToPtr(int8(gdpr))
	}
	if usPrivacy, err := jsonparser.GetString(signalRegs.Ext, "us_privacy"); err == nil && usPrivacy != "" {
		request.Regs.USPrivacy = usPrivacy
	}

	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "dsa", "dsarequired")
	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "dsa", "pubrender")
	request.Regs.Ext, _ = sdkutils.CopyPath(signalRegs.Ext, request.Regs.Ext, "dsa", "datatopub")
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

	if len(signalApp.Cat) > 0 {
		request.App.Cat = signalApp.Cat
	}

	if signalApp.Paid != nil {
		request.App.Paid = signalApp.Paid
	}

	if len(signalApp.Keywords) > 0 {
		request.App.Keywords = signalApp.Keywords
	}

	if signalApp.Name != "" {
		request.App.Name = signalApp.Name
	}

	if signalApp.Ver != "" {
		request.App.Ver = signalApp.Ver
	}

	if len(request.App.StoreURL) == 0 {
		request.App.StoreURL = signalApp.StoreURL
	}

	if signalApp.Content != nil && len(signalApp.Content.URL) > 0 {
		if request.App.Content == nil {
			request.App.Content = &openrtb2.Content{}
		}
		request.App.Content.URL = signalApp.Content.URL
	}
}

func modifyDevice(request *openrtb2.BidRequest, signalDevice *openrtb2.Device) {
	if signalDevice == nil {
		return
	}

	request.Device = sdkutils.MergeDevice(request.Device, signalDevice)

	request.Device.Ext, _ = sdkutils.CopyPath(signalDevice.Ext, request.Device.Ext, "atts")
	request.Device.Ext = sdkutils.CopyIFV(signalDevice.Ext, request.Device.Ext)
}

func modifyUser(request *openrtb2.BidRequest, signalUser *openrtb2.User) {
	if signalUser == nil {
		return
	}

	if request.User == nil {
		request.User = &openrtb2.User{}
	}

	if signalUser.Data != nil {
		request.User.Data = signalUser.Data
	}

	if signalUser.Yob > 0 {
		request.User.Yob = signalUser.Yob
	}

	if signalUser.Gender != "" {
		request.User.Gender = signalUser.Gender
	}

	if signalUser.Keywords != "" {
		request.User.Keywords = signalUser.Keywords
	}

	if consent, err := jsonparser.GetString(signalUser.Ext, "consent"); err == nil && consent != "" {
		request.User.Consent = consent
	}

	request.User.Ext, _ = sdkutils.CopyPath(signalUser.Ext, request.User.Ext, "sessionduration")
	request.User.Ext, _ = sdkutils.CopyPath(signalUser.Ext, request.User.Ext, "eids")
	request.User.Ext, _ = sdkutils.CopyPath(signalUser.Ext, request.User.Ext, "lastadomain")
	request.User.Ext = sdkutils.SetIfKeysExists(signalUser.Ext, request.User.Ext, "impdepth")
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
