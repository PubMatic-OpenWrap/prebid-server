package vastbidder

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// TagBidder is default implementation of ITagBidder
type TagBidder struct {
	adapters.Bidder
	bidderName    openrtb_ext.BidderName
	adapterConfig *config.Adapter
}

// MakeRequests will contains default definition for processing queries
func (a *TagBidder) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
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
func (a *TagBidder) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	handler := newResponseHandler(internalRequest, externalRequest, response)

	if err := handler.Validate(); len(err) > 0 {
		return nil, err[:]
	}

	return handler.MakeBids()
}

// NewTagBidder is an constructor for TagBidder
func NewTagBidder(bidderName openrtb_ext.BidderName, config config.Adapter) *TagBidder {
	obj := &TagBidder{
		bidderName:    bidderName,
		adapterConfig: &config,
	}
	return obj
}

// Builder builds a new instance of the 33Across adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, _ config.Server) (adapters.Bidder, error) {
	return NewTagBidder(bidderName, config), nil
}
