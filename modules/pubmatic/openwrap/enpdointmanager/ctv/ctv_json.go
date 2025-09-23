package ctvendpointmanager

import (
	"encoding/json"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type CTVJSON struct {
	metricsEngine metrics.MetricsEngine
	cache         cache.Cache
}

func NewCTVJSON(metricsEngine metrics.MetricsEngine, cache cache.Cache) *CTVJSON {
	return &CTVJSON{
		metricsEngine: metricsEngine,
		cache:         cache,
	}
}

func (cj *CTVJSON) HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.EntrypointPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.EntrypointPayload], error) {
	if len(rCtx.ResponseFormat) > 0 {
		if rCtx.ResponseFormat != models.ResponseFormatJSON && rCtx.ResponseFormat != models.ResponseFormatRedirect {
			result.NbrCode = int(nbr.InvalidResponseFormat)
			result.Errors = append(result.Errors, "Invalid response format, must be 'json' or 'redirect'")
			return rCtx, result, nil
		}
	}

	return rCtx, result, nil
}

func (cj *CTVJSON) HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawAuctionRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	if len(rCtx.RedirectURL) == 0 {
		rCtx.RedirectURL = models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.OwRedirectURL)
	}

	if len(rCtx.RedirectURL) > 0 {
		rCtx.RedirectURL = strings.TrimSpace(rCtx.RedirectURL)
		if rCtx.ResponseFormat == models.ResponseFormatRedirect && !utils.IsValidURL(rCtx.RedirectURL) {
			result.NbrCode = int(nbr.InvalidRedirectURL)
			result.Errors = append(result.Errors, "Invalid redirect URL")
			return rCtx, result, nil
		}
	}

	if rCtx.ResponseFormat == models.ResponseFormatRedirect && len(rCtx.RedirectURL) == 0 {
		result.NbrCode = int(nbr.MissingOWRedirectURL)
		result.Errors = append(result.Errors, "owRedirectURL is missing")
		return rCtx, result, nil
	}

	videoAdDuration := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationKey)
	policy := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationMatchingKey)
	if len(videoAdDuration) > 0 {
		rCtx.AdpodProfileConfig = &models.AdpodProfileConfig{
			AdserverCreativeDurations:              utils.GetIntArrayFromString(videoAdDuration, models.ArraySeparator),
			AdserverCreativeDurationMatchingPolicy: policy,
		}
	}

	err := validateVideoImpressions(payload.BidRequest)
	if err != nil {
		result.NbrCode = int(nbr.InvalidVideoRequest)
		result.Errors = append(result.Errors, err.Error())
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*payload.BidRequest) // for wrapper logger sz
		return rCtx, result, nil
	}

	setIncludeBrandCategory(rCtx)

	for _, imp := range payload.BidRequest.Imp {
		impCtx, ok := rCtx.ImpBidCtx[imp.ID]
		if !ok {
			continue
		}

		podID := imp.Video.PodID
		if podID == "" {
			podID = imp.ID
		}

		_, ok = rCtx.AdpodCtx[podID]
		//Adding default durations for CTV Test requests
		if rCtx.IsTestRequest > 0 && ok && rCtx.AdpodProfileConfig == nil {
			rCtx.AdpodProfileConfig = &models.AdpodProfileConfig{
				AdserverCreativeDurations:              []int{5, 10},
				AdserverCreativeDurationMatchingPolicy: openrtb_ext.OWRoundupVideoAdDurationMatching,
			}
		}

		rCtx.ImpBidCtx[imp.ID] = impCtx
	}

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		if ep.BidRequest.Source != nil && ep.BidRequest.Source.SChain != nil {
			err := isValidSchain(ep.BidRequest.Source.SChain)
			if err != nil {
				schainBytes, _ := json.Marshal(ep.BidRequest.Source.SChain)
				glog.Errorf(ctv.ErrSchainValidationFailed, models.SChainKey, err.Error(), rCtx.PubIDStr, rCtx.ProfileIDStr, string(schainBytes))
				ep.BidRequest.Source.SChain = nil
			}
		}

		err := filterNonVideoImpressions(ep.BidRequest)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		}

		ep.BidRequest = adpod.ApplyAdpodConfigs(rCtx, ep.BidRequest)

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-json-before-validation")

	return rCtx, result, nil
}

func (cj *CTVJSON) HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BidderRequestPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawBidderResponsePayload]) (models.RequestCtx, hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]) (models.RequestCtx, hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload]) (models.RequestCtx, hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ExitpointPaylaod]) (models.RequestCtx, hookstage.HookResult[hookstage.ExitpointPaylaod], error) {
	response, ok := payload.Response.(*openrtb2.BidResponse)
	if !ok {
		return rCtx, result, nil
	}

	if response.NBR != nil {
		return rCtx, result, nil
	}

	return rCtx, result, nil
}

func checkRedirectResponse(rCtx models.RequestCtx) bool {
	if rCtx.Debug {
		return false
	}

	if rCtx.RedirectURL != "" && rCtx.ResponseFormat == models.ResponseFormatRedirect {
		return true
	}

	return false
}
