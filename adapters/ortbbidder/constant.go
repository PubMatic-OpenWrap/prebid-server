package ortbbidder

// constants required for oRTB adapter
const (
	impKey     = "imp"
	extKey     = "ext"
	bidderKey  = "bidder"
	appsiteKey = "appsite"
	siteKey    = "site"
	appKey     = "app"
)

const (
	urlMacroPrefix          = "{{."
	urlMacroNoValue         = "<no value>"
	multiRequestBuilderType = "multi"
	locationIndexMacro      = "#"
	endpointTemplate        = "endpointTemplate"
	templateOption          = "missingkey=zero"
	oRTBPrefix              = "owortb_"
)

// constants to retrieve values from oRTB request/response
const (
	seatBidKey      = "seatbid"
	bidKey          = "bid"
	ortbCurrencyKey = "cur"
)

// constants to set values in adapter response
const (
	currencyKey = "Currency"
	typedbidKey = "Bid"
	bidsKey     = "Bids"
)
