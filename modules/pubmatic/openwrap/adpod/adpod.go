package adpod

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adpodconfig"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/ortb"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func setDefaultValues(adpodConfig *models.AdPod) {
	if adpodConfig.MinAds == 0 {
		adpodConfig.MinAds = models.DefaultMinAds
	}

	if adpodConfig.MaxAds == 0 {
		adpodConfig.MaxAds = models.DefaultMaxAds
	}

	if adpodConfig.AdvertiserExclusionPercent == nil {
		adpodConfig.AdvertiserExclusionPercent = ptrutil.ToPtr(models.DefaultAdvertiserExclusionPercent)
	}

	if adpodConfig.IABCategoryExclusionPercent == nil {
		adpodConfig.IABCategoryExclusionPercent = ptrutil.ToPtr(models.DefaultIABCategoryExclusionPercent)
	}

}

func GetV25AdpodConfigs(impVideo *openrtb2.Video, requestExtConfigs *models.ExtRequestAdPod, adUnitConfig *adunitconfig.AdConfig, partnerConfigMap map[int]map[string]string, pubId string, me metrics.MetricsEngine) (*models.AdPod, error) {
	adpodConfigs, ok, err := resolveV25AdpodConfigs(impVideo, requestExtConfigs, adUnitConfig, pubId, me)
	if !ok || err != nil {
		return nil, err
	}

	// Set default value if adpod object does not exists
	setDefaultValues(adpodConfigs)

	return adpodConfigs, nil
}

func resolveV25AdpodConfigs(impVideo *openrtb2.Video, requestExtConfigs *models.ExtRequestAdPod, adUnitConfig *adunitconfig.AdConfig, pubId string, me metrics.MetricsEngine) (*models.AdPod, bool, error) {
	var adpodConfig *models.AdPod

	// Check in impression extension
	if impVideo != nil && impVideo.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(impVideo.Ext, models.Adpod)
		if err == nil && len(adpodBytes) > 0 {
			me.RecordCTVReqImpsWithReqConfigCount(pubId)
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			return adpodConfig, true, err
		}
	}

	// Check in adunit config
	if adUnitConfig != nil && adUnitConfig.Video != nil && adUnitConfig.Video.Config != nil && adUnitConfig.Video.Config.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(adUnitConfig.Video.Config.Ext, models.Adpod)
		if err == nil && len(adpodBytes) > 0 {
			me.RecordCTVReqImpsWithDbConfigCount(pubId)
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			return adpodConfig, true, err
		}
	}

	return nil, false, nil
}

func ValidateV25Configs(rCtx models.RequestCtx, config *models.AdPod) error {
	if config == nil {
		return nil
	}

	if config.MinAds <= 0 {
		return errors.New("adpod.minads must be positive number")
	}

	if config.MaxAds <= 0 {
		return errors.New("adpod.maxads must be positive number")
	}

	if config.MinDuration <= 0 {
		return errors.New("adpod.adminduration must be positive number")
	}

	if config.MaxDuration <= 0 {
		return errors.New("adpod.admaxduration must be positive number")
	}

	if (config.AdvertiserExclusionPercent != nil) && (*config.AdvertiserExclusionPercent < 0 || *config.AdvertiserExclusionPercent > 100) {
		return errors.New("adpod.excladv must be number between 0 and 100")
	}

	if (config.IABCategoryExclusionPercent != nil) && (*config.IABCategoryExclusionPercent < 0 || *config.IABCategoryExclusionPercent > 100) {
		return errors.New("adpod.excliabcat must be number between 0 and 100")
	}

	if config.MinAds > config.MaxAds {
		return errors.New("adpod.minads must be less than adpod.maxads")
	}

	if config.MinDuration > config.MaxDuration {
		return errors.New("adpod.adminduration must be less than adpod.admaxduration")
	}

	if rCtx.AdpodProfileConfig != nil && len(rCtx.AdpodProfileConfig.AdserverCreativeDurations) > 0 {
		validDurations := false
		for _, videoDuration := range rCtx.AdpodProfileConfig.AdserverCreativeDurations {
			if videoDuration >= config.MinDuration && videoDuration <= config.MaxDuration {
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

func GetAdpodConfigs(rctx models.RequestCtx, cache cache.Cache, adunit *adunitconfig.AdConfig) ([]models.PodConfig, error) {
	// Fetch Adpod Configs from UI
	pods, err := cache.GetAdpodConfig(rctx.PubID, rctx.ProfileID, rctx.DisplayVersionID)
	if err != nil {
		return nil, err
	}

	var uiAdpodConfigs []models.PodConfig
	if pods != nil {
		uiAdpodConfigs = append(uiAdpodConfigs, decouplePodConfigs(pods)...)
	}

	// Vmap adpod configs
	var adrules []models.PodConfig
	if adunit != nil && adunit.Adrule != nil {
		for _, rule := range adunit.Adrule {
			if rule != nil {
				adrules = append(adrules, models.PodConfig{
					PodID:       rule.PodID,
					PodDur:      rule.PodDur,
					MaxSeq:      rule.MaxSeq,
					MinDuration: rule.MinDuration,
					MaxDuration: rule.MaxDuration,
					RqdDurs:     rule.RqdDurs,
				})
			}
		}
	}

	var podConfigs []models.PodConfig
	if len(uiAdpodConfigs) > 0 && adunit.Video != nil && adunit.Video.UsePodConfig != nil && *adunit.Video.UsePodConfig {
		podConfigs = append(podConfigs, uiAdpodConfigs...)
	} else if len(adrules) > 0 && rctx.AdruleFlag {
		podConfigs = append(podConfigs, adrules...)
	}

	return podConfigs, nil
}

func decouplePodConfigs(pods *adpodconfig.AdpodConfig) []models.PodConfig {
	if pods == nil {
		return nil
	}

	var podConfigs []models.PodConfig
	// Add all dynamic adpods
	for i, dynamic := range pods.Dynamic {
		podConfigs = append(podConfigs, models.PodConfig{
			PodID:       fmt.Sprintf("dynamic-%d", i+1),
			PodDur:      dynamic.PodDur,
			MaxSeq:      dynamic.MaxSeq,
			MinDuration: dynamic.MinDuration,
			MaxDuration: dynamic.MaxDuration,
			RqdDurs:     dynamic.RqdDurs,
		})
	}

	// Add all structured adpods
	for i, structured := range pods.Structured {
		podConfigs = append(podConfigs, models.PodConfig{
			PodID:       fmt.Sprintf("structured-%d", i+1),
			MinDuration: structured.MinDuration,
			MaxDuration: structured.MaxDuration,
			RqdDurs:     structured.RqdDurs,
		})
	}

	// Add all hybrid adpods
	for i, hybrid := range pods.Hybrid {
		pod := models.PodConfig{
			PodID:       fmt.Sprintf("hybrid-%d", i+1),
			MinDuration: hybrid.MinDuration,
			MaxDuration: hybrid.MaxDuration,
			RqdDurs:     hybrid.RqdDurs,
		}

		if hybrid.PodDur != nil {
			pod.PodDur = *hybrid.PodDur
		}

		if hybrid.MaxSeq != nil {
			pod.MaxSeq = *hybrid.MaxSeq
		}

		podConfigs = append(podConfigs, pod)
	}

	return podConfigs
}

func ValidateAdpodConfigs(configs []models.PodConfig) error {
	for _, config := range configs {
		if config.RqdDurs == nil && config.MinDuration == 0 && config.MaxDuration == 0 {
			return errors.New("slot duration is missing in adpod config")
		}

		if config.MinDuration > config.MaxDuration {
			return errors.New("min duration should be less than max duration")
		}

		if config.RqdDurs == nil && config.MaxDuration <= 0 {
			return errors.New("max duration should be greater than 0")
		}

		if config.PodDur < 0 {
			return errors.New("pod duration should be positive number")
		}

		if config.MaxSeq < 0 {
			return errors.New("max sequence should be positive number")
		}
	}

	return nil
}

func ApplyAdpodConfigs(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest) *openrtb2.BidRequest {
	if len(rctx.ImpAdPodConfig) == 0 {
		return bidRequest
	}

	imps := make([]openrtb2.Imp, 0)
	for _, imp := range bidRequest.Imp {
		if imp.Video == nil {
			imps = append(imps, imp)
			continue
		}

		// Give priority to adpod config in request
		if len(imp.Video.PodID) > 0 {
			imps = append(imps, imp)
			continue
		}

		impPodConfig, ok := rctx.ImpAdPodConfig[imp.ID]
		if !ok || len(impPodConfig) == 0 {
			imps = append(imps, imp)
			continue
		}

		// Apply adpod config
		for i, podConfig := range impPodConfig {
			impCopy := ortb.DeepCloneImpression(&imp)
			impCopy.ID = fmt.Sprintf("%s-%s-%d", impCopy.ID, podConfig.PodID, i)
			impCopy.Video.PodID = podConfig.PodID
			impCopy.Video.MaxSeq = podConfig.MaxSeq
			impCopy.Video.PodDur = podConfig.PodDur
			impCopy.Video.MinDuration = podConfig.MinDuration
			impCopy.Video.MaxDuration = podConfig.MaxDuration
			impCopy.Video.RqdDurs = podConfig.RqdDurs

			impCtx := rctx.ImpBidCtx[imp.ID]
			impCtxCopy := impCtx.DeepCopy()
			impCtxCopy.Video = impCopy.Video

			rctx.ImpBidCtx[impCopy.ID] = impCtxCopy
			imps = append(imps, *impCopy)
		}

		// Delete original imp from context
		delete(rctx.ImpBidCtx, imp.ID)
	}

	bidRequest.Imp = imps
	return bidRequest
}
