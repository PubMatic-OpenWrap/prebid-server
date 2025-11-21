package openwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/currency"
	"github.com/prebid/prebid-server/v3/floors"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/customdimensions"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	modelsAdunitConfig "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/ortb"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/googlesdk"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/sdkutils"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func (m OpenWrap) handleBeforeValidationHook(
	_ context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	// Validate module context, fetch request context and endpoint hook manager
	rCtx, endpointHookManager, result, ok := validateModuleContextBeforeValidationHook(moduleCtx)
	if !ok {
		return result, nil
	}

	defer func() {
		moduleCtx.ModuleContext.Set("rctx", rCtx)
		if result.Reject {
			m.metricEngine.RecordBadRequests(rCtx.Endpoint, rCtx.PubIDStr, getPubmaticErrorCode(openrtb3.NoBidReason(result.NbrCode)))
			m.metricEngine.RecordNobidErrPrebidServerRequests(rCtx.PubIDStr, result.NbrCode)
			if glog.V(models.LogLevelDebug) {
				bidRequest, _ := json.Marshal(payload.BidRequest)
				glog.Infof("[bad_request] pubid:[%d] profid:[%d] endpoint:[%s] nbr:[%d] bidrequest:[%s]",
					rCtx.PubID, rCtx.ProfileID, rCtx.Endpoint, result.NbrCode, string(bidRequest))
			}
		}
	}()

	//Do not execute the module for requests processed in SSHB(8001)
	if rCtx.Sshb == "1" {
		result.Reject = false
		return result, nil
	}

	if rCtx.Endpoint == models.EndpointHybrid {
		//TODO: Add bidder params fix
		result.Reject = false
		return result, nil
	}

	var err error
	// TODO: move this to entrypoint hook
	m.metricEngine.RecordPublisherProfileRequests(rCtx.PubIDStr, rCtx.ProfileIDStr)

	// Validate Bid Request
	result, ok = m.validateBidRequest(rCtx, result, payload.BidRequest)
	if !ok {
		return result, nil
	}

	// Get Request Extension
	rCtx.NewReqExt, err = models.GetRequestExt(payload.BidRequest.Ext)
	if err != nil {
		result.NbrCode = int(nbr.InvalidRequestExt)
		result.Errors = append(result.Errors, "failed to get request ext: "+err.Error())
		return result, nil
	}

	// Analytics flags to throttle trackers
	m.setAnalyticsFlags(&rCtx)

	// Get profile data
	rCtx.PartnerConfigMap, err = m.getProfileData(rCtx, *payload.BidRequest)
	if err != nil || len(rCtx.PartnerConfigMap) == 0 {
		// TODO: seperate DB fetch errors as internal errors
		result.NbrCode = int(nbr.InvalidProfileConfiguration)
		m.addDefaultRequestContextValues(payload.BidRequest, &rCtx)
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*payload.BidRequest) // for wrapper logger sz
		m.metricEngine.RecordPublisherInvalidProfileRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.ProfileIDStr)
		return result, errors.New("invalid profile data")
	}

	// Populate request context
	m.populateRequestContext(&rCtx, payload.BidRequest)

	// Country filter
	result, ok = handleCountryFiltering(rCtx, result)
	if !ok {
		return result, nil
	}

	// Platform
	result, ok = m.processPlatform(&rCtx, result, payload.BidRequest)
	if !ok {
		return result, nil
	}

	// Process request extension
	processRequestExtension(&rCtx)

	// Record publisher request to platform.
	m.metricEngine.RecordPublisherRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.Platform)

	if newPartnerConfigMap, ok := ABTestProcessing(rCtx); ok {
		rCtx.ABTestConfigApplied = 1
		rCtx.PartnerConfigMap = newPartnerConfigMap
		result.Warnings = append(result.Warnings, "update the rCtx.PartnerConfigMap with ABTest data")
	}

	// To check if VAST unwrap needs to be enabled for given request
	if isVastUnwrapEnabled(rCtx.PartnerConfigMap, m.cfg.Features.VASTUnwrapPercent) {
		//rCtx.ABTestConfigApplied = 1 // Re-use AB Test flag for VAST unwrap feature
		rCtx.VastUnWrap.Enabled = true
		rCtx.VastUnWrap.IsPrivacyEnforced = isPrivacyEnforced(payload.BidRequest.Regs, payload.BidRequest.Device)
	}

	//TMax should be updated after ABTest processing
	rCtx.TMax = m.setTimeout(rCtx, payload.BidRequest)

	// Partner throttling
	result, ok = m.processPartnerThrottling(&rCtx, result)
	if !ok {
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*payload.BidRequest) // for wrapper logger sz
		return result, nil
	}

	// Bidder Filtering
	result, ok = m.processBidderFiltering(&rCtx, result, payload.BidRequest)
	if !ok {
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*payload.BidRequest) // for wrapper logger sz
		return result, nil
	}

	// Price Granularity
	result, ok = processPriceGranularity(&rCtx, result, payload.BidRequest)
	if !ok {
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*payload.BidRequest) // for wrapper logger sz
		return result, nil
	}

	// Adunit Config
	rCtx.AdUnitConfig = m.cache.GetAdunitConfigFromCache(payload.BidRequest, rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)

	// Get currency rates conversions and store in rctx for tracker/logger calculation
	conversions := currency.GetAuctionCurrencyRates(m.rateConvertor, rCtx.NewReqExt.Prebid.CurrencyConversions)
	rCtx.CurrencyConversion = func(from, to string, value float64) (float64, error) {
		rate, err := conversions.GetRate(from, to)
		if err == nil {
			return value * rate, nil
		}
		return 0, err
	}

	// Adpod processing
	result, ok = m.processAdpod(&rCtx, result, payload.BidRequest)
	if !ok {
		return result, nil
	}

	// process impressions
	result, ok = m.processImpressions(&rCtx, result, payload.BidRequest)
	if !ok {
		return result, nil
	}

	// Content Transparency Object
	if cto := setContentTransparencyObject(rCtx, rCtx.NewReqExt); cto != nil {
		rCtx.NewReqExt.Prebid.Transparency = cto
	}

	// Floors
	adunitconfig.UpdateFloorsExtObjectFromAdUnitConfig(rCtx, rCtx.NewReqExt)
	setFloorsExt(rCtx.NewReqExt, &rCtx, m.pubFeatures.IsDynamicFloorEnabledPublisher(rCtx.PubID))

	// Google SDK
	rCtx.GoogleSDK.SDKRenderedAdID = googlesdk.SetSDKRenderedAdID(payload.BidRequest.App, rCtx.Endpoint)

	// Execute Endpoint specific before validation hook
	rCtx, result, err = endpointHookManager.HandleBeforeValidationHook(payload, rCtx, result, moduleCtx)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	rCtx.NewReqExt.Wrapper = nil
	rCtx.NewReqExt.Bidder = nil

	// if rCtx.Debug {
	// 	newImp, _ := json.Marshal(rCtx.ImpBidCtx)
	// 	result.DebugMessages = append(result.DebugMessages, "new imp: "+string(newImp))
	// 	newReqExt, _ := json.Marshal(rCtx.NewReqExt)
	// 	result.DebugMessages = append(result.DebugMessages, "new request.ext: "+string(newReqExt))
	// }

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		rctx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "failed to get request context in handleBeforeValidationHook mutation")
			return ep, nil
		}

		defer func() {
			moduleCtx.ModuleContext.Set("rctx", rctx)
		}()

		var err error
		ep.BidRequest, err = m.applyProfileChanges(rctx, ep.BidRequest)
		if err != nil {
			result.Errors = append(result.Errors, "failed to apply profile changes: "+err.Error())
		}

		if rctx.IsApplovinSchainABTestEnabled && ep.BidRequest.Source != nil {
			m.updateAppLovinMaxRequestSchain(&rctx, ep.BidRequest)
		}
		return ep, err
	}, hookstage.MutationUpdate, "request-body-with-profile-data")

	result.Reject = false
	return result, nil
}

// applyProfileChanges copies and updates BidRequest with required values from http header and partnetConfigMap
func (m *OpenWrap) applyProfileChanges(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest) (*openrtb2.BidRequest, error) {
	if rctx.IsTestRequest > 0 {
		bidRequest.Test = 1
	}

	if sdkutils.IsSdkIntegration(rctx.Endpoint) && rctx.AppStoreUrl != "" {
		bidRequest.App.StoreURL = rctx.AppStoreUrl
	}

	// Remove app.ext.token
	if rctx.Endpoint == models.EndpointUnityLevelPlay {
		bidRequest.App.Ext = jsonparser.Delete(bidRequest.App.Ext, "token")
	}

	googleSSUFeatureEnabled := models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.GoogleSSUFeatureEnabledKey) == models.Enabled
	if googleSSUFeatureEnabled {
		if rctx.NewReqExt == nil {
			rctx.NewReqExt = &models.RequestExt{}
		}
		rctx.NewReqExt.Prebid.GoogleSSUFeatureEnabled = googleSSUFeatureEnabled
	}
	if cur, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID][models.AdServerCurrency]; ok {
		bidRequest.Cur = append(bidRequest.Cur, cur)
	}
	if bidRequest.TMax == 0 {
		bidRequest.TMax = rctx.TMax
	}

	if bidRequest.Source == nil {
		bidRequest.Source = &openrtb2.Source{}
	}
	bidRequest.Source.TID = bidRequest.ID

	for i := 0; i < len(bidRequest.Imp); i++ {
		if rctx.Endpoint != models.EndpointAMP {
			m.applyBannerAdUnitConfig(rctx, &bidRequest.Imp[i])
		}
		m.applyVideoAdUnitConfig(rctx, &bidRequest.Imp[i])
		m.applyNativeAdUnitConfig(rctx, &bidRequest.Imp[i])
		m.applyImpChanges(rctx, &bidRequest.Imp[i])
	}

	setSChainInRequest(rctx.NewReqExt, bidRequest.Source, rctx.PartnerConfigMap)

	adunitconfig.ReplaceAppObjectFromAdUnitConfig(rctx, bidRequest.App)
	adunitconfig.ReplaceDeviceTypeFromAdUnitConfig(rctx, &bidRequest.Device)
	bidRequest.Device.IP = rctx.DeviceCtx.IP
	bidRequest.Device.Language = getValidLanguage(bidRequest.Device.Language)
	amendDeviceObject(bidRequest.Device, &rctx.DeviceCtx)

	if bidRequest.User == nil {
		bidRequest.User = &openrtb2.User{}
	}
	if bidRequest.User.CustomData == "" && rctx.KADUSERCookie != nil {
		bidRequest.User.CustomData = rctx.KADUSERCookie.Value
	}
	UpdateUserExtWithValidValues(bidRequest.User)

	for i := 0; i < len(bidRequest.WLang); i++ {
		bidRequest.WLang[i] = getValidLanguage(bidRequest.WLang[i])
	}

	if bidRequest.Site != nil && bidRequest.Site.Content != nil {
		bidRequest.Site.Content.Language = getValidLanguage(bidRequest.Site.Content.Language)
	} else if bidRequest.App != nil && bidRequest.App.Content != nil {
		bidRequest.App.Content.Language = getValidLanguage(bidRequest.App.Content.Language)
	}

	var err error
	var requestExtjson json.RawMessage
	if rctx.NewReqExt != nil {
		requestExtjson, err = json.Marshal(rctx.NewReqExt)
		bidRequest.Ext = requestExtjson
	}
	return bidRequest, err
}

func (m *OpenWrap) applyVideoAdUnitConfig(rCtx models.RequestCtx, imp *openrtb2.Imp) {
	//For AMP request, if AmpVideoEnabled is true then crate a empty video object and update with adunitConfigs
	if rCtx.AmpVideoEnabled {
		imp.Video = &openrtb2.Video{}
	}

	if imp.Video == nil {
		return
	}

	adUnitCfg := rCtx.ImpBidCtx[imp.ID].VideoAdUnitCtx.AppliedSlotAdUnitConfig
	if adUnitCfg == nil {
		return
	}

	impBidCtx := rCtx.ImpBidCtx[imp.ID]
	imp.BidFloor, imp.BidFloorCur = getImpBidFloorParams(rCtx, adUnitCfg, imp, m.rateConvertor.Rates())
	impBidCtx.BidFloor = imp.BidFloor
	impBidCtx.BidFloorCur = imp.BidFloorCur

	rCtx.ImpBidCtx[imp.ID] = impBidCtx

	if adUnitCfg.Exp != nil {
		imp.Exp = int64(*adUnitCfg.Exp)
	}

	if adUnitCfg.Video == nil {
		return
	}

	//check if video is disabled, if yes then remove video from imp object
	if adUnitCfg.Video.Enabled != nil && !*adUnitCfg.Video.Enabled {
		imp.Video = nil
		impBidCtx.Video = nil
		rCtx.ImpBidCtx[imp.ID] = impBidCtx
		return
	}

	//For AMP request if AmpVideoEnabled is true then, update the imp.video object with adunitConfig and if adunitConfig is not present then update with default values
	if rCtx.AmpVideoEnabled {
		if adUnitCfg.Video.Config != nil {
			updateImpVideoWithVideoConfig(imp, adUnitCfg.Video.Config)
		}
		updateAmpImpVideoWithDefault(imp)
		return
	}

	if adUnitCfg.Video.Config != nil {
		updateImpVideoWithVideoConfig(imp, adUnitCfg.Video.Config)
	}
}

func (m *OpenWrap) applyImpChanges(rCtx models.RequestCtx, imp *openrtb2.Imp) {
	if imp.BidFloor == 0 {
		imp.BidFloorCur = ""
	} else if imp.BidFloorCur == "" {
		imp.BidFloorCur = models.USD
	}

	if imp.Video != nil {
		m.applyImpVideoChanges(rCtx, imp.Video)
	}

	//update secure for applovin
	if rCtx.Endpoint == models.EndpointAppLovinMax {
		imp.Secure = openrtb2.Int8Ptr(1)
	}

	//update impression extensions
	imp.Ext = rCtx.ImpBidCtx[imp.ID].NewExt
}

func (m *OpenWrap) applyImpVideoChanges(rCtx models.RequestCtx, video *openrtb2.Video) {
	//update protocols
	if rCtx.NewReqExt != nil && rCtx.NewReqExt.Prebid.GoogleSSUFeatureEnabled {
		video.Protocols = UpdateImpProtocols(video.Protocols)
	}

	//update video.plcmt from video.placements
	if video.Placement > 0 && video.Plcmt == 0 {
		//TODO: move to ConvertUpTo26 once upgraded to prebid3.x version
		switch video.Placement {
		case adcom1.VideoPlacementInStream:
			video.Plcmt = adcom1.VideoPlcmtInstream
		case adcom1.VideoPlacementInBanner:
			video.Plcmt = adcom1.VideoPlcmtNoContent
		case adcom1.VideoPlacementAlwaysVisible:
			video.Plcmt = adcom1.VideoPlcmtInterstitial
		}
	}
}

func getImpBidFloorParams(rCtx models.RequestCtx, adUnitCfg *modelsAdunitConfig.AdConfig, imp *openrtb2.Imp, conversions currency.Conversions) (float64, string) {
	bidfloor := imp.BidFloor
	bidfloorcur := imp.BidFloorCur

	if rCtx.IsMaxFloorsEnabled && adUnitCfg.BidFloor != nil {
		bidfloor, bidfloorcur, _ = floors.GetMaxFloorValue(bidfloor, bidfloorcur, *adUnitCfg.BidFloor, *adUnitCfg.BidFloorCur, conversions)
	} else {
		if bidfloor == 0 && adUnitCfg.BidFloor != nil {
			//use adunitconfig bidfloor and bidfloorcur
			bidfloor = *adUnitCfg.BidFloor
			bidfloorcur = ""

			if adUnitCfg.BidFloorCur != nil {
				bidfloorcur = *adUnitCfg.BidFloorCur
			}
		}
	}

	if bidfloor == 0 {
		//no bidfloor value
		return 0, ""
	}
	if bidfloorcur == "" {
		bidfloorcur = models.USD
	}
	return bidfloor, bidfloorcur
}

func (m *OpenWrap) applyBannerAdUnitConfig(rCtx models.RequestCtx, imp *openrtb2.Imp) {
	if imp.Banner == nil {
		return
	}

	adUnitCfg := rCtx.ImpBidCtx[imp.ID].BannerAdUnitCtx.AppliedSlotAdUnitConfig
	if adUnitCfg == nil {
		return
	}

	impBidCtx := rCtx.ImpBidCtx[imp.ID]
	imp.BidFloor, imp.BidFloorCur = getImpBidFloorParams(rCtx, adUnitCfg, imp, m.rateConvertor.Rates())
	impBidCtx.BidFloor = imp.BidFloor
	impBidCtx.BidFloorCur = imp.BidFloorCur
	rCtx.ImpBidCtx[imp.ID] = impBidCtx

	if adUnitCfg.Exp != nil {
		imp.Exp = int64(*adUnitCfg.Exp)
	}

	if adUnitCfg.Banner == nil {
		return
	}

	if adUnitCfg.Banner.Enabled != nil && !*adUnitCfg.Banner.Enabled {
		imp.Banner = nil
		return
	}
}

// isVastUnwrapEnabled return whether to enable vastunwrap or not
func isVastUnwrapEnabled(partnerConfigMap map[int]map[string]string, vastUnwrapTraffic int) bool {
	trafficPercentage := vastUnwrapTraffic
	unwrapEnabled := models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.VastUnwrapperEnableKey) == models.Enabled
	if unwrapEnabled {
		if value := models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.VastUnwrapTrafficPercentKey); len(value) > 0 {
			if trafficPercentDB, err := strconv.Atoi(value); err == nil {
				trafficPercentage = trafficPercentDB
			}
		}
	}
	return unwrapEnabled && GetRandomNumberIn1To100() <= trafficPercentage
}

func getDomainFromUrl(pageUrl string) string {
	u, err := url.Parse(pageUrl)
	if err != nil {
		return ""
	}

	return u.Host
}

// NYC: make this generic. Do we need this?. PBS now has auto_gen_source_tid generator. We can make it to wiid for pubmatic adapter in pubmatic.go
func updateRequestExtBidderParamsPubmatic(bidderParams json.RawMessage, cookie, loggerID, bidderCode string, sendBurl bool) (json.RawMessage, error) {
	bidderParamsMap := make(map[string]map[string]interface{})
	_ = json.Unmarshal(bidderParams, &bidderParamsMap) // ignore error, incoming might be nil for now but we still have data to put

	bidderParamsMap[bidderCode] = map[string]interface{}{
		models.WrapperLoggerImpID: loggerID,
	}

	if len(cookie) != 0 {
		bidderParamsMap[bidderCode][models.COOKIE] = cookie
	}

	if sendBurl {
		bidderParamsMap[bidderCode][models.SendBurl] = true
	}

	return json.Marshal(bidderParamsMap)
}

func getPageURL(bidRequest *openrtb2.BidRequest) string {
	if bidRequest.App != nil && bidRequest.App.StoreURL != "" {
		return bidRequest.App.StoreURL
	} else if bidRequest.Site != nil && bidRequest.Site.Page != "" {
		return bidRequest.Site.Page
	}
	return ""
}

// getVASTEventMacros populates macros with PubMatic specific macros
// These marcros is used in replacing with actual values of Macros in case of Video Event tracke URLs
// If this function fails to determine value of any macro then it continues with next macro setup
// returns true when at least one macro is added to map
func getVASTEventMacros(rctx models.RequestCtx) map[string]string {
	macros := map[string]string{
		string(models.MacroProfileID):           fmt.Sprintf("%d", rctx.ProfileID),
		string(models.MacroProfileVersionID):    fmt.Sprintf("%d", rctx.DisplayVersionID),
		string(models.MacroUnixTimeStamp):       fmt.Sprintf("%d", rctx.StartTime),
		string(models.MacroPlatform):            fmt.Sprintf("%d", rctx.DeviceCtx.Platform),
		string(models.MacroWrapperImpressionID): rctx.LoggerImpressionID,
	}

	if rctx.SSAI != "" {
		macros[string(models.MacroSSAI)] = rctx.SSAI
	}

	return macros
}

func updateAliasGVLIds(aliasgvlids map[string]uint16, bidderCode string, partnerConfig map[string]string) {
	if vendorID, ok := partnerConfig[models.VENDORID]; ok && vendorID != "" {
		vid, err := strconv.ParseUint(vendorID, 10, 64)
		if err != nil {
			return
		}

		if vid == 0 {
			return
		}
		aliasgvlids[bidderCode] = uint16(vid)
	}
}

// setTimeout - This utility returns timeout applicable for a profile
func (m OpenWrap) setTimeout(rCtx models.RequestCtx, req *openrtb2.BidRequest) int64 {
	var auctionTimeout int64

	// BidRequest.TMax has highest priority
	if req.TMax != 0 {
		auctionTimeout = req.TMax
	} else {
		//check for ssTimeout in the partner config
		ssTimeout := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.SSTimeoutKey)
		if ssTimeout != "" {
			ssTimeoutDB, err := strconv.Atoi(ssTimeout)
			if err == nil {
				auctionTimeout = int64(ssTimeoutDB)
			}
		}
	}

	// found tmax value in request or db
	if auctionTimeout != 0 {
		if auctionTimeout < m.cfg.Timeout.MinTimeout {
			return m.cfg.Timeout.MinTimeout
		} else if auctionTimeout > m.cfg.Timeout.MaxTimeout {
			return m.cfg.Timeout.MaxTimeout
		}
		return auctionTimeout
	}

	//Below piece of code is applicable for older profiles where ssTimeout is not set
	//Here we will check the partner timeout and select max timeout considering timeout range
	auctionTimeout = m.cfg.Timeout.MinTimeout
	for _, partnerConfig := range rCtx.PartnerConfigMap {
		partnerTO, _ := strconv.Atoi(partnerConfig[models.TIMEOUT])
		if int64(partnerTO) > m.cfg.Timeout.MaxTimeout {
			auctionTimeout = m.cfg.Timeout.MaxTimeout
			break
		}
		if int64(partnerTO) >= m.cfg.Timeout.MinTimeout && auctionTimeout < int64(partnerTO) {
			auctionTimeout = int64(partnerTO)

		}
	}
	return auctionTimeout
}

// isSendAllBids returns true in below cases:
// if ssauction flag is set 0 in the request
// if ssauction flag is not set and platform is dislay, then by default send all bids
// if ssauction flag is not set and platform is in-app, then check if profile setting sendAllBids is set to 1
func isSendAllBids(rctx models.RequestCtx) bool {
	//for webs2s endpoint SendAllBids is always true
	if rctx.Endpoint == models.EndpointWebS2S {
		return true
	}
	//if ssauction is set to 0 in the request
	if rctx.SSAuction == 0 {
		return true
	} else if rctx.SSAuction == -1 && rctx.Platform == models.PLATFORM_APP {
		// if platform is in-app, then check if profile setting sendAllBids is set to 1
		if models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.SendAllBidsKey) == "1" {
			return true
		}
	}
	return false
}

func getValidLanguage(language string) string {
	if len(language) > 2 {
		lang := language[0:2]
		if models.ValidCode(lang) {
			return lang
		}
	}
	return language
}

func isSlotEnabled(imp openrtb2.Imp, videoAdUnitCtx, bannerAdUnitCtx, nativeAdUnitCtx models.AdUnitCtx) bool {
	videoEnabled := true
	if imp.Video == nil || (videoAdUnitCtx.AppliedSlotAdUnitConfig != nil && videoAdUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
		videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Enabled != nil && !*videoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Enabled) {
		videoEnabled = false
	}

	bannerEnabled := true
	if imp.Banner == nil || (bannerAdUnitCtx.AppliedSlotAdUnitConfig != nil && bannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner != nil &&
		bannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Enabled != nil && !*bannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Enabled) {
		bannerEnabled = false
	}

	nativeEnabled := true
	if imp.Native == nil || (nativeAdUnitCtx.AppliedSlotAdUnitConfig != nil && nativeAdUnitCtx.AppliedSlotAdUnitConfig.Native != nil &&
		nativeAdUnitCtx.AppliedSlotAdUnitConfig.Native.Enabled != nil && !*nativeAdUnitCtx.AppliedSlotAdUnitConfig.Native.Enabled) {
		nativeEnabled = false
	}

	return videoEnabled || bannerEnabled || nativeEnabled
}

func getPubID(bidRequest openrtb2.BidRequest) (pubID int, err error) {
	if bidRequest.Site != nil && bidRequest.Site.Publisher != nil && bidRequest.Site.Publisher.ID != "" {
		pubID, err = strconv.Atoi(bidRequest.Site.Publisher.ID)
	} else if bidRequest.App != nil && bidRequest.App.Publisher != nil && bidRequest.App.Publisher.ID != "" {
		pubID, err = strconv.Atoi(bidRequest.App.Publisher.ID)
	}
	return pubID, err
}

func getTagID(imp openrtb2.Imp, impExt *models.ImpExtension) string {
	//priority for tagId is imp.ext.gpid > imp.TagID > imp.ext.data.pbadslot
	if impExt.GpId != "" {
		if idx := strings.Index(impExt.GpId, "#"); idx != -1 {
			return impExt.GpId[:idx]
		}
		return impExt.GpId
	} else if imp.TagID != "" {
		return imp.TagID
	}
	return impExt.Data.PbAdslot
}

func (m OpenWrap) setAnalyticsFlags(rCtx *models.RequestCtx) {
	rCtx.LoggerDisabled, rCtx.TrackerDisabled = m.pubFeatures.IsAnalyticsTrackingThrottled(rCtx.PubID, rCtx.ProfileID)

	if rCtx.LoggerDisabled {
		rCtx.MetricsEngine.RecordAnalyticsTrackingThrottled(strconv.Itoa(rCtx.PubID), strconv.Itoa(rCtx.ProfileID), models.AnanlyticsThrottlingLoggerType)
	}

	if rCtx.TrackerDisabled {
		rCtx.MetricsEngine.RecordAnalyticsTrackingThrottled(strconv.Itoa(rCtx.PubID), strconv.Itoa(rCtx.ProfileID), models.AnanlyticsThrottlingTrackerType)
	}
}

func updateImpVideoWithVideoConfig(imp *openrtb2.Imp, configObjInVideoConfig *modelsAdunitConfig.VideoConfig) {
	if len(imp.Video.MIMEs) == 0 {
		imp.Video.MIMEs = configObjInVideoConfig.MIMEs
	}

	if imp.Video.MinDuration == 0 {
		imp.Video.MinDuration = configObjInVideoConfig.MinDuration
	}

	if imp.Video.MaxDuration == 0 {
		imp.Video.MaxDuration = configObjInVideoConfig.MaxDuration
	}

	if imp.Video.Skip == nil {
		imp.Video.Skip = configObjInVideoConfig.Skip
	}

	if imp.Video.SkipMin == 0 {
		imp.Video.SkipMin = configObjInVideoConfig.SkipMin
	}

	if imp.Video.SkipAfter == 0 {
		imp.Video.SkipAfter = configObjInVideoConfig.SkipAfter
	}

	if len(imp.Video.BAttr) == 0 {
		imp.Video.BAttr = configObjInVideoConfig.BAttr
	}

	if imp.Video.MinBitRate == 0 {
		imp.Video.MinBitRate = configObjInVideoConfig.MinBitRate
	}

	if imp.Video.MaxBitRate == 0 {
		imp.Video.MaxBitRate = configObjInVideoConfig.MaxBitRate
	}

	if imp.Video.MaxExtended == 0 {
		imp.Video.MaxExtended = configObjInVideoConfig.MaxExtended
	}

	if imp.Video.StartDelay == nil {
		imp.Video.StartDelay = configObjInVideoConfig.StartDelay
	}

	if imp.Video.Placement == 0 {
		imp.Video.Placement = configObjInVideoConfig.Placement
	}

	if imp.Video.Plcmt == 0 {
		imp.Video.Plcmt = configObjInVideoConfig.Plcmt
	}

	if imp.Video.Linearity == 0 {
		imp.Video.Linearity = configObjInVideoConfig.Linearity
	}

	if imp.Video.Protocol == 0 {
		imp.Video.Protocol = configObjInVideoConfig.Protocol
	}

	if len(imp.Video.Protocols) == 0 {
		imp.Video.Protocols = configObjInVideoConfig.Protocols
	}

	if imp.Video.W == nil {
		imp.Video.W = configObjInVideoConfig.W
	}

	if imp.Video.H == nil {
		imp.Video.H = configObjInVideoConfig.H
	}

	if imp.Video.Sequence == 0 {
		imp.Video.Sequence = configObjInVideoConfig.Sequence
	}

	if imp.Video.BoxingAllowed == nil {
		imp.Video.BoxingAllowed = configObjInVideoConfig.BoxingAllowed
	}

	if len(imp.Video.PlaybackMethod) == 0 {
		imp.Video.PlaybackMethod = configObjInVideoConfig.PlaybackMethod
	}

	if imp.Video.PlaybackEnd == 0 {
		imp.Video.PlaybackEnd = configObjInVideoConfig.PlaybackEnd
	}

	if imp.Video.Delivery == nil {
		imp.Video.Delivery = configObjInVideoConfig.Delivery
	}

	if imp.Video.Pos == nil {
		imp.Video.Pos = configObjInVideoConfig.Pos
	}

	if len(imp.Video.API) == 0 {
		imp.Video.API = configObjInVideoConfig.API
	}

	if len(imp.Video.CompanionType) == 0 {
		imp.Video.CompanionType = configObjInVideoConfig.CompanionType
	}

	if imp.Video.CompanionAd == nil {
		imp.Video.CompanionAd = configObjInVideoConfig.CompanionAd
	}
}

func updateAmpImpVideoWithDefault(imp *openrtb2.Imp) {
	if imp.Video.W == nil {
		imp.Video.W = getW(imp)
	}
	if imp.Video.H == nil {
		imp.Video.H = getH(imp)
	}
	if imp.Video.MIMEs == nil {
		imp.Video.MIMEs = []string{"video/mp4"}
	}
	if imp.Video.MinDuration == 0 {
		imp.Video.MinDuration = 0
	}
	if imp.Video.MaxDuration == 0 {
		imp.Video.MaxDuration = 30
	}
	if imp.Video.StartDelay == nil {
		imp.Video.StartDelay = adcom1.StartPreRoll.Ptr()
	}
	if imp.Video.Protocols == nil {
		imp.Video.Protocols = []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper}
	}
	if imp.Video.Placement == 0 {
		imp.Video.Placement = adcom1.VideoPlacementInBanner
	}
	if imp.Video.Plcmt == 0 {
		imp.Video.Plcmt = adcom1.VideoPlcmtNoContent
	}
	if imp.Video.Linearity == 0 {
		imp.Video.Linearity = adcom1.LinearityLinear
	}
	if imp.Video.Skip == nil {
		imp.Video.Skip = ptrutil.ToPtr[int8](0)
	}
	if imp.Video.PlaybackMethod == nil {
		imp.Video.PlaybackMethod = []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOff}
	}
	if imp.Video.PlaybackEnd == 0 {
		imp.Video.PlaybackEnd = adcom1.PlaybackCompletion
	}
	if imp.Video.Delivery == nil {
		imp.Video.Delivery = []adcom1.DeliveryMethod{adcom1.DeliveryProgressive, adcom1.DeliveryDownload}
	}
}

func getW(imp *openrtb2.Imp) *int64 {
	if imp.Banner != nil {
		if imp.Banner.W != nil {
			return imp.Banner.W
		}
		for _, format := range imp.Banner.Format {
			if format.W != 0 {
				return &format.W
			}
		}
	}
	return nil
}

func getH(imp *openrtb2.Imp) *int64 {
	if imp.Banner != nil {
		if imp.Banner.H != nil {
			return imp.Banner.H
		}
		for _, format := range imp.Banner.Format {
			if format.H != 0 {
				return &format.H
			}
		}
	}
	return nil
}

func getProfileAppStoreUrl(rctx models.RequestCtx) (string, bool) {
	isValidAppStoreUrl := false
	appStoreUrl := rctx.PartnerConfigMap[models.VersionLevelConfigID][models.AppStoreUrl]
	if appStoreUrl == "" {
		glog.Errorf("[PubID]: %d [ProfileID]: %d [Error]: app store url not present in DB", rctx.PubID, rctx.ProfileID)
		return appStoreUrl, isValidAppStoreUrl
	}
	appStoreUrl = strings.TrimSpace(appStoreUrl)
	if !utils.IsValidURL(appStoreUrl) {
		glog.Errorf("[PubID]: %d [ProfileID]: %d [AppStoreUrl]: %s [Error]: Invalid app store url", rctx.PubID, rctx.ProfileID, appStoreUrl)
		return appStoreUrl, isValidAppStoreUrl
	}
	isValidAppStoreUrl = true
	return appStoreUrl, isValidAppStoreUrl
}

func updateSkadnSourceapp(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest, impExt *models.ImpExtension) {
	if bidRequest.Device == nil || strings.ToLower(bidRequest.Device.OS) != "ios" {
		return
	}

	if impExt == nil || impExt.SKAdnetwork == nil {
		glog.Errorf("[PubID]: %d [ProfileID]: %d [Error]: skadn is missing in imp.ext", rctx.PubID, rctx.ProfileID)
		return
	}

	itunesID := extractItunesIdFromAppStoreUrl(rctx.AppStoreUrl)
	if itunesID == "" {
		rctx.MetricsEngine.RecordFailedParsingItuneID(rctx.PubIDStr, rctx.ProfileIDStr)
		glog.Errorf("[PubID]: %d [ProfileID]: %d [AppStoreUrl]: %s [Error]: itunes id is missing in app store url", rctx.PubID, rctx.ProfileID, rctx.AppStoreUrl)
		return
	}

	if updatedSKAdnetwork, err := jsonparser.Set(impExt.SKAdnetwork, []byte(strconv.Quote(itunesID)), "sourceapp"); err != nil {
		glog.Errorf("[PubID]: %d [ProfileID]: %d [AppStoreUrl]: %s [Error]: %s", rctx.PubID, rctx.ProfileID, rctx.AppStoreUrl, err.Error())
	} else {
		impExt.SKAdnetwork = updatedSKAdnetwork
	}
}

func extractItunesIdFromAppStoreUrl(url string) string {
	url = strings.TrimSuffix(url, "/")
	itunesID := ""
	for i := len(url) - 1; i >= 0; i-- {
		char := rune(url[i])
		if unicode.IsDigit(char) {
			itunesID = string(char) + itunesID
		} else {
			break
		}
	}
	return itunesID
}

func (m *OpenWrap) applyNativeAdUnitConfig(rCtx models.RequestCtx, imp *openrtb2.Imp) {
	if imp.Native == nil {
		return
	}

	impCtx, ok := rCtx.ImpBidCtx[imp.ID]
	if !ok {
		return
	}
	adUnitCfg := impCtx.NativeAdUnitCtx.AppliedSlotAdUnitConfig
	if adUnitCfg == nil {
		return
	}

	impBidCtx := rCtx.ImpBidCtx[imp.ID]
	imp.BidFloor, imp.BidFloorCur = getImpBidFloorParams(rCtx, adUnitCfg, imp, m.rateConvertor.Rates())
	impBidCtx.BidFloor = imp.BidFloor
	impBidCtx.BidFloorCur = imp.BidFloorCur
	rCtx.ImpBidCtx[imp.ID] = impBidCtx

	if adUnitCfg.Exp != nil {
		imp.Exp = int64(*adUnitCfg.Exp)
	}

	if adUnitCfg.Native == nil {
		return
	}

	if adUnitCfg.Native.Enabled != nil && !*adUnitCfg.Native.Enabled {
		imp.Native = nil
		return
	}
}

// validateModuleContext validates that required context is available
func validateModuleContextBeforeValidationHook(moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	result := hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
		Reject: true,
	}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBeforeValidationHook()")
		return models.RequestCtx{}, nil, result, false
	}

	rCtxInterface, ok := moduleCtx.ModuleContext.Get("rctx")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return models.RequestCtx{}, nil, result, false
	}
	rCtx, ok := rCtxInterface.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return models.RequestCtx{}, nil, result, false
	}

	endpointHookManagerInterface, ok := moduleCtx.ModuleContext.Get("endpointhookmanager")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBeforeValidationHook()")
		return models.RequestCtx{}, nil, result, false
	}
	endpointHookManager, ok := endpointHookManagerInterface.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBeforeValidationHook()")
		return models.RequestCtx{}, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}

// validateBidRequest performs basic validation of the bid request
func (m OpenWrap) validateBidRequest(rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bidRequest *openrtb2.BidRequest) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	// Check for empty impressions or missing site/app
	if len(bidRequest.Imp) == 0 || (bidRequest.Site == nil && bidRequest.App == nil) {
		result.Reject = false
		m.metricEngine.RecordBadRequests(rCtx.Endpoint, rCtx.PubIDStr, getPubmaticErrorCode(nbr.InvalidRequestExt))
		m.metricEngine.RecordNobidErrPrebidServerRequests(rCtx.PubIDStr, int(nbr.InvalidRequestExt))
		return result, false
	}

	return result, true
}

// populateRequestContext populates the request context with data from the bid request
func (m OpenWrap) populateRequestContext(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) {
	rCtx.Source, rCtx.Origin = getSourceAndOrigin(bidRequest)
	rCtx.PageURL = getPageURL(bidRequest)
	rCtx.Platform = getPlatformFromRequest(bidRequest)
	rCtx.HostName = m.cfg.Server.HostName
	rCtx.ReturnAllBidStatus = rCtx.NewReqExt.Prebid.ReturnAllBidStatus

	// Device
	rCtx.DeviceCtx.UA = getUserAgent(bidRequest, rCtx.DeviceCtx.UA)
	rCtx.DeviceCtx.IP = getIP(bidRequest, rCtx.DeviceCtx.IP)
	rCtx.DeviceCtx.Country = getCountry(bidRequest)
	rCtx.DeviceCtx.DerivedCountryCode, _ = m.getCountryCodes(rCtx.DeviceCtx.IP)
	rCtx.DeviceCtx.Platform = getDevicePlatform(*rCtx, bidRequest)
	populateDeviceContext(&rCtx.DeviceCtx, bidRequest.Device)

	// Features
	rCtx.IsMaxFloorsEnabled = rCtx.Endpoint == models.EndpointAppLovinMax && m.pubFeatures.IsMaxFloorsEnabled(rCtx.PubID)
	rCtx.IsTBFFeatureEnabled = m.pubFeatures.IsTBFFeatureEnabled(rCtx.PubID, rCtx.ProfileID)
	rCtx.CustomDimensions = customdimensions.GetCustomDimensions(rCtx.NewReqExt.Prebid.BidderParams)
	rCtx.IsApplovinSchainABTestEnabled = rCtx.Endpoint == models.EndpointAppLovinMax && getApplovinSchainABTestEnabled(m.pubFeatures.GetApplovinSchainABTestPercentage())

	// Set test request flag if present in bid request
	if bidRequest.Test != 0 {
		rCtx.IsTestRequest = bidRequest.Test
	}

	// Display version
	if ver, err := strconv.Atoi(models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.DisplayVersionID)); err == nil {
		rCtx.DisplayVersionID = ver
	}

	//set the profile MetaData for logging and tracking
	rCtx.ProfileType = getProfileType(rCtx.PartnerConfigMap)
	rCtx.ProfileTypePlatform = getProfileTypePlatform(rCtx.PartnerConfigMap, m.profileMetaData)
	rCtx.AppPlatform = getAppPlatform(rCtx.PartnerConfigMap)
	rCtx.AppIntegrationPath = ptrutil.ToPtr(getAppIntegrationPath(rCtx.PartnerConfigMap, m.profileMetaData))
	rCtx.AppSubIntegrationPath = ptrutil.ToPtr(getAppSubIntegrationPath(rCtx.PartnerConfigMap, m.profileMetaData))
}

// processRequestExtension extracts and processes the request extension
func processRequestExtension(rCtx *models.RequestCtx) {
	rCtx.NewReqExt.Prebid.Debug = rCtx.Debug
	rCtx.NewReqExt.Prebid.DebugOverride = rCtx.WakandaDebug.IsEnable()
	rCtx.NewReqExt.Prebid.SupportDeals = rCtx.SupportDeals && rCtx.IsCTVRequest // TODO: verify usecase of Prefered deals vs Support details
	rCtx.NewReqExt.Prebid.ExtOWRequestPrebid.TrackerDisabled = rCtx.TrackerDisabled
	rCtx.NewReqExt.Prebid.AlternateBidderCodes, rCtx.MarketPlaceBidders = getMarketplaceBidders(rCtx.NewReqExt.Prebid.AlternateBidderCodes, rCtx.PartnerConfigMap)
	// TODO: Check if we can directly accept keyVal in prebid ext
	if rCtx.NewReqExt.Wrapper != nil && rCtx.NewReqExt.Wrapper.KeyValues != nil {
		rCtx.NewReqExt.Prebid.KeyVal = rCtx.NewReqExt.Wrapper.KeyValues
	}
	rCtx.NewReqExt.Prebid.BidAdjustmentFactors = map[string]float64{}
}

// handleCountryFiltering applies country filtering if needed
func handleCountryFiltering(rCtx models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload]) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	if shouldApplyCountryFilter(rCtx.Endpoint) && rCtx.DeviceCtx.DerivedCountryCode != "" {
		mode, countryCodes := getCountryFilterConfig(rCtx.PartnerConfigMap)
		if !isCountryAllowed(rCtx.DeviceCtx.DerivedCountryCode, mode, countryCodes) {
			result.NbrCode = int(nbr.RequestBlockedGeoFiltered)
			result.Errors = append(result.Errors, "Request rejected due to country filter")
			return result, false
		}
	}

	return result, true
}

// validatePlatform validates and process the platform information
func (m OpenWrap) processPlatform(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bidRequest *openrtb2.BidRequest) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	platform := rCtx.GetVersionLevelKey(models.PLATFORM_KEY)
	if platform == "" {
		result.NbrCode = int(nbr.InvalidPlatform)
		result.Errors = append(result.Errors, "failed to get platform data")
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*bidRequest) // for wrapper logger sz
		m.metricEngine.RecordPublisherInvalidProfileRequests(rCtx.Endpoint, rCtx.PubIDStr, rCtx.ProfileIDStr)
		return result, false
	}
	rCtx.Platform = platform
	rCtx.DeviceCtx.Platform = getDevicePlatform(*rCtx, bidRequest)
	rCtx.SendAllBids = isSendAllBids(*rCtx)
	return result, true
}

// processPartnerThrottling applies partner throttling
func (m OpenWrap) processPartnerThrottling(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload]) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	var allPartnersThrottledFlag bool

	rCtx.AdapterThrottleMap, allPartnersThrottledFlag = m.applyPartnerThrottling(*rCtx)
	if allPartnersThrottledFlag {
		result.NbrCode = int(nbr.RequestBlockedGeoFiltered)
		result.Errors = append(result.Errors, "All adapters Blocked due to Geo Filtering")
		glog.V(models.LogLevelDebug).Info("All adapters Blocked due to Geo Filtering")
		return result, false
	}

	// Get adapter throttle map
	rCtx.AdapterThrottleMap, allPartnersThrottledFlag = GetAdapterThrottleMap(rCtx.PartnerConfigMap, rCtx.AdapterThrottleMap)
	if allPartnersThrottledFlag {
		result.NbrCode = int(nbr.AllPartnerThrottled)
		result.Errors = append(result.Errors, "All adapters throttled")
		return result, false
	}

	return result, true
}

// processBidderFiltering get filtered bidders
func (m OpenWrap) processBidderFiltering(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bidRequest *openrtb2.BidRequest) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	var allPartnersFilteredFlag bool
	rCtx.AdapterFilteredMap, allPartnersFilteredFlag = m.getFilteredBidders(*rCtx, bidRequest)
	result.SeatNonBid = getSeatNonBid(rCtx.AdapterFilteredMap, bidRequest)
	if allPartnersFilteredFlag {
		result.NbrCode = int(nbr.AllPartnersFiltered)
		result.Errors = append(result.Errors, "All partners filtered")
		return result, false
	}

	return result, true
}

// processPriceGranularity processes price granularity
func processPriceGranularity(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bidRequest *openrtb2.BidRequest) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	priceGranularity, err := computePriceGranularity(*rCtx)
	if err != nil {
		result.NbrCode = int(nbr.InvalidPriceGranularityConfig)
		result.Errors = append(result.Errors, "failed to price granularity details: "+err.Error())
		rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*bidRequest) // for wrapper logger sz
		return result, false
	}
	rCtx.PriceGranularity = &priceGranularity

	// Set targeting
	rCtx.NewReqExt.Prebid.Targeting = &openrtb_ext.ExtRequestTargeting{
		PriceGranularity:  &priceGranularity,
		IncludeBidderKeys: ptrutil.ToPtr(true),
		IncludeWinners:    ptrutil.ToPtr(true),
	}
	return result, true
}

type ImpressionMeta struct {
	aliasGVLIDs              map[string]uint16
	disabledSlots            map[string]struct{}
	bidAdjustmentFactors     map[string]float64
	serviceSideBidderPresent bool
	displayManager           string
	displayManagerVer        string
}

// processImpressions processes all impressions in the bid request
func (m OpenWrap) processImpressions(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bidRequest *openrtb2.BidRequest) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	rCtx.MultiFloors = make(map[string]*models.MultiFloors)
	impMeta := &ImpressionMeta{
		aliasGVLIDs:          make(map[string]uint16),
		disabledSlots:        make(map[string]struct{}),
		bidAdjustmentFactors: make(map[string]float64),
	}

	// Display manager
	impMeta.displayManager, impMeta.displayManagerVer = getDisplayManagerAndVer(bidRequest.App)

	for _, imp := range bidRequest.Imp {
		var ok bool
		result, ok = m.processImpression(rCtx, result, bidRequest, &imp, impMeta)
		if !ok {
			return result, false
		}

		if (rCtx.Platform == models.PLATFORM_APP || rCtx.Platform == models.PLATFORM_VIDEO || rCtx.Platform == models.PLATFORM_DISPLAY) && imp.Video != nil {
			if bidRequest.App != nil && bidRequest.App.Content != nil {
				m.metricEngine.RecordReqImpsWithContentCount(rCtx.PubIDStr, models.ContentTypeApp)
			}
			if bidRequest.Site != nil && bidRequest.Site.Content != nil {
				m.metricEngine.RecordReqImpsWithContentCount(rCtx.PubIDStr, models.ContentTypeSite)
			}
		}

	}

	if len(impMeta.disabledSlots) == len(bidRequest.Imp) {
		result.NbrCode = int(nbr.AllSlotsDisabled)
		err := errors.New("all slots disabled")
		result.Errors = append(result.Errors, err.Error())
		return result, false
	}

	if !impMeta.serviceSideBidderPresent {
		result.NbrCode = int(nbr.ServerSidePartnerNotConfigured)
		err := errors.New("server side partner not found")
		result.Errors = append(result.Errors, err.Error())
		return result, false
	}

	// update request extension
	rCtx.NewReqExt.Prebid.AliasGVLIDs = impMeta.aliasGVLIDs
	rCtx.NewReqExt.Prebid.BidAdjustmentFactors = impMeta.bidAdjustmentFactors
	if len(rCtx.Aliases) != 0 && rCtx.NewReqExt.Prebid.Aliases == nil {
		rCtx.NewReqExt.Prebid.Aliases = make(map[string]string)
	}
	for k, v := range rCtx.Aliases {
		rCtx.NewReqExt.Prebid.Aliases[k] = v
	}

	// update bidder params for pubmatic
	if _, ok := rCtx.AdapterThrottleMap[string(openrtb_ext.BidderPubmatic)]; !ok {
		rCtx.NewReqExt.Prebid.BidderParams, _ = updateRequestExtBidderParamsPubmatic(rCtx.NewReqExt.Prebid.BidderParams, rCtx.Cookies, rCtx.LoggerImpressionID, string(openrtb_ext.BidderPubmatic), rCtx.SendBurl)
	}
	for bidderCode, coreBidder := range rCtx.Aliases {
		if coreBidder == string(openrtb_ext.BidderPubmatic) {
			if _, ok := rCtx.AdapterThrottleMap[bidderCode]; !ok {
				rCtx.NewReqExt.Prebid.BidderParams, _ = updateRequestExtBidderParamsPubmatic(rCtx.NewReqExt.Prebid.BidderParams, rCtx.Cookies, rCtx.LoggerImpressionID, bidderCode, rCtx.SendBurl)
			}
		}
	}

	return result, true
}

// processImpression processes a single impression
func (m OpenWrap) processImpression(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bidRequest *openrtb2.BidRequest, imp *openrtb2.Imp, impMeta *ImpressionMeta) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	// Parse impression extension
	impExt := &models.ImpExtension{}
	if len(imp.Ext) != 0 {
		if err := json.Unmarshal(imp.Ext, impExt); err != nil {
			result.NbrCode = int(openrtb3.NoBidInvalidRequest)
			result.Errors = append(result.Errors, "failed to parse imp.ext: "+imp.ID)
			rCtx.ImpBidCtx = map[string]models.ImpCtx{} // do not create "s" object in owlogger
			return result, false
		}
	}

	// Handle tag ID
	if rCtx.Endpoint == models.EndpointWebS2S {
		imp.TagID = getTagID(*imp, impExt)
	}

	if imp.TagID == "" {
		result.NbrCode = int(nbr.InvalidImpressionTagID)
		result.Errors = append(result.Errors, "tagid missing for imp: "+imp.ID)
		rCtx.ImpBidCtx = map[string]models.ImpCtx{} // do not create "s" object in owlogger
		return result, false
	}

	// Get div from wrapper
	div := ""
	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	// Handle reward
	var reward *int8
	if imp.Rwdd == 1 {
		reward = openrtb2.Int8Ptr(1)
	}
	if reward == nil && impExt.Reward != nil {
		reward = impExt.Reward
		impExt.Prebid.IsRewardedInventory = reward
	}

	// Set pbadslot if absent
	if len(impExt.Data.PbAdslot) == 0 {
		impExt.Data.PbAdslot = imp.TagID
	}

	// TODO: Move this to entrypoint hook
	// Add size 300x600 for interstitial banner
	if (sdkutils.IsSdkIntegration(rCtx.Endpoint) || rCtx.Endpoint == models.EndpointV25) && imp.Instl == 1 {
		sdkutils.AddSize300x600ForInterstitialBanner(imp)
	}

	// Process ad unit configurations
	videoAdUnitCtx, bannerAdUnitCtx, nativeAdUnitCtx := m.processAdUnitConfig(rCtx, imp, div)

	slotType := "banner"
	// Handle video specific logic
	if imp.Video != nil {
		slotType = "video"
		m.processVideoImpression(rCtx, imp, rCtx.NewReqExt)
	}

	// Get slot information
	incomingSlots := models.GetIncomingSlots(*imp, videoAdUnitCtx)
	slotName := models.GetSlotName(imp.TagID, impExt)
	adUnitName := models.GetAdunitName(imp.TagID, impExt)

	// Check if slot is enabled
	if !isSlotEnabled(*imp, videoAdUnitCtx, bannerAdUnitCtx, nativeAdUnitCtx) {
		impMeta.disabledSlots[imp.ID] = struct{}{}
		rCtx.ImpBidCtx[imp.ID] = models.ImpCtx{ // for wrapper logger sz
			IncomingSlots:     incomingSlots,
			AdUnitName:        adUnitName,
			SlotName:          slotName,
			IsRewardInventory: reward,
		}
		return result, true
	}

	// Process multi-floors
	rCtx.MultiFloors[imp.ID] = m.getMultiFloors(*rCtx, reward, *imp)

	// Process bidders for this impression
	bidderMeta, nonMapped, result := m.processBidders(rCtx, result, imp, impExt, impMeta)

	// update the imp.ext with bidder params for this
	if impExt.Prebid.Bidder == nil {
		impExt.Prebid.Bidder = make(map[string]json.RawMessage)
	}
	for bidder, meta := range bidderMeta {
		impExt.Prebid.Bidder[bidder] = meta.Params
	}

	adserverURL := ""
	if impExt.Wrapper != nil {
		adserverURL = impExt.Wrapper.AdServerURL
	}

	if rCtx.Endpoint == models.EndpointAppLovinMax || rCtx.Endpoint == models.EndpointUnityLevelPlay {
		if len(impExt.GpId) == 0 {
			impExt.GpId = imp.TagID
		}
	}

	// Handle SDK integration specific logic
	if sdkutils.IsSdkIntegration(rCtx.Endpoint) {
		handleStoreURL(rCtx, bidRequest, impExt)
	}

	// Clean up and prepare impression extension
	impExt.Wrapper = nil
	impExt.Reward = nil
	impExt.Bidder = nil
	impExt.OWSDK = nil

	newImpExt, err := json.Marshal(impExt)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to update bidder params for impression %s", imp.ID))
		return result, false
	}

	// Update display manager
	displaymanager := impMeta.displayManager
	if imp.DisplayManager != "" {
		displaymanager = imp.DisplayManager
	}
	displaymanagerVer := impMeta.displayManagerVer
	if imp.DisplayManagerVer != "" {
		displaymanagerVer = imp.DisplayManagerVer
	}

	// Create or update impression context
	if _, ok := rCtx.ImpBidCtx[imp.ID]; !ok {
		rCtx.ImpBidCtx[imp.ID] = models.ImpCtx{
			ImpID:             imp.ID,
			TagID:             imp.TagID,
			Div:               div,
			IsRewardInventory: reward,
			BidFloor:          imp.BidFloor,
			BidFloorCur:       imp.BidFloorCur,
			Type:              slotType,
			IsBanner:          imp.Banner != nil,
			Banner:            ortb.DeepCopyImpBanner(imp.Banner),
			Video:             imp.Video,
			Native:            imp.Native,
			IncomingSlots:     incomingSlots,
			BidCtx:            make(map[string]models.BidCtx),
			NewExt:            newImpExt,
			SlotName:          slotName,
			AdUnitName:        adUnitName,
			AdserverURL:       adserverURL,
			DisplayManager:    displaymanager,
			DisplayManagerVer: displaymanagerVer,
			Bidders:           bidderMeta,
			NonMapped:         nonMapped,
			VideoAdUnitCtx:    videoAdUnitCtx,
			BannerAdUnitCtx:   bannerAdUnitCtx,
			NativeAdUnitCtx:   nativeAdUnitCtx,
		}
	}

	return result, true
}

// processAdUnitConfig processes ad unit configuration for an impression
func (m OpenWrap) processAdUnitConfig(rCtx *models.RequestCtx, imp *openrtb2.Imp, div string) (models.AdUnitCtx, models.AdUnitCtx, models.AdUnitCtx) {
	var videoAdUnitCtx, bannerAdUnitCtx, nativeAdUnitCtx models.AdUnitCtx

	// Update video object with ad unit config
	videoAdUnitCtx = adunitconfig.UpdateVideoObjectWithAdunitConfig(*rCtx, *imp, div, rCtx.DeviceCtx.ConnectionType)

	// Handle AMP video
	if rCtx.Endpoint == models.EndpointAMP && m.pubFeatures.IsAmpMultiformatEnabled(rCtx.PubID) && isVideoEnabledForAMP(videoAdUnitCtx.AppliedSlotAdUnitConfig) {
		rCtx.AmpVideoEnabled = true
		imp.Video = &openrtb2.Video{}
	}

	// Update banner object with ad unit config (except for AMP)
	if rCtx.Endpoint != models.EndpointAMP {
		bannerAdUnitCtx = adunitconfig.UpdateBannerObjectWithAdunitConfig(*rCtx, *imp, div)
	}

	// Update native object with ad unit config
	nativeAdUnitCtx = adunitconfig.UpdateNativeObjectWithAdunitConfig(*rCtx, *imp, div)

	return videoAdUnitCtx, bannerAdUnitCtx, nativeAdUnitCtx
}

// processVideoImpression handles video-specific impression processing
func (m OpenWrap) processVideoImpression(rCtx *models.RequestCtx, imp *openrtb2.Imp, requestExt *models.RequestExt) {
	// Record video interstitial stats
	if imp.Instl == 1 {
		m.metricEngine.RecordVideoInstlImpsStats(rCtx.PubIDStr, rCtx.ProfileIDStr)
	}

	// Set VAST event macros if not already set
	if len(requestExt.Prebid.Macros) == 0 {
		requestExt.Prebid.Macros = getVASTEventMacros(*rCtx)
	}
}

// handleStoreURL handles store url specific logic
func handleStoreURL(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest, impExt *models.ImpExtension) {
	appStoreUrl, isValidAppStoreUrl := getProfileAppStoreUrl(*rCtx)
	if !isValidAppStoreUrl && bidRequest.App != nil && bidRequest.App.StoreURL != "" {
		appStoreUrl = bidRequest.App.StoreURL
	}
	rCtx.AppStoreUrl = appStoreUrl
	rCtx.PageURL = appStoreUrl
	if appStoreUrl != "" {
		updateSkadnSourceapp(*rCtx, bidRequest, impExt)
	}
}

// processBidders processes bidders for an impression
func (m OpenWrap) processBidders(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], imp *openrtb2.Imp, impExt *models.ImpExtension, impMeta *ImpressionMeta) (map[string]models.PartnerData, map[string]struct{}, hookstage.HookResult[hookstage.BeforeValidationRequestPayload]) {
	bidderMeta := make(map[string]models.PartnerData)
	nonMapped := make(map[string]struct{})

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
		rCtx.PrebidBidderCode[bidderCode] = prebidBidderCode

		// Skip filtered or throttled bidders
		if _, ok := rCtx.AdapterFilteredMap[bidderCode]; ok {
			result.Warnings = append(result.Warnings, "Dropping adapter due to bidder filtering: "+bidderCode)
			continue
		}
		if _, ok := rCtx.AdapterThrottleMap[bidderCode]; ok {
			result.Warnings = append(result.Warnings, "Dropping throttled adapter from auction: "+bidderCode)
			continue
		}

		// Prepare bidder parameters
		slot, kgpv, isRegex, bidderParams, matchedSlotKeysVAST, err := m.prepareBidderParams(rCtx, imp, impExt, partnerID, prebidBidderCode)
		if err != nil || len(bidderParams) == 0 {
			nonMapped[bidderCode] = struct{}{}
			m.metricEngine.RecordPartnerConfigErrors(rCtx.PubIDStr, rCtx.ProfileIDStr, bidderCode, models.PartnerErrSlotNotMapped)
			if prebidBidderCode != string(openrtb_ext.BidderPubmatic) && prebidBidderCode != string(models.BidderPubMaticSecondaryAlias) {
				continue
			}
		}

		m.metricEngine.RecordPlatformPublisherPartnerReqStats(rCtx.Platform, rCtx.PubIDStr, bidderCode)

		if rCtx.NewReqExt.Prebid.SupportDeals && impExt.Bidder != nil {
			var bidderParamsMap map[string]interface{}
			err := json.Unmarshal(bidderParams, &bidderParamsMap)
			if err == nil {
				if bidderExt, ok := impExt.Bidder[bidderCode]; ok && bidderExt != nil && bidderExt.DealTier != nil {
					bidderParamsMap[models.DEAL_TIER_KEY] = bidderExt.DealTier
				}
				newBidderParams, err := json.Marshal(bidderParamsMap)
				if err == nil {
					bidderParams = newBidderParams
				}
			}
		}

		// Create partner data
		bidderMeta[bidderCode] = models.PartnerData{
			PartnerID:        partnerID,
			PrebidBidderCode: prebidBidderCode,
			MatchedSlot:      slot, // KGPSV
			Params:           bidderParams,
			KGP:              rCtx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN], // acutual slot
			KGPV:             kgpv,                                                     // regex pattern, use this field for pubmatic default unmapped slot as well using isRegex
			IsRegex:          isRegex,                                                  // regex pattern
		}

		// Handle VAST tag flags
		if len(matchedSlotKeysVAST) > 0 {
			meta := bidderMeta[bidderCode]
			meta.VASTTagFlags = make(map[string]bool)
			bidderMeta[bidderCode] = meta
		}

		// Handle aliases
		isAlias := handleBidderAlias(rCtx, bidderCode, partnerConfig)
		if isAlias || prebidBidderCode == models.BidderVASTBidder {
			updateAliasGVLIds(impMeta.aliasGVLIDs, bidderCode, partnerConfig)
		}

		// Set bid adjustment factor
		revShare := models.GetRevenueShare(rCtx.PartnerConfigMap[partnerID])
		impMeta.bidAdjustmentFactors[bidderCode] = models.GetBidAdjustmentValue(revShare)
		// Set service side bidder present flag
		impMeta.serviceSideBidderPresent = true
	}

	return bidderMeta, nonMapped, result
}

// prepareBidderParams prepares bidder parameters based on bidder code
func (m OpenWrap) prepareBidderParams(rCtx *models.RequestCtx, imp *openrtb2.Imp, impExt *models.ImpExtension, partnerID int, prebidBidderCode string) (string, string, bool, json.RawMessage, []string, error) {
	var (
		slot, kgpv          string
		isRegex             bool
		bidderParams        json.RawMessage
		err                 error
		matchedSlotKeysVAST []string
	)

	switch prebidBidderCode {
	case string(openrtb_ext.BidderPubmatic), models.BidderPubMaticSecondaryAlias:
		slot, kgpv, isRegex, bidderParams, err = bidderparams.PreparePubMaticParamsV25(*rCtx, m.cache, *imp, *impExt, partnerID)
	case models.BidderVASTBidder:
		slot, bidderParams, matchedSlotKeysVAST, err = bidderparams.PrepareVASTBidderParams(*rCtx, m.cache, *imp, *impExt, partnerID)
	default:
		slot, kgpv, isRegex, bidderParams, err = bidderparams.PrepareAdapterParamsV25(*rCtx, m.cache, *imp, *impExt, partnerID)
	}

	return slot, kgpv, isRegex, bidderParams, matchedSlotKeysVAST, err
}

// handleBidderAlias handles bidder aliases
func handleBidderAlias(rCtx *models.RequestCtx, bidderCode string, partnerConfig map[string]string) bool {
	isAlias := false

	// Check if bidder is an alias from partner config
	if alias, ok := partnerConfig[models.IsAlias]; ok && alias == "1" {
		if prebidPartnerName, ok := partnerConfig[models.PREBID_PARTNER_NAME]; ok {
			rCtx.Aliases[bidderCode] = adapters.ResolveOWBidder(prebidPartnerName)
			isAlias = true
		}
	}

	// Check if bidder is a known alias
	if alias, ok := IsAlias(bidderCode); ok {
		rCtx.Aliases[bidderCode] = alias
		isAlias = true
	}

	return isAlias
}

func (m OpenWrap) processAdpod(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bidRequest *openrtb2.BidRequest) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], bool) {
	if !rCtx.IsCTVRequest {
		return result, true
	}

	rCtx.AdpodCtx = make(map[string]models.AdpodConfig)
	for _, imp := range bidRequest.Imp {
		if imp.Video == nil {
			continue
		}

		if imp.Video.PodID != "" {
			// TODO: Retrieve pod config from DB
			// For now, we are using the pod config from the request
			rCtx.AdpodCtx.AddAdpodConfig(&imp)
		} else {
			// Get V25 Adpod configs
			adpodV25, err := adpod.GetV25AdpodConfigs(rCtx, &imp)
			if err != nil {
				result.NbrCode = int(nbr.InvalidAdpodConfig)
				result.Errors = append(result.Errors, "failed to get adpod configurations for "+imp.ID+" reason: "+err.Error())
				rCtx.ImpBidCtx = models.GetDefaultImpBidCtx(*bidRequest)
				return result, false
			}

			if adpodV25 == nil {
				continue
			}

			var domainExclusion, categoryExclusion bool
			if adpodV25.AdvertiserExclusionPercent != nil && *adpodV25.AdvertiserExclusionPercent == 0 {
				domainExclusion = true
			}
			if adpodV25.IABCategoryExclusionPercent != nil && *adpodV25.IABCategoryExclusionPercent == 0 {
				categoryExclusion = true
			}

			rCtx.AdpodCtx[imp.ID] = models.AdpodConfig{
				PodID:   imp.ID,
				PodType: models.PodTypeDynamic,
				Exclusion: models.ExclusionConfig{
					AdvertiserDomainExclusion: domainExclusion,
					IABCategoryExclusion:      categoryExclusion,
				},
				Slots: []models.SlotConfig{
					{
						Id:                          imp.ID,
						MinDuration:                 int64(adpodV25.MinDuration),
						MaxDuration:                 int64(adpodV25.MaxDuration),
						PodDur:                      imp.Video.MaxDuration,
						MaxSeq:                      int64(adpodV25.MaxAds),
						MinAds:                      int64(adpodV25.MinAds),
						MaxAds:                      int64(adpodV25.MaxAds),
						MinPodDuration:              imp.Video.MinDuration,
						MaxPodDuration:              imp.Video.MaxDuration,
						IABCategoryExclusionPercent: adpodV25.IABCategoryExclusionPercent,
						AdvertiserExclusionPercent:  adpodV25.AdvertiserExclusionPercent,
						Flexible:                    true,
					},
				},
			}
		}
	}
	return result, true
}

func getApplovinSchainABTestEnabled(percentage int) bool {
	if percentage > 0 && GetRandomNumberIn1To100() <= percentage {
		return true
	}
	return false
}

func (m *OpenWrap) addDefaultRequestContextValues(bidRequest *openrtb2.BidRequest, rCtx *models.RequestCtx) {
	rCtx.Source, rCtx.Origin = getSourceAndOrigin(bidRequest)
	rCtx.PageURL = getPageURL(bidRequest)
	rCtx.Platform = getPlatformFromRequest(bidRequest)
	rCtx.HostName = m.cfg.Server.HostName
	rCtx.ReturnAllBidStatus = rCtx.NewReqExt.Prebid.ReturnAllBidStatus

	// Device
	rCtx.DeviceCtx.UA = getUserAgent(bidRequest, rCtx.DeviceCtx.UA)
	rCtx.DeviceCtx.IP = getIP(bidRequest, rCtx.DeviceCtx.IP)
	rCtx.DeviceCtx.Country = getCountry(bidRequest)
	rCtx.DeviceCtx.DerivedCountryCode, _ = m.getCountryCodes(rCtx.DeviceCtx.IP)
	rCtx.DeviceCtx.Platform = getDevicePlatform(*rCtx, bidRequest)
	populateDeviceContext(&rCtx.DeviceCtx, bidRequest.Device)
}
