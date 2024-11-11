package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adpod"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func (m OpenWrap) HandleProcessedAuctionHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.ProcessedAuctionRequestPayload,
) (hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	result := hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{}

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleProcessedAuctionHook()")
		return result, nil
	}
	rctx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleProcessedAuctionHook()")
		return result, nil
	}

	//Do not execute the module for requests processed in SSHB(8001)
	if rctx.Sshb == "1" || rctx.Endpoint == models.EndpointHybrid {
		return result, nil
	}
	defer func() {
		moduleCtx.ModuleContext["rctx"] = rctx
	}()

	var imps []*openrtb_ext.ImpWrapper
	//var errs []error
	// if rctx.IsCTVRequest {
	// 	imps, errs = impressions.GenerateImpressions(payload.Request, rctx.ImpBidCtx, rctx.AdpodProfileConfig, rctx.PubIDStr, m.metricEngine)
	// 	if len(errs) > 0 {
	// 		for i := range errs {
	// 			result.Warnings = append(result.Warnings, errs[i].Error())
	// 		}
	// 	}
	// 	adapters.FilterImpsVastTagsByDuration(imps, rctx.ImpBidCtx)
	// }
	if rctx.IsCTVRequest {
		for _, imp := range payload.Request.Imp {
			impCtx, ok := rctx.ImpBidCtx[imp.ID]
			if !ok {
				continue
			}
			if imp.Video != nil {
				switch adpod.GetPodType(impCtx) {
				case models.Dynamic:
					podId := imp.Video.PodID
					if impCtx.AdpodConfig != nil {
						podId = imp.ID
					}
					rctx.AdpodCtx[podId] = adpod.NewDynamicAdpod(podId, imp, impCtx, rctx.AdpodProfileConfig, rctx.NewReqExt.AdPod)
					// case models.Structured:
					// 	if _, ok := rctx.AdpodCtx[imp.Video.PodID]; !ok {
					// 		rctx.AdpodCtx[imp.Video.PodID] = adpod.NewStructuredAdpod(imp.Video.PodID, impCtx, rctx.AdpodProfileConfig, rctx.NewReqExt.AdPod)
					// 	}
				}
			}
		}
	}

	if rctx.IsCTVRequest {
		for _, impWrapper := range payload.Request.GetImp() {
			impCtx, ok := rctx.ImpBidCtx[impWrapper.ID]
			if !ok {
				continue
			}
			if impWrapper.Video != nil {
				switch adpod.GetPodType(impCtx) {
				case models.Dynamic:
					podId := impWrapper.Video.PodID
					if impCtx.AdpodConfig != nil {
						podId = impWrapper.ID
					}
					dynamicAdpod := rctx.AdpodCtx[podId].(*adpod.DynamicAdpod)
					generatedImps := dynamicAdpod.GetImpressions()
					for i := range generatedImps {
						rctx.ImpToPodId[generatedImps[i].ID] = podId
					}
					imps = append(imps, generatedImps...)
					// case models.Structured:
					// 	structuredAdpod := rctx.AdpodCtx[impWrapper.Video.PodID].(*adpod.StructuredAdpod)
					// 	structuredAdpod.AddImpressions(*impWrapper.Imp)
					// 	rctx.ImpToPodId[impWrapper.ID] = impWrapper.Video.PodID
					// 	imps = append(imps, impWrapper)
					// }
				}
			}
			//TODO: Check if we require this for structured adpod
			adapters.FilterImpsVastTagsByDuration(imps, rctx.ImpBidCtx)
		}
	}
	ip := rctx.IP

	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		if parp.Request != nil && parp.Request.BidRequest.Device != nil && (parp.Request.BidRequest.Device.IP == "" && parp.Request.BidRequest.Device.IPv6 == "") {
			parp.Request.BidRequest.Device.IP = ip
		}

		if rctx.IsCTVRequest {
			if len(imps) > 0 {
				parp.Request.SetImp(imps)
			}
		}
		return parp, nil
	}, hookstage.MutationUpdate, "update-device-ip")

	return result, nil
}
