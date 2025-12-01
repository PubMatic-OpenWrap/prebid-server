package openwrap

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/hooks/hookexecution"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	v25 "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/openrtb/v25"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/googlesdk"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/sdkutils"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/unitylevelplay"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/wakanda"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/usersync"
	uuid "github.com/satori/go.uuid"
)

const (
	OpenWrapAuction      = "/pbs/openrtb2/auction"
	OpenWrapV25          = "/openrtb/2.5"
	OpenWrapV25Video     = "/openrtb/2.5/video"
	OpenWrapOpenRTBVideo = "/video/openrtb"
	OpenWrapVAST         = "/video/vast"
	OpenWrapJSON         = "/video/json"
	OpenWrapAmp          = "/amp"
)

func (m OpenWrap) handleEntrypointHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (result hookstage.HookResult[hookstage.EntrypointPayload], err error) {
	var (
		rCtx                models.RequestCtx
		endpointHookManager endpointmanager.EndpointHookManager
		pubid               int
		body                []byte
		ok                  bool
	)

	defer func() {
		if result.Reject {
			if rCtx.PubIDStr == "" {
				rCtx.PubIDStr = "0"
			}
			m.metricEngine.RecordBadRequests(rCtx.Endpoint, rCtx.PubIDStr, getPubmaticErrorCode(openrtb3.NoBidReason(result.NbrCode)))
			if glog.V(models.LogLevelDebug) {
				glog.Infof("[bad_request] pubid:[%d] profid:[%d] endpoint:[%s] nbr:[%d] query_params:[%s] body:[%s]",
					rCtx.PubID, rCtx.ProfileID, rCtx.Endpoint, result.NbrCode, payload.Request.URL.Query().Encode(), string(body))
			}
			return
		}

		if result.ModuleContext == nil {
			result.ModuleContext = hookstage.NewModuleContext()
		}
		result.ModuleContext.Set("rctx", rCtx)
		result.ModuleContext.Set("endpointhookmanager", endpointHookManager)
	}()

	// Intialise configs
	rCtx.StartTime = time.Now().Unix()
	rCtx.TMax = m.cfg.Timeout.MaxTimeout

	// process request
	processHTTPRequest(&rCtx, payload.Request)

	//Intialise endpoint Hook Manager based on endpoint
	endpointHookManager = endpointmanager.NewEndpointManager(rCtx.Endpoint, m.metricEngine, m.cache, m.creativeCache)

	if rCtx.Endpoint == models.EndpointHybrid {
		rCtx.Endpoint = models.EndpointHybrid
		return result, nil
	}

	//Do not execute the module for requests processed in SSHB(8001)
	if rCtx.Sshb == models.Enabled {
		rCtx.VastUnWrap.Enabled = getVastUnwrapperEnable(payload.Request.Context(), models.VastUnwrapperEnableKey)
		return result, nil
	}

	// Initialise static values
	rCtx.MetricsEngine = m.metricEngine
	rCtx.DCName = m.cfg.Server.DCName
	rCtx.TrackerEndpoint = m.cfg.Tracker.Endpoint
	rCtx.VideoErrorTrackerEndpoint = m.cfg.Tracker.VideoErrorTrackerEndpoint
	rCtx.WakandaDebug = &wakanda.Debug{
		Config: m.cfg.Wakanda,
	}

	// Preserve original request body for wakanda
	originalRequestBody := payload.Body

	// Execute endpoint specific entrypoint hook
	body, result, ok = endpointHookManager.HandleEntrypointHook(&rCtx, payload, miCtx, result)
	if !ok {
		result.Reject = true
		return result, nil
	}

	if rCtx.Endpoint == models.EndpointAppLovinMax {
		// updating body locally to access updated fields from signal
		body = updateAppLovinMaxRequest(body, rCtx)
		result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
			ep.Body = body
			return ep, nil
		}, hookstage.MutationUpdate, "update-max-app-lovin-request")
	}

	if rCtx.Endpoint == models.EndpointGoogleSDK {
		// Update fields from signal
		body = googlesdk.ModifyRequestWithGoogleSDKParams(body, rCtx, m.features)
		result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
			ep.Body = body
			return ep, nil
		}, hookstage.MutationUpdate, "update-google-sdk-request")
	}

	if rCtx.Endpoint == models.EndpointUnityLevelPlay {
		// Update fields from signal
		ulp := unitylevelplay.NewLevelPlay(m.metricEngine)
		body = ulp.ModifyRequestWithUnityLevelPlayParams(body)
		result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
			ep.Body = body
			return ep, nil
		}, hookstage.MutationUpdate, "update-unity-level-play-request")
	}

	// init default for all modules
	result.Reject = true

	// process wrapper extension
	result, ok = processWrapperExtension(&rCtx, payload.Request, body, result)
	if !ok {
		return result, nil
	}

	// Initialize request context
	initializeRequestContext(&rCtx)

	// Debug
	requestDebug, _ := jsonparser.GetBoolean(body, "ext", "prebid", "debug")
	if !rCtx.Debug {
		rCtx.Debug = requestDebug
	}

	// Features
	rCtx.SendBurl = getSendBurl(body, rCtx.Endpoint)
	rCtx.GoogleSDK = models.GoogleSDK{StartTime: time.Now()}

	pubIdStr, _, _, errs := getAccountIdFromRawRequest(false, nil, body)
	if len(errs) > 0 {
		result.NbrCode = int(nbr.InvalidPublisherID)
		result.Errors = append(result.Errors, errs[0].Error())
		return result, errs[0]
	}

	rCtx.PubID, err = strconv.Atoi(pubIdStr)
	if err != nil {
		result.NbrCode = int(nbr.InvalidPublisherID)
		result.Errors = append(result.Errors, "ErrInvalidPublisherID")
		return result, fmt.Errorf("invalid publisher id : %v", err)
	}
	rCtx.PubIDStr = pubIdStr

	// only http.ErrNoCookie is returned, we can ignore it
	rCtx.UidCookie, _ = payload.Request.Cookie(models.UidCookieName)
	if rCtx.UidCookie == nil {
		m.metricEngine.RecordUidsCookieNotPresentErrorStats(rCtx.PubIDStr, rCtx.ProfileIDStr)
	}

	rCtx.KADUSERCookie, _ = payload.Request.Cookie(models.KADUSERCOOKIE)
	if originCookie, _ := payload.Request.Cookie("origin"); originCookie != nil {
		rCtx.OriginCookie = originCookie.Value
	}

	if rCtx.LoggerImpressionID == "" {
		rCtx.LoggerImpressionID = uuid.NewV4().String()
	}

	// temp, for AMP, etc
	if pubid != 0 {
		rCtx.PubID = pubid
	}

	rCtx.WakandaDebug.EnableIfRequired(pubIdStr, rCtx.ProfileIDStr)
	if rCtx.WakandaDebug.IsEnable() {
		rCtx.WakandaDebug.SetHTTPRequestData(payload.Request, originalRequestBody)
	}

	result.Reject = false

	return result, nil
}

func getRequestWrapper(r *http.Request, body []byte, endpoint string, result hookstage.HookResult[hookstage.EntrypointPayload]) (models.RequestExtWrapper, error) {
	var requestExtWrapper models.RequestExtWrapper
	var err error
	switch endpoint {
	case models.EndpintInappVideo:
		requestExtWrapper, err = v25.ConvertVideoToAuctionRequest(r, body, &result)
	case models.EndpointAMP:
		requestExtWrapper, err = models.GetQueryParamRequestExtWrapper(r)
	case models.EndpointV25:
		fallthrough
	case models.EndpointVideo, models.EndpointORTB, models.EndpointVAST, models.EndpointJson:
		requestExtWrapper, err = models.GetRequestExtWrapper(body, "ext", "wrapper")
	case models.EndpointAppLovinMax:
		fallthrough
	case models.EndpointWebS2S:
		fallthrough
	default:
		requestExtWrapper, err = models.GetRequestExtWrapper(body)
	}

	return requestExtWrapper, err
}

func GetEndpoint(path, source string, agent string) string {
	switch path {
	case hookexecution.EndpointAuction:
		switch source {
		case "pbjs":
			return models.EndpointWebS2S
		case "owsdk":
			switch agent {
			case models.AppLovinMaxAgent:
				return models.EndpointAppLovinMax
			case models.GoogleSDKAgent:
				return models.EndpointGoogleSDK
			case models.UnityLevelPlayAgent:
				return models.EndpointUnityLevelPlay
			}
			return models.EndpointV25
		default:
			return models.EndpointHybrid
		}
	case OpenWrapAuction:
		return models.EndpointHybrid
	case OpenWrapV25:
		return models.EndpointV25
	case OpenWrapV25Video:
		return models.EndpintInappVideo
	case OpenWrapAmp:
		return models.EndpointAMP
	case OpenWrapOpenRTBVideo:
		return models.EndpointORTB
	case OpenWrapVAST:
		return models.EndpointVAST
	case OpenWrapJSON:
		return models.EndpointJson
	}
	return ""
}

func getSendBurl(request []byte, endpoint string) bool {
	if sdkutils.IsSdkIntegration(endpoint) {
		return true
	}

	//ignore error, default is false
	sendBurl, _ := jsonparser.GetBoolean(request, "ext", "prebid", "bidderparams", "pubmatic", "sendburl")
	return sendBurl
}

func processWrapperExtension(rCtx *models.RequestCtx, r *http.Request, body []byte, result hookstage.HookResult[hookstage.EntrypointPayload]) (hookstage.HookResult[hookstage.EntrypointPayload], bool) {
	requestExtWrapper, err := getRequestWrapper(r, body, rCtx.Endpoint, result)
	if err != nil {
		result.NbrCode = int(nbr.InvalidRequestWrapperExtension)
		result.Errors = append(result.Errors, err.Error())
		return result, false
	}

	if requestExtWrapper.ProfileId <= 0 {
		result.NbrCode = int(nbr.InvalidProfileID)
		result.Errors = append(result.Errors, "ErrMissingProfileID")
		return result, false
	}

	// validate redirect url
	if len(requestExtWrapper.AdServerURL) > 0 {
		if !utils.IsValidURL(requestExtWrapper.AdServerURL) {
			result.NbrCode = int(nbr.InvalidRedirectURL)
			result.Errors = append(result.Errors, "Invalid redirect URL")
			return result, false
		}
	}

	rCtx.ProfileID = requestExtWrapper.ProfileId
	rCtx.DisplayID = requestExtWrapper.VersionId
	rCtx.DisplayVersionID = requestExtWrapper.VersionId
	rCtx.SupportDeals = requestExtWrapper.SupportDeals
	rCtx.ABTestConfig = requestExtWrapper.ABTestConfig
	rCtx.SSAuction = requestExtWrapper.SSAuctionFlag
	rCtx.SummaryDisable = requestExtWrapper.SumryDisableFlag
	rCtx.LoggerImpressionID = requestExtWrapper.LoggerImpressionID
	rCtx.ClientConfigFlag = requestExtWrapper.ClientConfigFlag
	rCtx.SSAI = requestExtWrapper.SSAI
	rCtx.AdruleFlag = requestExtWrapper.Video.AdruleFlag
	rCtx.ProfileIDStr = strconv.Itoa(requestExtWrapper.ProfileId)

	return result, true
}

func initializeRequestContext(rCtx *models.RequestCtx) {
	rCtx.Aliases = make(map[string]string)
	rCtx.ImpBidCtx = make(map[string]models.ImpCtx)
	rCtx.PrebidBidderCode = make(map[string]string)
	rCtx.BidderResponseTimeMillis = make(map[string]int)
	rCtx.SeatNonBids = make(map[string][]openrtb_ext.NonBid)
	rCtx.ImpCountingMethodEnabledBidders = make(map[string]struct{})
}

func processHTTPRequest(rCtx *models.RequestCtx, r *http.Request) {
	queryParams := r.URL.Query()
	source := queryParams.Get("source") //source query param to identify /openrtb2/auction type
	rCtx.Sshb = queryParams.Get("sshb")
	rCtx.Debug = queryParams.Get(models.Debug) == "1"
	rCtx.ResponseFormat = strings.ToLower(strings.TrimSpace(queryParams.Get(models.ResponseFormatKey)))
	rCtx.RedirectURL = queryParams.Get(models.OWRedirectURLKey)

	endpoint := GetEndpoint(r.URL.Path, source, queryParams.Get(models.Agent))
	rCtx.Endpoint = endpoint

	rCtx.Header = r.Header
	rCtx.Method = r.Method
	rCtx.DeviceCtx = models.DeviceCtx{
		UA: r.Header.Get("User-Agent"),
		IP: models.GetIP(r),
	}
	rCtx.ParsedUidCookie = usersync.ReadCookie(r, usersync.Base64Decoder{}, &config.HostCookie{})
}
