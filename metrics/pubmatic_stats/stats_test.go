package pubmaticstats

import "testing"

func TestIncBidResponseByDealCountInPBS(t *testing.T) {
	IncBidResponseByDealCountInPBS("some_publisher_id", "some_profile_id", "some_alias_bidder", "some_dealid")
}

func TestIncPartnerTimeoutInPBS(t *testing.T) {
	IncPartnerTimeoutInPBS("some_publisher_id", "some_profile_id", "some_alias_bidder")
}
