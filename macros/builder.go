package macros

import (
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
	"golang.org/x/exp/maps"
)

const (
	BidIDKey       = "PBS-BIDID"
	AppBundleKey   = "PBS-APPBUNDLE"
	DomainKey      = "PBS-DOMAIN"
	PubDomainkey   = "PBS-PUBDOMAIN"
	PageURLKey     = "PBS-PAGEURL"
	AccountIDKey   = "PBS-ACCOUNTID"
	LmtTrackingKey = "PBS-LIMITADTRACKING"
	ConsentKey     = "PBS-GDPRCONSENT"
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
	// Build returns the macros map
	Build() map[string]string
	// CleanUp will remove the bid and vast event specific keys
	CleanUp()
}

type macroBuilder struct {
	// macros stores request level macros
	macros map[string]string
}

func NewBuilder() Builder {
	return &macroBuilder{}
}

func (b *macroBuilder) WithBidRequest(reqWrapper *openrtb_ext.RequestWrapper) {
	reqExt, _ := reqWrapper.GetRequestExt()
	if reqExt != nil && reqExt.GetPrebid() != nil {
		maps.Copy(b.macros, reqExt.GetPrebid().Macros)
	}

	b.macros[AppBundleKey] = reqWrapper.App.Bundle

	if reqWrapper.App.Domain != "" {
		b.macros[DomainKey] = reqWrapper.App.Domain
	}

	if reqWrapper.Site.Domain != "" {
		b.macros[DomainKey] = reqWrapper.Site.Domain
	}

	if reqWrapper.Site.Publisher.Domain != "" {
		b.macros[PubDomainkey] = reqWrapper.Site.Publisher.Domain
	}

	if reqWrapper.App.Publisher.Domain != "" {
		b.macros[PubDomainkey] = reqWrapper.App.Publisher.Domain
	}

	b.macros[PageURLKey] = reqWrapper.Site.Page
	userExt, _ := reqWrapper.GetUserExt()
	b.macros[ConsentKey] = *userExt.GetConsent()
	if reqWrapper.Device.Lmt != nil {
		b.macros[LmtTrackingKey] = strconv.Itoa(int(*reqWrapper.Device.Lmt))
	}

	b.macros[AccountIDKey] = reqWrapper.ID
	if reqWrapper.Site.Publisher.ID != "" {
		b.macros[AccountIDKey] = reqWrapper.Site.Publisher.ID
	}

	if reqWrapper.App.Publisher.ID != "" {
		b.macros[AccountIDKey] = reqWrapper.App.Publisher.ID
	}

}
func (b *macroBuilder) WithBidResponse(bid *openrtb2.Bid, bidderName string) {
	b.macros = map[string]string{}
	b.macros[BidIDKey] = bid.ID
}

func (b *macroBuilder) WithImpression(openrtb2.Imp) {

}

func (b *macroBuilder) Build() map[string]string {
	return b.macros
}

func (b *macroBuilder) WithEventDetails() {}

func (b *macroBuilder) CleanUp() {
	keys := []string{BidIDKey}

	for _, key := range keys {
		delete(b.macros, key)
	}
}
