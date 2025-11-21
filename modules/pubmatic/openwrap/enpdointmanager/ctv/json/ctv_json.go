package ctvjson

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod"
	impressions "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/legacy/impressions"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/auction"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/creativecache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	ctvutils "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager/ctv/utils"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

type CTVJSON struct {
	metricsEngine metrics.MetricsEngine
	creativeCache creativecache.Client
}

func NewCTVJSON(metricsEngine metrics.MetricsEngine, creativeCache creativecache.Client) *CTVJSON {
	return &CTVJSON{
		metricsEngine: metricsEngine,
		creativeCache: creativeCache,
	}
}

func (cj *CTVJSON) HandleEntrypointHook(payload hookstage.EntrypointPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.EntrypointPayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.EntrypointPayload], error) {
	cj.metricsEngine.RecordCTVHTTPMethodRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Method)
	if len(rCtx.ResponseFormat) > 0 {
		if rCtx.ResponseFormat != models.ResponseFormatJSON && rCtx.ResponseFormat != models.ResponseFormatRedirect {
			result.NbrCode = int(nbr.InvalidResponseFormat)
			return rCtx, result, errors.New("Invalid response format, must be 'json' or 'redirect'")
		}
	}

	return rCtx, result, nil
}

func (cj *CTVJSON) HandleRawAuctionHook(payload hookstage.RawAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawAuctionRequestPayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleBeforeValidationHook(payload hookstage.BeforeValidationRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	if len(rCtx.RedirectURL) == 0 {
		rCtx.RedirectURL = models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.OwRedirectURL)
	}

	if len(rCtx.RedirectURL) > 0 {
		rCtx.RedirectURL = strings.TrimSpace(rCtx.RedirectURL)
		if rCtx.ResponseFormat == models.ResponseFormatRedirect && !utils.IsValidURL(rCtx.RedirectURL) {
			result.NbrCode = int(nbr.InvalidRedirectURL)
			return rCtx, result, errors.New("Invalid redirect URL")
		}
	}

	if rCtx.ResponseFormat == models.ResponseFormatRedirect && len(rCtx.RedirectURL) == 0 {
		result.NbrCode = int(nbr.MissingOWRedirectURL)
		return rCtx, result, errors.New("owRedirectURL is missing")
	}

	videoAdDuration := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationKey)
	policy := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VideoAdDurationMatchingKey)
	if len(videoAdDuration) > 0 {
		rCtx.AdpodProfileConfig = &models.AdpodProfileConfig{
			AdserverCreativeDurations:              utils.GetIntArrayFromString(videoAdDuration, models.ArraySeparator),
			AdserverCreativeDurationMatchingPolicy: policy,
		}
	}

	err := ctvutils.ValidateVideoImpressions(payload.BidRequest)
	if err != nil {
		result.NbrCode = int(nbr.InvalidVideoRequest)
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*payload.BidRequest) // for wrapper logger sz
		return rCtx, result, err
	}

	// filter imps with invalid adserver url
	filterImpsWithInvalidAdserverURL(&rCtx, payload.BidRequest)

	ctvutils.SetIncludeBrandCategory(rCtx)

	// Add multibid configuration
	var multibid []*openrtb_ext.ExtMultiBid
	for _, partnerConfig := range rCtx.PartnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		partneridstr, ok := partnerConfig[models.PARTNER_ID]
		if !ok {
			continue
		}
		partnerID, err := strconv.Atoi(partneridstr)
		if err != nil || partnerID == models.VersionLevelConfigID {
			continue
		}

		// bidderCode is in context with pubmatic. Ex. it could be appnexus-1, appnexus-2, etc.
		bidderCode := partnerConfig[models.BidderCode]

		// prebidBidderCode is equivalent of PBS-Core's bidderCode
		prebidBidderCode := partnerConfig[models.PREBID_PARTNER_NAME]

		multibidConfig := &openrtb_ext.ExtMultiBid{
			Bidder:                 prebidBidderCode,
			Alias:                  bidderCode,
			MaxBids:                ptrutil.ToPtr(int(openrtb_ext.MaxBidLimit)),
			TargetBidderCodePrefix: bidderCode,
		}
		multibid = append(multibid, multibidConfig)
	}
	rCtx.NewReqExt.Prebid.MultiBid = multibid

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
		rCtx, ok := utils.GetRequestContext(miCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleBeforeValidationHook mutation")
			return ep, nil
		}

		defer func() {
			miCtx.ModuleContext.Set("rctx", rCtx)
		}()

		if ep.BidRequest.Source != nil && ep.BidRequest.Source.SChain != nil {
			err := ctvutils.IsValidSchain(ep.BidRequest.Source.SChain)
			if err != nil {
				schainBytes, _ := json.Marshal(ep.BidRequest.Source.SChain)
				glog.Errorf(ctv.ErrSchainValidationFailed, models.SChainKey, err.Error(), rCtx.PubIDStr, rCtx.ProfileIDStr, string(schainBytes))
				ep.BidRequest.Source.SChain = nil
			}
		}

		err := ctvutils.FilterNonVideoImpressions(ep.BidRequest)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		}

		// Remove adpod data from ext
		ctvutils.RemoveAdpodDataFromExt(ep.BidRequest)

		// Add GAM URL configs
		err = ApplyGAMURLConfig(&rCtx, ep.BidRequest)
		if err != nil {
			result.Warnings = append(result.Warnings, "Failed to apply GAM URL configs: "+err.Error())
		}

		// Set Default values for dynamic adpod configs
		adpod.SetDefaultValuesToPodConfig(&rCtx)

		// Validate adpod configs
		if err := adpod.ValidateV25Configs(&rCtx); err != nil {
			result.NbrCode = int(nbr.InvalidAdpodConfig)
			result.Errors = append(result.Errors, "Invalid adpod configuration: "+err.Error())
			return ep, nil
		}

		// Enable when UI support is added
		// ep.BidRequest = adpod.ApplyAdpodConfigs(rCtx, ep.BidRequest)

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-json-before-validation")

	return rCtx, result, nil
}

func (cj *CTVJSON) HandleProcessedAuctionHook(payload hookstage.ProcessedAuctionRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		rCtx, ok := utils.GetRequestContext(miCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleProcessedAuctionHook mutation")
			return parp, nil
		}

		defer func() {
			miCtx.ModuleContext.Set("rctx", rCtx)
		}()

		imps, errs := impressions.GenerateImpressions(rCtx, payload.Request)
		if len(errs) > 0 {
			for i := range errs {
				result.Warnings = append(result.Warnings, errs[i].Error())
			}
		}

		if len(imps) > 0 {
			parp.Request.SetImp(imps)
		}

		// filter vast tag durations
		adapters.FilterImpsVastTagsByDuration(rCtx, parp.Request)

		return parp, nil
	}, hookstage.MutationUpdate, "update-ctv-impressions")
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleBidderRequestHook(payload hookstage.BidderRequestPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BidderRequestPayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	result.ChangeSet.AddMutation(func(ep hookstage.BidderRequestPayload) (hookstage.BidderRequestPayload, error) {
		rCtx, ok := utils.GetRequestContext(miCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleBidderRequestHook mutation")
			return ep, nil
		}

		defer func() {
			miCtx.ModuleContext.Set("rctx", rCtx)
		}()

		// if payload.BidderInfo.OpenRTB.Version != "2.6" && len(rCtx.AdpodCtx) > 0 {
		// 	adpod.ConvertDownTo25(ep.Request)
		// }

		// Remove Multibid object from ext
		reqExt, err := ep.Request.GetRequestExt()
		if err != nil {
			result.Errors = append(result.Errors, "failed to get request ext in CTV handleBidderRequestHook mutation")
			return ep, nil
		}
		prebidExt := reqExt.GetPrebid()
		prebidExt.MultiBid = nil
		reqExt.SetPrebid(prebidExt)

		return ep, nil
	}, hookstage.MutationUpdate, "ctv-json-bidder-request")

	return rCtx, result, nil
}

func (cj *CTVJSON) HandleRawBidderResponseHook(payload hookstage.RawBidderResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.RawBidderResponsePayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleAllProcessedBidResponsesHook(payload hookstage.AllProcessedBidResponsesPayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload], error) {
	result.ChangeSet.AddMutation(func(apbrp hookstage.AllProcessedBidResponsesPayload) (hookstage.AllProcessedBidResponsesPayload, error) {
		rCtx, ok := utils.GetRequestContext(miCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleAllProcessedBidResponsesHook mutation")
			return apbrp, nil
		}

		defer func() {
			miCtx.ModuleContext.Set("rctx", rCtx)
		}()

		// Move to Raw bidder response hook once 2.6 fully supported
		adpod.ConvertUpTo26(rCtx, apbrp.Responses)
		return apbrp, nil
	}, hookstage.MutationUpdate, "update-bid-duration")
	return rCtx, result, nil
}

func (cj *CTVJSON) HandleAuctionResponseHook(payload hookstage.AuctionResponsePayload, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	// Add targetting keys
	for _, seatBid := range payload.BidResponse.SeatBid {
		targettingSeq := 1
		for _, bid := range seatBid.Bid {
			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			if bidCtx.Prebid == nil {
				bidCtx.Prebid = &openrtb_ext.ExtBidPrebid{}
			}

			if bidCtx.Prebid.Targeting == nil {
				bidCtx.Prebid.Targeting = make(map[string]string)
			}

			value := ctvutils.GetTargeting(openrtb_ext.CategoryDurationKey, openrtb_ext.BidderName(seatBid.Seat), bidCtx, targettingSeq)
			if value != "" {
				ctvutils.AddTargetingKey(bidCtx, openrtb_ext.CategoryDurationKey, value)
			}

			value = ctvutils.GetTargeting(openrtb_ext.PbKey, openrtb_ext.BidderName(seatBid.Seat), bidCtx, targettingSeq)
			if value != "" {
				ctvutils.AddTargetingKey(bidCtx, openrtb_ext.PbKey, value)
			}

			// increment targetting seq
			targettingSeq++

			impCtx.BidCtx[bid.ID] = bidCtx
			rCtx.ImpBidCtx[bid.ImpID] = impCtx
		}
	}

	// perform adpod auction
	if len(rCtx.AdpodCtx) > 0 {
		var ok bool
		result, ok = auction.AdpodAuction(&rCtx, result, payload.BidResponse)
		if !ok {
			return rCtx, result, nil
		}
	}

	result.ChangeSet.AddMutation(func(arp hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
		rCtx, ok := utils.GetRequestContext(miCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleExitpointHook mutation")
			return arp, nil
		}

		defer func() {
			miCtx.ModuleContext.Set("rctx", rCtx)
		}()

		for _, seatBid := range arp.BidResponse.SeatBid {
			for _, bid := range seatBid.Bid {
				ctvutils.AddPWTTargetingKeysForAdpod(rCtx, &bid, seatBid.Seat)
			}
		}
		return arp, nil
	}, hookstage.MutationUpdate, "add-pwt-targeting-keys-for-adpod")

	return rCtx, result, nil
}

func (cj *CTVJSON) HandleExitpointHook(payload hookstage.ExitpointPaylaod, rCtx models.RequestCtx, result hookstage.HookResult[hookstage.ExitpointPaylaod], miCtx hookstage.ModuleInvocationContext) (models.RequestCtx, hookstage.HookResult[hookstage.ExitpointPaylaod], error) {
	response, ok := payload.Response.(*openrtb2.BidResponse)
	if !ok {
		return rCtx, result, nil
	}

	if response.NBR != nil {
		return rCtx, result, nil
	}

	adpodBids := formCTVJSONResponse(&rCtx, response, cj.creativeCache)

	result.ChangeSet.AddMutation(func(ep hookstage.ExitpointPaylaod) (hookstage.ExitpointPaylaod, error) {
		rCtx, ok := utils.GetRequestContext(miCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in CTV handleExitpointHook mutation")
			return ep, nil
		}

		defer func() {
			miCtx.ModuleContext.Set("rctx", rCtx)
		}()
		ep.Response = bidResponseAdpod{AdPodBids: adpodBids, Ext: response.Ext}
		ep.W.Header().Set("Content-Type", "application/json")
		ep.W.Header().Set("Content-Options", "nosniff")
		ctvutils.SetCORSHeaders(ep.W, rCtx.Header)
		if checkRedirectResponse(rCtx) {
			redirectURL := rCtx.RedirectURL
			if len(adpodBids) > 0 {
				redirectURL = updateAdServerURL(adpodBids[0].Targeting, rCtx.RedirectURL)
			}
			ep.W.Header().Set("Location", redirectURL)
			ep.W.WriteHeader(http.StatusFound)

		}
		return ep, nil
	}, hookstage.MutationUpdate, "ctv-json-exitpoint")

	return rCtx, result, nil
}
