package nbr

const (
	// Refer below link for standard codes.
	// https://github.com/InteractiveAdvertisingBureau/openrtb/blob/2c3bf2bb2bc81ce0b5260f2e82c59938ea05b74a/extensions/community_extensions/seat-non-bid.md#list-non-bid-status-codes

	//  Internal Technical Error
	InternalError int = 1
	// Invalid Request
	InvalidRequest int = 2

	// 500+ Vendor-specific codes.
	// 5xx already in use by seat non bid. https://github.com/PubMatic-OpenWrap/prebid-openrtb/blob/main/openrtb3/non_bid_status_code.go#L53
	InvalidRequestWrapperExtension int = 600 + iota
	InvalidPublisherID
	InvalidProfileID
	InvalidProfileConfiguration
	AllPartnerThrottled
	InvalidPriceGranularityConfig
	InvalidImpressionTagID
	ServerSidePartnerNotConfigured
	AllSlotsDisabled
	InvalidVideoRequest
	InvalidPlatform
)
