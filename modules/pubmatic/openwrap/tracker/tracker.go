package tracker

import (
	"fmt"
	"net/url"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func GetTrackerInfo(rCtx models.RequestCtx, responseExt openrtb_ext.ExtBidResponse) string {
	if rCtx.TrackerDisabled {
		return ""
	}

	floorsDetails := models.GetFloorsDetails(responseExt)
	tracker := models.Tracker{
		PubID:             rCtx.PubID,
		ProfileID:         fmt.Sprintf("%d", rCtx.ProfileID),
		VersionID:         fmt.Sprintf("%d", rCtx.DisplayVersionID),
		PageURL:           rCtx.PageURL,
		Timestamp:         rCtx.StartTime,
		IID:               rCtx.LoggerImpressionID,
		Platform:          int(rCtx.DeviceCtx.Platform),
		Origin:            rCtx.Origin,
		TestGroup:         rCtx.ABTestConfigApplied,
		FloorModelVersion: floorsDetails.FloorModelVersion,
		FloorProvider:     floorsDetails.FloorProvider,
		FloorType:         floorsDetails.FloorType,
		FloorSkippedFlag:  floorsDetails.Skipfloors,
		FloorSource:       floorsDetails.FloorSource,
	}

	if rCtx.DeviceCtx.Ext != nil {
		tracker.ATTS, _ = rCtx.DeviceCtx.Ext.GetAtts()
	}

	constructedURLString := constructTrackerURL(rCtx, tracker)

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
