package ortbbidder

import (
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

type (
	ParserFunc         func(p Parser, bid, typeBid map[string]any, path string)
	ResponseParserFunc func(p Parser, adapterRespnse map[string]any, path string)
)
type Parser interface {
	MType(bid, typeBid map[string]any, path string)
	Dur(bid, typeBid map[string]any, path string)
	Fledge(adapterRespnse map[string]any, path string)
}

type ParserImpl struct {
	bidResponse map[string]any
}

func (p *ParserImpl) MType(bid, typeBid map[string]any, path string) {
	typeBid["bidType"] = func(bid map[string]any, path string) openrtb_ext.BidType {

		mType, ok := bid["mtype"].(string)
		if ok {
			return openrtb_ext.BidType(mType)
		}

		value, ok := getValueFromLocation(p.bidResponse, path)
		if ok {
			mType, ok := value.(string)
			if ok {
				return openrtb_ext.BidType(mType)
			}
		}
		return ""
	}
}

func (p *ParserImpl) Dur(bid, typeBid map[string]any, path string) {
}

func (p *ParserImpl) Fledge(adapterResponse map[string]any, path string) {
}

type parserFactory interface {
	getBidParamParser() map[string]ParserFunc
	getResponseParamParser() map[string]ResponseParserFunc
	NewParser(bidResponse map[string]any) Parser
}

type parserFactoryImpl struct {
}

func (parserFactoryImpl) NewParser(bidResponse map[string]any) Parser {
	return &ParserImpl{
		bidResponse: bidResponse,
	}
}

func (parserFactoryImpl) getBidParamParser() map[string]ParserFunc {
	return map[string]ParserFunc{
		"mtype": Parser.MType,
		"dur":   Parser.Dur,
	}
}

func (parserFactoryImpl) getResponseParamParser() map[string]ResponseParserFunc {
	return map[string]ResponseParserFunc{
		"fledge": Parser.Fledge,
	}
}
