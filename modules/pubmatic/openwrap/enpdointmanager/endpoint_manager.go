package endpointmanager

import (
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/creativecache"
	ctvjson "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/json"
	ctvopenrtb "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/openrtb"
	ctvvast "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/vast"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type EndpointHookManager interface {
	HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.EntrypointPayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.RawAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.BeforeValidationRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.BidderRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.RawBidderResponsePayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.AuctionResponsePayload], moduleCtx hookstage.ModuleInvocationContext) bool
	HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.ExitpointPaylaod], moduleCtx hookstage.ModuleInvocationContext) bool
}

func NewEndpointManager(endpoint string, metricsEngine metrics.MetricsEngine, cache cache.Cache, creativeCache creativecache.Client) EndpointHookManager {
	switch endpoint {
	case models.EndpointORTB:
		return ctvopenrtb.NewCTVOpenRTB(metricsEngine)
	case models.EndpointJson:
		return ctvjson.NewCTVJSON(metricsEngine, creativeCache)
	case models.EndpointVAST:
		return ctvvast.NewCTVVAST(metricsEngine)
	default:
		return &NilEndpointManager{}
	}
}

type NilEndpointManager struct{}

func (n *NilEndpointManager) HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.EntrypointPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.RawAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.BeforeValidationRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.BidderRequestPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.RawBidderResponsePayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.AuctionResponsePayload], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}

func (n *NilEndpointManager) HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.ExitpointPaylaod], moduleCtx hookstage.ModuleInvocationContext) bool {
	return true
}
