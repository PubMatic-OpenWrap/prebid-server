package vastbidder

import (
	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

// ITagRequestHandler describes creating new requests for external demands
type ITagRequestHandler interface {
	MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error)
}

//TagRequestHandler default implementation of request handling
type TagRequestHandler struct {
	ITagRequestHandler
	bidderName    openrtb_ext.BidderName
	adapterConfig *config.Adapter
}

//MakeRequests will contains default definition for processing queries
func (a *TagRequestHandler) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	bidderMacro := GetBidderMacro(a.bidderName)
	bidderMapper := GetBidderMapper(a.bidderName)
	macroProcessor := GetBidderMacroProcessor(a.bidderName, bidderMacro, bidderMapper)

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

//GetTagRequestHandler factory method to get any request handler
func GetTagRequestHandler(bidderName openrtb_ext.BidderName, config *config.Adapter) ITagRequestHandler {
	return &TagRequestHandler{
		bidderName:    bidderName,
		adapterConfig: config,
	}
}
