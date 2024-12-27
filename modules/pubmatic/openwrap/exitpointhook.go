package openwrap

import (
	"context"
	"encoding/json"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adpod"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) handleExitpointHook(
	_ context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.ExitPointPayload,
) (result hookstage.HookResult[hookstage.ExitPointPayload], err error) {

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}

	if rCtx.Endpoint != models.EndpointORTB && rCtx.Endpoint != models.EndpointVAST && rCtx.Endpoint != models.EndpointJson {
		return result, nil
	}

	responder := adpod.NewResponder(rCtx)

	// converting raw response to bid response for forming endpoint response
	var bidResp openrtb2.BidResponse
	err = json.Unmarshal(payload.RawResponse, &bidResp)
	if err != nil {
		// ToDo: Form error response as per the endpoint
		return result, nil
	}

	response := responder.FormResponse(&bidResp)
	updatedRawResponse, err := json.Marshal(response)
	if err != nil {
		// ToDo: Form error response as per the endpoint
		return result, nil
	}

	result.ChangeSet.AddMutation(func(ep hookstage.ExitPointPayload) (hookstage.ExitPointPayload, error) {
		ep.RawResponse = updatedRawResponse
		return ep, nil
	}, hookstage.MutationUpdate, "update-response-and-headers")

	return result, nil
}
