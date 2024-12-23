package hookstage

import (
	"context"
	"net/http"
)

type ExitPoint interface {
	HandleExitPointHook(
		context.Context,
		ModuleInvocationContext,
		ExitPointPayload,
	) (HookResult[ExitPointPayload], error)
}

// RawBidderResponsePayload consists of a list of adapters.TypedBid
// objects representing bids returned by a particular bidder.
// Hooks are allowed to modify bids using mutations.
type ExitPointPayload struct {
	RawResponse []byte
	Headers     http.Header
}
