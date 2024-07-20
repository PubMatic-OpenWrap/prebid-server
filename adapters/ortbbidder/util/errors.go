package util

import (
	"errors"
	"fmt"

	"github.com/prebid/prebid-server/v2/errortypes"
)

// list of constant errors
var (
	ErrImpMissing        error = errors.New("imp object not found in request")
	ErrNilBidderParamCfg error = errors.New("found nil bidderParamsConfig")
)

func NewBadInputError(message string, args ...any) error {
	return &errortypes.BadInput{
		Message: fmt.Sprintf(message, args...),
	}
}

func NewBadServerResponseError(message string, args ...any) error {
	return &errortypes.BadServerResponse{
		Message: fmt.Sprintf(message, args...),
	}
}

func NewWarning(message string, args ...any) error {
	return &errortypes.Warning{
		Message: fmt.Sprintf(message, args...),
	}
}
