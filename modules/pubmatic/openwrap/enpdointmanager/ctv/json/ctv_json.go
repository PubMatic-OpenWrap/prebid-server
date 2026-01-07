package ctvjson

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod"
	impressions "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/legacy/impressions"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/auction"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/creativecache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	ctvutils "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/utils"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/stage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type CTVJSON struct {
	metricsEngine metrics.MetricsEngine
	creativeCache creativecache.Client
}

func NewCTVJSON(metricsEngine metrics.MetricsEngine, creativeCache creativecache.Client) *CTVJSON {
	return &CTVJSON{
		metricsEngine: metricsEngine,
		creativeCache: creativeCache,
	}
}

func (cj *CTVJSON) HandleGETEndpoint(
	rCtx *models.RequestCtx,
	payload stage.EntrypointPayload,
	moduleCtx stage.ModuleContext,
	result stage.EntrypointResult,
) ([]byte, stage.EntrypointResult, bool) {
	// Set IsCTVRequest flag
	rCtx.IsCTVRequest = models.IsCTVAPIRequest(payload.Request.URL.Path)

	if payload.Request.Method != http.MethodGet {
		return payload.Body, result, true
	}

	bidRequest, err := ctv.NewOpenRTB(payload.Request).ParseORTBRequest(ctv.GetORTBParserMap())
	if err != nil {
		nbr := openrtb3.NoBidInvalidRequest.Ptr()
		if cerr, ok := err.(*ctv.ParseError); ok {
			nbr = cerr.NBR()
		}
		result.NbrCode = int(*nbr)
		result.Errors = append(result.Errors, err.Error())
		return payload.Body, result, false
	}

	body, err := json.Marshal(bidRequest)
	if err != nil {
		result.NbrCode = int(openrtb3.NoBidTechnicalError)
		result.Errors = append(result.Errors, "error occured in request processing")
		return payload.Body, result, false
	}

	result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
		if payload.Request.Method == http.MethodGet {
			ep.Body = body
		}
		return ep, nil
	}, hookstage.MutationUpdate, "ctv-get-endpoint")

	return body, result, true
}

func (cj *CTVJSON) HandleEntrypointHook(
	rCtx *models.RequestCtx,
	payload stage.EntrypointPayload,
	moduleCtx stage.ModuleContext,
	result stage.EntrypointResult,
) (stage.EntrypointResult, bool) {
	cj.metricsEngine.RecordCTVHTTPMethodRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Method)
	if len(rCtx.ResponseFormat) > 0 {
		if rCtx.ResponseFormat != models.ResponseFormatJSON && rCtx.ResponseFormat != models.ResponseFormatRedirect {
			result.NbrCode = int(nbr.InvalidResponseFormat)
			result.Errors = append(result.Errors, "Invalid response format, must be 'json' or 'redirect'")
			return result, false
		}
	}

	// SSAuction will be always 1 for CTV request
	rCtx.SSAuction = 1
	rCtx.ImpAdPodConfig = make(map[string][]models.PodConfig)

	return result, true
}

func (cj *CTVJSON) HandleRawAuctionHook(
	rCtx *models.RequestCtx,
	payload stage.RawAuctionPayload,
	moduleCtx stage.ModuleContext,
	result stage.RawAuctionResult,
) (stage.RawAuctionResult, bool) {
	return result, true
}

func (cj *CTVJSON) HandleBeforeValidationHook(
	rCtx *models.RequestCtx,
	payload stage.BeforeValidationPayload,
	moduleCtx stage.ModuleContext,
	result stage.BeforeValidationResult,
) (stage.BeforeValidationResult, bool) {
	// Redirect URL
	result, ok := processRedirectURL(rCtx, result)
	if !ok {
		return result, false
	}

	// Validate video request
	err := ctvutils.ValidateVideoImpressions(payload.BidRequest)
	if err != nil {
		result.NbrCode = int(nbr.InvalidVideoRequest)
		result.Errors = append(result.Errors, err.Error())
		return result, false
	}

	// Populate rctx with ctv features
	ctvutils.PopulateRequestContextWithCTVFeatures(rCtx)

	// update Adpod Configs from json endpoint features
	errs := updateAdpodConfigs(rCtx, payload.BidRequest)
	if len(errs) > 0 {
		for _, err := range errs {
			result.Warnings = append(result.Warnings, err.Error())
		}
	}

	// Set Default values to V25 dynamic adpod configs
	adpod.SetDefaultValuesToAdpodConfig(rCtx)

	// Adpod config Validation
	err = adpod.ValidateAdpodConfigs(rCtx)
	if err != nil {
		result.NbrCode = int(nbr.InvalidAdpodConfig)
		result.Errors = append(result.Errors, err.Error())
		return result, false
	}

	result.ChangeSet.AddMutation(func(ep stage.BeforeValidationPayload) (stage.BeforeValidationPayload, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleBeforeValidationHook mutation")
			return ep, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		// filter imps with invalid adserver url
		filterImpsWithInvalidAdserverURL(&rCtx, payload.BidRequest)

		if ep.BidRequest.Source != nil && ep.BidRequest.Source.SChain != nil {
			err := ctvutils.IsValidSchain(ep.BidRequest.Source.SChain)
			if err != nil {
				schainBytes, _ := json.Marshal(ep.BidRequest.Source.SChain)
				glog.Errorf(ctv.ErrSchainValidationFailed, models.SChainKey, err.Error(), rCtx.PubIDStr, rCtx.ProfileIDStr, string(schainBytes))
				ep.BidRequest.Source.SChain = nil
			}
		}

		err := ctvutils.FilterNonVideoImpressions(ep.BidRequest)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		}

		// Remove adpod data from ext
		ctvutils.RemoveAdpodDataFromExt(ep.BidRequest)

		// Add GAM URL configs
		err = adpod.ApplyGAMURLConfig(&rCtx, ep.BidRequest)
		if err != nil {
			result.Warnings = append(result.Warnings, "Failed to apply GAM URL configs: "+err.Error())
		}

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-json-before-validation")

	return result, true
}

func (cj *CTVJSON) HandleProcessedAuctionHook(
	rCtx *models.RequestCtx,
	payload stage.ProcessedAuctionPayload,
	moduleCtx stage.ModuleContext,
	result stage.ProcessedAuctionResult,
) (stage.ProcessedAuctionResult, bool) {
	result.ChangeSet.AddMutation(func(parp stage.ProcessedAuctionPayload) (stage.ProcessedAuctionPayload, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleProcessedAuctionHook mutation")
			return parp, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		imps, errs := impressions.GenerateImpressions(rCtx, payload.Request)
		if len(errs) > 0 {
			for i := range errs {
				result.Warnings = append(result.Warnings, errs[i].Error())
			}
		}

		if len(imps) > 0 {
			parp.Request.SetImp(imps)
		}

		// filter vast tag durations
		adapters.FilterImpsVastTagsByDuration(rCtx, parp.Request)

		return parp, nil
	}, hookstage.MutationUpdate, "update-ctv-impressions")

	return result, true
}

func (cj *CTVJSON) HandleBidderRequestHook(
	rCtx *models.RequestCtx,
	payload stage.BidderRequestPayload,
	moduleCtx stage.ModuleContext,
	result stage.BidderRequestResult,
) (stage.BidderRequestResult, bool) {
	result.ChangeSet.AddMutation(func(ep stage.BidderRequestPayload) (stage.BidderRequestPayload, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleBidderRequestHook mutation")
			return ep, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		// if payload.BidderInfo.OpenRTB.Version != "2.6" && len(rCtx.AdpodCtx) > 0 {
		// 	adpod.ConvertDownTo25(ep.Request)
		// }

		// Remove Multibid object from ext
		reqExt, err := ep.Request.GetRequestExt()
		if err != nil {
			result.Errors = append(result.Errors, "failed to get request ext in CTV handleBidderRequestHook mutation")
			return ep, nil
		}
		prebidExt := reqExt.GetPrebid()
		prebidExt.MultiBid = nil
		reqExt.SetPrebid(prebidExt)

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-json-bidder-request")

	return result, true
}

func (cj *CTVJSON) HandleRawBidderResponseHook(
	rCtx *models.RequestCtx,
	payload stage.RawBidderResponsePayload,
	moduleCtx stage.ModuleContext,
	result stage.RawBidderResponseResult,
) (stage.RawBidderResponseResult, bool) {
	return result, true
}

func (cj *CTVJSON) HandleAllProcessedBidResponsesHook(
	rCtx *models.RequestCtx,
	payload stage.AllProcessedBidResponsesPayload,
	moduleCtx stage.ModuleContext,
	result stage.AllProcessedBidResponsesResult,
) (stage.AllProcessedBidResponsesResult, bool) {
	result.ChangeSet.AddMutation(func(apbrp stage.AllProcessedBidResponsesPayload) (stage.AllProcessedBidResponsesPayload, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleAllProcessedBidResponsesHook mutation")
			return apbrp, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		// Move to Raw bidder response hook once 2.6 fully supported
		adpod.ConvertUpTo26(rCtx, apbrp.Responses)
		return apbrp, nil
	}, hookstage.MutationUpdate, "update-bid-duration")

	return result, true
}

func (cj *CTVJSON) HandleAuctionResponseHook(
	rCtx *models.RequestCtx,
	payload stage.AuctionResponsePayload,
	moduleCtx stage.ModuleContext,
	result stage.AuctionResponseResult,
) (stage.AuctionResponseResult, bool) {
	// Add targetting keys
	for _, seatBid := range payload.BidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			if bidCtx.Prebid == nil {
				bidCtx.Prebid = &openrtb_ext.ExtBidPrebid{}
			}

			if bidCtx.Prebid.Targeting == nil {
				bidCtx.Prebid.Targeting = make(map[string]string)
			}

			value := ctvutils.GetTargeting(openrtb_ext.CategoryDurationKey, openrtb_ext.BidderName(seatBid.Seat), bidCtx)
			if value != "" {
				ctvutils.AddTargetingKey(bidCtx, openrtb_ext.CategoryDurationKey, value)
			}

			value = ctvutils.GetTargeting(openrtb_ext.PbKey, openrtb_ext.BidderName(seatBid.Seat), bidCtx)
			if value != "" {
				ctvutils.AddTargetingKey(bidCtx, openrtb_ext.PbKey, value)
			}

			impCtx.BidCtx[bid.ID] = bidCtx
			rCtx.ImpBidCtx[bid.ImpID] = impCtx
		}
	}

	// perform adpod auction
	if len(rCtx.AdpodCtx) > 0 {
		auction.AdpodAuction(rCtx, payload.BidResponse, result)
	}

	result.ChangeSet.AddMutation(func(arp stage.AuctionResponsePayload) (stage.AuctionResponsePayload, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleExitpointHook mutation")
			return arp, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		for _, seatBid := range arp.BidResponse.SeatBid {
			for _, bid := range seatBid.Bid {
				ctvutils.AddPWTTargetingKeysForAdpod(rCtx, &bid, seatBid.Seat)
			}
		}
		return arp, nil
	}, hookstage.MutationUpdate, "add-pwt-targeting-keys-for-adpod")

	return result, true
}

func (cj *CTVJSON) HandleExitpointHook(
	rCtx *models.RequestCtx,
	payload stage.ExitpointPayload,
	moduleCtx stage.ModuleContext,
	result stage.ExitpointResult,
) (stage.ExitpointResult, bool) {
	response, ok := payload.Response.(*openrtb2.BidResponse)
	if !ok {
		return result, true
	}

	if response.NBR != nil {
		return result, true
	}

	adpodBids := formResponse(rCtx, response, cj.creativeCache)

	result.ChangeSet.AddMutation(func(ep stage.ExitpointPayload) (stage.ExitpointPayload, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleExitpointHook mutation")
			return ep, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()
		ep.Response = bidResponseAdpod{AdPodBids: adpodBids, Ext: response.Ext}
		ep.W.Header().Set("Content-Type", "application/json")
		ep.W.Header().Set("Content-Options", "nosniff")
		ctvutils.SetCORSHeaders(ep.W, rCtx.Header)
		if checkRedirectResponse(rCtx) {
			redirectURL := rCtx.RedirectURL
			if len(adpodBids) > 0 {
				redirectURL = updateAdServerURL(adpodBids[0].Targeting, rCtx.RedirectURL)
			}
			ep.W.Header().Set("Location", redirectURL)
			ep.W.WriteHeader(http.StatusFound)

		}
		return ep, nil
	}, hookstage.MutationUpdate, "ctv-json-exitpoint")

	return result, true
}
