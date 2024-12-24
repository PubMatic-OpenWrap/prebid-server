package ortbbidder

import (
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// struct to build the single request containing multi impressions when requestType="multi"
type singleRequestBuilder struct {
	requestBuilderImpl
	newRequest map[string]any
	imps       []map[string]any
	impIDs     []string
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

	imps, ok := rb.newRequest[impKey].([]any)
	if !ok {
		return util.ErrImpMissing
	}
	for index, imp := range imps {
		imp, ok := imp.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid imp found at index:%d", index)
		}

		if _, ok := imp[idKey].(string); !ok {
			return fmt.Errorf("invalid imp found error while paring imp id at index :%d", index)
		}
		rb.impIDs = append(rb.impIDs, imp[idKey].(string))
		rb.imps = append(rb.imps, imp)
	}
	return
}

// makeRequest constructs the endpoint URL and maps the bidder-parameters in request to create the RequestData objects.
// it create single RequestData object for all impressions.
func (rb *singleRequestBuilder) makeRequest() (requestData []*adapters.RequestData, errs []error) {
	if len(rb.imps) == 0 {
		errs = append(errs, util.NewBadInputError(util.ErrImpMissing.Error()))
		return
	}

	var (
		endpoint string
		err      error
	)

	//step 1: get endpoint
	if endpoint, err = rb.getEndpoint(getImpExtBidderParams(rb.imps[0])); err != nil {
		errs = append(errs, util.NewBadInputError(err.Error()))
		return
	}

	//step 2: replace parameters
	// iterate through imps in reverse order to ensure setRequestParams prioritizes
	// the parameters from imp[0].ext.bidder over those from imp[1..N].ext.bidder.
	for index := len(rb.imps) - 1; index >= 0; index-- {
		setRequestParams(rb.newRequest, getImpExtBidderParams(rb.imps[index]), rb.requestParams, []int{index})
	}

	//step 3: append new request data
	if requestData, err = appendRequestData(requestData, rb.newRequest, endpoint, rb.impIDs); err != nil {
		errs = append(errs, util.NewBadInputError(err.Error()))
	}
	return
}
