package adunitconfig

import (
	"encoding/json"
	"strings"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func selectSlot(rCtx models.RequestCtx, h, w int64, tagid, div, source string) (slotAdUnitConfig *adunitconfig.AdConfig, slotName string, isRegex bool, matchedRegex string) {
	slotName = bidderparams.GenerateSlotName(h, w, rCtx.AdUnitConfig.ConfigPattern, tagid, div, rCtx.Source)

	if slotAdUnitConfig, ok := rCtx.AdUnitConfig.Config[strings.ToLower(slotName)]; ok {
		return slotAdUnitConfig, slotName, false, ""
	} else if rCtx.AdUnitConfig.Regex {
		if matchedRegex = getRegexMatch(rCtx, strings.ToLower(slotName)); matchedRegex != "" {
			return rCtx.AdUnitConfig.Config[matchedRegex], slotName, true, matchedRegex
		}
	}

	return nil, "", false, ""
}

/*GetClientConfigForMediaType function fetches the client config data from the ad unit config JSON for the given media type*/
func GetClientConfigForMediaType(rctx models.RequestCtx, impID string, mediaType string) json.RawMessage {
	if rctx.AdUnitConfig == nil || rctx.AdUnitConfig.Config == nil {
		return nil
	}

	impData, ok := rctx.ImpBidCtx[impID]
	if !ok {
		return nil
	}

	if mediaType == models.AdunitConfigSlotBannerKey {
		if impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig != nil &&
			impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner != nil &&
			impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Config != nil {
			if impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Enabled != nil &&
				*impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Enabled == false {
				return nil
			}
			return impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Config.ClientConfig
		}
	} else if mediaType == models.AdunitConfigSlotVideoKey {
		if impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig != nil &&
			impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
			impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config != nil {
			if impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Enabled != nil &&
				*impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Enabled == false {
				return nil
			}
			return impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.ClientConfig
		}
	}

	return nil
}
