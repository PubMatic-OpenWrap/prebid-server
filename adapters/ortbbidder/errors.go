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

// newBadInputError returns the error of type bad-input
func newBadInputError(message string, args ...any) error {
	return &errortypes.BadServerResponse{
		Message: fmt.Sprintf(message, args...),
	}
}
