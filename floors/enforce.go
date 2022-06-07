package floors

import (
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func ShouldEnforceFloors(bidRequest *openrtb2.BidRequest, floorExt *openrtb_ext.PriceFloorRules, configEnforceRate int, f func(int) int) bool {

	if floorExt != nil && floorExt.Skipped != nil && *floorExt.Skipped {
		var floorInRequest bool
		for i := range bidRequest.Imp {
			if bidRequest.Imp[i].BidFloor > 0 {
				floorInRequest = true
				break
			}
		}
		if !floorInRequest {
			return floorInRequest
		}
	}

	if floorExt.Enforcement != nil && floorExt.Enforcement.EnforcePBS != nil && !*floorExt.Enforcement.EnforcePBS {
		return *floorExt.Enforcement.EnforcePBS
	}

	if floorExt.Enforcement != nil && floorExt.Enforcement.EnforceRate > 0 {
		configEnforceRate = floorExt.Enforcement.EnforceRate
	}

	shouldEnforce := configEnforceRate > f(ENFORCE_RATE_MAX+1)
	if floorExt.Enforcement == nil {
		floorExt.Enforcement = new(openrtb_ext.PriceFloorEnforcement)
	}
	floorExt.Enforcement.EnforcePBS = &shouldEnforce

	return shouldEnforce
}
