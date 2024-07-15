package ortbbidder

import (
	"errors"
	"fmt"

	"github.com/prebid/prebid-server/v2/errortypes"
)

// list of constant errors
var (
	errImpMissing        error = errors.New("imp object not found in request")
	errNilBidderParamCfg error = errors.New("found nil bidderParamsConfig")
)

func newBadInputError(message string, args ...any) error {
	return &errortypes.BadServerResponse{
		Message: fmt.Sprintf(message, args...),
	}
}

func newBadServerResponseError(message string, args ...any) error {
	return &errortypes.BadServerResponse{
		Message: fmt.Sprintf(message, args...),
	}
}

func newWarning(message string, args ...any) error {
	return &errortypes.Warning{
		Message: fmt.Sprintf(message, args...),
	}
}
