package resolver

import (
	"github.com/prebid/openrtb/v20/openrtb2"
)

var (
	resolvers = resolverMap{
		fledgeAuctionConfig:      &fledgeResolver{},
		bidType:                  &bidTypeResolver{},
		bidDealPriority:          &bidDealPriorityResolver{},
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
	getFromORTBObject(sourceNode map[string]any) (any, bool)
	retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool)
	autoDetect(request *openrtb2.BidRequest, sourceNode map[string]any) (any, bool)
	setValue(targetNode map[string]any, value any) bool
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

func (r *paramResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	return nil, false
}

func (r *paramResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	return nil, false
}

func (r *paramResolver) autoDetect(request *openrtb2.BidRequest, bid map[string]any) (any, bool) {
	return nil, false
}

// Resolve fetches a parameter value from sourceNode or bidderResponse and sets it in targetNode.
// The order of lookup is as follows:
// 1) ORTB standard field
// 2) Location from JSON file (bidder params)
// 3) Auto-detection
// If the value is found, it is set in the targetNode.
func (pr *paramResolver) Resolve(sourceNode, targetNode map[string]any, path string, param parameter) bool {
	if sourceNode == nil || targetNode == nil || pr.bidderResponse == nil {
		return false
	}
	resolver, ok := resolvers[param]
	if !ok {
		return false
	}

	// get the value from the ORTB object
	value, found := resolver.getFromORTBObject(sourceNode)
	if !found {
		// get the value from the bidder response using the location
		value, found = resolver.retrieveFromBidderParamLocation(pr.bidderResponse, path)
		if !found {
			// auto detect value
			value, found = resolver.autoDetect(pr.request, sourceNode)
			if !found {
				return false
			}
		}
	}

	return resolver.setValue(targetNode, value)
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
