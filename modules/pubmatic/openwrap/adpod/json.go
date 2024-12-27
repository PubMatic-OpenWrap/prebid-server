package adpod

import "github.com/prebid/openrtb/v20/openrtb2"

type jsonResp struct {
}

func newJsonResponder() Responder {
	return &jsonResp{}
}

func (json *jsonResp) FormResponse(bidResponse *openrtb2.BidResponse) interface{} {
	return nil
}
