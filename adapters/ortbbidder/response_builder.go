package ortbbidder

import (
	"encoding/json"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/resolver"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

type responseBuilder struct {
	bidderResponse map[string]any                            // Raw response from the bidder.
	adapterRespone map[string]any                            // Response in the prebid format.
	responseParams map[string]bidderparams.BidderParamMapper // Bidder response parameters.
	request        *openrtb2.BidRequest                      // Bid request.
}

func newResponseBuilder(responseParams map[string]bidderparams.BidderParamMapper, request *openrtb2.BidRequest) *responseBuilder {
	return &responseBuilder{
		responseParams: responseParams,
		request:        request,
	}
}

// parseResponse parses the bidder response from the given JSON raw message.
// It unmarshals the JSON into the rb.bidderResponse struct.
func (rb *responseBuilder) parseResponse(bidderResponseBytes json.RawMessage) error {
	return jsonutil.UnmarshalValid(bidderResponseBytes, &rb.bidderResponse)
}

// setPrebidBidderResponse builds the adapter response based on the given response parameters.
// It resolves the response level and bid level parameters using the provided responseParams map.
// The resolved response is stored in the adapterResponse map.
// If any invalid seatbid or bid is found in the response, an error is returned.
func (rb *responseBuilder) setPrebidBidderResponse() error {
	// Create a new ParamResolver with the bidder response.
	paramResolver := resolver.New(rb.request, rb.bidderResponse)
	// Initialize the adapter response with the currency from the bidder response.
	adapterResponse := map[string]any{
		currencyKey: rb.bidderResponse[ortbCurrencyKey],
	}

	// Resolve the  adapter response level parameters.
	paramMapper := rb.responseParams[resolver.Fledge.String()]
	paramResolver.Resolve(rb.bidderResponse, adapterResponse, paramMapper.Location, resolver.Fledge)

	// Extract the seat bids from the bidder response.
	seatBids, ok := rb.bidderResponse[seatBidKey].([]any)
	if !ok {
		return newBadServerResponseError("invalid seatbid array found in response, seatbids:[%v]", rb.bidderResponse[seatBidKey])
	}
	// Initialize the list of type bids.
	typeBids := make([]any, 0)
	for seatIndex, seatBid := range seatBids {
		seatBid, ok := seatBid.(map[string]any)
		if !ok {
			return newBadServerResponseError("invalid seatbid found in seatbid array, seatbid:[%v]", seatBids[seatIndex])
		}
		bids, ok := seatBid[bidKey].([]any)
		if !ok {
			return newBadServerResponseError("invalid bid array found in seatbid, bids:[%v]", seatBid[bidKey])
		}
		for bidIndex, bid := range bids {
			bid, ok := bid.(map[string]any)
			if !ok {
				return newBadServerResponseError("invalid bid found in bids array, bid:[%v]", bids[bidIndex])
			}
			// Initialize the type bid with the bid.
			typeBid := map[string]any{
				typeBidKey: bid,
			}
			// Resolve the type bid level parameters.
			paramMapper := rb.responseParams[resolver.BidType.String()]
			path := util.ReplaceLocationMacro(paramMapper.Location, []int{seatIndex, bidIndex})
			paramResolver.Resolve(bid, typeBid, path, resolver.BidType)

			// Add the type bid to the list of type bids.
			typeBids = append(typeBids, typeBid)
		}
	}
	// Add the type bids to the adapter response.
	adapterResponse[bidsKey] = typeBids
	// Set the adapter response in the response builder.
	rb.adapterRespone = adapterResponse
	return nil
}

// buildAdapterResponse converts the responseBuilder's adapter response to a prebid format.
// Returns the BidderResponse and any error encountered during the conversion.
func (rb *responseBuilder) buildAdapterResponse() (resp *adapters.BidderResponse, err error) {
	var adapterResponeBytes json.RawMessage
	adapterResponeBytes, err = jsonutil.Marshal(rb.adapterRespone)
	if err != nil {
		return
	}

	err = jsonutil.UnmarshalValid(adapterResponeBytes, &resp)
	if err != nil {
		return nil, err
	}
	return
}
