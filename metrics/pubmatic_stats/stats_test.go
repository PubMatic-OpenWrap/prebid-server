package pubmaticstats

import "testing"

func TestIncDealCount(t *testing.T) {
	IncDealBidCount("some_publisher_id", "some_profile_id", "some_alias_bidder", "some_dealid")
}
