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

// setPrebidBidderResponse determines and construct adapters.BidderResponse and adapters.TypedBid object with the help
// of response parameter mappings defined in static/bidder-response-params
func (rb *responseBuilder) setPrebidBidderResponse(bidderResponseBytes json.RawMessage) (errs []error) {

	err := jsonutil.UnmarshalValid(bidderResponseBytes, &rb.bidderResponse)
	if err != nil {
		return []error{util.NewBadServerResponseError(err.Error())}
	}
	// Create a new ParamResolver with the bidder response.
	paramResolver := resolver.New(rb.request, rb.bidderResponse)
	// Initialize the adapter response with the currency from the bidder response.
	adapterResponse := map[string]any{
		currencyKey: rb.bidderResponse[ortbCurrencyKey],
	}
	// Resolve the  adapter response level parameters.
	for _, param := range resolver.ResponseParams {
		bidderParam := rb.responseParams[param.String()]
		resolverErrors := paramResolver.Resolve(rb.bidderResponse, adapterResponse, bidderParam.Location, param)
		errs = append(errs, resolverErrors...)
	}
	// Extract the seat bids from the bidder response.
	seatBids, ok := rb.bidderResponse[seatBidKey].([]any)
	if !ok {
		return []error{util.NewBadServerResponseError("invalid seatbid array found in response, seatbids:[%v]", rb.bidderResponse[seatBidKey])}
	}
	// Initialize the list of typed bids.
	typedBids := make([]any, 0)
	for seatIndex, seatBid := range seatBids {
		seatBid, ok := seatBid.(map[string]any)
		if !ok {
			return []error{util.NewBadServerResponseError("invalid seatbid found in seatbid array, seatbid:[%v]", seatBids[seatIndex])}
		}
		bids, ok := seatBid[bidKey].([]any)
		if !ok {
			return []error{util.NewBadServerResponseError("invalid bid array found in seatbid, bids:[%v]", seatBid[bidKey])}
		}
		for bidIndex, bid := range bids {
			bid, ok := bid.(map[string]any)
			if !ok {
				return []error{util.NewBadServerResponseError("invalid bid found in bids array, bid:[%v]", bids[bidIndex])}
			}
			// Initialize the typed bid with the bid.
			typedBid := map[string]any{
				typedbidKey: bid,
			}
			// Resolve the typed bid level parameters.
			for _, params := range resolver.TypedBidParams {
				paramMapper := rb.responseParams[params.String()]
				location := util.ReplaceLocationMacro(paramMapper.Location, []int{seatIndex, bidIndex})
				resolverErrors := paramResolver.Resolve(bid, typedBid, location, params)
				errs = append(errs, resolverErrors...)
			}
			// Add the type bid to the list of typed bids.
			typedBids = append(typedBids, typedBid)
		}
	}
	// Add the type bids to the adapter response.
	adapterResponse[bidsKey] = typedBids
	// Set the adapter response in the response builder.
	rb.adapterRespone = adapterResponse
	return errs
}

// buildAdapterResponse converts the responseBuilder's adapter response to a prebid format.
// Returns the BidderResponse and any error encountered during the conversion.
func (rb *responseBuilder) buildAdapterResponse() (resp *adapters.BidderResponse, err error) {
	var adapterResponeBytes json.RawMessage
	adapterResponeBytes, err = jsonutil.Marshal(rb.adapterRespone)
	if err != nil {
		return nil, util.NewBadServerResponseError(err.Error())
	}

	err = jsonutil.UnmarshalValid(adapterResponeBytes, &resp)
	if err != nil {
		return nil, util.NewBadServerResponseError(err.Error())
	}
	return
}
