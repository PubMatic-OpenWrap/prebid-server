package tracker

import (
	"math"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func getFloorsDetails(responseExt openrtb_ext.ExtBidResponse) (*int, int, *int, string) {
	var skipfloors, floorSource *int
	floorType, floorModelVersion := 0, ""
	if responseExt.Prebid != nil && responseExt.Prebid.Floors != nil {
		floors := responseExt.Prebid.Floors
		if floors.Skipped != nil {
			skipfloors = ptrutil.ToPtr(0)
			if *floors.Skipped {
				skipfloors = ptrutil.ToPtr(1)
			}
		}
		if floors.Data != nil && len(floors.Data.ModelGroups) > 0 {
			floorModelVersion = floors.Data.ModelGroups[0].ModelVersion
		}
		if len(floors.PriceFloorLocation) > 0 {
			if source, ok := models.FloorSourceMap[floors.PriceFloorLocation]; ok {
				floorSource = &source
			}
		}
		if floors.Enforcement != nil && floors.Enforcement.EnforcePBS != nil && *floors.Enforcement.EnforcePBS {
			floorType = models.HardFloor
		}
	}
	return skipfloors, floorType, floorSource, floorModelVersion
}

func getRewardedInventoryFlag(reward *int8) int {
	if reward != nil {
		return int(*reward)
	}
	return 0
}

// Round value to 2 digit
func roundToTwoDigit(value float64) float64 {
	output := math.Pow(10, float64(2))
	return float64(math.Round(value*output)) / output
}
