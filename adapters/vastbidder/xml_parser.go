package vastbidder

import "github.com/prebid/prebid-server/v2/openrtb_ext"

type xmlParser interface {
	Parse([]byte) error
	SetVASTTag(vastTag *openrtb_ext.ExtImpVASTBidderTag)
	GetAdvertiser() []string
	GetPricingDetails() (float64, string)
	GetCreativeID() string
	GetDuration() (int, error)
}

type xmlParserType int

const (
	unknownXMLParserType xmlParserType = iota
	etreeXMLParserType
	fastXMLParserType
)

func getXMLParser(ty xmlParserType) xmlParser {
	if ty == fastXMLParserType {
		return newFastXMLParser()
	}
	return newETreeXMLParser()
}
