package models

import (
	"encoding/json"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
)

func GetDefaultImpBidCtx(request openrtb2.BidRequest) map[string]ImpCtx {
	impBidCtx := make(map[string]ImpCtx)
	for _, imp := range request.Imp {
		impExt := &ImpExtension{}
		json.Unmarshal(imp.Ext, impExt)

		impBidCtx[imp.ID] = ImpCtx{
			IncomingSlots:     GetIncomingSlots(imp, AdUnitCtx{}),
			AdUnitName:        GetAdunitName(imp.TagID, impExt),
			SlotName:          GetSlotName(imp.TagID, impExt),
			IsRewardInventory: impExt.Reward,
		}
	}
	return impBidCtx
}

func GetIncomingSlots(imp openrtb2.Imp, videoAdUnitCtx AdUnitCtx) []string {
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
		if imp.Video.W != nil && imp.Video.H != nil {
			sizes[fmt.Sprintf("%dx%d", *imp.Video.W, *imp.Video.H)] = struct{}{}
		} else if videoAdUnitCtx.AppliedSlotAdUnitConfig != nil && videoAdUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
			videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config != nil &&
			videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.W != nil && videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.H != nil {
			sizes[fmt.Sprintf("%dx%d", *videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.W, *videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.H)] = struct{}{}
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

/*
getAdunitName will return adunit name according to below priority
 1. imp.ext.data.adserver.adslot if imp.ext.data.adserver.name == "gam"
 2. imp.ext.data.pbadslot
 3. imp.tagid
*/
func GetAdunitName(tagId string, impExt *ImpExtension) string {
	if impExt == nil {
		return tagId
	}
	if impExt.Data.AdServer != nil && impExt.Data.AdServer.Name == GamAdServer && impExt.Data.AdServer.AdSlot != "" {
		return impExt.Data.AdServer.AdSlot
	}
	if len(impExt.Data.PbAdslot) > 0 {
		return impExt.Data.PbAdslot
	}
	return tagId
}

/*
getSlotName will return slot name according to below priority
 1. imp.ext.gpid
 2. imp.tagid
 3. imp.ext.data.pbadslot
 4. imp.ext.prebid.storedrequest.id
*/
func GetSlotName(tagId string, impExt *ImpExtension) string {
	if impExt == nil {
		return tagId
	}

	if len(impExt.GpId) > 0 {
		return impExt.GpId
	}

	if len(tagId) > 0 {
		return tagId
	}

	if len(impExt.Data.PbAdslot) > 0 {
		return impExt.Data.PbAdslot
	}

	var storeReqId string
	if impExt.Prebid.StoredRequest != nil {
		storeReqId = impExt.Prebid.StoredRequest.ID
	}

	return storeReqId
}
