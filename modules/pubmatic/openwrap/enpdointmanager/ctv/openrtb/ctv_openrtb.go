package ctvopenrtb

import (
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type CTVOpenRTB struct {
	metricsEngine metrics.MetricsEngine
	cache         cache.Cache
}

func NewCTVOpenRTB(metricsEngine metrics.MetricsEngine, cache cache.Cache) *CTVOpenRTB {
	return &CTVOpenRTB{
		metricsEngine: metricsEngine,
		cache:         cache,
	}
}

// EntrypointHook
func (co *CTVOpenRTB) HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.EntrypointPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.EntrypointPayload], error) {
	return rCtx, result, nil
}

// RawAuctionHook
func (co *CTVOpenRTB) HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	return rCtx, result, nil
}

// BeforeValidationHook
func (co *CTVOpenRTB) HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	co.metricsEngine.RecordCTVHTTPMethodRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Method)
	return rCtx, result, nil
}

// ProcessedAuctionHook
func (co *CTVOpenRTB) HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	return rCtx, result, nil
}

// BidderRequestHook
func (co *CTVOpenRTB) HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BidderRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	return rCtx, result, nil
}

// RawBidderResponseHook
func (co *CTVOpenRTB) HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawBidderResponsePayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	return rCtx, result, nil
}

// AllProcessedBidResponsesHook
func (co *CTVOpenRTB) HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error) {
	return rCtx, result, nil
}

// AuctionResponseHook
func (co *CTVOpenRTB) HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	return rCtx, result, nil
}

// ExitpointHook
func (co *CTVOpenRTB) HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ExitpointPaylaod], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ExitpointPaylaod], error) {
	return rCtx, result, nil
}
