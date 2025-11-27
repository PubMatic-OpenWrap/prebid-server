package ctvopenrtb

import (
	"encoding/json"

	"github.com/golang/glog"
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
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/stage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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
) (stage.EntrypointResult, bool) {
	co.metricsEngine.RecordCTVHTTPMethodRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Method)
	return result, true
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
		return result, false
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

		// Enable when UI support is added
		// ep.BidRequest = adpod.ApplyAdpodConfigs(rCtx, ep.BidRequest)

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

		if payload.Bidder == models.BidderVASTBidder {
			adapters.FilterImpsVastTagsByDuration(rCtx, ep.Request)
		}
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
	return result, true
}
