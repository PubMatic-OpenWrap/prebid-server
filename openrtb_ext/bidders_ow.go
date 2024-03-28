package openrtb_ext

import "strings"

const oRTBBidderPrefix = "owortb_"

func NormalizeBidderNameWithORTBBidder(name string) (BidderName, bool) {
	if normalized, exists := NormalizeBidderName(name); exists {
		return normalized, true
	}
	if strings.HasPrefix(name, "owortb_") {
		return BidderName(name), true
	}
	return BidderName(name), false
}

func IsORTBBidder(bidder string) bool {
	return strings.HasPrefix(bidder, oRTBBidderPrefix)
}
