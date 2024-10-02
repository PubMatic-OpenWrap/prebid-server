package adpod

import (
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adpodconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/utils/ortb"
)

func GetPodType(impCtx models.ImpCtx) models.PodType {
	if impCtx.AdpodConfig != nil {
		return models.Dynamic
	}

	if len(impCtx.Video.PodID) > 0 && impCtx.Video.PodDur > 0 {
		return models.Dynamic
	}

	if len(impCtx.Video.PodID) > 0 {
		return models.Structured
	}

	return models.NotAdpod
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
