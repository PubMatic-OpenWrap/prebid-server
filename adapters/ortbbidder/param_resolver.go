package ortbbidder

var (
	bidLevelParams      = [...]string{"mtype", "duration", "meta"}
	responseLevelParams = [...]string{"fledge"}
)

type BidderParamProcessor interface {
	ResolveParam(targetNode map[string]any, node map[string]any, ortbResponse map[string]any, location, param string)
}

type paramResolver interface {
	fromOriginalObject(node map[string]any) (any, bool)
	fromParamLocation(responseNode map[string]any, path string) (any, bool)
	autoDetect(node map[string]any) (any, bool)
	setValue(targetNode map[string]any, value any)
}

type paramResolvers map[string]paramResolver

type paramProcessor struct {
	resolvers paramResolvers
}

func NewParamProcessor() *paramProcessor {
	resolvers := paramResolvers{
		"mtype": &mtypeResolver{},
	}
	return &paramProcessor{resolvers: resolvers}
}

func (ps *paramProcessor) ResolveParam(targetNode map[string]any, node map[string]any, responseNode map[string]any, location, param string) {
	resolver, ok := ps.resolvers[param]
	if !ok {
		return
	}

	value, found := resolver.fromOriginalObject(node)
	if !found {
		value, found = resolver.fromParamLocation(responseNode, location)
		if !found {
			value, found = resolver.autoDetect(node)
			if !found {
				return
			}
		}
	}

	resolver.setValue(targetNode, value)
}
