package ortbbidder

import (
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/util"
)

// setRequestParams updates the request object by mapping bidderParams at expected location.
func setRequestParams(request, params map[string]any, paramsMapper map[string]bidderparams.BidderParamMapper, paramIndices []int) bool {
	updatedRequest := false
	for paramName, paramValue := range params {
		paramMapper, ok := paramsMapper[paramName]
		if !ok {
			continue
		}
		// add index in path by replacing # macro
		location := addIndicesInPath(paramMapper.Location, paramIndices)
		// set the value in the request according to the mapping details
		// remove the parameter from bidderParams after successful mapping
		if util.SetValue(request, location, paramValue) {
			delete(params, paramName)
			updatedRequest = true
		}
	}
	return updatedRequest
}

// addIndicesInPath updates the path by replacing # by arrayIndices
func addIndicesInPath(path string, indices []int) []string {
	parts := strings.Split(path, ".")
	j := 0
	for i, part := range parts {
		if part == locationIndexMacro {
			if j >= len(indices) {
				break
			}
			parts[i] = strconv.Itoa(indices[j])
			j++
		}
	}
	return parts
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
