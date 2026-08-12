package adapters

const (
	errMandatoryParameterMissingFormat     = `adapter:[%s] message:[missing_mandatory_param] key:[%v]`
	errInvalidS2SPartnerFormat             = `adapter:[%s] message:[invalid_s2s_adapter] slotkey:[%s]`
	errDefaultBidderParameterMissingFormat = `adapter:[%s] message:[default_bidder_missing_manadatory_param] param:[%s] applicable-key:[%s]`
)

var ignoreAppnexusKeys = map[string]bool{
	"generate_ad_pod_id": true,
	"invCode":            true,
	"inv_code":           true,
	"keywords":           true,
	"member":             true,
	"placementId":        true,
	"placement_id":       true,
	"private_sizes":      true,
	"reserve":            true,
	"usePaymentRule":     true,
	"use_pmt_rule":       true,
	"video":              true,
}

// Bidder Params
const (
	BidderParamApacdex_siteId       = "siteId"
	BidderParamApacdex_placementId  = "placementId"
	BidderParamApacdex_geo          = "geo"
	BidderParamApacdex_floorPrice   = "floorPrice"
	BidderParamBoldwinPlacementID   = "placementId"
	BidderParamBoldwinEndpointID    = "endpointId"
	BidderParamColossusTagID        = "TagID"
	BidderParamColossusgroupID      = "groupId"
	BidderNextmillenniumPlacementID = "placement_id"
	BidderNextmillenniumgroupID     = "group_id"
	BidderParamRiseOrg              = "org"
	BidderParamRisePublisherID      = "publisher_id"
	BidderKaroPlacementID           = "placementId"
	BidderKargoAdSlotID             = "adSlotID"
	BidderParamAidemPublisherId     = "publisherId"
	BidderParamAidemSiteId          = "siteId"
	BidderParamAidemPlacementId     = "placementId"
)
