package tagbidder

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/PubMatic-OpenWrap/openrtb"
)

func getHTTPRequest(tagBidder ITagBidder, mp *MacroProcessor, bidRequest *openrtb.BidRequest) {
	for i := range bidRequest.Imp {
		tagBidder.LoadImpression(&bidRequest.Imp[i])
		queryString := mp.Process(tagBidder.URL())
		fmt.Printf("Query:%v\n", queryString)
	}
}

func TestBidder(t *testing.T) {
	var bidRequest *openrtb.BidRequest
	tagBidder := NewSpotxBidderMacro(bidRequest)
	mapper := GetBidderMapper(tagBidder.Name())
	mp := NewMacroProcessor(mapper)

	getHTTPRequest(tagBidder, mp, bidRequest)
}

type IRequestHandler interface {
	GetHeaders() http.Header
	GetHTTP
}
