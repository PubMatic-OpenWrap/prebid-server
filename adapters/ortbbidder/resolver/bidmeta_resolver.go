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
		return nil, nil
	}
	bidMeta, err := validateBidMeta(value)
	if err != nil {
		return nil, util.NewWarning("failed to map response-param:[bidMeta] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}

	adomains, ok := validateDataTypeSlice[string](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaAdvertiserDomains] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	advId, ok := validateNumber[int](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaAdvertiserId] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	advName, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaAdvertiserName] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	agencyId, ok := validateNumber[int](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaAgencyId] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	agencyName, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaAgencyName] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	brandId, ok := validateNumber[int](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaBrandId] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	brandName, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaBrandName] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	dChain, ok := validateMap(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaDchain] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	demandSource, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaDemandSource] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	mediaType, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaMediaType] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	networkId, ok := validateNumber[int](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaNetworkId] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	networkName, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaNetworkName] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	categoryId, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaPrimaryCatId] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	rendererName, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaRendererName] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	rendererVersion, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaRendererVersion] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	rendererData, ok := validateMap(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaRendererData] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	rendererUrl, ok := validateString(value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaRendererUrl] method:[response_param_location] value:[%v]", value)
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
		return nil, nil
	}
	secondaryCategories, ok := validateDataTypeSlice[string](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidMetaSecondaryCatIds] method:[response_param_location] value:[%v]", value)
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
		return util.NewWarning("failed to set key:[%s] in BidMeta, value:[%+v] error:[incorrect data type]", key, value)
	}
	typedMeta[key] = value
	return nil
}
