package ortbbidder

import (
	"github.com/prebid/openrtb/v20/openrtb2"
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
	getMType := func(bid map[string]any, path string) openrtb_ext.BidType {
		// check in ortb bid.MType
		mType, ok := bid["mtype"].(float64)
		if ok {
			return getMediaTypeForBidFromMType(openrtb2.MarkupType(mType))
		}

		// get from bidder param location
		value, ok := getValueFromLocation(p.bidResponse, path)
		if ok {
			mType, ok := value.(string)
			if ok {
				return openrtb_ext.BidType(mType)
			}
		}

		// auto detection logic here
		return ""
	}
	typeBid["BidType"] = getMType(bid, path)
}

func (p *ParserImpl) Dur(bid, typeBid map[string]any, path string) {

}

func (p *ParserImpl) Fledge(adapterResponse map[string]any, path string) {

}

type ParserFactory interface {
	GetBidParamParser() map[string]ParserFunc
	GetResponseParamParser() map[string]ResponseParserFunc
	NewParser(bidResponse map[string]any) Parser
}

type ParserFactoryImpl struct {
}

func (ParserFactoryImpl) NewParser(bidResponse map[string]any) Parser {
	return &ParserImpl{
		bidResponse: bidResponse,
	}
}

var (
	bidParamParser = map[string]ParserFunc{
		"mtype": Parser.MType,
		"dur":   Parser.Dur,
	}
	responseParamParser = map[string]ResponseParserFunc{
		"fledge": Parser.Fledge,
	}
)

func (ParserFactoryImpl) GetBidParamParser() map[string]ParserFunc {
	return bidParamParser
}

func (ParserFactoryImpl) GetResponseParamParser() map[string]ResponseParserFunc {
	return responseParamParser
}
