package ortbbidder

import (
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
)

// setRequestParams updates the request object by mapping bidderParams at expected location.
func setRequestParams(request, params map[string]any, paramsMapper map[string]bidderparams.BidderParamMapper, paramIndices []int) {
	for paramName, paramValue := range params {
		paramMapper, ok := paramsMapper[paramName]
		if !ok {
			continue
		}
		// add index in path by replacing # macro
		location := addIndicesInPath(paramMapper.GetLocation(), paramIndices)
		// set the value in the request according to the mapping details
		// remove the parameter from bidderParams after successful mapping
		if setValue(request, location, paramValue) {
			delete(params, paramName)
		}
	}
}

// addIndicesInPath updates the path by replacing # by arrayIndices
func addIndicesInPath(path string, indices []int) []string {
	parts := strings.Split(path, ".")
	j := 0
	for i, part := range parts {
		if part == "#" {
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
