package tagbidder

import (
	"errors"
	"github.com/mxmCherry/openrtb/v15/openrtb2"

	"github.com/prebid/prebid-server/adapters"
)

//ITagResponseHandler parse bidder response
type ITagResponseHandler interface {
	Validate(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) []error
	MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error)
}

//GetResponseHandler returns response handler
func GetResponseHandler(responseType ResponseHandlerType) (ITagResponseHandler, error) {
	switch responseType {
	case VASTTagResponseHandlerType:
		return NewVASTTagResponseHandler(), nil
	}
	return nil, errors.New(`Unkown Response Handler`)
}
