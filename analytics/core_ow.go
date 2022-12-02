package analytics

import (
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/mxmCherry/openrtb/v16/openrtb3"
)

// RejectedBid contains oRTB Bid object with
// rejection reason and seat information
type RejectedBid struct {
	RejectionReason openrtb3.LossReason
	Bid             *openrtb2.Bid
	Seat            string
	BidderName      string
}
