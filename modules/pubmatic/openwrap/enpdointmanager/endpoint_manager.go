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
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/stage"
)

type EndpointHookManager interface {
	HandleEntrypointHook(rCtx *models.RequestCtx, payload stage.EntrypointPayload, moduleCtx stage.ModuleContext, result stage.EntrypointResult) (stage.EntrypointResult, bool)
	HandleRawAuctionHook(rCtx *models.RequestCtx, payload stage.RawAuctionPayload, moduleCtx stage.ModuleContext, result stage.RawAuctionResult) (stage.RawAuctionResult, bool)
	HandleBeforeValidationHook(rCtx *models.RequestCtx, payload stage.BeforeValidationPayload, moduleCtx stage.ModuleContext, result stage.BeforeValidationResult) (stage.BeforeValidationResult, bool)
	HandleProcessedAuctionHook(rCtx *models.RequestCtx, payload stage.ProcessedAuctionPayload, moduleCtx stage.ModuleContext, result stage.ProcessedAuctionResult) (stage.ProcessedAuctionResult, bool)
	HandleBidderRequestHook(rCtx *models.RequestCtx, payload stage.BidderRequestPayload, moduleCtx stage.ModuleContext, result stage.BidderRequestResult) (stage.BidderRequestResult, bool)
	HandleRawBidderResponseHook(rCtx *models.RequestCtx, payload stage.RawBidderResponsePayload, moduleCtx stage.ModuleContext, result stage.RawBidderResponseResult) (stage.RawBidderResponseResult, bool)
	HandleAllProcessedBidResponsesHook(rCtx *models.RequestCtx, payload stage.AllProcessedBidResponsesPayload, moduleCtx hookstage.ModuleInvocationContext, result stage.AllProcessedBidResponsesResult) (stage.AllProcessedBidResponsesResult, bool)
	HandleAuctionResponseHook(rCtx *models.RequestCtx, payload stage.AuctionResponsePayload, moduleCtx hookstage.ModuleInvocationContext, result stage.AuctionResponseResult) (stage.AuctionResponseResult, bool)
	HandleExitpointHook(rCtx *models.RequestCtx, payload stage.ExitpointPayload, moduleCtx stage.ModuleContext, result stage.ExitpointResult) (stage.ExitpointResult, bool)
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

func (n *NilEndpointManager) HandleEntrypointHook(rCtx *models.RequestCtx, payload stage.EntrypointPayload, moduleCtx stage.ModuleContext, result stage.EntrypointResult) (stage.EntrypointResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleRawAuctionHook(rCtx *models.RequestCtx, payload stage.RawAuctionPayload, moduleCtx stage.ModuleContext, result stage.RawAuctionResult) (stage.RawAuctionResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleBeforeValidationHook(rCtx *models.RequestCtx, payload stage.BeforeValidationPayload, moduleCtx stage.ModuleContext, result stage.BeforeValidationResult) (stage.BeforeValidationResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleProcessedAuctionHook(rCtx *models.RequestCtx, payload stage.ProcessedAuctionPayload, moduleCtx stage.ModuleContext, result stage.ProcessedAuctionResult) (stage.ProcessedAuctionResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleBidderRequestHook(rCtx *models.RequestCtx, payload stage.BidderRequestPayload, moduleCtx stage.ModuleContext, result stage.BidderRequestResult) (stage.BidderRequestResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleRawBidderResponseHook(rCtx *models.RequestCtx, payload stage.RawBidderResponsePayload, moduleCtx stage.ModuleContext, result stage.RawBidderResponseResult) (stage.RawBidderResponseResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleAllProcessedBidResponsesHook(rCtx *models.RequestCtx, payload stage.AllProcessedBidResponsesPayload, moduleCtx stage.ModuleContext, result stage.AllProcessedBidResponsesResult) (stage.AllProcessedBidResponsesResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleAuctionResponseHook(rCtx *models.RequestCtx, payload stage.AuctionResponsePayload, moduleCtx stage.ModuleContext, result stage.AuctionResponseResult) (stage.AuctionResponseResult, bool) {
	return result, true
}

func (n *NilEndpointManager) HandleExitpointHook(rCtx *models.RequestCtx, payload stage.ExitpointPayload, moduleCtx stage.ModuleContext, result stage.ExitpointResult) (stage.ExitpointResult, bool) {
	return result, true
}
