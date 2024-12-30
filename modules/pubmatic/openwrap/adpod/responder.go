package adpod

import (
	"net/http"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type Responder interface {
	FormResponse(bidResponse *openrtb2.BidResponse, headers http.Header) (interface{}, http.Header, error)
}

func NewResponder(rctx models.RequestCtx) Responder {
	switch rctx.Endpoint {
	case "vast":
		return newVastResponder(rctx)
	case "ortb":
		return newOrtbResponder(rctx)
	case "json":
		return newJsonResponder(rctx)
	default:
		return nil
	}
}
