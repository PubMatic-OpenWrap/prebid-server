package adpod

import (
	"encoding/json"
	"errors"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
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

func ResolveAdpodConfigs(impVideo *openrtb2.Video, requestExtConfigs *models.ExtRequestAdPod, adUnitConfig *adunitconfig.AdConfig) (*models.AdPod, error) {
	var adpodConfig models.AdPod

	// Check in impression extension
	if impVideo != nil && impVideo.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(impVideo.Ext, "adpod")
		if len(adpodBytes) > 0 && err == nil {
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			return &adpodConfig, err
		}
	}

	// Check in request extension
	if requestExtConfigs != nil {
		adpodConfig = requestExtConfigs.AdPod
		return &adpodConfig, nil
	}

	// Check in adunit config
	if adUnitConfig.Video != nil && adUnitConfig.Video.Config != nil && adUnitConfig.Video.Config.Ext != nil {
		adpodBytes, _, _, err := jsonparser.Get(adUnitConfig.Video.Config.Ext, "adpod")
		if len(adpodBytes) > 0 && err == nil {
			err := json.Unmarshal(adpodBytes, &adpodConfig)
			return &adpodConfig, err
		}
	}

	return nil, errors.New("no adpod configs found")

}

func IsValidAdPod(pod *models.AdPod) error {
	if pod == nil {
		return errors.New("empty adpod object")
	}

	if pod.MinAds <= 0 {
		return errors.New("adpod.minads must be positive number")
	}

	if pod.MaxAds <= 0 {
		return errors.New("adpod.maxads must be positive number")
	}

	if pod.MinDuration <= 0 {
		return errors.New("adpod.adminduration must be positive number")
	}

	if pod.MaxDuration <= 0 {
		return errors.New("adpod.admaxduration must be positive number")
	}

	if pod.AdvertiserExclusionPercent < 0 || pod.AdvertiserExclusionPercent > 100 {
		return errors.New("adpod.excladv must be number between 0 and 100")
	}

	if pod.IABCategoryExclusionPercent < 0 || pod.IABCategoryExclusionPercent > 100 {
		return errors.New("adpod.excliabcat must be number between 0 and 100")
	}

	if pod.MinAds > pod.MaxAds {
		return errors.New("adpod.minads must be less than adpod.maxads")
	}

	if pod.MinDuration > pod.MaxDuration {
		return errors.New("adpod.adminduration must be less than adpod.admaxduration")
	}

	return nil
}
