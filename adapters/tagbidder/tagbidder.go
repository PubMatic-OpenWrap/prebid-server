package tagbidder

import (
	"net/http"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
)

type ITagBidder interface {
	GetURI() string
	GetHeaders() http.Header
}

type TagBidder struct {
	ITagBidder
}

func NewTagBidder() *TagBidder {
	return &TagBidder{}
}

func (a *TagBidder) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo, bidderMacro IBidderMacro, bidderMapper Mapper, flags Flags) ([]*adapters.RequestData, []error) {
	macroProcessor := NewMacroProcessor(bidderMapper)

	bidderMacro.InitBidRequest(request)

	requestData := []*adapters.RequestData{}
	for i := range request.Imp {
		if err := bidderMacro.LoadImpression(&request.Imp[i]); nil != err {
			continue
		}
		
		uri := macroProcessor.ProcessURL(a.ITagBidder.GetURI(), flags)

		requestData = append(requestData, &adapters.RequestData{
			ImpIndex: i,
			Method:   `GET`,
			Uri:      uri,
			Headers:  a.ITagBidder.GetHeaders(),
		})
	}

	return requestData, nil
}
