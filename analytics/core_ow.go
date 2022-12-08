package analytics

import (
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/openrtb/v17/openrtb3"
)

// RejectedBid contains oRTB Bid object with
// rejection reason and seat information
type RejectedBid struct {
	RejectionReason openrtb3.LossReason
	Bid             *openrtb2.Bid
	Seat            string
	BidderName      string
}
