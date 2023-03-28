package analytics

import (
	"github.com/prebid/prebid-server/exchange/entities"
	"github.com/prebid/openrtb/v17/openrtb3"
)

// RejectedBid contains oRTB Bid object with
// rejection reason and seat information
type RejectedBid struct {
	RejectionReason openrtb3.NonBidStatusCode
	Bid             *entities.PbsOrtbBid
	Seat            string
}
