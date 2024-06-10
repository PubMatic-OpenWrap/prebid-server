package resolver

var (
	BidLevelParams      = [...]string{"mtype", "duration", "meta"}
	ResponseLevelParams = [...]string{"currency", "fledge"}
)

var (
	resolvers = resolverMap{
		"mtype":    &mtypeResolver{},
		"currency": &currencyResolver{},
	}
)

type resolver interface {
	getFromORTBObject(node map[string]any) (any, bool)
	getUsingBidderParamLocation(responseNode map[string]any, path string) (any, bool)
	autoDetect(node map[string]any) (any, bool)
	setValue(targetNode map[string]any, value any)
}

type resolverMap map[string]resolver

type paramResolver struct {
	bidderResponse map[string]any
}

func NewParamResolver(bidderResponse map[string]any) *paramResolver {
	return &paramResolver{
		bidderResponse: bidderResponse,
	}
}

func (pr *paramResolver) Resolve(sourceNode, targetNode map[string]any, location, param string) {
	if sourceNode == nil || targetNode == nil || pr.bidderResponse == nil {
		return
	}
	resolver, ok := resolvers[param]
	if !ok {
		return
	}

	value, found := resolver.getFromORTBObject(sourceNode)
	if !found {
		value, found = resolver.getUsingBidderParamLocation(pr.bidderResponse, location)
		if !found {
			value, found = resolver.autoDetect(sourceNode)
			if !found {
				return
			}
		}
	}

	resolver.setValue(targetNode, value)
}
