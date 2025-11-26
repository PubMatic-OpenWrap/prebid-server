package ctvutils

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func ProcessAdpodProfileConfig(rCtx *models.RequestCtx) {
	videoAdDuration := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationKey)
	policy := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationMatchingKey)
	if len(videoAdDuration) > 0 {
		rCtx.AdpodProfileConfig = &models.AdpodProfileConfig{
			AdserverCreativeDurations:              utils.GetIntArrayFromString(videoAdDuration, models.ArraySeparator),
			AdserverCreativeDurationMatchingPolicy: policy,
		}
	}

	if rCtx.IsTestRequest > 0 && rCtx.AdpodProfileConfig == nil && len(rCtx.ImpAdPodConfig) > 0 {
		rCtx.AdpodProfileConfig = &models.AdpodProfileConfig{
			AdserverCreativeDurations:              []int{5, 10},
			AdserverCreativeDurationMatchingPolicy: openrtb_ext.OWRoundupVideoAdDurationMatching,
		}
	}
}
