package ctvjson

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
)

func filterImpsWithInvalidAdserverURL(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) {
	var invalidImpIds []string
	for impId, impCtx := range rCtx.ImpBidCtx {
		if len(impCtx.AdserverURL) == 0 {
			continue
		}

		impCtx.AdserverURL = strings.TrimSpace(impCtx.AdserverURL)
		if !utils.IsValidURL(impCtx.AdserverURL) {
			invalidImpIds = append(invalidImpIds, impId)
		}
	}

	// Remove Invalid Imps
	for _, impId := range invalidImpIds {
		delete(rCtx.ImpBidCtx, impId)
	}

	validImps := make([]openrtb2.Imp, 0, len(rCtx.ImpBidCtx))
	for _, imp := range bidRequest.Imp {
		if slices.Contains(invalidImpIds, imp.ID) {
			continue
		}
		validImps = append(validImps, imp)
	}

	bidRequest.Imp = validImps
}

func ApplyGAMURLConfig(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) error {
	if rCtx.RedirectURL == "" {
		return nil
	}

	if rCtx.AdUnitConfig == nil || rCtx.AdUnitConfig.Config == nil ||
		rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey] == nil ||
		!rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey].EnableGAMUrlLookup {
		return nil
	}

	gamRedirectURL, err := url.Parse(rCtx.RedirectURL)
	if err != nil {
		return err
	}

	gamQueryParams := gamRedirectURL.Query()
	setDeviceParams(rCtx, bidRequest, gamQueryParams)
	setAppParams(rCtx, bidRequest, gamQueryParams)
	setImpVideoParams(rCtx, bidRequest, gamQueryParams)

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

func setAppParams(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest, queryParams url.Values) {
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

func setImpVideoParams(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest, queryParams url.Values) {
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

		podId := imp.Video.PodID
		if podId == "" {
			podId = imp.ID
		}

		podConfig, ok := rCtx.AdpodCtx[podId]
		if !ok {
			podConfig = models.AdpodConfig{
				PodID:   podId,
				PodSeq:  imp.Video.PodSeq,
				PodType: models.PodTypeDynamic,
				Slots: []models.SlotConfig{
					{
						Flexible: true,
					},
				},
			}
		}

		if podConfig.Slots[0].MaxAds == 0 {
			if maxAds := queryParams.Get(models.GAMAdpodMaxAds); len(maxAds) > 0 {
				maxAdsInt, err := strconv.Atoi(maxAds)
				if err == nil {
					podConfig.Slots[0].MaxAds = int64(maxAdsInt)
				}
			}
		}

		if podConfig.Slots[0].MinPodDuration == 0 {
			if minPodDuration := queryParams.Get(models.GAMVideoMinDuration); len(minPodDuration) > 0 {
				minPodDurationInt, err := strconv.Atoi(minPodDuration)
				if err == nil {
					podConfig.Slots[0].MinPodDuration = int64(minPodDurationInt)
				}
			}
		}

		if podConfig.Slots[0].MaxPodDuration == 0 {
			if maxPodDuration := queryParams.Get(models.GAMVideoMaxDuration); len(maxPodDuration) > 0 {
				maxPodDurationInt, err := strconv.Atoi(maxPodDuration)
				if err == nil {
					podConfig.Slots[0].MaxPodDuration = int64(maxPodDurationInt)
				}
			}
		}

		if podConfig.Slots[0].MinDuration == 0 {
			if minDuration := queryParams.Get(models.GAMAdMinDuration); len(minDuration) > 0 {
				minDurationInt, err := strconv.Atoi(minDuration)
				if err == nil {
					podConfig.Slots[0].MinDuration = int64(minDurationInt)
				}
			}
		}

		if podConfig.Slots[0].MaxDuration == 0 {
			if maxDuration := queryParams.Get(models.GAMAdMaxDuration); len(maxDuration) > 0 {
				maxDurationInt, err := strconv.Atoi(maxDuration)
				if err == nil {
					podConfig.Slots[0].MaxDuration = int64(maxDurationInt)
				}
			}
		}

		rCtx.AdpodCtx[podId] = podConfig
	}
}

func getDimension(size string) []string {
	if !strings.Contains(size, models.Pipe) {
		return strings.Split(size, models.DelimiterX)
	}
	return strings.Split(strings.Split(size, models.Pipe)[0], models.DelimiterX)
}
