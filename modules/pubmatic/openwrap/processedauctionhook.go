package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adpod/impressions"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (m OpenWrap) HandleProcessedAuctionHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.ProcessedAuctionRequestPayload,
) (hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	result := hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{}

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	rctx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	defer func() {
		moduleCtx.ModuleContext["rctx"] = rctx
	}()

	var imps []*openrtb_ext.ImpWrapper
	var errs []error
	if rctx.IsCTVRequest {
		imps, errs = impressions.GenerateImpressions(payload.RequestWrapper, rctx.ImpBidCtx)
		if len(errs) > 0 {
			for i := range errs {
				result.Warnings = append(result.Warnings, errs[i].Error())
			}
		}
	}

	ip := rctx.IP

	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		if parp.RequestWrapper != nil && parp.RequestWrapper.BidRequest.Device != nil && (parp.RequestWrapper.BidRequest.Device.IP == "" && parp.RequestWrapper.BidRequest.Device.IPv6 == "") {
			parp.RequestWrapper.BidRequest.Device.IP = ip
		}

		if rctx.IsCTVRequest {
			if len(imps) > 0 {
				parp.RequestWrapper.SetImp(imps)
			}
		}
		return parp, nil
	}, hookstage.MutationUpdate, "update-device-ip")

	return result, nil
}
