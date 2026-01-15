package adpod

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
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

func ApplyGAMURLAdpodConfig(rCtx *models.RequestCtx, adpodV25 *models.AdPod) (minPodDuration int64, maxPodDuration int64) {
	if !checkGAMURLLookupEnabled(rCtx) {
		return
	}

	gamRedirectURL, err := url.Parse(rCtx.RedirectURL)
	if err != nil {
		glog.Error("failed to parse redirect URL in ApplyGAMURLAdpodConfig: " + err.Error())
		return
	}

	queryParams := gamRedirectURL.Query()
	setAdpodConfigFromGAMQuery(queryParams, adpodV25)

	if minPodDur := queryParams.Get(models.GAMVideoMinDuration); len(minPodDur) > 0 {
		minPodDurationInt, err := strconv.Atoi(minPodDur)
		if err == nil {
			minPodDuration = int64(minPodDurationInt)
		}
	}

	if maxPodDur := queryParams.Get(models.GAMVideoMaxDuration); len(maxPodDur) > 0 {
		maxPodDurationInt, err := strconv.Atoi(maxPodDur)
		if err == nil {
			maxPodDuration = int64(maxPodDurationInt)
		}
	}

	return
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

func setAdpodConfigFromGAMQuery(queryParams url.Values, v25Config *models.AdPod) {
	if maxAds := queryParams.Get(models.GAMAdpodMaxAds); len(maxAds) > 0 && v25Config.MaxAds == nil {
		maxAdsInt, err := strconv.Atoi(maxAds)
		if err == nil {
			v25Config.MaxAds = ptrutil.ToPtr(int64(maxAdsInt))

		}
	}

	if minDuration := queryParams.Get(models.GAMAdMinDuration); len(minDuration) > 0 && v25Config.MinDuration == nil {
		minDurationInt, err := strconv.Atoi(minDuration)
		if err == nil {
			v25Config.MinDuration = ptrutil.ToPtr(int64(minDurationInt))

		}
	}

	if maxDuration := queryParams.Get(models.GAMAdMaxDuration); len(maxDuration) > 0 && v25Config.MaxDuration == nil {
		maxDurationInt, err := strconv.Atoi(maxDuration)
		if err == nil {
			v25Config.MaxDuration = ptrutil.ToPtr(int64(maxDurationInt))

		}
	}
}
