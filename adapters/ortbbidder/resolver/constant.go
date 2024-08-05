package resolver

type parameter string

func (s parameter) String() string {
	return string(s)
}

// constant parameters defined in response-param.json file
const (
	bidType                  parameter = "bidType"
	bidDealPriority          parameter = "bidDealPriority"
	bidVideo                 parameter = "bidVideo"
	bidVideoDuration         parameter = "bidVideoDuration"
	bidVideoPrimaryCategory  parameter = "bidVideoPrimaryCategory"
	fledgeAuctionConfig      parameter = "fledgeAuctionConfig"
	bidMeta                  parameter = "bidMeta"
	bidMetaAdvertiserDomains parameter = "bidMetaAdvertiserDomains"
	bidMetaAdvertiserId      parameter = "bidMetaAdvertiserId"
	bidMetaAdvertiserName    parameter = "bidMetaAdvertiserName"
	bidMetaAgencyId          parameter = "bidMetaAgencyId"
	bidMetaAgencyName        parameter = "bidMetaAgencyName"
	bidMetaBrandId           parameter = "bidMetaBrandId"
	bidMetaBrandName         parameter = "bidMetaBrandName"
	bidMetaDChain            parameter = "bidMetaDchain"
	bidMetaDemandSource      parameter = "bidMetaDemandSource"
	bidMetaMediaType         parameter = "bidMetaMediaType"
	bidMetaNetworkId         parameter = "bidMetaNetworkId"
	bidMetaNetworkName       parameter = "bidMetaNetworkName"
	bidMetaPrimaryCatId      parameter = "bidMetaPrimaryCatId"
	bidMetaRendererName      parameter = "bidMetaRendererName"
	bidMetaRendererVersion   parameter = "bidMetaRendererVersion"
	bidMetaRenderedData      parameter = "bidMetaRendererData"
	bidMetaRenderedUrl       parameter = "bidMetaRendererUrl"
	bidMetaSecondaryCatId    parameter = "bidMetaSecondaryCatIds"
)

// constants used to look up for standard ortb fields in bidResponse
const (
	ortbFieldMtype    = "mtype"
	ortbFieldDuration = "dur"
	ortbFieldCurrency = "cur"
	ortbFieldAdM      = "adm"
	ortbFieldCategory = "cat"
	ortbFieldImpId    = "impid"
	ortbFieldBidder   = "bidder"
	ortbFieldAdapter  = "adapter"
	ortbFieldConfig   = "config"
)

// constants used to set keys in the BidderResponse map
const (
	bidVideoPrimaryCategoryKey  = "primary_category"
	bidVideoDurationKey         = "duration"
	bidVideoKey                 = "BidVideo"
	bidTypeKey                  = "BidType"
	currencyKey                 = "Currency"
	fledgeAuctionConfigKey      = "FledgeAuctionConfigs"
	bidDealPriorityKey          = "DealPriority"
	bidMetaKey                  = "BidMeta"
	bidMetaAdvertiserDomainsKey = "advertiserDomains"
	bidMetaAdvertiserIdKey      = "advertiserId"
	bidMetaAdvertiserNameKey    = "advertiserName"
	bidMetaAgencyIdKey          = "agencyId"
	bidMetaAgencyNameKey        = "agencyName"
	bidMetaBrandIdKey           = "brandId"
	bidMetaBrandNameKey         = "brandName"
	bidMetaDChainKey            = "dchain"
	bidMetaDemandSourceKey      = "demandSource"
	bidMetaMediaTypeKey         = "mediaType"
	bidMetaNetworkIdKey         = "networkId"
	bidMetaNetworkNameKey       = "networkName"
	bidMetaPrimaryCatIdKey      = "primaryCatId"
	bidMetaRendererNameKey      = "rendererName"
	bidMetaRendererVersionKey   = "rendererVersion"
	bidMetaRenderedDataKey      = "rendererData"
	bidMetaRenderedUrlKey       = "rendererUrl"
	bidMetaSecondaryCatIdKey    = "secondaryCatIds"
)
