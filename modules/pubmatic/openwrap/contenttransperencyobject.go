package openwrap

import (
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// setContentObjectTransparencyObject from request or AdUnit Object
// setContentObjectTransparencyObject from request or AdUnit Object
func setContentTransparencyObject(rctx models.RequestCtx, reqExt *models.RequestExt) (prebidTransparency *openrtb_ext.TransparencyExt) {
	if reqExt.Prebid.Transparency != nil {
		return
	}

	for _, impCtx := range rctx.ImpBidCtx {
		var transparency *adunitconfig.Transparency

		if impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig != nil && impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Transparency != nil {
			transparency = impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Transparency
		} else if impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig != nil && impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Transparency != nil {
			transparency = impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Transparency
		}

		if transparency == nil || len(transparency.Content.Mappings) == 0 {
			continue
		}

		content := make(map[string]openrtb_ext.TransparencyRule)

		for _, partnerConfig := range rctx.PartnerConfigMap {
			bidder := partnerConfig[models.BidderCode]

			_, ok := rctx.AdapterThrottleMap[bidder]
			if ok || partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
				continue
			}

			for _, rule := range getRules(rctx.Source, bidder) {
				if transparencyRule, ok := transparency.Content.Mappings[rule]; ok {
					content[bidder] = transparencyRule
					break
				}
			}
		}

		if len(content) > 0 {
			return &openrtb_ext.TransparencyExt{
				Content: content,
			}
		}
	}

	return nil
}

func getRules(source, bidder string) []string {
	return []string{source + "|" + bidder, "*|" + bidder, source + "|*", "*|*"}
}
