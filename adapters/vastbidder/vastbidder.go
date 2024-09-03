package vastbidder

import (
	"time"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// VASTBidder is default implementation of ITagBidder
type VASTBidder struct {
	adapters.Bidder
	bidderName        openrtb_ext.BidderName
	adapterConfig     *config.Adapter
	fastXMLExperiment bool
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
	parser := getXMLParser(etreeXMLParserType)
	handler := newResponseHandler(internalRequest, externalRequest, response, parser)

	_start := time.Now()
	if err := handler.Validate(); len(err) > 0 {
		return nil, err[:]
	}

	responseData, errs := handler.MakeBids()

	if a.fastXMLExperiment && len(errs) == 0 {
		a.fastXMLTesting(
			newResponseHandler(internalRequest, externalRequest, response, getXMLParser(fastXMLParserType)),
			responseData,
			time.Since(_start))
	}

	return responseData, errs
}

func (a *VASTBidder) fastXMLTesting(handler *responseHandler, responseData *adapters.BidderResponse, etreeParserTime time.Duration) {
	var (
		handlerTime    time.Duration
		isVASTMismatch bool
	)

	_start := time.Now()
	if err := handler.Validate(); len(err) == 0 {
		_, errs := handler.MakeBids()
		handlerTime = time.Since(_start)
		if len(errs) > 0 {
			isVASTMismatch = true
		}
	}

	xmlParsingMetrics := &openrtb_ext.FastXMLMetrics{
		XMLParserTime:   handlerTime,
		EtreeParserTime: etreeParserTime,
		IsRespMismatch:  isVASTMismatch,
	}

	responseData.FastXMLMetrics = xmlParsingMetrics
}

// NewTagBidder is an constructor for TagBidder
func NewTagBidder(bidderName openrtb_ext.BidderName, config config.Adapter, enableFastXML bool) *VASTBidder {
	obj := &VASTBidder{
		bidderName:        bidderName,
		adapterConfig:     &config,
		fastXMLExperiment: enableFastXML,
	}
	return obj
}

// Builder builds a new instance of the 33Across adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, serverConfig config.Server) (adapters.Bidder, error) {
	return NewTagBidder(bidderName, config, serverConfig.EnableFastXML), nil
}
