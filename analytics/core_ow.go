package analytics

import (
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/exchange/entities"
)

// RejectedBid contains oRTB Bid object with
// rejection reason and seat information
type RejectedBid struct {
	RejectionReason openrtb3.NonBidStatusCode
	Bid             *entities.PbsOrtbBid
	Seat            string
}
