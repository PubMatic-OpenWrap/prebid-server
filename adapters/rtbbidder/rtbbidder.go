package rtbbidder

import (
	"encoding/json"
	"fmt"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/prebid/openrtb/v19/openrtb2"
)

// reference: https://docs.prebid.org/prebid-server/developers/add-new-bidder-go.html
type RTBBidder struct {
	RequestMode RequestMode
	Uri         string
}

type RequestMode int

const (
	Multi RequestMode = iota // default is multi
	Single
)

// oRTB 2.6
func (r *RTBBidder) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	fmt.Println("Making RTB Requests")
	requestData := make([]*adapters.RequestData, 0)
	var errs []error
	var bidderUrl string = ""
	/* Iterate over each impression and determine this bidder specific param value */
	for _, imp := range request.Imp {
		impExt := adapters.ExtImpBidder{}
		if err := json.Unmarshal(imp.Ext, &impExt); err != nil {
			errs = append(errs, err)
		}
		paramsMap := map[string]string{}
		json.Unmarshal(impExt.Bidder, &paramsMap)
		if bidderUrl == "" {
			// assuming same url is present across all bidder params
			bidderUrl = paramsMap["uri"]
		}
		if r.RequestMode == Single {
			clonedRequest := request
			clonedRequest.Imp = []openrtb2.Imp{imp}
			body, err := json.Marshal(clonedRequest)
			if err == nil {
				requestData = append(requestData, &adapters.RequestData{
					Method: "POST",
					Uri:    paramsMap["uri"],
					Body:   body,
				})
			}
		}
		if r.RequestMode == Multi {
			body, err := json.Marshal(request)
			if err == nil {
				requestData = append(requestData, &adapters.RequestData{
					Method: "POST",
					Uri:    bidderUrl,
					Body:   body,
				})
			}
			break
		}
	}

	return requestData, errs
}

func (r *RTBBidder) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Println("Making RTB Response")
	var ortbResponse *openrtb2.BidResponse
	var errors []error
	var bidderResponse *adapters.BidderResponse
	if err := json.Unmarshal(response.Body, &ortbResponse); err != nil {
		errors = append(errors, err)
	} else {
		if ortbResponse != nil {
			bidderResponse = &adapters.BidderResponse{
				Currency: ortbResponse.Cur,
				Bids:     getTypeBids(*ortbResponse),
			}
		}
	}
	return bidderResponse, errors
}

// Builder builds a new instance of the Pubmatic adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	bidder := &RTBBidder{
		Uri: config.Endpoint,
	}
	return bidder, nil
}

func getTypeBids(bidResponse openrtb2.BidResponse) []*adapters.TypedBid {
	tBids := make([]*adapters.TypedBid, 0)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			tBids = append(tBids, &adapters.TypedBid{
				Bid: &bid,
			})
		}
	}
	return tBids
}
