package ortbbidder

import (
	"errors"

	"github.com/prebid/prebid-server/v2/errortypes"
)

// list of constant errors
var (
	errImpMissing        error = errors.New("imp object not found in request")
	errNilBidderParamCfg error = errors.New("found nil bidderParamsConfig")
)

// newBadInputError returns the error of type bad-input
func newBadInputError(message string) error {
	return &errortypes.BadInput{
		Message: message,
	}
}
