package tagbidder

import (
	"errors"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

//TagBidder is default implementation of ITagBidder
type TagBidder struct {
	adapters.Bidder
	bidderName    string
	bidderConfig  *BidderConfig
	adapterConfig *config.Adapter
}

//NewTagBidder is an constructor for TagBidder
func NewTagBidder(bidderName openrtb_ext.BidderName, config config.Adapter) (*TagBidder, error) {
	obj := &TagBidder{
		bidderName:    string(bidderName),
		adapterConfig: &config,
		bidderConfig:  GetBidderConfig(string(bidderName)),
	}
	if nil == obj.bidderConfig {
		return nil, errors.New(`missing bidder config`)
	}
	return obj, nil
}

//NewTestTagBidder is an constructor for TagBidder
func NewTestTagBidder(bidderName openrtb_ext.BidderName, config config.Adapter) *TagBidder {
	obj := &TagBidder{
		bidderName:    string(bidderName),
		adapterConfig: &config,
		bidderConfig:  GetBidderConfig(string(bidderName)),
	}
	if nil == obj.bidderConfig {
		return nil
	}
	return obj
}

//MakeRequests will contains default definition for processing queries
func (a *TagBidder) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	bidderMacro, err := GetNewBidderMacro(a.bidderName)
	if nil != err {
		return nil, []error{err}
	}

	bidderMapper := GetBidderMapper(a.bidderName)
	if nil == bidderMapper {
		return nil, []error{errors.New(`missing bidder mapper`)}
	}

	macroProcessor := NewMacroProcessor(bidderMacro, bidderMapper)

	//Setting config parameters
	bidderMacro.SetAdapterConfig(a.adapterConfig)
	bidderMacro.SetBidderConfig(a.bidderConfig)
	bidderMacro.InitBidRequest(request)

	requestData := []*adapters.RequestData{}
	for i := range request.Imp {
		if err := bidderMacro.LoadImpression(&request.Imp[i]); nil != err {
			continue
		}

		uri := macroProcessor.ProcessURL(bidderMacro.GetURI(), a.bidderConfig.Flags)

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
	handler, err := GetResponseHandler(a.bidderConfig.ResponseType)
	if nil != err {
		return nil, []error{err}
	}
	return handler.MakeBids(internalRequest, externalRequest, response)
}

// Builder builds a new instance of the 33Across adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	return NewTagBidder(bidderName, config)
}
