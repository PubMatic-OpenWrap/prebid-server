package ortbbidder

import (
	"github.com/PubMatic-OpenWrap/prebid-server/v2/openrtb_ext"
)

type ParserFunc func(p Parser, currNode, newNode map[string]any, location []string)

type Parser interface {
	MType(currNode, newNode map[string]any, location []string)
	Dur(currNode, newNode map[string]any, location []string)
	Fledge(currNode, newNode map[string]any, location []string)
}

type ParserImpl struct {
}

func (p *ParserImpl) MType(bid, typeBid map[string]any, location []string) {
	var getMType = func(bid map[string]any, location []string) openrtb_ext.BidType {

		mType, ok := bid["mtype"].(string)
		if ok {
			return openrtb_ext.BidType(mType)
		}

		value, ok := getValueFromLocation(bid, location[2:])
		if ok {
			mType, ok := value.(string)
			if ok {
				return openrtb_ext.BidType(mType)
			}
		}
		return ""
	}

	typeBid["bidType"] = getMType(bid, location)
}

func (p *ParserImpl) Dur(currNode, newNode map[string]any, location []string) {
}

func (p *ParserImpl) Fledge(currNode, newNode map[string]any, location []string) {
}

func getBidParamParser() map[string]ParserFunc {
	return map[string]ParserFunc{
		"mtype": func(p Parser, currNode, newNode map[string]any, location []string) {
			p.MType(currNode, newNode, location)
		},
		"dur": func(p Parser, currNode, newNode map[string]any, location []string) {
			p.Dur(currNode, newNode, location)
		},
	}
}

func getRequestParamParser() map[string]ParserFunc {
	return map[string]ParserFunc{
		"fledge": func(p Parser, currNode, newNode map[string]any, location []string) {
			p.Fledge(currNode, newNode, location)
		},
	}
}
