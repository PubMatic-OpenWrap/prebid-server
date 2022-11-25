package analytics

import "github.com/mxmCherry/openrtb/v16/openrtb2"

// RejectedBid contains oRTB Bid object with
// rejection reason and seat information
type RejectedBid struct {
	RejectionReason int
	Bid             *openrtb2.Bid
	Seat            string
}
