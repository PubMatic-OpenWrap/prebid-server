package adpod

import (
	"encoding/json"
	"errors"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func GetV25AdpodConfigs(rctx *models.RequestCtx, imp *openrtb2.Imp) ([]models.PodConfig, error) {
	adpodConfigV25, ok, err := resolveV25AdpodConfigs(rctx, imp)
	if !ok || err != nil {
		return nil, err
	}

	minPodDuration := imp.Video.MinDuration // For V25, minpodDuration was imp.video.minduration
	maxPodDuration := imp.Video.MaxDuration // For V25, maxpodDuration was imp.video.maxduration

	impCtx := rctx.ImpBidCtx[imp.ID]

	if minPodDuration == 0 {
		adUnitConfig := impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig
		if adUnitConfig != nil && adUnitConfig.Video != nil &&
			adUnitConfig.Video.Enabled != nil && *adUnitConfig.Video.Enabled {
			minPodDuration = adUnitConfig.Video.Config.MinDuration
		}
	}

	if maxPodDuration == 0 {
		adUnitConfig := impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig
		if adUnitConfig != nil && adUnitConfig.Video != nil &&
			adUnitConfig.Video.Enabled != nil && *adUnitConfig.Video.Enabled {
			maxPodDuration = adUnitConfig.Video.Config.MaxDuration
		}
	}

	podConfig := models.PodConfig{
		PodID:       imp.ID,
		MinDuration: int64(adpodConfigV25.MinDuration),
		MaxDuration: int64(adpodConfigV25.MaxDuration),
		AdpodConfigV25: &models.AdpodConfigV25{
			MinAds:                      int64(adpodConfigV25.MinAds),
			MaxAds:                      int64(adpodConfigV25.MaxAds),
			MinPodDuration:              minPodDuration,
			MaxPodDuration:              maxPodDuration,
			IABCategoryExclusionPercent: adpodConfigV25.IABCategoryExclusionPercent,
			AdvertiserExclusionPercent:  adpodConfigV25.AdvertiserExclusionPercent,
		},
	}

	return []models.PodConfig{podConfig}, nil
}

func resolveV25AdpodConfigs(rctx *models.RequestCtx, imp *openrtb2.Imp) (*models.AdPod, bool, error) {
	var adpodConfig *models.AdPod

	// Check in impression extension
	if imp.Video != nil && imp.Video.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(imp.Video.Ext, models.Adpod)
		if err == nil && len(adpodBytes) > 0 {
			rctx.MetricsEngine.RecordCTVReqImpsWithReqConfigCount(rctx.PubIDStr)
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			return adpodConfig, true, err
		}
	}

	// Check in request extension
	if rctx.NewReqExt != nil && rctx.NewReqExt.AdPod != nil {
		adpodConfig = &models.AdPod{
			MinAds:                      rctx.NewReqExt.AdPod.MinAds,
			MaxAds:                      rctx.NewReqExt.AdPod.MaxAds,
			MinDuration:                 rctx.NewReqExt.AdPod.MinDuration,
			MaxDuration:                 rctx.NewReqExt.AdPod.MaxDuration,
			AdvertiserExclusionPercent:  rctx.NewReqExt.AdPod.AdvertiserExclusionPercent,
			IABCategoryExclusionPercent: rctx.NewReqExt.AdPod.IABCategoryExclusionPercent,
		}
		rctx.MetricsEngine.RecordCTVReqCountWithAdPod(rctx.PubIDStr, rctx.ProfileIDStr)
		return adpodConfig, true, nil
	}

	impCtx, ok := rctx.ImpBidCtx[imp.ID]
	if !ok {
		return nil, false, nil
	}

	adUnitConfig := impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig
	if adUnitConfig != nil && adUnitConfig.Video != nil && adUnitConfig.Video.Config != nil && adUnitConfig.Video.Config.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(adUnitConfig.Video.Config.Ext, models.Adpod)
		if err == nil && len(adpodBytes) > 0 {
			rctx.MetricsEngine.RecordCTVReqImpsWithDbConfigCount(rctx.PubIDStr)
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			return adpodConfig, true, err
		}
	}

	return nil, false, nil
}

func setDefaultValuesToV25PodConfig(config *models.AdpodConfigV25) {
	if config.MinAds == 0 {
		config.MinAds = models.DefaultMinAds
	}

	if config.MaxAds == 0 {
		config.MaxAds = models.DefaultMaxAds
	}

	if config.AdvertiserExclusionPercent == nil {
		config.AdvertiserExclusionPercent = ptrutil.ToPtr(models.DefaultAdvertiserExclusionPercent)
	}

	if config.IABCategoryExclusionPercent == nil {
		config.IABCategoryExclusionPercent = ptrutil.ToPtr(models.DefaultIABCategoryExclusionPercent)
	}
}

func validateV25Configs(config *models.PodConfig, adpodProfileConfig *models.AdpodProfileConfig) error {
	if config.AdpodConfigV25.MinAds <= 0 {
		return errors.New("adpod.minads must be positive number")
	}

	if config.AdpodConfigV25.MaxAds <= 0 {
		return errors.New("adpod.maxads must be positive number")
	}

	if config.MinDuration <= 0 {
		return errors.New("adpod.adminduration must be positive number")
	}

	if config.MaxDuration <= 0 {
		return errors.New("adpod.admaxduration must be positive number")
	}

	if (config.AdpodConfigV25.AdvertiserExclusionPercent != nil) && (*config.AdpodConfigV25.AdvertiserExclusionPercent < 0 || *config.AdpodConfigV25.AdvertiserExclusionPercent > 100) {
		return errors.New("adpod.excladv must be number between 0 and 100")
	}

	if (config.AdpodConfigV25.IABCategoryExclusionPercent != nil) && (*config.AdpodConfigV25.IABCategoryExclusionPercent < 0 || *config.AdpodConfigV25.IABCategoryExclusionPercent > 100) {
		return errors.New("adpod.excliabcat must be number between 0 and 100")
	}

	if config.AdpodConfigV25.MinAds > config.AdpodConfigV25.MaxAds {
		return errors.New("adpod.minads must be less than adpod.maxads")
	}

	if config.MinDuration > config.MaxDuration {
		return errors.New("adpod.adminduration must be less than adpod.admaxduration")
	}

	if adpodProfileConfig != nil && len(adpodProfileConfig.AdserverCreativeDurations) > 0 {
		validDurations := false
		for _, videoDuration := range adpodProfileConfig.AdserverCreativeDurations {
			if videoDuration >= int(config.MinDuration) && videoDuration <= int(config.MaxDuration) {
				validDurations = true
				break
			}
		}

		if !validDurations {
			return errors.New("videoAdDuration values should be between adpod.adminduration and dpod.adminduration")
		}
	}

	return nil
}
