package openwrap

import (
	"context"
	"encoding/json"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
)

func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
	return initOpenWrap(rawCfg, deps)
}

// HandleRawBidderResponseHook unwraps VAST creatives if vast un-wrapper feature is enabled
func (m OpenWrap) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	result := hookstage.HookResult[hookstage.RawBidderResponsePayload]{}

	if m.cfg.OpenWrap.Vastunwrap.Enabled {
		return handleRawBidderResponseHook(payload, miCtx.ModuleContext)
	}
	return result, nil

}

// HandleEntrypointHook retrieves vast un-wrapper flag and User-agent proivded in request context
func (m OpenWrap) HandleEntrypointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	if m.cfg.OpenWrap.Vastunwrap.Enabled {
		return handleEntrypointHook(ctx, miCtx, payload)
	}
	return hookstage.HookResult[hookstage.EntrypointPayload]{}, nil
}
