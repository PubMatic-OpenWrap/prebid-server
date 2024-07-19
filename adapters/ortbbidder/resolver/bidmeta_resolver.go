package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// bidMetaResolver retrieves the meta object of the bid using the bidder param location.
// The determined bidMeta is subsequently assigned to adapterresponse.typedbid.bidmeta
type bidMetaResolver struct {
	paramResolver
}

func (b *bidMetaResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta]", path)
	}
	bidMeta, err := validateBidMeta(value)
	if err != nil {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta]", path)
	}
	return bidMeta, nil
}

func validateBidMeta(value any) (any, error) {
	bidMetaBytes, err := jsonutil.Marshal(value)
	if err != nil {
		return nil, err
	}

	var bidMeta openrtb_ext.ExtBidPrebidMeta
	err = jsonutil.UnmarshalValid(bidMetaBytes, &bidMeta)
	if err != nil {
		return nil, err
	}

	var bidMetaMap map[string]any
	err = jsonutil.UnmarshalValid(bidMetaBytes, &bidMetaMap)
	if err != nil {
		return nil, err
	}
	return bidMetaMap, nil
}

func (b *bidMetaResolver) setValue(adapterBid map[string]any, value any) error {
	adapterBid[bidMetaKey] = value
	return nil
}

// bidMetaAdvDomainsResolver retrieves the advertiserDomains of the bid using the bidder param location.
// The determined advertiserDomains is subsequently assigned to adapterresponse.typedbid.bidmeta.advertiserDomains
type bidMetaAdvDomainsResolver struct {
	paramResolver
}

func (b *bidMetaAdvDomainsResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, ok := util.GetValueFromLocation(responseNode, path)
	if !ok {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.advertiserDomains]", path)
	}

	adomains, ok := validateDataTypeSlice[string](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.advertiserDomains]", path)
	}
	return adomains, nil
}

func (b *bidMetaAdvDomainsResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaAdvertiserDomainsKey, value)
}

// bidMetaAdvIDResolver retrieves the advertiserId of the bid using the bidder param location.
// The determined advertiserId is subsequently assigned to adapterresponse.typedbid.bidmeta.advertiserId
type bidMetaAdvIDResolver struct {
	paramResolver
}

func (b *bidMetaAdvIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.advertiserId]", path)
	}
	advId, ok := validateNumber[int](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.advertiserId]", path)
	}
	return advId, nil
}

func (b *bidMetaAdvIDResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaAdvertiserIdKey, value)
}

// bidMetaAdvNameResolver retrieves the advertiserName of the bid using the bidder param location.
// The determined advertiserName is subsequently assigned to adapterresponse.typedbid.bidmeta.AdvertiserName
type bidMetaAdvNameResolver struct {
	paramResolver
}

func (b *bidMetaAdvNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.advertiserName]", path)
	}
	advName, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.advertiserName]", path)

	}
	return advName, nil
}

func (b *bidMetaAdvNameResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaAdvertiserNameKey, value)
}

// bidMetaAgencyIDResolver retrieves the AgencyID of the bid using the bidder param location.
// The determined AgencyID is subsequently assigned to adapterresponse.typedbid.bidmeta.AgencyID
type bidMetaAgencyIDResolver struct {
	paramResolver
}

func (b *bidMetaAgencyIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.agencyID]", path)
	}
	agencyId, ok := validateNumber[int](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.agencyID]", path)

	}
	return agencyId, nil
}

func (b *bidMetaAgencyIDResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaAgencyIdKey, value)
}

// bidMetaAgencyNameResolver retrieves the AgencyName of the bid using the bidder param location.
// The determined AgencyName is subsequently assigned to adapterresponse.typedbid.bidmeta.AgencyName
type bidMetaAgencyNameResolver struct {
	paramResolver
}

func (b *bidMetaAgencyNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.agencyName]", path)
	}
	agencyName, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.agencyName]", path)

	}
	return agencyName, nil
}

func (b *bidMetaAgencyNameResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaAgencyNameKey, value)
}

// bidMetaBrandIDResolver retrieves the BrandID of the bid using the bidder param location.
// The determined BrandID is subsequently assigned to adapterresponse.typedbid.bidmeta.BrandID
type bidMetaBrandIDResolver struct {
	paramResolver
}

func (b *bidMetaBrandIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.brandID]", path)
	}
	brandId, ok := validateNumber[int](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.brandID]", path)

	}
	return brandId, nil
}

func (b *bidMetaBrandIDResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaBrandIdKey, value)
}

// bidMetaBrandNameResolver retrieves the BrandName of the bid using the bidder param location.
// The determined BrandName is subsequently assigned to adapterresponse.typedbid.bidmeta.BrandName
type bidMetaBrandNameResolver struct {
	paramResolver
}

func (b *bidMetaBrandNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.brandName]", path)
	}
	brandName, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.brandName]", path)

	}
	return brandName, nil
}

func (b *bidMetaBrandNameResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaBrandNameKey, value)
}

// bidMetaDChainResolver retrieves the Dchain of the bid using the bidder param location.
// The determined Dchain is subsequently assigned to adapterresponse.typedbid.bidmeta.DChain
type bidMetaDChainResolver struct {
	paramResolver
}

func (b *bidMetaDChainResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.dchain]", path)
	}
	dChain, ok := validateMap(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.dchain]", path)

	}
	return dChain, nil
}

func (b *bidMetaDChainResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaDChainKey, value)
}

// bidMetaDemandSourceResolver retrieves the DemandSource of the bid using the bidder param location.
// The determined DemandSource is subsequently assigned to adapterresponse.typedbid.bidmeta.DemandSource
type bidMetaDemandSourceResolver struct {
	paramResolver
}

func (b *bidMetaDemandSourceResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.demandSource]", path)
	}
	demandSource, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.demandSource]", path)

	}
	return demandSource, nil
}

func (b *bidMetaDemandSourceResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaDemandSourceKey, value)
}

// bidMetaMediaTypeResolver retrieves the MediaType of the bid using the bidder param location.
// The determined MediaType is subsequently assigned to adapterresponse.typedbid.bidmeta.MediaType
type bidMetaMediaTypeResolver struct {
	paramResolver
}

func (b *bidMetaMediaTypeResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.mediaType]", path)
	}
	mediaType, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.mediaType]", path)

	}
	return mediaType, nil
}

func (b *bidMetaMediaTypeResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaMediaTypeKey, value)
}

// bidMetaNetworkIDResolver retrieves the NetworkID of the bid using the bidder param location.
// The determined NetworkID is subsequently assigned to adapterresponse.typedbid.bidmeta.NetworkID
type bidMetaNetworkIDResolver struct {
	paramResolver
}

func (b *bidMetaNetworkIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.networkId]", path)
	}
	networkId, ok := validateNumber[int](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.networkId]", path)

	}
	return networkId, nil
}

func (b *bidMetaNetworkIDResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaNetworkIdKey, value)
}

// bidMetaNetworkNameResolver retrieves the NetworkName of the bid using the bidder param location.
// The determined NetworkName is subsequently assigned to adapterresponse.typedbid.bidmeta.NetworkName
type bidMetaNetworkNameResolver struct {
	paramResolver
}

func (b *bidMetaNetworkNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.networkName]", path)
	}
	networkName, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.networkName]", path)

	}
	return networkName, nil
}

func (b *bidMetaNetworkNameResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaNetworkNameKey, value)
}

// bidMetaPrimaryCategoryIDResolver retrieves the PrimaryCategory of the bid using the bidder param location.
// The determined PrimaryCategory is subsequently assigned to adapterresponse.typedbid.bidmeta.PrimaryCategory
type bidMetaPrimaryCategoryIDResolver struct {
	paramResolver
}

func (b *bidMetaPrimaryCategoryIDResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.primaryCategory]", path)
	}
	categoryId, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.primaryCategory]", path)

	}
	return categoryId, nil
}

func (b *bidMetaPrimaryCategoryIDResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaPrimaryCatIdKey, value)
}

// bidMetaRendererNameResolver retrieves the RendererName of the bid using the bidder param location.
// The determined RendererName is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererName
type bidMetaRendererNameResolver struct {
	paramResolver
}

func (b *bidMetaRendererNameResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererName]", path)
	}
	rendererName, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererName]", path)

	}
	return rendererName, nil
}

func (b *bidMetaRendererNameResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaRendererNameKey, value)
}

// bidMetaRendererVersionResolver retrieves the RendererVersion of the bid using the bidder param location.
// The determined RendererVersion is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererVersion
type bidMetaRendererVersionResolver struct {
	paramResolver
}

func (b *bidMetaRendererVersionResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererVersion]", path)
	}
	rendererVersion, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererVersion]", path)

	}
	return rendererVersion, nil
}

func (b *bidMetaRendererVersionResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaRendererVersionKey, value)
}

// bidMetaRendererDataResolver retrieves the RendererData of the bid using the bidder param location.
// The determined RendererData is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererData
type bidMetaRendererDataResolver struct {
	paramResolver
}

func (b *bidMetaRendererDataResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererData]", path)
	}
	rendererData, ok := validateMap(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererData]", path)

	}
	return rendererData, nil
}

func (b *bidMetaRendererDataResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaRenderedDataKey, value)
}

// bidMetaRendererUrlResolver retrieves the RendererUrl of the bid using the bidder param location.
// The determined RendererUrl is subsequently assigned to adapterresponse.typedbid.bidmeta.RendererUrl
type bidMetaRendererUrlResolver struct {
	paramResolver
}

func (b *bidMetaRendererUrlResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererUrl]", path)
	}
	rendererUrl, ok := validateString(value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.rendererUrl]", path)

	}
	return rendererUrl, nil
}

func (b *bidMetaRendererUrlResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaRenderedUrlKey, value)
}

// bidMetaSecondaryCategoryIDsResolver retrieves the secondary-category ids of the bid using the bidder param location.
// The determined secondary-category id are subsequently assigned to adapterresponse.typedbid.bidmeta.secondaryCatIds
type bidMetaSecondaryCategoryIDsResolver struct {
	paramResolver
}

func (b *bidMetaSecondaryCategoryIDsResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.meta.secondaryCategoryIds]", path)
	}
	secondaryCategories, ok := validateDataTypeSlice[string](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.meta.secondaryCategoryIds]", path)

	}
	return secondaryCategories, nil
}

func (b *bidMetaSecondaryCategoryIDsResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidMeta(adapterBid, bidMetaSecondaryCatIdKey, value)
}

// setKeyValueInBidMeta sets the key and value in bidMeta object
// it creates the bidMeta object if required.
func setKeyValueInBidMeta(adapterBid map[string]any, key string, value any) error {
	meta, found := adapterBid[bidMetaKey]
	if !found {
		meta = map[string]any{}
		adapterBid[bidMetaKey] = meta
	}
	typedMeta, ok := meta.(map[string]any)
	if !ok || typedMeta == nil {
		return NewValidationFailedError("failed to set key:[%s] in BidMeta, error:[incorrect data type]", key)
	}
	typedMeta[key] = value
	return nil
}
