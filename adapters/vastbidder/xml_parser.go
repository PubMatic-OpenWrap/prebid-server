package vastbidder

import "github.com/prebid/prebid-server/v3/openrtb_ext"

type xmlParser interface {
	Name() string
	Parse([]byte) error
	SetVASTTag(vastTag *openrtb_ext.ExtImpVASTBidderTag)
	GetAdvertiser() []string
	GetPricingDetails() (float64, string)
	GetCreativeID() string
	GetDuration() (int, error)
}

func getXMLParser() xmlParser {
	if openrtb_ext.IsFastXMLEnabled() {
		return newFastXMLParser()
	}
	return newETreeXMLParser()
}
