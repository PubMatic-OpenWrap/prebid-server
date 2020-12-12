package spotxtag

import (
	"net/http"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters/tagbidder"
)

//SpotxAdapter partner adapter
type SpotxAdapter struct {
	*tagbidder.TagBidder
	uri string
}

//NewSpotxAdapter new object
func NewSpotxAdapter(bidderName, uri string, flags tagbidder.Flags) *SpotxAdapter {
	return &SpotxAdapter{
		TagBidder: tagbidder.NewTagBidder(bidderName, flags),
		uri:       uri,
	}
}

//GetURI get URL
func (a *SpotxAdapter) GetURI() string {
	return a.uri
}

//GetHeaders GetHeaders
func (a *SpotxAdapter) GetHeaders() http.Header {
	return http.Header{}
}

//MakeRequests make new requests
func (a *SpotxAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	//request validation can be done here independently
	return a.TagBidder.MakeRequests(request, reqInfo)
}

//MakeBids makes bids
func (a *SpotxAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	//response validation can be done here independently
	var handler tagbidder.ITagResponseHandler
	handler = tagbidder.NewVASTTagResponseHandler()
	return handler.MakeBids(internalRequest, externalRequest, response)
}

func init() {
	tagbidder.RegisterNewTagBidder(`spotx`, NewSpotxMacro, spotxMapperJSON)
}
