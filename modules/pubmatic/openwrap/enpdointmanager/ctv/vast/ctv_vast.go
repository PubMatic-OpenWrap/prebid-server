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
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type CTVVAST struct {
	metricsEngine metrics.MetricsEngine
}

func NewCTVVAST(metricsEngine metrics.MetricsEngine) *CTVVAST {
	return &CTVVAST{
		metricsEngine: metricsEngine,
	}
}

func (cv *CTVVAST) HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.EntrypointPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.EntrypointPayload], error) {
	cv.metricsEngine.RecordCTVHTTPMethodRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Method)
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	videoAdDuration := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationKey)
	policy := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationMatchingKey)
	if len(videoAdDuration) > 0 {
		rCtx.AdpodProfileConfig = &models.AdpodProfileConfig{
			AdserverCreativeDurations:              utils.GetIntArrayFromString(videoAdDuration, models.ArraySeparator),
			AdserverCreativeDurationMatchingPolicy: policy,
		}
	}

	err := ctvutils.ValidateVideoImpressions(payload.BidRequest)
	if err != nil {
		result.NbrCode = int(nbr.InvalidVideoRequest)
		result.Errors = append(result.Errors, err.Error())
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*payload.BidRequest) // for wrapper logger sz
		return rCtx, result, nil
	}

	ctvutils.SetIncludeBrandCategory(rCtx)

	for _, imp := range payload.BidRequest.Imp {
		impCtx, ok := rCtx.ImpBidCtx[imp.ID]
		if !ok {
			continue
		}

		podID := imp.Video.PodID
		if podID == "" {
			podID = imp.ID
		}

		_, ok = rCtx.AdpodCtx[podID]
		//Adding default durations for CTV Test requests
		if rCtx.IsTestRequest > 0 && ok && rCtx.AdpodProfileConfig == nil {
			rCtx.AdpodProfileConfig = &models.AdpodProfileConfig{
				AdserverCreativeDurations:              []int{5, 10},
				AdserverCreativeDurationMatchingPolicy: openrtb_ext.OWRoundupVideoAdDurationMatching,
			}
		}

		rCtx.ImpBidCtx[imp.ID] = impCtx
	}

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		rCtxInterface, _ := moduleCtx.ModuleContext.Get("rctx")
		rCtx := rCtxInterface.(models.RequestCtx)
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

		// Enable when UI support is added
		// ep.BidRequest = adpod.ApplyAdpodConfigs(rCtx, ep.BidRequest)

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-vast-before-validation")
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
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
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BidderRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BidderRequestPayload], error) {
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
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawBidderResponsePayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error) {
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
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	// perform adpod auction
	if len(rCtx.AdpodCtx) > 0 {
		var ok bool
		result, ok = auction.AdpodAuction(&rCtx, result, payload.BidResponse)
		if !ok {
			return rCtx, result, nil
		}
	}
	return rCtx, result, nil
}

func (cv *CTVVAST) HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ExitpointPaylaod], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ExitpointPaylaod], error) {
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
	return rCtx, result, nil
}
