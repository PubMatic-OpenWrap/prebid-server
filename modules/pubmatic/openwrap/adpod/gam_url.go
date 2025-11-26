package adpod

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func checkGAMURLLookupEnabled(rCtx *models.RequestCtx) bool {
	if rCtx.RedirectURL == "" {
		return false
	}

	return rCtx.AdUnitConfig != nil &&
		rCtx.AdUnitConfig.Config != nil &&
		rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey] != nil &&
		rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey].EnableGAMUrlLookup
}

func ApplyGAMURLConfig(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) error {
	if !checkGAMURLLookupEnabled(rCtx) {
		return nil
	}

	gamRedirectURL, err := url.Parse(rCtx.RedirectURL)
	if err != nil {
		return err
	}

	gamQueryParams := gamRedirectURL.Query()
	setDeviceParams(rCtx, bidRequest, gamQueryParams)
	setAppParams(bidRequest, gamQueryParams)
	setImpVideoParams(bidRequest, gamQueryParams)

	return nil
}

func ApplyGAMURLAdpodConfig(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) error {
	if !checkGAMURLLookupEnabled(rCtx) {
		return nil
	}

	gamRedirectURL, err := url.Parse(rCtx.RedirectURL)
	if err != nil {
		return err
	}

	gamQueryParams := gamRedirectURL.Query()
	config, ok := getAdpodConfigFromGAMQuery(gamQueryParams)
	if !ok {
		return nil
	}

	for _, imp := range bidRequest.Imp {
		podConfigs, ok := rCtx.ImpAdPodConfig[imp.ID]
		if !ok {
			podConfigs = []models.PodConfig{config}
		} else {
			setAdpodConfigFromGAMQuery(podConfigs, config)
		}
		rCtx.ImpAdPodConfig[imp.ID] = podConfigs
	}

	return nil
}

func setDeviceParams(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest, queryParams url.Values) {
	if bidRequest.Device == nil {
		bidRequest.Device = &openrtb2.Device{}
	}

	if rCtx.DeviceCtx.DeviceIFA == "" {
		rCtx.DeviceCtx.DeviceIFA = queryParams.Get(models.GAMDeviceIFA)
		bidRequest.Device.IFA = rCtx.DeviceCtx.DeviceIFA
	}

	if rCtx.DeviceCtx.Language == "" {
		rCtx.DeviceCtx.Language = queryParams.Get(models.GAMDeviceLanguage)
		bidRequest.Device.Language = rCtx.DeviceCtx.Language
	}
}

func setAppParams(bidRequest *openrtb2.BidRequest, queryParams url.Values) {
	if bidRequest.App == nil {
		return
	}

	if bidRequest.App.ID == "" {
		bidRequest.App.ID = queryParams.Get(models.GAMAppID)
	}

	if bidRequest.App.Bundle == "" {
		bidRequest.App.Bundle = queryParams.Get(models.GAMAppBundle)
	}

	if bidRequest.App.StoreURL == "" {
		bidRequest.App.StoreURL = queryParams.Get(models.GAMAppStoreUrl)
	}
}

func setImpVideoParams(bidRequest *openrtb2.BidRequest, queryParams url.Values) {
	for _, imp := range bidRequest.Imp {
		if imp.Video == nil {
			continue
		}

		if imp.Video.Linearity == 0 {
			if linearity := queryParams.Get(models.GAMVideoLinearity); len(linearity) > 0 {
				switch linearity {
				case models.GAMVideoLinear:
					imp.Video.Linearity = adcom1.LinearityLinear
				case models.GAMVideoNonLinear:
					imp.Video.Linearity = adcom1.LinearityNonLinear
				}
			}
		}

		if imp.Video.W == nil && imp.Video.H == nil {
			if dimension := queryParams.Get(models.GAMVideoDimensions); len(dimension) > 0 {
				sizeMap := getDimension(dimension)
				if len(sizeMap) > 1 {
					w, err := strconv.Atoi(sizeMap[0])
					if err == nil {
						h, err := strconv.Atoi(sizeMap[1])
						if err == nil {
							w := int64(w)
							h := int64(h)
							imp.Video.W = &w
							imp.Video.H = &h
						}
					}
				}
			}
		}
	}
}

func getDimension(size string) []string {
	if !strings.Contains(size, models.Pipe) {
		return strings.Split(size, models.DelimiterX)
	}
	return strings.Split(strings.Split(size, models.Pipe)[0], models.DelimiterX)
}

func getAdpodConfigFromGAMQuery(queryParams url.Values) (models.PodConfig, bool) {
	config := models.PodConfig{
		AdpodConfigV25: &models.AdpodConfigV25{},
	}

	var isConfigSet bool
	if maxAds := queryParams.Get(models.GAMAdpodMaxAds); len(maxAds) > 0 {
		maxAdsInt, err := strconv.Atoi(maxAds)
		if err == nil {
			config.AdpodConfigV25.MaxAds = int64(maxAdsInt)
			isConfigSet = true
		}
	}

	if minPodDuration := queryParams.Get(models.GAMVideoMinDuration); len(minPodDuration) > 0 {
		minPodDurationInt, err := strconv.Atoi(minPodDuration)
		if err == nil {
			config.AdpodConfigV25.MinPodDuration = int64(minPodDurationInt)
			isConfigSet = true
		}
	}

	if maxPodDuration := queryParams.Get(models.GAMVideoMaxDuration); len(maxPodDuration) > 0 {
		maxPodDurationInt, err := strconv.Atoi(maxPodDuration)
		if err == nil {
			config.AdpodConfigV25.MaxPodDuration = int64(maxPodDurationInt)
			isConfigSet = true
		}
	}

	if minDuration := queryParams.Get(models.GAMAdMinDuration); len(minDuration) > 0 {
		minDurationInt, err := strconv.Atoi(minDuration)
		if err == nil {
			config.MinDuration = int64(minDurationInt)
			isConfigSet = true
		}
	}

	if maxDuration := queryParams.Get(models.GAMAdMaxDuration); len(maxDuration) > 0 {
		maxDurationInt, err := strconv.Atoi(maxDuration)
		if err == nil {
			config.MaxDuration = int64(maxDurationInt)
			isConfigSet = true
		}
	}

	return config, isConfigSet
}

func setAdpodConfigFromGAMQuery(podConfigs []models.PodConfig, gamAdpodConfig models.PodConfig) {
	for i := range podConfigs {
		if podConfigs[i].AdpodConfigV25 == nil {
			continue
		}

		if podConfigs[i].AdpodConfigV25.MaxAds == 0 {
			podConfigs[i].AdpodConfigV25.MaxAds = gamAdpodConfig.AdpodConfigV25.MaxAds
		}

		if podConfigs[i].AdpodConfigV25.MinPodDuration == 0 {
			podConfigs[i].AdpodConfigV25.MinPodDuration = gamAdpodConfig.AdpodConfigV25.MinPodDuration
		}

		if podConfigs[i].AdpodConfigV25.MaxPodDuration == 0 {
			podConfigs[i].AdpodConfigV25.MaxPodDuration = gamAdpodConfig.AdpodConfigV25.MaxPodDuration
		}

		if podConfigs[i].MinDuration == 0 {
			podConfigs[i].MinDuration = gamAdpodConfig.MinDuration
		}

		if podConfigs[i].MaxDuration == 0 {
			podConfigs[i].MaxDuration = gamAdpodConfig.MaxDuration
		}
	}
}
