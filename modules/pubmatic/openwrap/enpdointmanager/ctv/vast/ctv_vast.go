package ctvvast

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod"
	impressions "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/legacy/impressions"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/auction"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	ctvutils "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/utils"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
)

type CTVVAST struct {
	metricsEngine metrics.MetricsEngine
}

func NewCTVVAST(metricsEngine metrics.MetricsEngine) *CTVVAST {
	return &CTVVAST{
		metricsEngine: metricsEngine,
	}
}

func (cv *CTVVAST) HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.EntrypointPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	cv.metricsEngine.RecordCTVHTTPMethodRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Method)
	return true
}

func (cv *CTVVAST) HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.RawAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (cv *CTVVAST) HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.BeforeValidationRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	// Validate video request
	err := ctvutils.ValidateVideoImpressions(payload.BidRequest)
	if err != nil {
		result.NbrCode = int(nbr.InvalidVideoRequest)
		result.Errors = append(result.Errors, err.Error())
		return false
	}

	// Populate rctx with ctv features
	ctvutils.SetIncludeBrandCategory(rCtx)
	ctvutils.AddMultiBidConfigurations(rCtx)
	ctvutils.ProcessAdpodProfileConfig(rCtx)

	// Set Default values to V25 dynamic adpod configs
	adpod.SetDefaultValuesToAdpodConfig(rCtx)

	// Adpod config Validation
	err = adpod.ValidateAdpodConfigs(rCtx)
	if err != nil {
		result.NbrCode = int(nbr.InvalidAdpodConfig)
		result.Errors = append(result.Errors, err.Error())
		return false
	}

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
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
	}, hookstage.MutationUpdate, "ctv-vast-before-validation")
	return true
}

func (cv *CTVVAST) HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		rCtxInterface, _ := moduleCtx.ModuleContext.Get("rctx")
		rCtx := rCtxInterface.(models.RequestCtx)
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
	return true
}

func (cv *CTVVAST) HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.BidderRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	result.ChangeSet.AddMutation(func(ep hookstage.BidderRequestPayload) (hookstage.BidderRequestPayload, error) {
		rCtxInterface, _ := moduleCtx.ModuleContext.Get("rctx")
		rCtx := rCtxInterface.(models.RequestCtx)
		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		// if payload.BidderInfo.OpenRTB.Version != "2.6" && len(rCtx.AdpodCtx) > 0 {
		// 	adpod.ConvertDownTo25(ep.Request)
		// }

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-openrtb-bidder-request")
	return true
}

func (cv *CTVVAST) HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.RawBidderResponsePayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (cv *CTVVAST) HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	result.ChangeSet.AddMutation(func(apbrp hookstage.AllProcessedBidResponsesPayload) (hookstage.AllProcessedBidResponsesPayload, error) {
		rCtxInterface, _ := moduleCtx.ModuleContext.Get("rctx")
		rCtx := rCtxInterface.(models.RequestCtx)
		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		// Move to Raw bidder response hook once 2.6 fully supported
		adpod.ConvertUpTo26(rCtx, apbrp.Responses)
		return apbrp, nil
	}, hookstage.MutationUpdate, "update-bid-duration")
	return true
}

func (cv *CTVVAST) HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.AuctionResponsePayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	// perform adpod auction
	if len(rCtx.AdpodCtx) > 0 {
		auction.AdpodAuction(rCtx, result, payload.BidResponse)
	}
	return true
}

func (cv *CTVVAST) HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.ExitpointPaylaod], moduleCtx hookstage.ModuleInvocationContext) bool {
	result.ChangeSet.AddMutation(func(ep hookstage.ExitpointPaylaod) (hookstage.ExitpointPaylaod, error) {
		rCtx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleExitpointHook mutation")
			return ep, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rCtx)
		}()

		ep.W.Header().Set("Content-Type", "application/xml")
		ep.W.Header().Set("Content-Options", "nosniff")

		response, ok := ep.Response.(*openrtb2.BidResponse)
		if !ok {
			ep.Response = EmptyVASTResponse
			return ep, nil
		}

		if nbr := response.NBR; nbr != nil {
			ep.Response = EmptyVASTResponse
			if rCtx.Debug {
				ep.W.Header().Set(HeaderOpenWrapStatus, fmt.Sprintf(NBRFormat, *nbr))
			}
			return ep, nil
		}

		var nbr *openrtb3.NoBidReason
		ep.Response, nbr = formVastResponse(&rCtx, response)
		if nbr != nil {
			if rCtx.Debug {
				ep.W.Header().Set(HeaderOpenWrapStatus, fmt.Sprintf(NBRFormat, *nbr))
			}
		}

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-vast-exitpoint")
	return true
}
