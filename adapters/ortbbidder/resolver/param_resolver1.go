package resolver

var (
	BidLevelParams      = [...]string{"mtype", "duration", "meta"}
	ResponseLevelParams = [...]string{"currency", "fledge"}
)

type resolver interface {
	getFromORTBObject(node map[string]any) (any, bool)
	getUsingBidderParam(responseNode map[string]any, path string) (any, bool)
	autoDetect(node map[string]any) (any, bool)
	setValue(targetNode map[string]any, value any)
}

type resolverMap map[string]resolver

type ParamResolver struct {
	BidderResponse map[string]any
}

var resolvers = resolverMap{
	"mtype":    &mtypeResolver{},
	"currency": &currencyResolver{},
}

func (pr *ParamResolver) Resolve(sourceNode, targetNode map[string]any, location, param string) {
	if sourceNode == nil || targetNode == nil || pr.BidderResponse == nil {
		return
	}
	resolver, ok := resolvers[param]
	if !ok {
		return
	}

	value, found := resolver.getFromORTBObject(sourceNode)
	if !found {
		value, found = resolver.getUsingBidderParam(pr.BidderResponse, location)
		if !found {
			value, found = resolver.autoDetect(sourceNode)
			if !found {
				return
			}
		}
	}

	resolver.setValue(targetNode, value)
}
