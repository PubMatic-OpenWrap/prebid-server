package tagbidder

import (
	"errors"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
)

//ITagResponseHandler parse bidder response
type ITagResponseHandler interface {
	Validate(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) []error
	MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error)
}

//GetResponseHandler returns response handler
func GetResponseHandler(responseType ResponseHandlerType) (ITagResponseHandler, error) {
	switch responseType {
	case VASTTagResponseHandlerType:
		return NewVASTTagResponseHandler(), nil
	}
	return nil, errors.New(`Unkown Response Handler`)
}
