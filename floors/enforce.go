package floors

import (
	"github.com/prebid/prebid-server/openrtb_ext"
)

func ShouldEnforceFloors(requestExt *openrtb_ext.PriceFloorRules, configEnforceRate int, f func(int) int) bool {

	if requestExt.Enforcement != nil && !requestExt.Enforcement.EnforcePBS {
		return false
	}

	if requestExt.Enforcement != nil && requestExt.Enforcement.EnforceRate > 0 {
		configEnforceRate = requestExt.Enforcement.EnforceRate
	}

	return configEnforceRate > f(ENFORCE_RATE_MAX+1)
}
