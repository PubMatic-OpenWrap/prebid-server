package ortbbidder

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// struct to build the request for multi request mode where single request supports multiple impressions
type singleRequestBuilder struct {
	requestBuilderImpl
	newRequest map[string]any
	imps       []map[string]any
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

	imps, ok := rb.newRequest[impKey].([]map[string]any)
	if !ok || len(imps) == 0 {
		//TODO append error
		return err
	}
	return
}

// makeRequest constructs the endpoint URL and maps the bidder-parameters in request to create the RequestData objects.
// it create single RequestData object for all impressions.
func (rb *singleRequestBuilder) makeRequest() (requestData []*adapters.RequestData, errs []error) {
	if len(rb.imps) == 0 {
		return nil, errs
	}

	var (
		endpoint   string
		newRequest map[string]any
		err        error
	)

	//step 1: get endpoint
	if endpoint, err = rb.getEndpoint(getImpExtBidderParams(rb.imps[0])); err != nil {
		errs = append(errs, newBadInputError(err.Error()))
		return nil, errs
	}

	//step 2: replace parameters
	// iterate through imps in reverse order to ensure setRequestParams prioritizes
	// the parameters from imp[0].ext.bidder over those from imp[1..N].ext.bidder.
	for index := len(rb.imps) - 1; index >= 0; index-- {
		setRequestParams(newRequest, getImpExtBidderParams(rb.imps[index]), rb.requestParams, []int{index})
	}

	//step 3: append new request data
	if requestData, err = appendRequestData(requestData, newRequest, endpoint); err != nil {
		errs = append(errs, newBadInputError(err.Error()))
	}
	return requestData, errs
}
