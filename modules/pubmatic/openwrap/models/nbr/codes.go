package nbr

import "github.com/prebid/openrtb/v19/openrtb3"

// vendor specific NoBidReasons (500+)
const (
	LossBidLostToHigherBid        openrtb3.NoBidReason = 501 // Response Rejected - Lost to Higher Bid
	LossBidLostToDealBid          openrtb3.NoBidReason = 502 // Response Rejected - Lost to a Bid for a Deal
	RequestBlockedSlotNotMapped   openrtb3.NoBidReason = 503
	RequestBlockedPartnerThrottle openrtb3.NoBidReason = 504
)

// Openwrap module specific codes
const (
	InvalidRequestWrapperExtension openrtb3.NoBidReason = 601
	InvalidProfileID               openrtb3.NoBidReason = 602
	InvalidPublisherID             openrtb3.NoBidReason = 603
	InvalidRequestExt              openrtb3.NoBidReason = 604
	InvalidProfileConfiguration    openrtb3.NoBidReason = 605
	InvalidPlatform                openrtb3.NoBidReason = 606
	AllPartnerThrottled            openrtb3.NoBidReason = 607
	InvalidPriceGranularityConfig  openrtb3.NoBidReason = 608
	InvalidImpressionTagID         openrtb3.NoBidReason = 609
	InternalError                  openrtb3.NoBidReason = 610
	AllSlotsDisabled               openrtb3.NoBidReason = 611
	ServerSidePartnerNotConfigured openrtb3.NoBidReason = 612
	InvalidVideoRequest            openrtb3.NoBidReason = 613
	EmptySeatBid                   openrtb3.NoBidReason = 614
	InvalidAdpodConfig             openrtb3.NoBidReason = 615
	InvalidRedirectURL             openrtb3.NoBidReason = 616
	InvalidResponseFormat          openrtb3.NoBidReason = 617
	MissingOWRedirectURL           openrtb3.NoBidReason = 618
)
