package floors

import (
	"github.com/prebid/openrtb/v20/openrtb2"
)

func RequestHasFloors(bidRequest *openrtb2.BidRequest) bool {
	for i := range bidRequest.Imp {
		if bidRequest.Imp[i].BidFloor > 0 {
			return true
		}
	}
	return false
}
