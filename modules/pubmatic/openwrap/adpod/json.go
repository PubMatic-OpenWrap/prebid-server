package adpod

import (
	"net/http"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type jsonResp struct {
	rctx models.RequestCtx
}

func newJsonResponder(rctx models.RequestCtx) Responder {
	return &jsonResp{rctx: rctx}
}

func (json *jsonResp) FormResponse(bidResponse *openrtb2.BidResponse, headers http.Header) (interface{}, http.Header, error) {
	return bidResponse, headers, nil
}
