package openwrap

import (
	"fmt"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type OpenwrapAdapter struct {
	Endpoint string
}

func (o *OpenwrapAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	fmt.Println("Make Request")
	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     o.Endpoint,
		Body:    nil,
		Headers: nil,
	}}, nil
}

func (o *OpenwrapAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Println("Make Bids")
	return nil, nil
}

// Builder builds a new instance of the Pubmatic adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &OpenwrapAdapter{
		Endpoint: config.Endpoint,
	}
	return bidder, nil
}
