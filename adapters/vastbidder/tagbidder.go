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
	handler := GetTagRequestHandler(a.bidderName, a.adapterConfig)
	return handler.MakeRequests(request, reqInfo)
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
