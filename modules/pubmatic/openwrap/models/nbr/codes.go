package nbr

import "github.com/prebid/openrtb/v20/openrtb3"

// vendor specific NoBidReasons (500+)
const (
	LossBidLostToHigherBid             openrtb3.NoBidReason = 501 // Response Rejected - Lost to Higher Bid
	LossBidLostToDealBid               openrtb3.NoBidReason = 502 // Response Rejected - Lost to a Bid for a Deal
	RequestBlockedSlotNotMapped        openrtb3.NoBidReason = 503
	RequestBlockedPartnerThrottle      openrtb3.NoBidReason = 504
	RequestBlockedPartnerFiltered      openrtb3.NoBidReason = 505
	LossBidLostInVastUnwrap            openrtb3.NoBidReason = 506
	LossBidLostInVastVersionValidation openrtb3.NoBidReason = 507
	RequestBlockedGeoFiltered          openrtb3.NoBidReason = 508
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
	AllPartnersFiltered            openrtb3.NoBidReason = 613
	InvalidVideoRequest            openrtb3.NoBidReason = 614
	EmptySeatBid                   openrtb3.NoBidReason = 615
	InvalidAdpodConfig             openrtb3.NoBidReason = 616
	InvalidRedirectURL             openrtb3.NoBidReason = 617
	InvalidResponseFormat          openrtb3.NoBidReason = 618
	MissingOWRedirectURL           openrtb3.NoBidReason = 619
	ResponseRejectedDSA            openrtb3.NoBidReason = 620 // Response Rejected - DSA
	ResponseRejectedMissingParam   openrtb3.NoBidReason = 621 // Response rejected due to missing required parameter
)
