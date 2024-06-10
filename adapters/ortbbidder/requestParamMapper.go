package ortbbidder

import (
	"encoding/json"
	"fmt"

	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

const (
	impKey     = "imp"
	extKey     = "ext"
	bidderKey  = "bidder"
	appsiteKey = "appsite"
	siteKey    = "site"
	appKey     = "app"
)

// setRequestParams updates the requestBody based on the requestParams mapping details.
func setRequestParams(requestBody []byte, requestParams map[string]bidderparams.BidderParamMapper) ([]byte, error) {
	if len(requestParams) == 0 {
		return requestBody, nil
	}
	request := map[string]any{}
	err := json.Unmarshal(requestBody, &request)
	if err != nil {
		return nil, err
	}
	imps, ok := request[impKey].([]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_imp_found_in_requestbody], imp:[%v]", request[impKey])
	}
	updatedRequest := false
	for ind, imp := range imps {
		request[impKey] = imp
		imp, ok := imp.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_imp_found_in_implist], imp:[%v]", request[impKey])
		}
		ext, ok := imp[extKey].(map[string]any)
		if !ok {
			continue
		}
		bidderParams, ok := ext[bidderKey].(map[string]any)
		if !ok {
			continue
		}
		for paramName, paramValue := range bidderParams {
			paramMapper, ok := requestParams[paramName]
			if !ok {
				continue
			}
			// set the value in the request according to the mapping details and remove the parameter.
			if util.SetValue(request, paramMapper.GetLocation(), paramValue) {
				delete(bidderParams, paramName)
				updatedRequest = true
			}
		}
		imps[ind] = request[impKey]
	}
	// update the impression list in the request
	request[impKey] = imps
	// if the request was modified, marshal it back to JSON.
	if updatedRequest {
		requestBody, err = json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("error:[fail_to_update_request_body] msg:[%s]", err.Error())
		}
	}
	return requestBody, nil
}
