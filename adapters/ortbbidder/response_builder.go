package ortbbidder

import (
	"encoding/json"
	"fmt"

	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/resolver"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

type responseBuilder struct {
	bidderResponse map[string]any
	adapterRespone map[string]any
	responseParams map[string]bidderparams.BidderParamMapper
}

func newResponseBuilder(responseParams map[string]bidderparams.BidderParamMapper) *responseBuilder {
	if responseParams == nil {
		responseParams = make(map[string]bidderparams.BidderParamMapper)
	}
	return &responseBuilder{
		responseParams: responseParams,
	}
}

// parseResponse parses the bidder response from the given JSON raw message.
// It unmarshals the JSON into the rb.bidderResponse struct.
func (rb *responseBuilder) parseResponse(bidderResponseBytes json.RawMessage) (err error) {
	err = jsonutil.UnmarshalValid(bidderResponseBytes, &rb.bidderResponse)
	return
}

// buildResponse builds the adapter response based on the given response parameters.
// It resolves the response level and bid level parameters using the provided responseParams map.
// The resolved response is stored in the adapterResponse map.
// If any invalid seatbid or bid is found in the response, an error is returned.
func (rb *responseBuilder) buildResponse() error {
	// Create a new ParamResolver with the bidder response.
	paramResolver := resolver.NewParamResolver(rb.bidderResponse)

	// Initialize the adapter response with the currency from the bidder response.
	adapterResponse := map[string]interface{}{
		"Currency": rb.bidderResponse["cur"],
	}

	// Loop over the response level parameters.
	// If the parameter exists in the response parameters, resolve it.s
	for _, paramName := range resolver.ResponseLevelParams {
		if paramMapper, ok := rb.responseParams[paramName]; ok {
			paramResolver.Resolve(rb.bidderResponse, adapterResponse, paramMapper.GetPath(), paramName)
		}
	}

	// Extract the seat bids from the bidder response.
	seatBids, ok := rb.bidderResponse["seatbid"].([]interface{})
	if !ok {
		return fmt.Errorf("error:[invalid_seatbid_found_in_responsebody], seatbid:[%v]", rb.bidderResponse["seatbid"])
	}
	// Initialize the list of type bids.
	typeBids := make([]interface{}, 0)
	for seatIndex, seatBid := range seatBids {
		seatBid, ok := seatBid.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error:[invalid_seatbid_found_in_seatbids_list], seatbid:[%v]", seatBids)
		}
		bids, ok := seatBid["bid"].([]interface{})
		if !ok {
			return fmt.Errorf("error:[invalid_bid_found_in_seatbid], bid:[%v]", seatBid["bid"])
		}
		for bidIndex, bid := range bids {
			bid, ok := bid.(map[string]interface{})
			if !ok {
				return fmt.Errorf("error:[invalid_bid_found_in_bids_list], bid:[%v]", seatBid["bid"])
			}
			// Initialize the type bid with the bid.
			typeBid := map[string]interface{}{
				"Bid": bid,
			}
			// Loop over the bid level parameters.
			// If the parameter exists in the response parameters, resolve it.
			for _, paramName := range resolver.BidLevelParams {
				if paramMapper, ok := rb.responseParams[paramName]; ok {
					path := util.GetPath(paramMapper.GetPath(), []int{seatIndex, bidIndex})
					paramResolver.Resolve(bid, typeBid, path, paramName)
				}
			}
			// Add the type bid to the list of type bids.
			typeBids = append(typeBids, typeBid)
		}
	}
	// Add the type bids to the adapter response.
	adapterResponse["Bids"] = typeBids
	// Set the adapter response in the response builder.
	rb.adapterRespone = adapterResponse
	return nil
}

// convertToAdapterResponse converts the responseBuilder's adapter response to a prebid format.
// Returns the BidderResponse and any error encountered during the conversion.
func (rb *responseBuilder) convertToAdapterResponse() (resp *adapters.BidderResponse, err error) {
	var adapterResponeBytes json.RawMessage
	adapterResponeBytes, err = jsonutil.Marshal(rb.adapterRespone)
	if err != nil {
		return
	}

	err = jsonutil.UnmarshalValid(adapterResponeBytes, &resp)
	return
}
