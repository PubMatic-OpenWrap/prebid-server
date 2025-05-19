package googlesdk

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/feature"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func GetFlexSlotSizes(banner *openrtb2.Banner, gsdkFeatures []feature.Feature) []openrtb2.Format {
	if banner == nil || banner.Ext == nil {
		return nil
	}

	var flexSlotSizes []openrtb2.Format
	var bannerExt openrtb_ext.ExtImpBanner
	if err := json.Unmarshal(banner.Ext, &bannerExt); err == nil && bannerExt.Flexslot != nil {
		var slotSizes []string
		for i := range gsdkFeatures {
			if gsdkFeatures[i].Name == feature.FeatureFlexSlot {
				if sizes, ok := gsdkFeatures[i].Data.([]string); ok {
					slotSizes = sizes
					break
				}
			}
		}

		if len(slotSizes) > 0 {
			flexSlotSizes = make([]openrtb2.Format, 0)
		}

		for _, size := range slotSizes {
			dimensions := strings.Split(size, "x")
			w, _ := strconv.Atoi(dimensions[0])
			h, _ := strconv.Atoi(dimensions[1])

			if bannerExt.Flexslot.Wmin <= int32(w) && bannerExt.Flexslot.Wmax >= int32(w) &&
				bannerExt.Flexslot.Hmin <= int32(h) && bannerExt.Flexslot.Hmax >= int32(h) {
				flexSlotSizes = append(flexSlotSizes, openrtb2.Format{
					W: int64(w),
					H: int64(h),
				})
			}
		}
	}

	return flexSlotSizes
}

func checkSlotSizeExists(bannerFormat []openrtb2.Format, inputFormat openrtb2.Format) bool {
	for _, format := range bannerFormat {
		if format.W == inputFormat.W && format.H == inputFormat.H {
			return true
		}
	}
	return false
}

func SetFlexSlotSizes(banner *openrtb2.Banner, rCtx models.RequestCtx) {
	if banner == nil {
		return
	}

	for i := range rCtx.GoogleSDK.FlexSlot {
		if !checkSlotSizeExists(banner.Format, rCtx.GoogleSDK.FlexSlot[i]) {
			banner.Format = append(banner.Format, rCtx.GoogleSDK.FlexSlot[i])
		}
	}
}
