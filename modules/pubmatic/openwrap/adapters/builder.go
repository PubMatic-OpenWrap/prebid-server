package adapters

import (
	"encoding/json"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// BidderParameters provides all properties requires for bidder to generate bidder json
type BidderParameters struct {
	//AdapterName, BidderCode should be passed in builder function
	ReqID                   string
	AdapterName, BidderCode string
	ImpExt                  *models.ImpExtension

	//bidder specific parameters
	FieldMap      JSONObject
	Width, Height *int64
	SlotKey       string
}

// JSONObject generic JSON object
type JSONObject = map[string]interface{}

// builder callback type
type builder func(params BidderParameters) (json.RawMessage, error)

// bidderBuilderFactor
var _bidderBuilderFactory map[string]builder

// initBidderBuilderFactory initialise all hard coded bidder builder
func initBidderBuilderFactory() {
	_bidderBuilderFactory = map[string]builder{
		string(openrtb_ext.BidderAdform):         builderAdform,
		string(openrtb_ext.BidderAdf):            builderAdform,
		string(openrtb_ext.BidderAppnexus):       builderAppNexus,
		string(openrtb_ext.BidderBeachfront):     builderBeachfront,
		string(openrtb_ext.BidderCriteo):         builderCriteo,
		string(openrtb_ext.BidderGumGum):         builderGumGum,
		string(openrtb_ext.BidderImprovedigital): builderImproveDigital,
		string(openrtb_ext.BidderIx):             builderIndex,
		string(openrtb_ext.BidderOpenx):          builderOpenx,
		string(openrtb_ext.BidderOutbrain):       builderOutbrain,
		string(openrtb_ext.BidderPangle):         builderPangle,
		string(openrtb_ext.BidderPubmatic):       builderPubMatic, /*this only gets used incase of hybrid case*/
		string(openrtb_ext.BidderPulsepoint):     builderPulsePoint,
		string(openrtb_ext.BidderRubicon):        builderRubicon,
		string(openrtb_ext.BidderSmaato):         builderSmaato,
		string(openrtb_ext.BidderSmartAdserver):  builderSmartAdServer,
		string(openrtb_ext.BidderSonobi):         builderSonobi,
		string(openrtb_ext.BidderSovrn):          builderSovrn,
		string(openrtb_ext.BidderApacdex):        builderApacdex,
	}
}

// getBuilder will return core bidder hard coded builder, if not found then returns default builder
func getBuilder(adapterName string) builder {
	//resolve hardcoded bidder alias
	adapterName = ResolveOWBidder(adapterName)

	if callback, ok := _bidderBuilderFactory[adapterName]; ok {
		return callback
	}
	return defaultBuilder
}

// InitBidders will initialise bidder alias, default bidder parameter json and builders for each bidder
func InitBidders(cfg config.Config) error {
	initBidderBuilderFactory()
	return parseBidderParams(cfg)
}
