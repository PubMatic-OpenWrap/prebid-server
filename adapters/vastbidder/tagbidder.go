package vastbidder

import (
	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

//TagBidder is default implementation of ITagBidder
type TagBidder struct {
	adapters.Bidder
	bidderName    openrtb_ext.BidderName
	adapterConfig *config.Adapter
}

//MakeRequests will contains default definition for processing queries
func (a *TagBidder) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	bidderMacro := GetNewBidderMacro(a.bidderName)
	bidderMapper := GetDefaultMapper()
	macroProcessor := NewMacroProcessor(bidderMacro, bidderMapper)

	//Setting config parameters
	//bidderMacro.SetBidderConfig(a.bidderConfig)
	bidderMacro.SetAdapterConfig(a.adapterConfig)
	bidderMacro.InitBidRequest(request)

	requestData := []*adapters.RequestData{}
	for i := range request.Imp {
		if err := bidderMacro.LoadImpression(&request.Imp[i]); nil != err {
			continue
		}

		//Setting Bidder Level Keys
		bidderKeys := bidderMacro.GetBidderKeys()
		macroProcessor.SetBidderKeys(bidderKeys)

		//uri := macroProcessor.ProcessURL(bidderMacro.GetURI(), a.bidderConfig.Flags)
		uri := macroProcessor.ProcessURL(bidderMacro.GetURI(), Flags{RemoveEmptyParam: true})

		requestData = append(requestData, &adapters.RequestData{
			ImpIndex: i,
			Method:   `GET`,
			Uri:      uri,
			Headers:  bidderMacro.GetHeaders(),
		})
	}

	return requestData, nil
}

//MakeBids makes bids
func (a *TagBidder) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	//response validation can be done here independently
	//handler, err := GetResponseHandler(a.bidderConfig.ResponseType)
	handler, err := GetResponseHandler(VASTTagResponseHandlerType)
	if nil != err {
		return nil, []error{err}
	}
	return handler.MakeBids(internalRequest, externalRequest, response)
}

//NewTagBidder is an constructor for TagBidder
func NewTagBidder(bidderName openrtb_ext.BidderName, config config.Adapter) *TagBidder {
	obj := &TagBidder{
		bidderName:    bidderName,
		adapterConfig: &config,
	}
	return obj
}

// Builder builds a new instance of the 33Across adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	return NewTagBidder(bidderName, config), nil
}
