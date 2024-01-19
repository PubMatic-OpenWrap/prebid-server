package tracker

import (
	"fmt"
	"net/url"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func GetTrackerInfo(rCtx models.RequestCtx, responseExt openrtb_ext.ExtBidResponse) string {
	floorsDetails := models.GetFloorsDetails(responseExt)
	tracker := models.Tracker{
		PubID:             rCtx.PubID,
		ProfileID:         fmt.Sprintf("%d", rCtx.ProfileID),
		VersionID:         fmt.Sprintf("%d", rCtx.DisplayID),
		PageURL:           rCtx.PageURL,
		Timestamp:         rCtx.StartTime,
		IID:               rCtx.LoggerImpressionID,
		Platform:          int(rCtx.DeviceCtx.Platform),
		Origin:            rCtx.Origin,
		TestGroup:         rCtx.ABTestConfigApplied,
		FloorModelVersion: floorsDetails.FloorModelVersion,
		FloorType:         floorsDetails.FloorType,
		FloorSkippedFlag:  floorsDetails.Skipfloors,
		FloorSource:       floorsDetails.FloorSource,
	}

	if rCtx.DeviceCtx.Ext != nil {
		tracker.ATTS = rCtx.DeviceCtx.Ext.ATTS
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
