package resolver

import (
	"github.com/prebid/openrtb/v20/openrtb2"
)

var (
	resolvers = resolverMap{
		fledgeAuctionConfig:      &fledgeResolver{},
		bidType:                  &bidTypeResolver{},
		bidDealPriority:          &bidDealPriorityResolver{},
		bidVideo:                 &bidVideoResolver{},
		bidVideoDuration:         &bidVideoDurationResolver{},
		bidVideoPrimaryCategory:  &bidVideoPrimaryCategoryResolver{},
		bidMeta:                  &bidMetaResolver{},
		bidMetaAdvertiserDomains: &bidMetaAdvDomainsResolver{},
		bidMetaAdvertiserId:      &bidMetaAdvIDResolver{},
		bidMetaAdvertiserName:    &bidMetaAdvNameResolver{},
		bidMetaAgencyId:          &bidMetaAgencyIDResolver{},
		bidMetaAgencyName:        &bidMetaAgencyNameResolver{},
		bidMetaBrandId:           &bidMetaBrandIDResolver{},
		bidMetaBrandName:         &bidMetaBrandNameResolver{},
		bidMetaDChain:            &bidMetaDChainResolver{},
		bidMetaDemandSource:      &bidMetaDemandSourceResolver{},
		bidMetaMediaType:         &bidMetaMediaTypeResolver{},
		bidMetaNetworkId:         &bidMetaNetworkIDResolver{},
		bidMetaNetworkName:       &bidMetaNetworkNameResolver{},
		bidMetaPrimaryCatId:      &bidMetaPrimaryCategoryIDResolver{},
		bidMetaRendererName:      &bidMetaRendererNameResolver{},
		bidMetaRendererVersion:   &bidMetaRendererVersionResolver{},
		bidMetaRenderedData:      &bidMetaRendererDataResolver{},
		bidMetaRenderedUrl:       &bidMetaRendererUrlResolver{},
		bidMetaSecondaryCatId:    &bidMetaSecondaryCategoryIDsResolver{},
	}
)

type resolver interface {
	getFromORTBObject(sourceNode map[string]any) (any, error)
	retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error)
	autoDetect(request *openrtb2.BidRequest, sourceNode map[string]any) (any, error)
	setValue(targetNode map[string]any, value any) error
}

type resolverMap map[parameter]resolver

type paramResolver struct {
	bidderResponse map[string]any
	request        *openrtb2.BidRequest
}

// New returns a new instance of paramResolver.
func New(request *openrtb2.BidRequest, bidderResponse map[string]any) *paramResolver {
	return &paramResolver{
		bidderResponse: bidderResponse,
		request:        request,
	}
}

func (r *paramResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	return nil, nil
}

func (r *paramResolver) getFromORTBObject(bid map[string]any) (any, error) {
	return nil, nil
}

func (r *paramResolver) autoDetect(request *openrtb2.BidRequest, bid map[string]any) (any, error) {
	return nil, nil
}

// Resolve fetches a parameter value from sourceNode or bidderResponse and sets it in targetNode.
// The order of lookup is as follows:
// 1) ORTB standard field
// 2) Location from JSON file (bidder params)
// 3) Auto-detection
// If the value is found, it is set in the targetNode.
func (pr *paramResolver) Resolve(sourceNode, targetNode map[string]any, path string, param parameter) (errs []error) {
	if sourceNode == nil || targetNode == nil || pr.bidderResponse == nil {
		return
	}
	resolver, ok := resolvers[param]
	if !ok {
		return
	}

	value, err := resolver.getFromORTBObject(sourceNode) // get the value from the ORTB object
	if err != nil {
		errs = append(errs, err)
	}

	if value == nil {
		value, err = resolver.retrieveFromBidderParamLocation(pr.bidderResponse, path) // get the value from the bidder response using the location
		if err != nil {
			errs = append(errs, err)
		}
	}

	if value == nil {
		value, err = resolver.autoDetect(pr.request, sourceNode) // auto detect value
		if err != nil {
			errs = append(errs, err)
		}
	}

	// return if value not found
	if value == nil {
		return errs
	}

	err = resolver.setValue(targetNode, value)
	if err != nil {
		errs = append(errs, err)
	}
	return errs
}

// list of parameters to be resolved at typedBid level.
// order of elements matters since child parameter's (BidMetaAdvertiserDomains) value overrides the parent parameter's (BidMeta.AdvertiserDomains) value.
var TypedBidParams = []parameter{
	bidType,
	bidDealPriority,
	bidVideo,
	bidVideoDuration,
	bidVideoPrimaryCategory,
	bidMeta,
	bidMetaAdvertiserDomains,
	bidMetaAdvertiserId,
	bidMetaAdvertiserName,
	bidMetaAgencyId,
	bidMetaAgencyName,
	bidMetaBrandId,
	bidMetaBrandName,
	bidMetaDChain,
	bidMetaDemandSource,
	bidMetaMediaType,
	bidMetaNetworkId,
	bidMetaNetworkName,
	bidMetaPrimaryCatId,
	bidMetaRendererName,
	bidMetaRendererVersion,
	bidMetaRenderedData,
	bidMetaRenderedUrl,
	bidMetaSecondaryCatId,
}

// list of parameters to be resolved at response level.
var ResponseParams = []parameter{
	fledgeAuctionConfig,
}
