package macros

import (
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
	"golang.org/x/exp/maps"
)

type Builder interface {
	// WithBidRequest extracts and stores request level macros from bid request
	WithBidRequest(*openrtb_ext.RequestWrapper)
	// WithBidResponse extracts and stores bid level macros from seatBid.bid
	WithBidResponse(*openrtb2.Bid, string)
	// WithImpression extracts and stores impression level macros.
	WithImpression(openrtb2.Imp)
	// WithEventDetails extracts and stores events level macros
	WithEventDetails()
	Build() map[string]string
}

type macroBuilder struct {
	// requestMacros stores request level macros
	requestMacros map[string]string
	// bidMacros stores bid level macros.
	// new instance will be created for every bid
	bidMacros map[string]string
	// eventMacros stores macros extracted from vast exl and vast events
	eventMacros map[string]string
}

func NewBuilder() Builder {
	return &macroBuilder{}
}

func (b *macroBuilder) WithBidRequest(reqWrapper *openrtb_ext.RequestWrapper) {
	reqExt, _ := reqWrapper.GetRequestExt()
	if reqExt != nil && reqExt.GetPrebid() != nil {
		maps.Copy(b.requestMacros, reqExt.GetPrebid().Macros)
	}

	b.requestMacros["PBS-APPBUNDLE"] = reqWrapper.App.Bundle

	if reqWrapper.App.Domain != "" {
		b.requestMacros["PBS-DOMAIN"] = reqWrapper.App.Domain
	}

	if reqWrapper.Site.Domain != "" {
		b.requestMacros["PBS-DOMAIN"] = reqWrapper.Site.Domain
	}

	if reqWrapper.Site.Publisher.Domain != "" {
		b.requestMacros["PBS-PUBDOMAIN"] = reqWrapper.Site.Publisher.Domain
	}

	if reqWrapper.App.Publisher.Domain != "" {
		b.requestMacros["PBS-PUBDOMAIN"] = reqWrapper.App.Publisher.Domain
	}

	b.requestMacros["PBS-PAGEURL"] = reqWrapper.Site.Page
	userExt, _ := reqWrapper.GetUserExt()
	b.requestMacros["PBS-GDPRCONSENT"] = *userExt.GetConsent()
	if reqWrapper.Device.Lmt != nil {
		b.requestMacros["PBS-LIMITADTRACKING"] = strconv.Itoa(int(*reqWrapper.Device.Lmt))
	}

	b.requestMacros["PBS-AUCTIONID"] = reqWrapper.ID
	if reqWrapper.Site.Publisher.ID != "" {
		b.requestMacros["PBS-ACCOUNTID"] = reqWrapper.Site.Publisher.ID
	}

	if reqWrapper.App.Publisher.ID != "" {
		b.requestMacros["PBS-ACCOUNTID"] = reqWrapper.App.Publisher.ID
	}

}
func (b *macroBuilder) WithBidResponse(bid *openrtb2.Bid, bidderName string) {
	b.bidMacros = map[string]string{}
	b.bidMacros["PBS-BIDID"] = bid.ID
}

func (b *macroBuilder) WithImpression(openrtb2.Imp) {

}

func (b *macroBuilder) Build() map[string]string {
	macros := map[string]string{}

	maps.Copy(macros, b.requestMacros)
	maps.Copy(macros, b.requestMacros)
	maps.Copy(macros, b.eventMacros)
	return macros
}

func (b *macroBuilder) WithEventDetails() {}
