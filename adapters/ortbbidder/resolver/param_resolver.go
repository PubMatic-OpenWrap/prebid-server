package resolver

var (
	BidLevelParams      = [...]string{"mtype", "duration", "meta"}
	ResponseLevelParams = [...]string{"currency", "fledge"}
)

// type BidderParamProcessor interface {
// 	Resolve(node, targetNode map[string]any, ortbResponse map[string]any, location, param string)
// }

type resolver interface {
	getFromORTBObject(node map[string]any) (any, bool)
	getUsingBidderParam(responseNode map[string]any, path string) (any, bool)
	autoDetect(node map[string]any) (any, bool)
	setValue(targetNode map[string]any, value any)
}

type paramResolvers map[string]resolver

type paramProcessor struct {
	responseNode map[string]any
}

var resolvers = paramResolvers{
	"mtype":    &mtypeResolver{},
	"currency": &currencyResolver{},
}

type ParamResolver struct {
	BidderResponse map[string]any
	SourceNode     map[string]any
	TargetNode     map[string]any
	Location       string
}

func (pr *ParamResolver) Resolve(param string) {
	if pr.SourceNode == nil || pr.TargetNode == nil || pr.BidderResponse == nil {
		return
	}
	resolver, ok := resolvers[param]
	if !ok {
		return
	}

	value, found := resolver.getFromORTBObject(pr.SourceNode)
	if !found {
		value, found = resolver.getUsingBidderParam(pr.BidderResponse, pr.Location)
		if !found {
			value, found = resolver.autoDetect(pr.SourceNode)
			if !found {
				return
			}
		}
	}

	resolver.setValue(pr.TargetNode, value)
}
