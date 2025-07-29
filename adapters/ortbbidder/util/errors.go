package util

import (
	"errors"
	"fmt"

	"github.com/prebid/prebid-server/v3/errortypes"
)

// list of constant errors
var (
	ErrImpMissing        error = errors.New("imp object not found in request")
	ErrNilBidderParamCfg error = errors.New("found nil bidderParamsConfig")
)

func NewBadInputError(message string, args ...any) error {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
	} else {
		msg = message
	}
	return &errortypes.BadInput{
		Message: msg,
	}
}

func NewBadServerResponseError(message string, args ...any) error {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
	} else {
		msg = message
	}
	return &errortypes.BadServerResponse{
		Message: msg,
	}
}

func NewWarning(message string, args ...any) error {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
	} else {
		msg = message
	}
	return &errortypes.Warning{
		Message: msg,
	}
}
