package openrtb_ext

import (
	"github.com/prebid/prebid-server/openrtb_ext"
)

func FetchRTBBidders() error {
	// var rtbBidders []openrtb_ext.BidderName // list of rtb bidders from wrapper_partner
	openrtb_ext.SetAliasBidderName("magnite", "rtbbidder")
	return nil
}
