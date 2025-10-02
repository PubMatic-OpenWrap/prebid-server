package openwrap

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/moduledeps"
)

// init openwrap module and its dependecies like config, cache, db connection, bidder cfg, etc.
func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
	return initOpenWrap(rawCfg, deps)
}

// temporary openwrap changes to support non-pbs apis like openrtb/2.5, openrtb/amp, etc
// temporary openwrap changes to support non-ortb fields like request.ext.wrapper
func (m OpenWrap) HandleEntrypointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			m.metricEngine.RecordOpenWrapServerPanicStats(m.cfg.Server.HostName, "HandleEntrypointHook")
			glog.Error("body:" + string(payload.Body) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleEntrypointHook(ctx, miCtx, payload)
}

// changes to init the request ctx with profile and request details
func (m OpenWrap) HandleBeforeValidationHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			m.metricEngine.RecordOpenWrapServerPanicStats(m.cfg.Server.HostName, "HandleBeforeValidationHook")
			request, err := json.Marshal(payload)
			if err != nil {
				glog.Error("request:" + string(request) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
				return
			}
			glog.Error("request:" + string(request) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleBeforeValidationHook(ctx, miCtx, payload)
}

func (m OpenWrap) HandleAllProcessedBidResponsesHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.AllProcessedBidResponsesPayload,
) (hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			m.metricEngine.RecordOpenWrapServerPanicStats(m.cfg.Server.HostName, "HandleAllProcessedBidResponsesHook")
			request, err := json.Marshal(payload)
			if err != nil {
				glog.Error("request:" + string(request) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
			}
			glog.Error("request:" + string(request) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleAllProcessedBidResponsesHook(ctx, miCtx, payload)
}

func (m OpenWrap) HandleBidderRequestHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BidderRequestPayload,
) (hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			m.metricEngine.RecordOpenWrapServerPanicStats(m.cfg.Server.HostName, "HandleBidderRequestHook")
			request, err := json.Marshal(payload)
			if err != nil {
				glog.Error("request:" + string(request) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
			}
			glog.Error("request:" + string(request) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleBidderRequestHook(ctx, miCtx, payload)
}

func (m OpenWrap) HandleAuctionResponseHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	defer func() {
		if r := recover(); r != nil {
			m.metricEngine.RecordOpenWrapServerPanicStats(m.cfg.Server.HostName, "HandleAuctionResponseHook")
			response, err := json.Marshal(payload)
			if err != nil {
				glog.Error("response:" + string(response) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
				return
			}
			glog.Error("response:" + string(response) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleAuctionResponseHook(ctx, miCtx, payload)
}

// HandleRawBidderResponseHook fetches rCtx and check for vast unwrapper flag to enable/disable vast unwrapping feature
func (m OpenWrap) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	defer func() {
		if r := recover(); r != nil {
			m.metricEngine.RecordOpenWrapServerPanicStats(m.cfg.Server.HostName, "HandleRawBidderResponseHook")
			response, err := json.Marshal(payload)
			if err != nil {
				glog.Error("response:" + string(response) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
				return
			}
			glog.Error("response:" + string(response) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleRawBidderResponseHook(miCtx, payload)
}

func (m OpenWrap) HandleExitpointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.ExitpointPaylaod,
) (hookstage.HookResult[hookstage.ExitpointPaylaod], error) {
	defer func() {
		if r := recover(); r != nil {
			m.metricEngine.RecordOpenWrapServerPanicStats(m.cfg.Server.HostName, "HandleExitpointHook")
			response, err := json.Marshal(payload)
			if err != nil {
				glog.Error("response:" + string(response) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
				return
			}
			glog.Error("response:" + string(response) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleExitpointHook(ctx, miCtx, payload)
}
