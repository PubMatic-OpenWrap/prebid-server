package processor

import (
	"net/url"
	"strconv"
	"time"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	BidIDKey          = "PBS-BIDID"
	AppBundleKey      = "PBS-APPBUNDLE"
	DomainKey         = "PBS-DOMAIN"
	PubDomainkey      = "PBS-PUBDOMAIN"
	PageURLKey        = "PBS-PAGEURL"
	AccountIDKey      = "PBS-ACCOUNTID"
	LmtTrackingKey    = "PBS-LIMITADTRACKING"
	ConsentKey        = "PBS-GDPRCONSENT"
	CustomMacroPrefix = "PBS-MACRO_"
	BidderKey         = "PBS-BIDDER"
	IntegrationKey    = "PBS-INTEGRATION"
	VastCRTIDKey      = "PBS-VASTCRTID"
	LineIDKey         = "PBS-LINEID"
	TimestampKey      = "PBS-TIMESTAMP"
	AuctionIDKey      = "PBS-AUCTIONID"
	ChannelKey        = "PBS-CHANNEL"
	EventTypeKey      = "PBS-EVENTTYPE"
	VastEventKey      = "PBS-VASTEVENT"
)

var (
	bidLevelKeys = []string{BidIDKey, BidderKey, VastEventKey, EventTypeKey, LineIDKey, VastCRTIDKey}
)

type MacroContext struct {
	Bid            *openrtb2.Bid
	Imp            *openrtb2.Imp
	Seat           string
	VastCreativeID string
	VastEventType  config.TrackingEventType
	EventElement   config.VASTEventElement
}

type Provider interface {
	// GetMacro returns the macro value for the given macro key
	GetMacro(key string) string
	// GetAllMacros return all the macros
	GetAllMacros(keys []string) map[string]string
	// SetContext set the bid and imp for the current provider
	SetContext(ctx MacroContext)
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
	b.macros[TimestampKey] = strconv.Itoa(int(time.Now().Unix()))
	reqExt, _ := reqWrapper.GetRequestExt()
	if reqExt != nil && reqExt.GetPrebid() != nil {
		for key, value := range reqExt.GetPrebid().Macros {
			customMacroKey := CustomMacroPrefix + key       // Adding prefix PBS-MACRO to custom macro keys
			b.macros[customMacroKey] = truncate(value, 100) // limit the custom macro value  to 100 chars only
		}

		b.macros[IntegrationKey] = reqExt.GetPrebid().Integration
		channel := reqExt.GetPrebid().Channel
		if channel != nil {
			b.macros[ChannelKey] = channel.Name
		}

	}
	b.macros[AuctionIDKey] = reqWrapper.ID
	if reqWrapper.App != nil && reqWrapper.App.Bundle != "" {
		b.macros[AppBundleKey] = reqWrapper.App.Bundle
	}

	if reqWrapper.App != nil && reqWrapper.App.Domain != "" {
		b.macros[DomainKey] = reqWrapper.App.Domain
	}

	if reqWrapper.Site != nil && reqWrapper.Site.Domain != "" {
		b.macros[DomainKey] = reqWrapper.Site.Domain
	}

	if reqWrapper.Site != nil && reqWrapper.Site.Publisher != nil && reqWrapper.Site.Publisher.Domain != "" {
		b.macros[PubDomainkey] = reqWrapper.Site.Publisher.Domain
	}

	if reqWrapper.App != nil && reqWrapper.App.Publisher != nil && reqWrapper.App.Publisher.Domain != "" {
		b.macros[PubDomainkey] = reqWrapper.App.Publisher.Domain
	}

	if reqWrapper.Site != nil {
		b.macros[PageURLKey] = reqWrapper.Site.Page
	}
	userExt, _ := reqWrapper.GetUserExt()
	if userExt != nil && userExt.GetConsent() != nil {
		b.macros[ConsentKey] = *userExt.GetConsent()
	}
	if reqWrapper.Device != nil && reqWrapper.Device.Lmt != nil {
		b.macros[LmtTrackingKey] = strconv.Itoa(int(*reqWrapper.Device.Lmt))
	}

	b.macros[AccountIDKey] = reqWrapper.ID
	if reqWrapper.Site != nil && reqWrapper.Site.Publisher != nil && reqWrapper.Site.Publisher.ID != "" {
		b.macros[AccountIDKey] = reqWrapper.Site.Publisher.ID
	}

	if reqWrapper.App != nil && reqWrapper.App.Publisher != nil && reqWrapper.App.Publisher.ID != "" {
		b.macros[AccountIDKey] = reqWrapper.App.Publisher.ID
	}
}

func (b *macroProvider) GetMacro(key string) string {
	return url.QueryEscape(b.macros[key])
}
func (b *macroProvider) GetAllMacros(keys []string) map[string]string {
	macroValues := map[string]string{}

	for _, key := range keys {
		macroValues[key] = url.QueryEscape(b.macros[key]) // encoding the macro values
	}
	return macroValues
}
func (b *macroProvider) SetContext(ctx MacroContext) {
	b.resetcontext()
	b.macros[BidIDKey] = ctx.Bid.ID
	b.macros[BidderKey] = ctx.Seat
	b.macros[VastCRTIDKey] = ctx.VastCreativeID
	b.macros[LineIDKey] = ctx.Bid.CID
	b.macros[VastEventKey] = string(ctx.EventElement)
	b.macros[EventTypeKey] = string(ctx.VastEventType)
}
func (b *macroProvider) resetcontext() {
	for _, key := range bidLevelKeys {
		delete(b.macros, key)
	}
}

func truncate(text string, width int) string {
	if width < 0 {
		return text
	}

	r := []rune(text)
	if len(r) < width {
		return text
	}
	trunc := r[:width]
	return string(trunc)
}

// macro1=##PBS-BIDID##&macro2=##PBS-APPBUNDLE##&macro3=##PBS-DOMAIN##&macro4=##PBS-PUBDOMAIN##&macro5=##PBS-PAGEURL##&macro6=##PBS-ACCOUNTID##&macro7=##PBS-LIMITADTRACKING##&macro8=##PBS-GDPRCONSENT##&macro9=##PBS-MACRO_##&macro10=##PBS-BIDDER##&macro11=##PBS-INTEGRATION##&macro12=##PBS-VASTCRTID##&macro13=##PBS-LINEID##&macro14=##PBS-TIMESTAMP##&macro15=##PBS-AUCTIONID##&macro16=##PBS-CHANNEL##&macro17=##PBS-EVENTTYPE##&macro18=##PBS-VASTEVENT##
