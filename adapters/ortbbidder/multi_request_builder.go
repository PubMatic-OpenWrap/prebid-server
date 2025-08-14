package ortbbidder

import (
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v3/util/jsonutil"
)

// struct to build the multi requests each containing sinlge impression when requestType="single"
type multiRequestBuilder struct {
	requestBuilderImpl
	imps []map[string]any
}

// parseRequest parse the incoming request and populates intermediate fields required for building requestData object
func (rb *multiRequestBuilder) parseRequest(request *openrtb2.BidRequest) (err error) {
	if len(request.Imp) == 0 {
		return util.ErrImpMissing
	}

	//get rawrequests without impression objects
	tmpImp := request.Imp[0:]
	request.Imp = nil
	if rb.rawRequest, err = jsonutil.Marshal(request); err != nil {
		return err
	}
	request.Imp = tmpImp[0:]

	//cache impression from request
	data, err := jsonutil.Marshal(tmpImp)
	if err != nil {
		return err
	}
	if err = jsonutil.Unmarshal(data, &rb.imps); err != nil {
		return err
	}

	return nil
}

// makeRequest constructs the endpoint URL and maps the bidder-parameters in request to create the RequestData objects.
// it processes a request to generate 'N' RequestData objects, one for each of the 'N' impressions
func (rb *multiRequestBuilder) makeRequest() (requestData []*adapters.RequestData, errs []error) {
	var (
		endpoint             string
		newRequest           map[string]any
		err                  error
		requestCloneRequired bool
	)

	requestCloneRequired = true

	for index := range rb.imps {
		//step 1: clone request
		if requestCloneRequired {
			if newRequest, err = cloneRequest(rb.rawRequest); err != nil {
				errs = append(errs, util.NewBadInputError("%s", err.Error()))
				continue
			}
		}

		//step 2: get impression extension
		imp := rb.imps[index]
		bidderParams := getImpExtBidderParams(imp)

		//step 3: get endpoint
		if endpoint, err = rb.getEndpoint(bidderParams); err != nil {
			errs = append(errs, util.NewBadInputError("%s", err.Error()))
			continue
		}

		//step 4: update the request object by mapping bidderParams at expected location.
		newRequest[impKey] = []any{imp}
		requestCloneRequired = setRequestParams(newRequest, bidderParams, rb.requestParams, []int{0})

		if _, ok := imp[idKey].(string); !ok {
			errs = append(errs, fmt.Errorf("invalid imp found error while paring imp id at index :%d", index))
			continue
		}
		//step 5: append new request data
		if requestData, err = appendRequestData(requestData, newRequest, endpoint, []string{imp[idKey].(string)}); err != nil {
			errs = append(errs, util.NewBadInputError("%s", err.Error()))
		}
	}
	return requestData, errs
}
