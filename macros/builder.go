package macros

import (
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
	"golang.org/x/exp/maps"
)

type Builder interface {
	// WithBidRequest determines value of macro
	// from openrtb_ext.RequestWrapper
	// For custom macros builder will merge entire
	// map present at openrtb_ext.RequestWrapper.RequestExt.prebid.macros
	WithBidRequest(*openrtb_ext.RequestWrapper)
	// WithBidResponse determines value of macro
	// from openrtb2.Bid, biddername, corebidder (Optional)
	WithBidResponse(*openrtb2.Bid, string)
	// WithImpression determines value of macro
	// from openrtb2.Imp
	WithImpression(openrtb2.Imp)
	// Build with return map of macro and value
	// macro will be with delimiters
	Build() map[string]string
}

type macroBuilder struct {
	requestMacros map[string]string
	bidMacros     map[string]string
	evenMacros    map[string]string
}

func NewBuilder() Builder {
	return &macroBuilder{}
}

func (b *macroBuilder) WithBidRequest(reqWrapper *openrtb_ext.RequestWrapper) {
	b.requestMacros["##PBS-APPBUNDLE##"] = reqWrapper.App.Bundle

	if reqWrapper.App.Domain != "" {
		b.requestMacros["##PBS-DOMAIN##"] = reqWrapper.App.Domain
	}

	if reqWrapper.Site.Domain != "" {
		b.requestMacros["##PBS-DOMAIN##"] = reqWrapper.Site.Domain
	}

	if reqWrapper.Site.Publisher.Domain != "" {
		b.requestMacros["##PBS-PUBDOMAIN##"] = reqWrapper.Site.Publisher.Domain
	}

	if reqWrapper.App.Publisher.Domain != "" {
		b.requestMacros["##PBS-PUBDOMAIN##"] = reqWrapper.App.Publisher.Domain
	}

	b.requestMacros["##PBS-PAGEURL##"] = reqWrapper.Site.Page
	userExt, _ := reqWrapper.GetUserExt()
	b.requestMacros["##PBS-GDPRCONSENT##"] = *userExt.GetConsent()
	if reqWrapper.Device.Lmt != nil {
		b.requestMacros["##PBS-LIMITADTRACKING##"] = strconv.Itoa(int(*reqWrapper.Device.Lmt))
	}

	b.requestMacros["##PBS-AUCTIONID##"] = reqWrapper.ID
	if reqWrapper.Site.Publisher.ID != "" {
		b.requestMacros["##PBS-ACCOUNTID##"] = reqWrapper.Site.Publisher.ID
	}

	if reqWrapper.App.Publisher.ID != "" {
		b.requestMacros["##PBS-ACCOUNTID##"] = reqWrapper.App.Publisher.ID
	}

	//b.requestMacros["##PBS-TIMESTAMP##"] =

}
func (b *macroBuilder) WithBidResponse(bid *openrtb2.Bid, bidderName string) {
	b.bidMacros = map[string]string{}
	b.bidMacros["##PBS-BIDID##"] = bid.ID
}

func (b *macroBuilder) WithImpression(openrtb2.Imp) {

}

func (b *macroBuilder) Build() map[string]string {
	macros := map[string]string{}

	maps.Copy(macros, b.requestMacros)
	maps.Copy(macros, b.requestMacros)
	maps.Copy(macros, b.evenMacros)
	return macros
}
