package endpointmanager

import (
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/creativecache"
	ctvjson "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/json"
	ctvopenrtb "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/openrtb"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type EndpointHookManager interface {
	HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.EntrypointPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.EntrypointPayload], error)
	HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawAuctionRequestPayload], error)
	HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error)
	HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error)
	HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BidderRequestPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BidderRequestPayload], error)
	HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawBidderResponsePayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawBidderResponsePayload], error)
	HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error)
	HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AuctionResponsePayload], error)
	HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ExitpointPaylaod], moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ExitpointPaylaod], error)
}

// type MutationManager interface {
// 	EntrypointMutation(ep hookstage.EntrypointPayload, rCtx models.RequestCtx) (hookstage.EntrypointPayload, error)
// 	RawAuctionMutation(rarp hookstage.RawAuctionRequestPayload, rCtx models.RequestCtx) (hookstage.RawAuctionRequestPayload, error)
// 	BeforeValidationMutation(bvrp hookstage.BeforeValidationRequestPayload, rCtx models.RequestCtx) (hookstage.BeforeValidationRequestPayload, error)
// 	ProcessedAuctionMutation(parp hookstage.ProcessedAuctionRequestPayload, rCtx models.RequestCtx) (hookstage.ProcessedAuctionRequestPayload, error)
// 	BidderRequestMutation(brp hookstage.BidderRequestPayload, rCtx models.RequestCtx) (hookstage.BidderRequestPayload, error)
// 	RawBidderResponseMutation(rbrp hookstage.RawBidderResponsePayload, rCtx models.RequestCtx) (hookstage.RawBidderResponsePayload, error)
// 	AllProcessedBidResponsesMutation(aprp hookstage.AllProcessedBidResponsesPayload, rCtx models.RequestCtx) (hookstage.AllProcessedBidResponsesPayload, error)
// 	AuctionResponseMutation(arp hookstage.AuctionResponsePayload, rCtx models.RequestCtx) (hookstage.AuctionResponsePayload, error)
// 	ExitpointMutation(ep hookstage.Exitpoint, rCtx models.RequestCtx) (hookstage.Exitpoint, error)
// }

func NewEndpointManager(endpoint string, metricsEngine metrics.MetricsEngine, cache cache.Cache, creativeCache creativecache.Client) EndpointHookManager {
	switch endpoint {
	case models.EndpointORTB:
		return ctvopenrtb.NewCTVOpenRTB(metricsEngine, cache)
	case models.EndpointJson:
		return ctvjson.NewCTVJSON(metricsEngine, cache, creativeCache)
	default:
		return nil
	}
}
