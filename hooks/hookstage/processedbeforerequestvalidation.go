package hookstage

import (
	"context"

	"github.com/prebid/openrtb/v19/openrtb2"
)

// BeforeValidationRequest
type BeforeValidationRequest interface {
	HandleBeforeValidationHook(
		context.Context,
		ModuleInvocationContext,
		BeforeValidationRequestPayload,
	) (HookResult[BeforeValidationRequestPayload], error)
}

// ProcessedBeforeRequestValidationPayload consists of the openrtb2.BidRequest object.
// Hooks are allowed to modify openrtb2.BidRequest using mutations.
type BeforeValidationRequestPayload struct {
	BidRequest *openrtb2.BidRequest
}
