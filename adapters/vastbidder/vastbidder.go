package vastbidder

import (
	"encoding/base64"
	"time"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

// VASTBidder is default implementation of ITagBidder
type VASTBidder struct {
	adapters.Bidder
	bidderName    openrtb_ext.BidderName
	adapterConfig *config.Adapter
}

// MakeRequests will contains default definition for processing queries
func (a *VASTBidder) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	bidderMacro := GetNewBidderMacro(a.bidderName)
	bidderMapper := GetDefaultMapper()
	macroProcessor := NewMacroProcessor(bidderMacro, bidderMapper)

	//Setting config parameters
	//bidderMacro.SetBidderConfig(a.bidderConfig)
	bidderMacro.SetAdapterConfig(a.adapterConfig)
	bidderMacro.InitBidRequest(request)

	requestData := []*adapters.RequestData{}
	for impIndex := range request.Imp {
		bidderExt, err := bidderMacro.LoadImpression(&request.Imp[impIndex])
		if nil != err {
			continue
		}

		//iterate each vast tags, and load vast tag
		for vastTagIndex, tag := range bidderExt.Tags {
			//load vasttag
			bidderMacro.LoadVASTTag(tag)

			//Setting Bidder Level Keys
			bidderKeys := bidderMacro.GetBidderKeys()
			macroProcessor.SetBidderKeys(bidderKeys)

			uri := macroProcessor.Process(bidderMacro.GetURI())

			// append custom headers if any
			headers := bidderMacro.getAllHeaders()

			requestData = append(requestData, &adapters.RequestData{
				Params: &adapters.BidRequestParams{
					ImpIndex:     impIndex,
					VASTTagIndex: vastTagIndex,
				},
				Method:  `GET`,
				Uri:     uri,
				Headers: headers,
			})
		}
	}

	return requestData, nil
}

// MakeBids makes bids
func (a *VASTBidder) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	parser := getXMLParser()
	handler := newResponseHandler(internalRequest, externalRequest, response, parser)

	_start := time.Now()
	if err := handler.Validate(); len(err) > 0 {
		return nil, err[:]
	}

	responseData, errs := handler.MakeBids()
	if len(errs) > 0 {
		openrtb_ext.XMLLogf(openrtb_ext.XMLLogFormat, handler.parser.Name(), "vastbidder", base64.StdEncoding.EncodeToString(handler.response.Body))
		return nil, errs[:]
	}

	responseData.XMLMetrics = &openrtb_ext.XMLMetrics{
		ParserName:  parser.Name(),
		ParsingTime: time.Since(_start),
	}

	return responseData, errs
}

// NewTagBidder is an constructor for TagBidder
func NewTagBidder(bidderName openrtb_ext.BidderName, config config.Adapter) *VASTBidder {
	obj := &VASTBidder{
		bidderName:    bidderName,
		adapterConfig: &config,
	}
	return obj
}

// Builder builds a new instance of the 33Across adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, serverConfig config.Server) (adapters.Bidder, error) {
	return NewTagBidder(bidderName, config), nil
}
