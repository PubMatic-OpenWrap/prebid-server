package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

// bidMetaResolver retrieves the meta object of the bid using the bidder param location.
// The determined bidMeta is subsequently assigned to adapterresponse.typedbid.bidmeta
type bidMetaResolver struct {
	defaultValueResolver
}

func (b *bidMetaResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateBidMeta(value)
}

func validateBidMeta(value any) (map[string]any, bool) {
	inputMeta, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}

	outputMeta := map[string]any{}
	for metaKey, metaValue := range inputMeta {
		switch metaKey {
		case bidMetaAdvertiserDomainsKey, bidMetaSecondaryCatIdKey:
			metaValue, ok = validateDataTypeSlice[string](metaValue)

		case bidMetaAdvertiserIdKey, bidMetaAgencyIdKey, bidMetaBrandIdKey, bidMetaNetworkIdKey:
			metaValue, ok = validateInt(metaValue)

		case bidMetaDChainKey, bidMetaRenderedDataKey: // TODO - verify this ???
			metaValue, ok = validateJSONRawMessage(metaValue)

		default:
			metaValue, ok = validateString(metaValue)
		}
		if !ok {
			continue
		}
		outputMeta[metaKey] = metaValue
	}

	return outputMeta, len(outputMeta) != 0
}

func (b *bidMetaResolver) setValue(adapterBid map[string]any, value any) bool {
	adapterBid[bidMetaKey] = value
	return true
}

// bidMetaAdvDomainsResolver retrieves the advertiserDomains of the bid using the bidder param location.
// The determined advertiserDomains is subsequently assigned to adapterresponse.typedbid.bidmeta.advertiserDomains
type bidMetaAdvDomainsResolver struct {
	defaultValueResolver
}

func (b *bidMetaAdvDomainsResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, ok := util.GetValueFromLocation(responseNode, path)
	if !ok {
		return nil, false
	}
	return validateDataTypeSlice[string](value)
}

func (b *bidMetaAdvDomainsResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaAdvertiserDomainsKey, value)
}

// bidMetaAdvIDResolver retrieves the advertiserId of the bid using the bidder param location.
// The determined advertiserId is subsequently assigned to adapterresponse.typedbid.bidmeta.advertiserId
type bidMetaAdvIDResolver struct {
	defaultValueResolver
}

func (b *bidMetaAdvIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateInt(value)
}

func (b *bidMetaAdvIDResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaAdvertiserIdKey, value)
}

// bidMetaAdvNameResolver retrieves the advertiserName of the bid using the bidder param location.
// The determined advertiserName is subsequently assigned to adapterresponse.typedbid.bidmeta.AdvertiserName
type bidMetaAdvNameResolver struct {
	defaultValueResolver
}

func (b *bidMetaAdvNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaAdvNameResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaAdvertiserNameKey, value)
}

// bidMetaAgencyIDResolver retrieves the AgencyID of the bid using the bidder param location.
// The determined AgencyID is subsequently assigned to adapterresponse.typedbid.bidmeta.AgencyID
type bidMetaAgencyIDResolver struct {
	defaultValueResolver
}

func (b *bidMetaAgencyIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateInt(value)
}

func (b *bidMetaAgencyIDResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaAgencyIdKey, value)
}

// bidMetaAgencyNameResolver retrieves the AgencyName of the bid using the bidder param location.
// The determined AgencyName is subsequently assigned to adapterresponse.typedbid.bidmeta.AgencyName
type bidMetaAgencyNameResolver struct {
	defaultValueResolver
}

func (b *bidMetaAgencyNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaAgencyNameResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaAgencyNameKey, value)
}

// bidMetaBrandIDResolver retrieves the BrandID of the bid using the bidder param location.
// The determined BrandID is subsequently assigned to adapterresponse.typedbid.bidmeta.BrandID
type bidMetaBrandIDResolver struct {
	defaultValueResolver
}

func (b *bidMetaBrandIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateInt(value)
}

func (b *bidMetaBrandIDResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaBrandIdKey, value)
}

// bidMetaBrandNameResolver retrieves the BrandName of the bid using the bidder param location.
// The determined BrandName is subsequently assigned to adapterresponse.typedbid.bidmeta.BrandName
type bidMetaBrandNameResolver struct {
	defaultValueResolver
}

func (b *bidMetaBrandNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaBrandNameResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaBrandNameKey, value)
}

// bidMetaDChainResolver retrieves the Dchain of the bid using the bidder param location.
// The determined Dchain is subsequently assigned to adapterresponse.typedbid.bidmeta.DChain
type bidMetaDChainResolver struct {
	defaultValueResolver
}

func (b *bidMetaDChainResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateJSONRawMessage(value)
}

func (b *bidMetaDChainResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaDChainKey, value)
}

// bidMetaDemandSourceResolver retrieves the DemandSource of the bid using the bidder param location.
// The determined DemandSource is subsequently assigned to adapterresponse.typedbid.bidmeta.DemandSource
type bidMetaDemandSourceResolver struct {
	defaultValueResolver
}

func (b *bidMetaDemandSourceResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaDemandSourceResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaDemandSourceKey, value)
}

// bidMetaMediaTypeResolver retrieves the MediaType of the bid using the bidder param location.
// The determined MediaType is subsequently assigned to adapterresponse.typedbid.bidmeta.MediaType
type bidMetaMediaTypeResolver struct {
	defaultValueResolver
}

func (b *bidMetaMediaTypeResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaMediaTypeResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaMediaTypeKey, value)
}

// bidMetaNetworkIDResolver retrieves the NetworkID of the bid using the bidder param location.
// The determined NetworkID is subsequently assigned to adapterresponse.typedbid.bidmeta.NetworkID
type bidMetaNetworkIDResolver struct {
	defaultValueResolver
}

func (b *bidMetaNetworkIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateInt(value)
}

func (b *bidMetaNetworkIDResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaNetworkIdKey, value)
}

// bidMetaNetworkNameResolver retrieves the NetworkName of the bid using the bidder param location.
// The determined NetworkName is subsequently assigned to adapterresponse.typedbid.bidmeta.NetworkName
type bidMetaNetworkNameResolver struct {
	defaultValueResolver
}

func (b *bidMetaNetworkNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaNetworkNameResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaNetworkNameKey, value)
}

// bidMetaPrimaryCategoryIDResolver retrieves the PrimaryCategory of the bid using the bidder param location.
// The determined PrimaryCategory is subsequently assigned to adapterresponse.typedbid.bidmeta.PrimaryCategory
type bidMetaPrimaryCategoryIDResolver struct {
	defaultValueResolver
}

func (b *bidMetaPrimaryCategoryIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaPrimaryCategoryIDResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaPrimaryCatIdKey, value)
}

// bidMetaRendererNameResolver retrieves the RendererName of the bid using the bidder param location.
// The determined RendererName is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererName
type bidMetaRendererNameResolver struct {
	defaultValueResolver
}

func (b *bidMetaRendererNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaRendererNameResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaRendererNameKey, value)
}

// bidMetaRendererVersionResolver retrieves the RendererVersion of the bid using the bidder param location.
// The determined RendererVersion is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererVersion
type bidMetaRendererVersionResolver struct {
	defaultValueResolver
}

func (b *bidMetaRendererVersionResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaRendererVersionResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaRendererVersionKey, value)
}

// bidMetaRendererDataResolver retrieves the RendererData of the bid using the bidder param location.
// The determined RendererData is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererData
type bidMetaRendererDataResolver struct {
	defaultValueResolver
}

func (b *bidMetaRendererDataResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateJSONRawMessage(value)
}

func (b *bidMetaRendererDataResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaRenderedDataKey, value)
}

// bidMetaRendererUrlResolver retrieves the RendererUrl of the bid using the bidder param location.
// The determined RendererUrl is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererUrl
type bidMetaRendererUrlResolver struct {
	defaultValueResolver
}

func (b *bidMetaRendererUrlResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateString(value)
}

func (b *bidMetaRendererUrlResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaRenderedUrlKey, value)
}

// bidMetaSecondaryCategoryIDsResolver retrieves the secondary-category ids of the bid using the bidder param location.
// The determined secondary-category id are subsequently assigned to adapterresponse.typedbid.bidmeta.secondaryCatIds
type bidMetaSecondaryCategoryIDsResolver struct {
	defaultValueResolver
}

func (b *bidMetaSecondaryCategoryIDsResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateDataTypeSlice[string](value)
}

func (b *bidMetaSecondaryCategoryIDsResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidMeta(adapterBid, bidMetaSecondaryCatIdKey, value)
}

// setKeyValueInBidMeta sets the key and value in bidMeta object
// it creates the bidMeta object if required.
func setKeyValueInBidMeta(adapterBid map[string]any, key string, value any) bool {
	meta, found := adapterBid[bidMetaKey]
	if !found {
		meta = map[string]any{}
		adapterBid[bidMetaKey] = meta
	}
	typedMeta, ok := meta.(map[string]any)
	if !ok || typedMeta == nil {
		return false
	}
	typedMeta[key] = value
	return true
}
