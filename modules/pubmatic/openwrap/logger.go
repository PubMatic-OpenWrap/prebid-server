package openwrap

import (
	"encoding/json"
	"fmt"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func getIncomingSlots(imp openrtb2.Imp, videoAdUnitCtx models.AdUnitCtx) []string {
	sizes := map[string]struct{}{}
	if imp.Banner == nil && imp.Video == nil && imp.Native != nil {
		return []string{"1x1"}
	}
	if imp.Banner != nil {
		if imp.Banner.W != nil && imp.Banner.H != nil {
			sizes[fmt.Sprintf("%dx%d", *imp.Banner.W, *imp.Banner.H)] = struct{}{}
		}

		for _, format := range imp.Banner.Format {
			sizes[fmt.Sprintf("%dx%d", format.W, format.H)] = struct{}{}
		}
	}

	videoSlotEnabled := true
	if videoAdUnitCtx.AppliedSlotAdUnitConfig != nil && videoAdUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
		videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Enabled != nil && !*videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Enabled {
		videoSlotEnabled = false
	}

	if imp.Video != nil && videoSlotEnabled {
		if imp.Video.W != 0 && imp.Video.H != 0 {
			sizes[fmt.Sprintf("%dx%d", imp.Video.W, imp.Video.H)] = struct{}{}
		} else if videoAdUnitCtx.AppliedSlotAdUnitConfig != nil && videoAdUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
			videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config != nil {
			sizes[fmt.Sprintf("%dx%d", videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.W, videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.H)] = struct{}{}
		} else {
			sizes[fmt.Sprintf("%dx%d", 0, 0)] = struct{}{}
		}
	}

	var s []string
	for k := range sizes {
		s = append(s, k)
	}
	return s
}

func getDefaultImpBidCtx(request openrtb2.BidRequest) map[string]models.ImpCtx {
	impBidCtx := make(map[string]models.ImpCtx)
	for _, imp := range request.Imp {
		impExt := &models.ImpExtension{}
		json.Unmarshal(imp.Ext, impExt)

		impBidCtx[imp.ID] = models.ImpCtx{
			IncomingSlots:     getIncomingSlots(imp, models.AdUnitCtx{}),
			AdUnitName:        getAdunitName(imp.TagID, impExt),
			SlotName:          getSlotName(imp.TagID, impExt),
			IsRewardInventory: impExt.Reward,
		}
	}
	return impBidCtx
}
