package adpod

import (
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/ortb"
)

func ApplyAdpodConfigs(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) (*openrtb2.BidRequest, error) {
	if len(rCtx.ImpAdPodConfig) == 0 {
		return bidRequest, nil
	}

	rCtx.AdpodCtx = make(models.AdpodCtx)
	imps := make([]openrtb2.Imp, 0)
	for _, imp := range bidRequest.Imp {
		if imp.Video == nil {
			continue
		}

		// Give priority to adpod config in request
		if len(imp.Video.PodID) > 0 {
			imps = append(imps, imp)
			rCtx.AdpodCtx.AddAdpodConfig(&imp)
			continue
		}

		impPodConfig, ok := rCtx.ImpAdPodConfig[imp.ID]
		if !ok || len(impPodConfig) == 0 {
			imps = append(imps, imp)
			continue
		}

		// Apply adpod config
		var shouldDeleteImp bool
		for i, podConfig := range impPodConfig {
			if podConfig.AdpodConfigV25 != nil {
				rCtx.AdpodCtx.AddAdpodV25Config(&imp, podConfig)
				imps = append(imps, imp)
				continue
			}

			impCopy := ortb.DeepCloneImpression(&imp)
			impCopy.ID = fmt.Sprintf("%s-%s-%d", podConfig.PodID, impCopy.ID, i)
			impCopy.Video.PodID = podConfig.PodID
			impCopy.Video.MaxSeq = podConfig.MaxSeq
			impCopy.Video.PodDur = podConfig.PodDur
			impCopy.Video.MinDuration = podConfig.MinDuration
			impCopy.Video.MaxDuration = podConfig.MaxDuration
			impCopy.Video.RqdDurs = podConfig.RqdDurs
			impCopy.Video.StartDelay = podConfig.StartDelay

			rCtx.AdpodCtx.AddAdpodConfig(impCopy)

			impCtx := rCtx.ImpBidCtx[imp.ID]
			impCtxCopy := impCtx.DeepCopy()
			impCtxCopy.ImpID = impCopy.ID
			impCtxCopy.Video = impCopy.Video

			rCtx.ImpBidCtx[impCopy.ID] = impCtxCopy
			imps = append(imps, *impCopy)
			shouldDeleteImp = true
		}

		// Delete original imp from context if we created copies
		if shouldDeleteImp {
			delete(rCtx.ImpBidCtx, imp.ID)
		}
	}

	bidRequest.Imp = imps
	return bidRequest, nil
}

func ValidateAdpodConfigs(rCtx *models.RequestCtx) error {
	for _, configs := range rCtx.ImpAdPodConfig {
		for _, config := range configs {
			if config.AdpodConfigV25 != nil {
				err := validateV25Configs(&config, rCtx.AdpodProfileConfig)
				if err != nil {
					return err
				}
				// In case of v2.5 config, do not validate 2.6 configs
				continue
			}

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
	}

	return nil
}

func SetDefaultValuesToAdpodConfig(rCtx *models.RequestCtx) {
	for impId, configs := range rCtx.ImpAdPodConfig {
		for i := range configs {
			if configs[i].AdpodConfigV25 != nil {
				setDefaultValuesToV25PodConfig(configs[i].AdpodConfigV25)
				if configs[i].MinDuration == 0 && configs[i].AdpodConfigV25.MinPodDuration > 0 {
					configs[i].MinDuration = int64(configs[i].AdpodConfigV25.MinPodDuration / 2)
				}
				if configs[i].MaxDuration == 0 && configs[i].AdpodConfigV25.MaxPodDuration > 0 {
					configs[i].MaxDuration = int64(configs[i].AdpodConfigV25.MaxPodDuration / 2)
				}
			}
		}
		rCtx.ImpAdPodConfig[impId] = configs
	}
}
