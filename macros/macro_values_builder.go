package macros

import (
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type macroValuesBuilder interface {
	// WithBidRequest determines value of macro
	// from openrtb_ext.RequestWrapper
	// For custom macros builder will merge entire
	// map present at openrtb_ext.RequestWrapper.RequestExt.prebid.macros
	WithBidRequest(openrtb_ext.RequestWrapper) macroValuesBuilder
	// WithBidResponse determines value of macro
	// from openrtb2.Bid, biddername, corebidder (Optional)
	WithBidResponse(openrtb2.Bid, string, string) macroValuesBuilder
	// WithImpression determines value of macro
	// from openrtb2.Imp
	WithImpression(openrtb2.Imp) macroValuesBuilder
	// Build with return map of macro and value
	// macro will be with delimiters
	Build() map[string]string
}

type DefaultBuilder struct {
	macroValuesBuilder
	m map[string]string
}

func NewBuilder() macroValuesBuilder {
	return DefaultBuilder{}
}

func resolveValues(req openrtb_ext.RequestWrapper, m map[string]string) {
	m["PBS_DOMAIN"] = req.Site.Domain
}
