package stage

import "github.com/prebid/prebid-server/v3/hooks/hookstage"

type (
	EntrypointPayload = hookstage.EntrypointPayload
	EntrypointResult  = hookstage.HookResult[hookstage.EntrypointPayload]

	RawAuctionPayload = hookstage.RawAuctionRequestPayload
	RawAuctionResult  = hookstage.HookResult[hookstage.RawAuctionRequestPayload]

	BeforeValidationPayload = hookstage.BeforeValidationRequestPayload
	BeforeValidationResult  = hookstage.HookResult[hookstage.BeforeValidationRequestPayload]

	ProcessedAuctionPayload = hookstage.ProcessedAuctionRequestPayload
	ProcessedAuctionResult  = hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]

	BidderRequestPayload = hookstage.BidderRequestPayload
	BidderRequestResult  = hookstage.HookResult[hookstage.BidderRequestPayload]

	RawBidderResponsePayload = hookstage.RawBidderResponsePayload
	RawBidderResponseResult  = hookstage.HookResult[hookstage.RawBidderResponsePayload]

	AllProcessedBidResponsesPayload = hookstage.AllProcessedBidResponsesPayload
	AllProcessedBidResponsesResult  = hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]

	AuctionResponsePayload = hookstage.AuctionResponsePayload
	AuctionResponseResult  = hookstage.HookResult[hookstage.AuctionResponsePayload]

	ExitpointPayload = hookstage.ExitpointPayload
	ExitpointResult  = hookstage.HookResult[hookstage.ExitpointPayload]

	ModuleContext = hookstage.ModuleInvocationContext
)
