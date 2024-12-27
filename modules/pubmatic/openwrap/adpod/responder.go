package adpod

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type Responder interface {
	FormResponse(bidResponse *openrtb2.BidResponse) interface{}
}

func NewResponder(rctx models.RequestCtx) Responder {
	switch rctx.Endpoint {
	case "vast":
		return newVastResponder()
	case "ortb":
		return newOrtbResponder()
	case "json":
		return newJsonResponder()
	default:
		return nil
	}
}
