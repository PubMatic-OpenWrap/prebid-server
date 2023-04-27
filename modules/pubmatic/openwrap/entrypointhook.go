package openwrap

import (
	"context"
	"net/http"
	"time"

	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/hooks/hookstage"
	v25 "github.com/prebid/prebid-server/modules/pubmatic/openwrap/endpoints/legacy/openrtb/v25"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	uuid "github.com/satori/go.uuid"
)

const (
	OpenWrapAuction  = "/pbs/openrtb2/auction"
	OpenWrapV25      = "/openrtb/2.5"
	OpenWrapV25Video = "/openrtb/2.5/video"
	OpenWrapAmp      = "/openrtb/amp"
)

func (m OpenWrap) handleEntrypointHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	// TODO marshal and log rCtx when request.ext.prebid.trace=verbose

	// TODO in all hooks
	result := hookstage.HookResult[hookstage.EntrypointPayload]{
		Reject: true,
	}

	var err error
	var requestExtWrapper models.RequestExtWrapper
	switch payload.Request.URL.Path {
	// NYC_TODO: Both hybid and non-hybrid flow should be under same API "/openrtb2/auction"
	// but modules should not executre of hybrid flow.
	// check isHybrid()
	case hookexecution.EndpointAuction:
		if !models.IsHybrid(payload.Body) {
			return result, nil
		}
		requestExtWrapper, err = models.GetRequestExtWrapper(payload.Body)
	case OpenWrapAuction:
		return result, nil
	case OpenWrapV25:
		requestExtWrapper, err = models.GetRequestExtWrapper(payload.Body, "ext", "wrapper")
	case OpenWrapV25Video:
		requestExtWrapper, err = v25.ConvertVideoToAuctionRequest(payload, &result)
		if err != nil {
			result.NbrCode = nbr.InvalidVideoRequest
			result.Errors = append(result.Errors, "InvalidVideoRequest")
			return result, err
		}
	case OpenWrapAmp:
		// requestExtWrapper, err = models.GetQueryParamRequestExtWrapper(payload.Body)
	}

	if err != nil || requestExtWrapper.ProfileId == 0 {
		result.NbrCode = nbr.InvalidProfileID
		result.Errors = append(result.Errors, "ErrMissingProfileID")
		return result, err
	}

	queryParams := payload.Request.URL.Query()

	rCtx := models.RequestCtx{
		ProfileID:                 requestExtWrapper.ProfileId,
		DisplayID:                 requestExtWrapper.VersionId,
		LogInfoFlag:               requestExtWrapper.LogInfoFlag,
		PreferDeals:               requestExtWrapper.SupportDeals,
		ABTestConfig:              requestExtWrapper.ABTestConfig,
		SSAuction:                 requestExtWrapper.SSAuctionFlag,
		SummaryDisable:            requestExtWrapper.SumryDisableFlag,
		LoggerImpressionID:        requestExtWrapper.LoggerImpressionID,
		ClientConfigFlag:          requestExtWrapper.ClientConfigFlag,
		SSAI:                      requestExtWrapper.SSAI,
		Aliases:                   make(map[string]string),
		IsCTVRequest:              models.IsCTVAPIRequest(payload.Request.URL.Path),
		UA:                        payload.Request.Header.Get("User-Agent"),
		Debug:                     queryParams.Get(models.Debug) == "1",
		StartTime:                 time.Now().Unix(),
		ImpBidCtx:                 make(map[string]models.ImpCtx),
		URL:                       m.cfg.OpenWrap.Logger.PublicEndpoint,
		IP:                        models.GetIP(payload.Request),
		PrebidBidderCode:          make(map[string]string),
		TrackerEndpoint:           m.cfg.OpenWrap.Tracker.Endpoint,
		VideoErrorTrackerEndpoint: m.cfg.OpenWrap.Tracker.VideoErrorTrackerEndpoint,
		BidderResponseTimeMillis:  make(map[string]int),
	}

	rCtx.UidCookie, err = payload.Request.Cookie(models.UidCookieName)
	if err != nil && err != http.ErrNoCookie {
		result.Errors = append(result.Errors, "failed to parse cookie: uids err: "+err.Error())
	}

	rCtx.KADUSERCookie, err = payload.Request.Cookie(models.KADUSERCOOKIE)
	if err != nil && err != http.ErrNoCookie {
		result.Errors = append(result.Errors, "failed to parse cookie: KADUSERCOOKIE err: "+err.Error())
	}

	originCookie, err := payload.Request.Cookie("origin")
	if err != nil && err != http.ErrNoCookie {
		result.Errors = append(result.Errors, "failed to parse cookie: origin err: "+err.Error())
	}
	if originCookie != nil {
		rCtx.OriginCookie = originCookie.Value
	}

	if rCtx.LoggerImpressionID == "" {
		rCtx.LoggerImpressionID = uuid.NewV4().String()
	}

	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext["rctx"] = rCtx

	result.Reject = false
	return result, nil
}
