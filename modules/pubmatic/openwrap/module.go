package openwrap

import (
	"context"
	"encoding/json"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
)

func Builder(_ json.RawMessage, _ moduledeps.ModuleDeps) (interface{}, error) {
	return Module{}, nil
}

type Module struct{}

// HandleBidderRequestHook updates blocking fields on the openrtb2.BidRequest.
// Fields are updated only if request satisfies conditions provided by the module config.
func (m Module) HandleBidderRequestHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BidderRequestPayload,
) (hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	var err error
	result := hookstage.HookResult[hookstage.BidderRequestPayload]{}
	// if len(miCtx.AccountConfig) == 0 {
	// 	return result, nil
	// }

	// cfg, err := newConfig(miCtx.AccountConfig)
	// if err != nil {
	// 	return result, err
	// }

	//return handleBidderRequestHook(cfg, payload)
	return result, err
}

// HandleRawBidderResponseHook rejects bids for a specific bidder if they fail the attribute check.
func (m Module) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	//result := hookstage.HookResult[hookstage.RawBidderResponsePayload]{}
	// var cfg config
	// if len(miCtx.AccountConfig) != 0 {
	// 	ncfg, err := newConfig(miCtx.AccountConfig)
	// 	if err != nil {
	// 		return result, err
	// 	}
	// 	cfg = ncfg
	// }

	return handleRawBidderResponseHook(payload, miCtx.ModuleContext)
}
