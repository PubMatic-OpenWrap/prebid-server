package tagbidder

import (
	"errors"
	"net/http"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
)

//ITagBidder interface will be used for specific bidder to set their headers
type ITagBidder interface {
	GetURI() string
	GetHeaders() http.Header
}

//TagBidder is default implementation of ITagBidder
type TagBidder struct {
	ITagBidder
	bidderName string
	flags      Flags
}

//NewTagBidder is an constructor for TagBidder
func NewTagBidder(bidderName string, flags Flags) *TagBidder {
	return &TagBidder{
		bidderName: bidderName,
		flags:      flags,
	}
}

//MakeRequests will contains default definition for processing queries
func (a *TagBidder) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	bidderMacro := GetNewBidderMacro(a.bidderName)
	if nil == bidderMacro {
		return nil, []error{errors.New(`invalid bidder macro defined`)}
	}

	bidderMapper := GetBidderMapper(a.bidderName)

	macroProcessor := NewMacroProcessor(bidderMacro, bidderMapper)

	bidderMacro.InitBidRequest(request)

	requestData := []*adapters.RequestData{}
	for i := range request.Imp {
		if err := bidderMacro.LoadImpression(&request.Imp[i]); nil != err {
			continue
		}

		uri := macroProcessor.ProcessURL(a.ITagBidder.GetURI(), a.flags)

		requestData = append(requestData, &adapters.RequestData{
			ImpIndex: i,
			Method:   `GET`,
			Uri:      uri,
			Headers:  a.ITagBidder.GetHeaders(),
		})
	}

	return requestData, nil
}

//MakeBids makes bids
func (a *TagBidder) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	//response validation can be done here independently
	var handler ITagResponseHandler
	handler = NewVASTTagResponseHandler()
	return handler.MakeBids(internalRequest, externalRequest, response)
}

//RegisterNewTagBidder will register new tag bidder
func RegisterNewTagBidder(bidderName string, bidderMacro func() IBidderMacro) {
	RegisterNewBidderMacroInitializer(bidderName, bidderMacro)
}
