package auction

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
)

type podBid struct {
	openrtb2.Bid
	Nbr               *openrtb3.NoBidReason
	DealTierSatisfied bool
}

func newPodBid(b openrtb2.Bid, dealtierSatisfied bool) *podBid {
	return &podBid{
		Bid:               b,
		Nbr:               nil,
		DealTierSatisfied: dealtierSatisfied,
	}
}
