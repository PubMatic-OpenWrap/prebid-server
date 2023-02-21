package processor

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

var (
	bidLevelKeys = []string{BidIDKey}
)

type Provider interface {
	// GetMacro returns the macro value for the given macro key
	GetMacro(key string) string
	// GetAllMacros return all the macros
	GetAllMacros(keys []string) map[string]string
	// SetContext set the bid and imp for the current provider
	SetContext(bid *openrtb2.Bid, imp *openrtb2.Imp)
}

type macroProvider struct {
	// macros map stores macros key values
	macros map[string]string
}

// NewBuilder returns the instance of macro buidler
func NewProvider(reqWrapper *openrtb_ext.RequestWrapper) Provider {

	macroProvider := &macroProvider{macros: map[string]string{}}
	macroProvider.populateRequestMacros(reqWrapper)
	return macroProvider
}

func (b *macroProvider) populateRequestMacros(reqWrapper *openrtb_ext.RequestWrapper) {
	reqExt, _ := reqWrapper.GetRequestExt()
	if reqExt != nil && reqExt.GetPrebid() != nil {
		maps.Copy(b.macros, reqExt.GetPrebid().Macros)
	}

	if reqWrapper.App.Bundle != "" {
		b.macros[AppBundleKey] = reqWrapper.App.Bundle
	}

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

func (b *macroProvider) GetMacro(key string) string {
	return b.macros[key]
}
func (b *macroProvider) GetAllMacros(keys []string) map[string]string {
	macroValues := map[string]string{}

	for _, key := range keys {
		macroValues[key] = b.macros[key]
	}
	return macroValues
}
func (b *macroProvider) SetContext(bid *openrtb2.Bid, imp *openrtb2.Imp) {
	b.resetcontext()
	b.macros[BidIDKey] = bid.ID
}
func (b *macroProvider) resetcontext() {
	for _, key := range bidLevelKeys {
		delete(b.macros, key)
	}
}
