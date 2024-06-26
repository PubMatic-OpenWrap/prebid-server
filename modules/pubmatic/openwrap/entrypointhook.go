package openwrap

import (
	"context"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/hooks/hookexecution"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	v25 "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/endpoints/legacy/openrtb/v25"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/usersync"
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
	queryParams := payload.Request.URL.Query()
	source := queryParams.Get("source") //source query param to identify /openrtb2/auction type

	rCtx := models.RequestCtx{}
	var endpoint string
	var pubid int
	var requestExtWrapper models.RequestExtWrapper
	defer func() {
		if result.Reject {
			m.metricEngine.RecordBadRequests(endpoint, getPubmaticErrorCode(openrtb3.NoBidReason(result.NbrCode)))
		} else {
			result.ModuleContext = make(hookstage.ModuleContext)
			result.ModuleContext["rctx"] = rCtx
		}
	}()

	rCtx.Sshb = queryParams.Get("sshb")
	//Do not execute the module for requests processed in SSHB(8001)
	if rCtx.Sshb == models.Enabled {
		rCtx.VastUnwrapEnabled = getVastUnwrapperEnable(payload.Request.Context(), models.VastUnwrapperEnableKey)
		return result, nil
	}
	endpoint = GetEndpoint(payload.Request.URL.Path, source, queryParams.Get(models.Agent))
	if endpoint == models.EndpointHybrid {
		rCtx.Endpoint = models.EndpointHybrid
		return result, nil
	}

	if endpoint == models.EndpointAppLovinMax {
		rCtx.MetricsEngine = m.metricEngine
		// updating body locally to access updated fields from signal
		payload.Body = updateAppLovinMaxRequest(payload.Body, rCtx)
		result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
			ep.Body = payload.Body
			return ep, nil
		}, hookstage.MutationUpdate, "update-max-app-lovin-request")
	}

	// init default for all modules
	result.Reject = true

	requestExtWrapper, err = GetRequestWrapper(payload, result, endpoint)
	if err != nil {
		result.NbrCode = int(nbr.InvalidRequestWrapperExtension)
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if requestExtWrapper.ProfileId <= 0 {
		result.NbrCode = int(nbr.InvalidProfileID)
		result.Errors = append(result.Errors, "ErrMissingProfileID")
		return result, err
	}

	requestDebug, _ := jsonparser.GetBoolean(payload.Body, "ext", "prebid", "debug")
	rCtx = models.RequestCtx{
		StartTime:                 time.Now().Unix(),
		Debug:                     queryParams.Get(models.Debug) == "1" || requestDebug,
		UA:                        payload.Request.Header.Get("User-Agent"),
		ProfileID:                 requestExtWrapper.ProfileId,
		DisplayID:                 requestExtWrapper.VersionId,
		DisplayVersionID:          requestExtWrapper.VersionId,
		LogInfoFlag:               requestExtWrapper.LogInfoFlag,
		SupportDeals:              requestExtWrapper.SupportDeals,
		ABTestConfig:              requestExtWrapper.ABTestConfig,
		SSAuction:                 requestExtWrapper.SSAuctionFlag,
		SummaryDisable:            requestExtWrapper.SumryDisableFlag,
		LoggerImpressionID:        requestExtWrapper.LoggerImpressionID,
		ClientConfigFlag:          requestExtWrapper.ClientConfigFlag,
		SSAI:                      requestExtWrapper.SSAI,
		IP:                        models.GetIP(payload.Request),
		IsCTVRequest:              models.IsCTVAPIRequest(payload.Request.URL.Path),
		TrackerEndpoint:           m.cfg.Tracker.Endpoint,
		VideoErrorTrackerEndpoint: m.cfg.Tracker.VideoErrorTrackerEndpoint,
		Aliases:                   make(map[string]string),
		ImpBidCtx:                 make(map[string]models.ImpCtx),
		PrebidBidderCode:          make(map[string]string),
		BidderResponseTimeMillis:  make(map[string]int),
		ProfileIDStr:              strconv.Itoa(requestExtWrapper.ProfileId),
		Endpoint:                  endpoint,
		MetricsEngine:             m.metricEngine,
		DCName:                    m.cfg.Server.DCName,
		SeatNonBids:               make(map[string][]openrtb_ext.NonBid),
		ParsedUidCookie:           usersync.ReadCookie(payload.Request, usersync.Base64Decoder{}, &config.HostCookie{}),
		TMax:                      m.cfg.Timeout.MaxTimeout,
	}

	// only http.ErrNoCookie is returned, we can ignore it
	rCtx.UidCookie, _ = payload.Request.Cookie(models.UidCookieName)
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

	result.Reject = false
	return result, nil
}

func GetRequestWrapper(payload hookstage.EntrypointPayload, result hookstage.HookResult[hookstage.EntrypointPayload], endpoint string) (models.RequestExtWrapper, error) {
	var requestExtWrapper models.RequestExtWrapper
	var err error
	switch endpoint {
	case models.EndpintInappVideo:
		requestExtWrapper, err = v25.ConvertVideoToAuctionRequest(payload, &result)
	case models.EndpointAMP:
		requestExtWrapper, err = models.GetQueryParamRequestExtWrapper(payload.Request)
	case models.EndpointV25:
		fallthrough
	case models.EndpointVideo, models.EndpointVAST, models.EndpointJson:
		requestExtWrapper, err = models.GetRequestExtWrapper(payload.Body, "ext", "wrapper")
	case models.EndpointWebS2S, models.EndpointAppLovinMax:
		fallthrough
	default:
		requestExtWrapper, err = models.GetRequestExtWrapper(payload.Body)
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
		return models.EndpointVideo
	case OpenWrapVAST:
		return models.EndpointVAST
	case OpenWrapJSON:
		return models.EndpointJson
	}
	return ""
}
