package pubmatic

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/openrtb/v19/openrtb2"
	podutil "github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"
	ow_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/gocache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// GetTestMode checks if request.Test is set as 2
func GetPubmaticTestMode(request *openrtb2.BidRequest) bool {
	return request.Test == 2
}

// IsPubmaticCorePartner returns true when the partner is pubmatic or internally an alias of pubmatic
func IsPubmaticCorePartner(partnerName string) bool {
	if partnerName == models.BidderPubMatic || partnerName == models.BidderPubMaticSecondaryAlias || partnerName == models.BidderGroupM {
		return true
	}
	return false
}

// // anyNonBidExistForBidder returns true if any non-bid exist for impression-bidder combination , false otherwise
// func anyNonBidExistForBidder(wlog *WloggerRecord, imp *openrtb2.Imp, bidExt *openrtb.ImpressionExt, bidder string) bool {

// 	if imp == nil || imp.Id == nil || bidExt == nil {
// 		return false
// 	}

// 	if bidderWrapper, ok := wlog.IsNonBidPresent[*imp.Id][bidder]; ok {
// 		// first check for vast bidder
// 		if bidExt.Prebid != nil && bidExt.Prebid.Video != nil && len(bidExt.Prebid.Video.VASTTagID) > 0 {
// 			return bidderWrapper.VASTagFlags[bidExt.Prebid.Video.VASTTagID]
// 		}
// 		// check for normal bidders in absence of vastTagId
// 		return bidderWrapper.Flag
// 	}

// 	return false
// }

// Round value to 2 digit
func roundToTwoDigit(value float64) float64 {
	output := math.Pow(10, float64(2))
	return float64(math.Round(value*output)) / output
}

func (partnerRecord *PartnerRecord) logSlotAdPodParameters(eachBid *openrtb2.Bid) {
	_, sequenceNumber := podutil.DecodeImpressionID(eachBid.ImpID)
	if sequenceNumber > 0 {

		//iab categories
		if len(eachBid.Cat) > 0 {
			partnerRecord.Cat = make([]string, len(eachBid.Cat)) //[]string `json:"apcat,omitempty"`
			copy(partnerRecord.Cat, eachBid.Cat)
		}

		//setting adpod sequence number
		partnerRecord.AdPodSequenceNumber = &sequenceNumber

		if nil != eachBid.Ext {
			/// AAA
			// ext, ok := eachBid.Ext.(*openrtb.ImpressionExt)
			// if ok {
			// 	//Set ReasonCode
			// 	if nil != ext.AdPod && nil != ext.AdPod.ReasonCode {
			// 		partnerRecord.NoBidReason = ext.AdPod.ReasonCode
			// 	}
			// }
		}

		//partnerRecord.NoBidReason //*int     `json:"aprc,omitempty"`
	}
}

func (wlog *WloggerRecord) logAdPodPercentage(adpod *ExtRequestAdPod) {
	if nil == adpod {
		return
	}

	percentage := &AdPodPercentage{}
	found := false

	if nil != adpod.CrossPodAdvertiserExclusionPercent {
		percentage.CrossPodAdvertiserExclusionPercent = adpod.CrossPodAdvertiserExclusionPercent
		found = true
	}

	if nil != adpod.CrossPodIABCategoryExclusionPercent {
		percentage.CrossPodIABCategoryExclusionPercent = adpod.CrossPodIABCategoryExclusionPercent
		found = true
	}

	if nil != adpod.IABCategoryExclusionWindow {
		percentage.IABCategoryExclusionWindow = adpod.IABCategoryExclusionWindow
		found = true
	}

	if nil != adpod.AdvertiserExclusionWindow {
		percentage.AdvertiserExclusionWindow = adpod.AdvertiserExclusionWindow
		found = true
	}

	if found {
		wlog.AdPodPercentage = percentage
	}
}

func getAdPodSlot(imp *openrtb2.Imp, ext *BidResponseAdPodExt) *AdPodSlot {
	if nil == imp || nil == imp.Video || nil == imp.Video.Ext || nil == ext || nil == ext.Config {
		return nil
	}

	config, ok := ext.Config[imp.ID]
	if !ok || nil == config.VideoExt || nil == config.VideoExt.AdPod {
		return nil
	}

	adpod := config.VideoExt.AdPod

	var adpodSlot AdPodSlot

	if nil != adpod.MinAds {
		adpodSlot.MinAds = adpod.MinAds
	}

	if nil != adpod.MaxAds {
		adpodSlot.MaxAds = adpod.MaxAds
	}

	if nil != adpod.MinDuration {
		adpodSlot.MinDuration = adpod.MinDuration
	}

	if nil != adpod.MaxDuration {
		adpodSlot.MaxDuration = adpod.MaxDuration
	}

	if nil != adpod.AdvertiserExclusionPercent {
		adpodSlot.AdvertiserExclusionPercent = adpod.AdvertiserExclusionPercent
	}

	if nil != adpod.IABCategoryExclusionPercent {
		adpodSlot.IABCategoryExclusionPercent = adpod.IABCategoryExclusionPercent
	}

	return &adpodSlot
}

// Get Universal Pixel Object from AdUniConfig
func getUniversalPixelFromAdUnit(cache ow_cache.Cache, request *openrtb2.BidRequest) []adunitconfig.UniversalPixel {
	if GetPubmaticTestMode(request) {
		return nil
	}

	///AAA: ???
	rctx := models.RequestCtx{}
	// rctx := ow_models.GetRequestCtxFromRequestObject(request)
	adUnitConfig := cache.GetAdunitConfigFromCache(&openrtb2.BidRequest{}, rctx.PubID, rctx.ProfileID, rctx.DisplayID)

	if adUnitConfig != nil && adUnitConfig.Config != nil {
		if defaultAdUnitConfig, ok := adUnitConfig.Config[models.AdunitConfigDefaultKey]; ok && defaultAdUnitConfig != nil && len(defaultAdUnitConfig.UniversalPixel) != 0 {
			return defaultAdUnitConfig.UniversalPixel
		}
	}

	return nil
}

// CreateLoggerRecordFromRequest creates logger and tracker records from request data
func (wlog *WloggerRecord) CreateLoggerRecordFromRequest(rctx *models.RequestCtx, uaFromHTTPReq string, ortbBidRequest *openrtb2.BidRequest, platform string, displayVersionID int) {

	pubid, _ := strconv.Atoi(GetString(GetValueFromRequest(ortbBidRequest, models.PublisherID)))
	wlog.SetPubID(pubid)
	wlog.SetProfileID(strconv.Itoa(rctx.ProfileID))
	wlog.SetVersionID(strconv.Itoa(displayVersionID))
	wlog.SetConsentString(GetString(GetValueFromRequest(ortbBidRequest, models.Consent)))
	wlog.SetGDPR(int8(GetInt(GetValueFromRequest(ortbBidRequest, models.GDPR))))
	wlog.SetUserAgent(GetString(GetValueFromRequest(ortbBidRequest, models.UserAgent)))
	wlog.SetIP(GetString(GetValueFromRequest(ortbBidRequest, models.IP)))
	wlog.SetPageURL(GetString(GetValueFromRequest(ortbBidRequest, models.StoreURL)))
	wlog.SetOrigin(GetString(GetValueFromRequest(ortbBidRequest, models.Origin)))

	//log device object
	wlog.logDeviceObject(rctx, uaFromHTTPReq, ortbBidRequest, platform)

	//log content object
	if nil != ortbBidRequest.Site {
		wlog.logContentObject(ortbBidRequest.Site.Content)
	} else if nil != ortbBidRequest.App {
		wlog.logContentObject(ortbBidRequest.App.Content)
	}

	//log adpod percentage object
	if nil != ortbBidRequest.Ext {
		var ext ExtRequest
		err := json.Unmarshal(ortbBidRequest.Ext, &ext)
		if err != nil {
			wlog.logAdPodPercentage(ext.AdPod)
			wlog.logFloorType(ext.Prebid)
		}
	}

	// logger.DebugWithBid(*ortbBidRequest.Id, "Initial Logger record formed from ortbBidRequest %+v", wlog)
}

// logFloorType will be used to log floor type
func (wlog *WloggerRecord) logFloorType(prebid *ExtRequestPrebid) {
	wlog.record.FloorType = models.SoftFloor
	if prebid != nil && prebid.Floors != nil &&
		prebid.Floors.Enabled != nil && *prebid.Floors.Enabled &&
		prebid.Floors.Enforcement != nil && prebid.Floors.Enforcement.EnforcePBS != nil && *prebid.Floors.Enforcement.EnforcePBS {
		wlog.record.FloorType = models.HardFloor
	}
}

func setCommonLogger(rctx *models.RequestCtx, loggerRecord *WloggerRecord, openRTB *openrtb2.BidRequest, platform string, displayVersionID, timeout int, testConfigApplied bool) {
	uaFromHTTPReq := rctx.UA
	loggerRecord.CreateLoggerRecordFromRequest(rctx, uaFromHTTPReq, openRTB, platform, displayVersionID)
	loggerRecord.SetTimeout(int(timeout))
	loggerRecord.SetIntegrationType(rctx.ReqAPI)
	if rctx.UidCookie != nil {
		loggerRecord.SetUID(rctx.UidCookie.Value)
	}
	if testConfigApplied {
		loggerRecord.SetTestConfigApplied(1)
	}
}
func prepareLogger(
	openRTB *openrtb2.BidRequest,
	// responseMap *openrtb.BidResponseMap,
	impWinningbidresponseMap map[string]openrtb2.BidResponse,
	partnerConfigMap map[int]map[string]string,
	platform string,
	timeout int,
	allPartnersThrottled, testConfigApplied bool) {

	// displayVersionID, _ := strconv.Atoi(models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.DisplayVersionID))
	// setCommonLogger(openRTB, platform, displayVersionID, timeout, testConfigApplied)

	// partnerCookieFlagMap := util.ParseRequestCookies(controller.HTTPRequest, partnerConfigMap)
	// partnerMap := controller.LoggerRecord.GetPartnerRecordMap(openRTB, partnerConfigMap, partnerCookieFlagMap)

	// //Log Partner Records
	// controller.LoggerRecord.AddCTVResponseAndDBConfigValues(
	// 	openRTB,
	// 	responseMap,
	// 	impWinningbidresponseMap,
	// 	partnerMap,
	// 	platform,
	// 	allPartnersThrottled)

	// //Inject Tracking URL's
	// controller.LoggerRecord.InjectTrackers(
	// 	openRTB,
	// 	responseMap,
	// 	impWinningbidresponseMap,
	// 	partnerMap,
	// 	controller.TrackerURL,
	// 	controller.VideoErrorURL,
	// 	controller.Cache,
	// 	platform,
	// 	allPartnersThrottled)
}

// GetPartnerRecordMap will return one time partner record map for request
func (wlog *WloggerRecord) GetPartnerRecordMap(
	rctx models.RequestCtx,
	request *openrtb2.BidRequest,
	partnerConfigMap map[int]map[string]string,
	partnerCookieFlagMap map[string]int) []map[string]PartnerRecord {

	pubmaticTestMode := GetPubmaticTestMode(request)
	partnerMap := make([]map[string]PartnerRecord, len(request.Imp))

	for index, eachImp := range request.Imp {
		/* form default bids */
		partnerMap[index] = make(map[string]PartnerRecord)

		for partnerID, partnerConfig := range partnerConfigMap {

			//ignore version level properties partner ID
			if partnerID == models.VersionLevelConfigID {
				continue
			}

			// AAA : use cache from OW module
			// slotMap := dbcache.GetCache().GetMappingsFromCacheV25(request, partnerConfig)
			slotMap := gocache.Cache.GetMappingsFromCacheV25(rctx, partnerID)
			//if no mappings found for a partner, dont form default bid
			if !pubmaticTestMode && nil == slotMap {
				// logger.DebugWithBid(request.ID, "slotMap is nil for partner: %s", partnerConfig[constant.BidderCode])
				continue
			}

			// AAA : use cache from OW module
			// slotMappingInfo := gocache.Cache.GetSlotToHashValueMapFromCacheV25(request, partnerConfig)
			slotMappingInfo := gocache.Cache.GetSlotToHashValueMapFromCacheV25(rctx, partnerID)
			partnerAdded := wlog.formDefaultPartnerRecord(rctx, request, &eachImp, partnerCookieFlagMap, partnerConfig, partnerMap[index], slotMap, slotMappingInfo, pubmaticTestMode, false)

			//if unmapped case for pubmatic/pubmatic-sec/groupm, form default mapping partner record, set defaultMapping flag = true
			if !partnerAdded && IsPubmaticCorePartner(partnerConfig[models.PREBID_PARTNER_NAME]) && partnerConfig[models.KEY_GEN_PATTERN] != models.REGEX_KGP {
				wlog.formDefaultPartnerRecord(rctx, request, &eachImp, partnerCookieFlagMap, partnerConfig, partnerMap[index], slotMap, slotMappingInfo, pubmaticTestMode, true)
			}
		}
	}
	return partnerMap
}

// formDefaultPartnerRecord returns default Partner record for first mapped case for each banner or video
func (wlog *WloggerRecord) formDefaultPartnerRecord(
	rctx models.RequestCtx,
	request *openrtb2.BidRequest,
	imp *openrtb2.Imp,
	partnerCookieFlagMap map[string]int,
	partnerConfig map[string]string,
	partnerMap map[string]PartnerRecord,
	slotMap map[string]models.SlotMapping,
	slotMappingInfo models.SlotMappingInfo,
	pubmaticTestMode bool,
	defaultMapping bool) bool {

	var partnerSlotAdded bool

	// check for valid mapping for banner.W and banner.H
	if imp.Banner != nil && imp.Banner.W != nil && imp.Banner.H != nil {
		partnerSlotAdded = addDefaultSlotForPartner(rctx, request, wlog, imp, partnerConfig, slotMap, partnerCookieFlagMap, partnerMap, pubmaticTestMode, slotMappingInfo, models.Banner, defaultMapping)
	}

	// check for valid mapping for banner.Format
	if !partnerSlotAdded && imp.Banner != nil && len(imp.Banner.Format) > 0 {
		for _, size := range imp.Banner.Format {
			// newImp := util.CreateImp(imp, size.W, size.H)
			newImp := imp
			newImp.Banner.W = getInt64Ptr(size.W)
			newImp.Banner.W = getInt64Ptr(size.H)
			partnerSlotAdded = true
			partnerSlotAdded = addDefaultSlotForPartner(rctx, request, wlog, newImp, partnerConfig, slotMap, partnerCookieFlagMap, partnerMap, pubmaticTestMode, slotMappingInfo, models.Banner, defaultMapping)
			if partnerSlotAdded {
				break
			}
		}
	}

	//check for video mapping
	if !partnerSlotAdded && imp.Video != nil {
		// newImp := util.CreateImp(imp, 0, 0)
		newImp := imp
		newImp.Banner.W = getInt64Ptr(0)
		newImp.Banner.H = getInt64Ptr(0)
		partnerSlotAdded = addDefaultSlotForPartner(rctx, request, wlog, newImp, partnerConfig, slotMap, partnerCookieFlagMap, partnerMap, pubmaticTestMode, slotMappingInfo, models.Video, defaultMapping)
	}

	//check for native mapping
	if !partnerSlotAdded && imp.Native != nil {
		newImp := imp
		newImp.Banner.W = getInt64Ptr(0)
		newImp.Banner.H = getInt64Ptr(0)
		partnerSlotAdded = addDefaultSlotForPartner(rctx, request, wlog, newImp, partnerConfig, slotMap, partnerCookieFlagMap, partnerMap, pubmaticTestMode, slotMappingInfo, models.Native, defaultMapping)
	}

	return partnerSlotAdded
}
func addDefaultSlotForPartner(
	rctx models.RequestCtx,
	request *openrtb2.BidRequest,
	wlog *WloggerRecord,
	eachImp *openrtb2.Imp,
	partnerConfig map[string]string,
	slotMap map[string]models.SlotMapping,
	partnerCookieFlagMap map[string]int,
	partnerMap map[string]PartnerRecord,
	pubmaticTestMode bool,
	slotMappingInfo models.SlotMappingInfo,
	adFormat string,
	defaultMappings bool) bool {

	var partnerSlotAdded bool
	bidderCode := partnerConfig[models.BidderCode]

	impExt := new(ImpExtension)
	err := json.Unmarshal(eachImp.Ext, impExt)
	if err != nil {
		return false
	}

	// rctx.Source = getSourceAndOrigin()
	slotKey := GenerateSlotName(*eachImp.Banner.H, *eachImp.Banner.W, partnerConfig[models.KEY_GEN_PATTERN], eachImp.TagID, *impExt.Wrapper.Div, rctx.Source)

	// slotKey := util.FormSlotKeyV25(request, eachImp, partnerConfig[models.KEY_GEN_PATTERN])
	slotKeyPresent := true

	//Check1: Empty Slot Key
	slotKeyPresent = (len(slotKey) > 0)

	dontSkipPartnerBid := false
	if defaultMappings || pubmaticTestMode {
		dontSkipPartnerBid = true
	}

	matchedPattern := slotKey
	//Check2: KGP Based
	if slotKeyPresent {
		if partnerConfig[models.KEY_GEN_PATTERN] == models.REGEX_KGP {
			//REGEX_KGP
			// var regexMap map[string]interface{}
			regexMap := ""

			// profileID, _ := strconv.Atoi(wlog.ProfileID)
			// versionID, _ := strconv.Atoi(wlog.VersionID)
			partnerID, _ := strconv.Atoi(partnerConfig[models.PARTNER_ID])
			rctx := models.RequestCtx{}
			regexMap, matchedPattern = bidderparams.GetRegexMatchingSlot(rctx, gocache.Cache, slotKey, slotMap, slotMappingInfo, partnerID)

			// regexMap, matchedPattern = util.RunRegexMatch(*request.Id, slotMap, slotMappingInfo, slotKey, wlog.PubID, partnerID, profileID, versionID, bidderCode)
			slotKeyPresent = regexMap != ""
		} else if partnerConfig[models.KEY_GEN_PATTERN] == models.ADUNIT_SOURCE_VASTTAG_KGP {
			//ADUNIT_SOURCE_VASTTAG_KGP, this will be handled outside
			//check for slot keys
			if adFormat == models.Video {
				slotKeyPresent = true
			}
		} else {
			//Others KGP: check slotmapping entry
			if _, ok := slotMap[strings.ToLower(slotKey)]; !ok {
				slotKeyPresent = false
			}
		}
	}

	if slotKeyPresent || dontSkipPartnerBid {
		revShare, _ := strconv.ParseFloat(partnerConfig[models.REVSHARE], 64)
		matchedImpression := 0
		if partnerCookieFlagMap[bidderCode] == 1 {
			matchedImpression = 1
		}
		// logger.DebugWithBid(*request.Id, "Forming default bid for tagid:%s and kgpv: %s", *eachImp.TagId, slotKey)
		partner := wlog.FormDefaultBidForPartner(eachImp.ID, partnerConfig[models.PREBID_PARTNER_NAME], bidderCode, matchedPattern, partnerConfig[models.KEY_GEN_PATTERN], revShare, matchedImpression, slotKey, adFormat)
		partnerMap[bidderCode] = partner
		partnerSlotAdded = true
	} else {
		// logger.DebugWithBid(*request.Id, "No entry in slotMap for biddercode:%s slotKey:%s", partnerConfig[bidderCode], slotKey)
	}

	return partnerSlotAdded
}

// FormDefaultBidForPartner create default bid(partner record) for a given parnter with given tagid and kgpv
func (wlog *WloggerRecord) FormDefaultBidForPartner(bidID string, partnerName, bidderCode string, kgpv string,
	kgp string, revShare float64, matchedImpression int, kgpsv, adFormat string) PartnerRecord {
	// logger.Debug("In FormDefaultBidForPartner method")
	partner := PartnerRecord{
		PartnerID:            partnerName,
		BidderCode:           bidderCode,
		PartnerSize:          "0x0",
		KGPV:                 kgpv,
		KGPSV:                kgpsv,
		NetECPM:              0,
		GrossECPM:            0,
		Latency1:             0,
		Latency2:             0,
		PostTimeoutBidStatus: 0,
		WinningBidStaus:      0,
		BidID:                bidID,
		OrigBidID:            bidID,
		DealID:               "-1",
		DealChannel:          "",
		DefaultBidStatus:     1,
		ServerSide:           1,
		RevShare:             revShare,
		KGP:                  kgp,
		MatchedImpression:    matchedImpression,
		Adformat:             adFormat,
	}
	return partner
}

// func prepareLoggerCTV(rctx *models.RequestCtx, bidrequest *openrtb2.BidRequest, platform string, partnerCfgMap map[int]map[string]string, timeout int, allPartnersThrottled, testConfigApplied bool) {

// 	loggerRecord := &WloggerRecord{}
// 	displayVersionID, _ := strconv.Atoi(models.GetVersionLevelPropertyFromPartnerConfig(partnerCfgMap, models.DisplayVersionID))
// 	setCommonLogger(rctx, loggerRecord, bidrequest, platform, displayVersionID, timeout, testConfigApplied)
// 	// setCommonLogger(controller.OpenRTB, controller.Params.Platform, displayVersionID, timeout, controller.Params.TestConfigApplied)

// 	partnerCookieFlagMap := util.ParseRequestCookies(controller.HTTPRequest, partnerCfgMap)

// 	partnerMap := controller.LoggerRecord.GetPartnerRecordMap(controller.OpenRTB, partnerCfgMap, partnerCookieFlagMap)

// 	//Log Partner Records
// 	controller.LoggerRecord.AddCTVResponseAndDBConfigValues(
// 		controller.OpenRTB,
// 		controller.impAllBidResponseMap,
// 		controller.impHighestBidResponseMap,
// 		partnerMap,
// 		controller.Params.Platform,
// 		allPartnersThrottled)

// }

func getInt64Ptr(v int64) *int64 {
	return &v
}
