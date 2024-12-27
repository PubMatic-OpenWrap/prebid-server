package adpod

import "github.com/prebid/openrtb/v20/openrtb2"

type vastResp struct {
}

func newVastResponder() Responder {
	return &vastResp{}
}

func (vast *vastResp) FormResponse(bidResponse *openrtb2.BidResponse) interface{} {
	return nil
}
