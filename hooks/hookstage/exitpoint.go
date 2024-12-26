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

// ExitPointPayload represents the payload data for the exit point hook stage.
// It contains the raw response body and the associated HTTP headers.
// Hooks are allowed to modify response using mutations.
type ExitPointPayload struct {
	RawResponse []byte
	Headers     http.Header
}
