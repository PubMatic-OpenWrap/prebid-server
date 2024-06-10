package resolver

// import (
// 	_ "github.com/golang/mock/mockgen/model"
// )

// var (
// 	bidLevelParams      = [...]string{"mtype", "duration", "meta"}
// 	responseLevelParams = [...]string{"currency", "fledge"}
// )

// type BidderParamProcessor interface {
// 	Resolve(node, targetNode map[string]any, ortbResponse map[string]any, location, param string)
// }

// type Resolver interface {
// 	fromOriginalObject(node map[string]any) (any, bool)
// 	fromParamLocation(responseNode map[string]any, path string) (any, bool)
// 	autoDetect(node map[string]any) (any, bool)
// 	setValue(targetNode map[string]any, value any)
// }

// type paramResolvers map[string]Resolver

// type paramProcessor struct {
// 	responseNode map[string]any
// }

// var resolvers = paramResolvers{
// 	"mtype":    &mtypeResolver{},
// 	"currency": &currencyResolver{},
// }

// // func NewParamProcessor() *paramProcessor {
// // 	resolvers := paramResolvers{
// // 		"mtype":    &mtypeResolver{},
// // 		"currency": &currencyResolver{},
// // 	}

// // 	return &paramProcessor{resolvers: resolvers}
// // }

// type paramResolver struct {
// 	bidderResponse map[string]any
// 	sourceNode     map[string]any
// 	targetNode     map[string]any
// 	location       string
// }

// func (pr *paramResolver) Resolve(param string) {
// 	if pr.sourceNode == nil || pr.targetNode == nil || pr.bidderResponse == nil {
// 		return
// 	}
// 	resolver, ok := resolvers[param]
// 	if !ok {
// 		return
// 	}

// 	value, found := resolver.fromOriginalObject(pr.sourceNode)
// 	if !found {
// 		value, found = resolver.fromParamLocation(pr.bidderResponse, pr.location)
// 		if !found {
// 			value, found = resolver.autoDetect(pr.sourceNode)
// 			if !found {
// 				return
// 			}
// 		}
// 	}

// 	resolver.setValue(pr.targetNode, value)
// }
