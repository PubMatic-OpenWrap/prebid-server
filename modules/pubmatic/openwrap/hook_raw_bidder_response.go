package openwrap

import (
	"fmt"

	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/hooks/hookstage"
)

func handleRawBidderResponseHook(
	payload hookstage.RawBidderResponsePayload,
	moduleCtx hookstage.ModuleContext,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {

	rCtx, ok := moduleCtx["rctx"].(RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	defer func() {
		moduleCtx["rctx"] = rCtx
	}()
	util.JLogf("VastUnwrapperEnable", rCtx.VastUnwrapFlag)

	if !rCtx.VastUnwrapFlag {
		fmt.Printf("\n **** VAST unwrapping Disabled **** !!!! ")
	} else {
		fmt.Printf("\n VAST unwrapping Enabled  !!!! ")
	}
	return result, nil
}
