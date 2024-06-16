package resolver

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

type resolveType string

func (s resolveType) String() string {
	return string(s)
}

const (
	BidType  resolveType = "bidtype"
	Duration resolveType = "duration"
	BidMeta  resolveType = "bidmeta"
	Fledge   resolveType = "fledge"
)

var (
	resolvers = resolverMap{
		BidType: &mtypeResolver{},
	}
)

type resolver interface {
	getFromORTBObject(sourceNode map[string]any) (any, bool)
	retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool)
	autoDetect(request *openrtb2.BidRequest, sourceNode map[string]any) (any, bool)
	setValue(targetNode map[string]any, value any)
}

type resolverMap map[resolveType]resolver

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

// Resolve fetches a parameter value from sourceNode or bidderResponse and sets it in targetNode.
// The order of lookup is as follows:
// 1) ORTB standard field
// 2) Location from JSON file (bidder params)
// 3) Auto-detection
// If the value is found, it is set in the targetNode.
func (pr *paramResolver) Resolve(sourceNode, targetNode map[string]any, path string, param resolveType) {
	if sourceNode == nil || targetNode == nil || pr.bidderResponse == nil {
		return
	}
	resolver, ok := resolvers[param]
	if !ok {
		return
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
				return
			}
		}
	}

	resolver.setValue(targetNode, value)
}

// valueResolver is a generic resolver to get values from the response node using location
type valueResolver struct{}

func (r *valueResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	return util.GetValueFromLocation(responseNode, path)
}
