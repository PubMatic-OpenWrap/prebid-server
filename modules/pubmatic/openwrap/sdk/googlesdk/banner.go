package googlesdk

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/feature"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

// GetFlexSlotSizes extracts compatible slot sizes for a banner based on
// flexible slot constraints defined in the banner extension and available
// banner sizes from the DB
func GetFlexSlotSizes(banner *openrtb2.Banner, features feature.Features) []openrtb2.Format {
	if banner == nil || banner.Ext == nil || features == nil {
		return nil
	}

	gsdkFeatures, ok := features[feature.FeatureNameGoogleSDK]
	if !ok || len(gsdkFeatures) == 0 {
		return nil
	}

	var bannerExt openrtb_ext.ExtImpBanner
	if err := json.Unmarshal(banner.Ext, &bannerExt); err != nil || bannerExt.Flexslot == nil {
		return nil
	}

	if bannerExt.Flexslot.Wmin > bannerExt.Flexslot.Wmax ||
		bannerExt.Flexslot.Hmin > bannerExt.Flexslot.Hmax {
		return nil
	}

	var bannerSizes []string
	for i := range gsdkFeatures {
		if gsdkFeatures[i].Name == feature.FeatureFlexSlot {
			if sizes, ok := gsdkFeatures[i].Data.([]string); ok {
				bannerSizes = sizes
				break
			}
		}
	}

	if len(bannerSizes) == 0 {
		return nil
	}

	flexSlotSizes := make([]openrtb2.Format, 0)
	for _, size := range bannerSizes {
		dimensions := strings.Split(size, "x")
		if len(dimensions) != 2 {
			continue
		}

		w, werr := strconv.Atoi(dimensions[0])
		h, herr := strconv.Atoi(dimensions[1])
		if werr != nil || herr != nil {
			continue
		}

		if bannerExt.Flexslot.Wmin <= int32(w) && bannerExt.Flexslot.Wmax >= int32(w) &&
			bannerExt.Flexslot.Hmin <= int32(h) && bannerExt.Flexslot.Hmax >= int32(h) {
			flexSlotSizes = append(flexSlotSizes, openrtb2.Format{
				W: int64(w),
				H: int64(h),
			})
		}
	}

	return flexSlotSizes
}

func SetFlexSlotSizes(banner *openrtb2.Banner, flexSlots []openrtb2.Format) {
	if banner == nil || len(flexSlots) == 0 {
		return
	}

	existing := make(map[[2]int64]struct{}, len(banner.Format))
	for _, f := range banner.Format {
		existing[[2]int64{f.W, f.H}] = struct{}{}
	}

	for _, slot := range flexSlots {
		key := [2]int64{slot.W, slot.H}
		if _, found := existing[key]; !found {
			banner.Format = append(banner.Format, slot)
			existing[key] = struct{}{}
		}
	}
}
