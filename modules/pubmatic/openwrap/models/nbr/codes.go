package nbr

const (
	// 500+ Vendor-specific codes.
	// 5xx already in use by seat non bid. https://github.com/PubMatic-OpenWrap/prebid-openrtb/blob/main/openrtb3/non_bid_status_code.go#L53
	InvalidRequestWrapperExtension int = 601 + iota
	InvalidProfileID
	InvalidPublisherID
	InvalidRequestExt
	InvalidProfileConfiguration
	InvalidPlatform
	AllPartnerThrottled
	InvalidPriceGranularityConfig
	InvalidImpressionTagID
	InternalError
	AllSlotsDisabled
	ServerSidePartnerNotConfigured
	AllPartnersFiltered
)
