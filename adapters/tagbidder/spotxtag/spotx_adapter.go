package spotxtag

import (
	"fmt"
	"net/http"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters/tagbidder"
)

type SpotxAdapter struct {
	*tagbidder.TagBidder
	uri string
}

func NewSpotxAdapter(uri string) *SpotxAdapter {
	return &SpotxAdapter{
		TagBidder: tagbidder.NewTagBidder(),
		uri:       uri,
	}
}

func (a *SpotxAdapter) GetURI() string {
	return a.uri
}

func (a *SpotxAdapter) GetHeaders() http.Header {
	return http.Header{}
}

func (a *SpotxAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	//request validation can be done here independently

	return a.TagBidder.MakeRequests(
		request, reqInfo,
		NewSpotxMacro(),
		spotxMapper,
		tagbidder.Flags{RemoveEmptyParam: true})
}

func (a *SpotxAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	//response validation can be done here independently

	return nil, []error{
		fmt.Errorf("No Bid"),
	}
}
