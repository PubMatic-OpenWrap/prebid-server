package adpod

import (
	"encoding/json"
	"errors"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/util/ptrutil"
)

// func getAdpodConfigsFromExt(ext json.RawMessage) (models.AdPod, error) {
// 	var adpodConfig models.AdPod
// 	adpodBytes, _, _, err := jsonparser.Get(ext, "adpod")
// 	if len(adpodBytes) > 0 && err == nil {
// 		err := json.Unmarshal(adpodBytes, &adpodConfig)
// 		return adpodConfig, err
// 	}

// 	return adpodConfig, fmt.Errorf("no adpod configs found")
// }

func setDefaultValues(adpodConfig *models.AdPod) {
	if adpodConfig.MinAds == 0 {
		adpodConfig.MinAds = 1
	}

	if adpodConfig.MaxAds == 0 {
		adpodConfig.MaxAds = 3
	}

	if adpodConfig.AdvertiserExclusionPercent == nil {
		adpodConfig.AdvertiserExclusionPercent = ptrutil.ToPtr(100)
	}

	if adpodConfig.IABCategoryExclusionPercent == nil {
		adpodConfig.IABCategoryExclusionPercent = ptrutil.ToPtr(100)
	}

}

func GetAdpodConfigs(impVideo *openrtb2.Video, requestExtConfigs *models.ExtRequestAdPod, adUnitConfig *adunitconfig.AdConfig, partnerConfigMap map[int]map[string]string) (*models.AdPod, error) {
	adpodConfigs, ok, err := resolveAdpodConfigs(impVideo, requestExtConfigs, adUnitConfig)
	if !ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	videoAdDuration := models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.VideoAdDurationKey)
	if len(videoAdDuration) > 0 {
		adpodConfigs.VideoAdDuration = utils.GetIntArrayFromString(videoAdDuration, models.ArraySeparator)
	}

	videoAdDurationMatchingPolicy := models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.VideoAdDurationMatchingKey)
	if len(videoAdDurationMatchingPolicy) > 0 {
		adpodConfigs.VideoAdDurationMatching = videoAdDurationMatchingPolicy
	}

	// Set default value if adpod object does not exists
	setDefaultValues(adpodConfigs)

	return adpodConfigs, nil

}

func resolveAdpodConfigs(impVideo *openrtb2.Video, requestExtConfigs *models.ExtRequestAdPod, adUnitConfig *adunitconfig.AdConfig) (*models.AdPod, bool, error) {
	var adpodConfig models.AdPod

	// Check in impression extension
	if impVideo != nil && impVideo.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(impVideo.Ext, "adpod")
		if len(adpodBytes) > 0 && err == nil {
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			if err != nil {
				return nil, true, err
			}
			return &adpodConfig, true, err
		}
	}

	// Check in request extension (Removed support for accepting from request)
	// if requestExtConfigs != nil {
	// 	adpodConfig = &requestExtConfigs.AdPod
	// 	return adpodConfig, nil
	// }

	// Check in adunit config
	if adUnitConfig.Video != nil && adUnitConfig.Video.Config != nil && adUnitConfig.Video.Config.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(adUnitConfig.Video.Config.Ext, "adpod")
		if len(adpodBytes) > 0 && err == nil {
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			if err != nil {
				return nil, true, err
			}
			return &adpodConfig, true, err
		}
	}

	return nil, false, nil

}

func Validate(config *models.AdPod) error {
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

	return nil
}
