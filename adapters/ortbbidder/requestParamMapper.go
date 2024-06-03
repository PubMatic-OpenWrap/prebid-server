package ortbbidder

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
)

// setRequestParams updates the request and imp object by mapping bidderParams at expected location.
func setRequestParams(request, imp, bidderParams map[string]any, paramsMapper map[string]bidderparams.BidderParamMapper) {
	for paramName, paramValue := range bidderParams {
		paramMapper, ok := paramsMapper[paramName]
		if !ok {
			continue
		}
		// set the value in the request according to the mapping details
		// remove the parameter from bidderParams after successful mapping
		if setValue(request, imp, paramMapper.GetLocation(), paramValue) {
			delete(bidderParams, paramName)
		}
	}
}

// getImpExtBidderParams returns imp.ext.bidder parameters
func getImpExtBidderParams(imp map[string]any) map[string]any {
	ext, ok := imp[extKey].(map[string]any)
	if !ok {
		return nil
	}
	bidderParams, ok := ext[bidderKey].(map[string]any)
	if !ok {
		return nil
	}
	return bidderParams
}
