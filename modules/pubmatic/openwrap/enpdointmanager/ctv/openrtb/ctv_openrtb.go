package ctvopenrtb

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod"
	impressions "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/legacy/impressions"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/auction"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	ctvutils "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/utils"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/stage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
)

type CTVOpenRTB struct {
	metricsEngine metrics.MetricsEngine
}

func NewCTVOpenRTB(metricsEngine metrics.MetricsEngine) *CTVOpenRTB {
	return &CTVOpenRTB{
		metricsEngine: metricsEngine,
	}
}

// EntrypointHook
func (co *CTVOpenRTB) HandleEntrypointHook(
	rCtx *models.RequestCtx,
	payload stage.EntrypointPayload,
	moduleCtx stage.ModuleContext,
	result stage.EntrypointResult,
) ([]byte, stage.EntrypointResult, bool) {
	co.metricsEngine.RecordCTVHTTPMethodRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Method)
	// SSAuction will be always 1 for CTV request
	rCtx.SSAuction = 1
	rCtx.ImpAdPodConfig = make(map[string][]models.PodConfig)
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
		result.Errors = append(result.Errors, "error occured in request proecessing")
		return payload.Body, result, false
	}

	result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
		if payload.Request.Method == http.MethodGet {
			ep.Body = body
		}
		return ep, nil
	}, hookstage.MutationUpdate, "ctv-openrtb-entrypoint")

	return body, result, true
}

// RawAuctionHook
func (co *CTVOpenRTB) HandleRawAuctionHook(
	rCtx *models.RequestCtx,
	payload stage.RawAuctionPayload,
	moduleCtx stage.ModuleContext,
	result stage.RawAuctionResult,
) (stage.RawAuctionResult, bool) {
	return result, true
}

// BeforeValidationHook
func (co *CTVOpenRTB) HandleBeforeValidationHook(
	rCtx *models.RequestCtx,
	payload stage.BeforeValidationPayload,
	moduleCtx stage.ModuleContext,
	result stage.BeforeValidationResult,
) (stage.BeforeValidationResult, bool) {
	// Validate video request
	err := ctvutils.ValidateVideoImpressions(payload.BidRequest)
	if err != nil {
		result.NbrCode = int(nbr.InvalidVideoRequest)
		result.Errors = append(result.Errors, err.Error())
		return result, false
	}

	// Populate rctx with ctv features
	ctvutils.PopulateRequestContextWithCTVFeatures(rCtx)

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

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-openrtb-before-validation")

	return result, true
}

// ProcessedAuctionHook
func (co *CTVOpenRTB) HandleProcessedAuctionHook(
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

		return parp, nil
	}, hookstage.MutationUpdate, "update-ctv-impressions")

	return result, true
}

// BidderRequestHook
func (co *CTVOpenRTB) HandleBidderRequestHook(
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
	}, hookstage.MutationUpdate, "ctv-openrtb-bidder-request")

	return result, true
}

// RawBidderResponseHook
func (co *CTVOpenRTB) HandleRawBidderResponseHook(
	rCtx *models.RequestCtx,
	payload stage.RawBidderResponsePayload,
	moduleCtx stage.ModuleContext,
	result stage.RawBidderResponseResult,
) (stage.RawBidderResponseResult, bool) {
	return result, true
}

// AllProcessedBidResponsesHook
func (co *CTVOpenRTB) HandleAllProcessedBidResponsesHook(
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

// AuctionResponseHook
func (co *CTVOpenRTB) HandleAuctionResponseHook(
	rCtx *models.RequestCtx,
	payload stage.AuctionResponsePayload,
	moduleCtx stage.ModuleContext,
	result stage.AuctionResponseResult,
) (stage.AuctionResponseResult, bool) {
	// perform adpod auction
	if len(rCtx.AdpodCtx) > 0 {
		auction.AdpodAuction(rCtx, payload.BidResponse, result)
	}

	return result, true
}

// ExitpointHook
func (co *CTVOpenRTB) HandleExitpointHook(
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

	response = formResponse(rCtx, response)

	result.ChangeSet.AddMutation(func(ep hookstage.ExitpointPaylaod) (hookstage.ExitpointPaylaod, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleExitpointHook mutation")
			return ep, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		ep.W.Header().Set("Content-Type", "application/json")
		ep.W.Header().Set("Content-Options", "nosniff")
		ctvutils.SetCORSHeaders(ep.W, rCtx.Header)

		// Update response
		ep.Response = response

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-openrtb-exitpoint")
	return result, true
}
