package ortbbidder

import (
	"fmt"

	"github.com/prebid/prebid-server/v2/errortypes"
)

func newBadServerResponseError(message string, args ...any) error {
	return &errortypes.BadServerResponse{
		Message: fmt.Sprintf(message, args...),
	}
}
