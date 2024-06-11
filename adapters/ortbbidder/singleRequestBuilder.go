package ortbbidder

import (
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// struct to build the single request containing multi impressions when requestMode="multi"
type singleRequestBuilder struct {
	requestBuilderImpl
	newRequest map[string]any
	imps       []any
}

// parseRequest parse the incoming request and populates intermediate fields required for building requestData object
func (rb *singleRequestBuilder) parseRequest(request *openrtb2.BidRequest) (err error) {
	rb.rawRequest, err = jsonutil.Marshal(request)
	if err != nil {
		return err
	}

	rb.newRequest, err = cloneRequest(rb.rawRequest)
	if err != nil {
		return err
	}

	var ok bool
	rb.imps, ok = rb.newRequest[impKey].([]any)
	if !ok || len(rb.imps) == 0 {
		return errImpMissing
	}
	return
}

// makeRequest constructs the endpoint URL and maps the bidder-parameters in request to create the RequestData objects.
// it create single RequestData object for all impressions.
func (rb *singleRequestBuilder) makeRequest() (requestData []*adapters.RequestData, errs []error) {
	if len(rb.imps) == 0 {
		errs = append(errs, newBadInputError(errImpMissing.Error()))
		return
	}

	var (
		endpoint string
		err      error
	)

	//step 1: get endpoint
	imp, ok := rb.imps[0].(map[string]any)
	if !ok {
		errs = append(errs, newBadInputError("invalid imp found at index:0"))
		return nil, errs
	}
	if endpoint, err = rb.getEndpoint(getImpExtBidderParams(imp)); err != nil {
		errs = append(errs, newBadInputError(err.Error()))
		return nil, errs
	}

	//step 2: replace parameters
	// iterate through imps in reverse order to ensure setRequestParams prioritizes
	// the parameters from imp[0].ext.bidder over those from imp[1..N].ext.bidder.
	for index := len(rb.imps) - 1; index >= 0; index-- {
		imp, ok := rb.imps[index].(map[string]any)
		if !ok {
			errs = append(errs, newBadInputError(fmt.Sprintf("invalid imp found at index:%d", index)))
			continue // ignore particular impression
		}
		setRequestParams(rb.newRequest, getImpExtBidderParams(imp), rb.requestParams, []int{index})
	}

	//step 3: append new request data
	if requestData, err = appendRequestData(requestData, rb.newRequest, endpoint); err != nil {
		errs = append(errs, newBadInputError(err.Error()))
	}
	return requestData, errs
}
