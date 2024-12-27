package adpod

import "github.com/prebid/openrtb/v20/openrtb2"

type ortbResp struct {
}

func newOrtbResponder() Responder {
	return &ortbResp{}
}

func (ortb *ortbResp) FormResponse(bidResponse *openrtb2.BidResponse) interface{} {
	return nil
}
