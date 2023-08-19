package pubmatic

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb2"
	podutil "github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"
	ow_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
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

// anyNonBidExistForBidder returns true if any non-bid exist for impression-bidder combination , false otherwise
func anyNonBidExistForBidder(wlog *WloggerRecord, imp *openrtb.Imp, bidExt *openrtb.ImpressionExt, bidder string) bool {

	if imp == nil || imp.Id == nil || bidExt == nil {
		return false
	}

	if bidderWrapper, ok := wlog.IsNonBidPresent[*imp.Id][bidder]; ok {
		// first check for vast bidder
		if bidExt.Prebid != nil && bidExt.Prebid.Video != nil && len(bidExt.Prebid.Video.VASTTagID) > 0 {
			return bidderWrapper.VASTagFlags[bidExt.Prebid.Video.VASTTagID]
		}
		// check for normal bidders in absence of vastTagId
		return bidderWrapper.Flag
	}

	return false
}

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
