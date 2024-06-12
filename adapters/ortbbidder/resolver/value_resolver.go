package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
)

// valueResolver is a generic resolver to get values from the response node using location
type valueResolver struct{}

func (r *valueResolver) getUsingBidderParamLocation(responseNode map[string]interface{}, path string) (interface{}, bool) {
	return util.GetValueFromLocation(responseNode, path)
}
