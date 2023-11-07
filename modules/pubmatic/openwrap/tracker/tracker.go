package tracker

import (
	"fmt"
	"net/url"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func GetTrackerInfo(rCtx models.RequestCtx, prebidExt *openrtb_ext.ExtResponsePrebid) string {

	tracker := models.Tracker{
		PubID:     rCtx.PubID,
		ProfileID: fmt.Sprintf("%d", rCtx.ProfileID),
		VersionID: fmt.Sprintf("%d", rCtx.DisplayID),
		PageURL:   rCtx.PageURL,
		Timestamp: rCtx.StartTime,
		IID:       rCtx.LoggerImpressionID,
		Platform:  int(rCtx.DevicePlatform),
		Origin:    rCtx.Origin,
		TestGroup: rCtx.ABTestConfigApplied,
	}

	setFloorsDetails(&tracker, prebidExt)
	constructedURLString := ConstructTrackerURL(rCtx, tracker)

	trackerURL, err := url.Parse(constructedURLString)
	if err != nil {
		return ""
	}

	params := trackerURL.Query()
	params.Set(models.TRKPartnerID, models.MacroPartnerName)
	params.Set(models.TRKBidderCode, models.MacroBidderCode)
	params.Set(models.TRKKGPV, models.MacroKGPV)
	params.Set(models.TRKGrossECPM, models.MacroGrossECPM)
	params.Set(models.TRKNetECPM, models.MacroNetECPM)
	params.Set(models.TRKBidID, models.MacroBidID)
	params.Set(models.TRKOrigBidID, models.MacroOrigBidID)
	params.Set(models.TRKSlotID, models.MacroSlotID)
	params.Set(models.TRKAdunit, models.MacroAdunit)
	params.Set(models.TRKRewardedInventory, models.MacroRewarded)
	trackerURL.RawQuery = params.Encode()

	return trackerURL.String()
}

// sets floors details in tracker
func setFloorsDetails(tracker *models.Tracker, prebidExt *openrtb_ext.ExtResponsePrebid) {
	if prebidExt != nil && prebidExt.Floors != nil {
		if prebidExt.Floors.Skipped != nil {
			skipfloors := ptrutil.ToPtr(0)
			if *prebidExt.Floors.Skipped {
				skipfloors = ptrutil.ToPtr(1)
			}
			tracker.FloorSkippedFlag = skipfloors
		}
		if prebidExt.Floors.Data != nil && len(prebidExt.Floors.Data.ModelGroups) > 0 {
			tracker.FloorModelVersion = prebidExt.Floors.Data.ModelGroups[0].ModelVersion
		}
		if len(prebidExt.Floors.PriceFloorLocation) > 0 {
			if source, ok := models.FloorSourceMap[prebidExt.Floors.PriceFloorLocation]; ok {
				tracker.FloorSource = &source
			}
		}
		if prebidExt.Floors.Enforcement != nil && prebidExt.Floors.Enforcement.EnforcePBS != nil && *prebidExt.Floors.Enforcement.EnforcePBS {
			tracker.FloorType = models.HardFloor
		}
	}
}
