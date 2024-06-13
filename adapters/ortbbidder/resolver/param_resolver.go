package resolver

import "github.com/prebid/openrtb/v20/openrtb2"

var (
	// TypeBidFields is a list of typebid fields that are populated using resolver framework
	TypeBidFields = [...]string{"bidtype", "duration", "bidmeta"}
	// AdapterResponseFields is a list of adapter response fields that are populated using resolver framework
	AdapterResponseFields = [...]string{"currency", "fledge"}
)

var (
	resolvers = resolverMap{
		"bidtype":  &mtypeResolver{},
		"currency": &currencyResolver{},
	}
)

type resolver interface {
	getFromORTBObject(sourceNode map[string]any) (any, bool)
	getUsingBidderParamLocation(responseNode map[string]any, path string) (any, bool)
	autoDetect(request *openrtb2.BidRequest, sourceNode map[string]any) (any, bool)
	setValue(targetNode map[string]any, value any)
}

type resolverMap map[string]resolver

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
func (pr *paramResolver) Resolve(sourceNode, targetNode map[string]any, location, param string) {
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
		value, found = resolver.getUsingBidderParamLocation(pr.bidderResponse, location)
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
