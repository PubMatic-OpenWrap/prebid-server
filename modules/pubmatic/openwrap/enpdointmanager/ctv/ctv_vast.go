package ctvendpointmanager

import (
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type CTVVAST struct {
	MetricsEngine metrics.MetricsEngine
}

func NewCTVVAST(metricsEngine metrics.MetricsEngine) *CTVVAST {
	return &CTVVAST{
		MetricsEngine: metricsEngine,
	}
}

func (co *CTVVAST) HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.EntrypointPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.EntrypointPayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawAuctionRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BidderRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawBidderResponsePayload]) (models.RequestCtx, hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload]) (models.RequestCtx, hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	return rCtx, result, nil
}

func (co *CTVVAST) HandleExitpointHook(payload hookstage.Exitpoint, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.Exitpoint]) (models.RequestCtx, hookstage.HookResult[hookstage.Exitpoint], error) {
	return rCtx, result, nil
}
