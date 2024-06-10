package ortbbidder

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// struct to build the request for single request mode where single imp is supported in a request
type multiRequestBuilder struct {
	requestBuilderImpl
	imps []map[string]any
}

// parseRequest parse the incoming request and populates intermediate fields required for building requestData object
func (rb *multiRequestBuilder) parseRequest(request *openrtb2.BidRequest) (err error) {
	if len(request.Imp) == 0 {
		//set errors
		return err
	}

	//get rawrequests without impression objects
	tmpImp := request.Imp[0:]
	request.Imp = nil
	if rb.rawRequest, err = jsonutil.Marshal(request); err != nil {
		return err
	}
	// request.Imp = tmpImp[0:] //resetting is not required

	//cache impression from request
	data, err := jsonutil.Marshal(tmpImp)
	if err != nil {
		return err
	}
	if err = jsonutil.Unmarshal(data, rb.imps); err != nil {
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
				continue
			}
		}

		//step 2: get impression extension
		// set "imp" object in request to empty to improve performance while creating deep copy of request
		imp := rb.imps[index]
		bidderParams := getImpExtBidderParams(imp)

		//step 3: get endpoint
		if endpoint, err = rb.getEndpoint(bidderParams); err != nil {
			continue
		}

		//step 4: update the request object by mapping bidderParams at expected location.
		newRequest[impKey] = []any{imp}
		requestCloneRequired = setRequestParams(newRequest, bidderParams, rb.requestParams, []int{0})

		//step 5: append new request data
		if requestData, err = appendRequestData(requestData, newRequest, endpoint); err != nil {
			errs = append(errs, newBadInputError(err.Error()))
		}
	}
	return requestData, errs
}
