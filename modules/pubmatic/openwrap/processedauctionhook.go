package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adpod/impressions"
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
	var errs []error
	if rctx.IsCTVRequest {
		imps, errs = impressions.GenerateImpressions(payload.Request, rctx.ImpBidCtx, rctx.AdpodProfileConfig, rctx.PubIDStr, m.metricEngine)
		if len(errs) > 0 {
			for i := range errs {
				result.Warnings = append(result.Warnings, errs[i].Error())
			}
		}
		adapters.FilterImpsVastTagsByDuration(imps, rctx.ImpBidCtx)
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
