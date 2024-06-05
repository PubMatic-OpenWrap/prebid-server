package openwrap

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/metrics"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

var (
	widthRegEx  *regexp.Regexp
	heightRegEx *regexp.Regexp
	auIDRegEx   *regexp.Regexp
	divRegEx    *regexp.Regexp

	openRTBDeviceOsAndroidRegex *regexp.Regexp
	androidUARegex              *regexp.Regexp
	iosUARegex                  *regexp.Regexp
	openRTBDeviceOsIosRegex     *regexp.Regexp
	mobileDeviceUARegex         *regexp.Regexp
	ctvRegex                    *regexp.Regexp
	accountIdSearchPath         = [...]struct {
		isApp  bool
		isDOOH bool
		key    []string
	}{
		{true, false, []string{"app", "publisher", "ext", openrtb_ext.PrebidExtKey, "parentAccount"}},
		{true, false, []string{"app", "publisher", "id"}},
		{false, false, []string{"site", "publisher", "ext", openrtb_ext.PrebidExtKey, "parentAccount"}},
		{false, false, []string{"site", "publisher", "id"}},
		{false, true, []string{"dooh", "publisher", "ext", openrtb_ext.PrebidExtKey, "parentAccount"}},
		{false, true, []string{"dooh", "publisher", "id"}},
	}
)

const (
	test = "_test"
)

func init() {
	widthRegEx = regexp.MustCompile(models.MACRO_WIDTH)
	heightRegEx = regexp.MustCompile(models.MACRO_HEIGHT)
	auIDRegEx = regexp.MustCompile(models.MACRO_AD_UNIT_ID)
	//auIndexRegEx := regexp.MustCompile(models.MACRO_AD_UNIT_INDEX)
	//integerRegEx := regexp.MustCompile(models.MACRO_INTEGER)
	divRegEx = regexp.MustCompile(models.MACRO_DIV)

	openRTBDeviceOsAndroidRegex = regexp.MustCompile(models.OpenRTBDeviceOsAndroidRegexPattern)
	androidUARegex = regexp.MustCompile(models.AndroidUARegexPattern)
	iosUARegex = regexp.MustCompile(models.IosUARegexPattern)
	openRTBDeviceOsIosRegex = regexp.MustCompile(models.OpenRTBDeviceOsIosRegexPattern)
	mobileDeviceUARegex = regexp.MustCompile(models.MobileDeviceUARegexPattern)
	ctvRegex = regexp.MustCompile(models.ConnectedDeviceUARegexPattern)
}

//	rCtx.DevicePlatform = getDevicePlatform(rCtx.UA, payload.BidRequest, rCtx.Platform, rCtx.PubIDStr, m.metricEngine)
//
// getDevicePlatform determines the device from which request has been generated
func getDevicePlatform(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest) models.DevicePlatform {
	userAgentString := rCtx.UA

	switch rCtx.Platform {
	case models.PLATFORM_AMP:
		return models.DevicePlatformMobileWeb

	case models.PLATFORM_APP:
		//Its mobile; now determine ios or android
		var os = ""
		if bidRequest != nil && bidRequest.Device != nil && len(bidRequest.Device.OS) != 0 {
			os = bidRequest.Device.OS
		}
		if isIos(os, userAgentString) {
			return models.DevicePlatformMobileAppIos
		} else if isAndroid(os, userAgentString) {
			return models.DevicePlatformMobileAppAndroid
		}

	case models.PLATFORM_DISPLAY:
		//Its web; now determine mobile or desktop
		var deviceType adcom1.DeviceType
		if bidRequest != nil && bidRequest.Device != nil && bidRequest.Device.DeviceType != 0 {
			deviceType = bidRequest.Device.DeviceType
		}
		if isMobile(deviceType, userAgentString) {
			return models.DevicePlatformMobileWeb
		}
		return models.DevicePlatformDesktop

	case models.PLATFORM_VIDEO:
		var deviceType adcom1.DeviceType
		if bidRequest != nil && bidRequest.Device != nil && bidRequest.Device.DeviceType != 0 {
			deviceType = bidRequest.Device.DeviceType
		}
		isCtv := isCTV(userAgentString)

		if deviceType != 0 {
			if deviceType == adcom1.DeviceTV || deviceType == adcom1.DeviceConnected || deviceType == adcom1.DeviceSetTopBox {
				return models.DevicePlatformConnectedTv
			}
		}

		if deviceType == 0 && isCtv {
			return models.DevicePlatformConnectedTv
		}

		if bidRequest != nil && bidRequest.Site != nil {
			//Its web; now determine mobile or desktop
			if isMobile(bidRequest.Device.DeviceType, userAgentString) {
				return models.DevicePlatformMobileWeb
			}
			return models.DevicePlatformDesktop
		}

		if bidRequest != nil && bidRequest.App != nil {
			//Its mobile; now determine ios or android
			var os = ""
			if bidRequest.Device != nil && len(bidRequest.Device.OS) != 0 {
				os = bidRequest.Device.OS
			}

			if isIos(os, userAgentString) {
				return models.DevicePlatformMobileAppIos
			} else if isAndroid(os, userAgentString) {
				return models.DevicePlatformMobileAppAndroid
			}
		}

	default:
		return models.DevicePlatformNotDefined

	}

	return models.DevicePlatformNotDefined
}

func isMobile(deviceType adcom1.DeviceType, userAgentString string) bool {
	if deviceType != 0 {
		return deviceType == adcom1.DeviceMobile || deviceType == adcom1.DeviceTablet || deviceType == adcom1.DevicePhone
	}

	if mobileDeviceUARegex.Match([]byte(strings.ToLower(userAgentString))) {
		return true
	}
	return false
}

func isIos(os string, userAgentString string) bool {
	if openRTBDeviceOsIosRegex.Match([]byte(strings.ToLower(os))) || iosUARegex.Match([]byte(strings.ToLower(userAgentString))) {
		return true
	}
	return false
}

func isAndroid(os string, userAgentString string) bool {
	if openRTBDeviceOsAndroidRegex.Match([]byte(strings.ToLower(os))) || androidUARegex.Match([]byte(strings.ToLower(userAgentString))) {
		return true
	}
	return false
}

// GetIntArray converts interface to int array if it is compatible
func GetIntArray(val interface{}) []int {
	intArray := make([]int, 0)
	valArray, ok := val.([]interface{})
	if !ok {
		return nil
	}
	for _, x := range valArray {
		var intVal int
		intVal = GetInt(x)
		intArray = append(intArray, intVal)
	}
	return intArray
}

// GetInt converts interface to int if it is compatible
func GetInt(val interface{}) int {
	var result int
	if val != nil {
		switch val.(type) {
		case int:
			result = val.(int)
		case float64:
			val := val.(float64)
			result = int(val)
		case float32:
			val := val.(float32)
			result = int(val)
		}
	}
	return result
}

func getSourceAndOrigin(bidRequest *openrtb2.BidRequest) (string, string) {
	var source, origin string
	if bidRequest.Site != nil {
		if len(bidRequest.Site.Domain) != 0 {
			source = bidRequest.Site.Domain
			origin = source
		} else if len(bidRequest.Site.Page) != 0 {
			source = getDomainFromUrl(bidRequest.Site.Page)
			origin = source

		}
	} else if bidRequest.App != nil {
		source = bidRequest.App.Bundle
		origin = source
	}
	return source, origin
}

// getHostName Generates server name from node and pod name in K8S  environment
func GetHostName() string {

	var nodeName string
	if nodeName, _ = os.LookupEnv(models.ENV_VAR_NODE_NAME); nodeName == "" {
		nodeName = models.DEFAULT_NODENAME
	} else {
		nodeName = strings.Split(nodeName, ".")[0]
	}

	podName := getPodName()

	return nodeName + ":" + podName
}

func getPodName() string {

	var podName string
	if podName, _ = os.LookupEnv(models.ENV_VAR_POD_NAME); podName == "" {
		podName = models.DEFAULT_PODNAME
	} else {
		podName = strings.TrimPrefix(podName, "ssheaderbidding-")
	}
	return podName
}

// RecordPublisherPartnerNoCookieStats parse request cookies and records the stats if cookie is not found for partner
func RecordPublisherPartnerNoCookieStats(rctx models.RequestCtx) {

	for _, partnerConfig := range rctx.PartnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] == "0" {
			continue
		}

		partnerName := partnerConfig[models.PREBID_PARTNER_NAME]
		syncer := models.SyncerMap[adapters.ResolveOWBidder(partnerName)]
		if syncer != nil {
			uid, _, _ := rctx.ParsedUidCookie.GetUID(syncer.Key())
			if uid != "" {
				continue
			}
		}
		rctx.MetricsEngine.RecordPublisherPartnerNoCookieStats(rctx.PubIDStr, partnerConfig[models.BidderCode])
	}
}

// getPubmaticErrorCode is temporary function which returns the pubmatic specific error code for standardNBR code
func getPubmaticErrorCode(standardNBR openrtb3.NoBidReason) int {
	switch standardNBR {
	case nbr.InvalidPublisherID:
		return 604 // ErrMissingPublisherID

	case nbr.InvalidRequestExt:
		return 18 // ErrBadRequest

	case nbr.InvalidProfileID:
		return 700 // ErrMissingProfileID

	case nbr.AllPartnerThrottled:
		return 11 // ErrAllPartnerThrottled

	case nbr.InvalidPriceGranularityConfig:
		return 26 // ErrPrebidInvalidCustomPriceGranularity

	case nbr.InvalidImpressionTagID:
		return 605 // ErrMissingTagID

	case nbr.InvalidProfileConfiguration, nbr.InvalidPlatform, nbr.AllSlotsDisabled, nbr.ServerSidePartnerNotConfigured:
		return 6 // ErrInvalidConfiguration

	case nbr.InternalError:
		return 17 // ErrInvalidImpression

	case nbr.AllPartnersFiltered:
		return 26
	}

	return -1
}

func isCTV(userAgent string) bool {
	return ctvRegex.Match([]byte(userAgent))
}

// getUserAgent returns value of bidRequest.Device.UA if present else returns empty string
func getUserAgent(bidRequest *openrtb2.BidRequest, defaultUA string) string {
	userAgent := defaultUA
	if bidRequest != nil && bidRequest.Device != nil && len(bidRequest.Device.UA) > 0 {
		userAgent = bidRequest.Device.UA
	}
	return userAgent
}

func getIP(bidRequest *openrtb2.BidRequest, defaultIP string) string {
	ip := defaultIP
	if bidRequest != nil && bidRequest.Device != nil {
		if len(bidRequest.Device.IP) > 0 {
			ip = bidRequest.Device.IP
		} else if len(bidRequest.Device.IPv6) > 0 {
			ip = bidRequest.Device.IPv6
		}
	}
	return ip
}

func getCountry(bidRequest *openrtb2.BidRequest) string {
	if bidRequest.Device != nil && bidRequest.Device.Geo != nil && bidRequest.Device.Geo.Country != "" {
		return bidRequest.Device.Geo.Country
	}
	if bidRequest.User != nil && bidRequest.User.Geo != nil && bidRequest.User.Geo.Country != "" {
		return bidRequest.User.Geo.Country
	}
	return ""
}

func getPlatformFromRequest(request *openrtb2.BidRequest) string {
	var platform string
	if request.Site != nil {
		return models.PLATFORM_DISPLAY
	}
	if request.App != nil {
		return models.PLATFORM_APP
	}
	return platform
}

// for AMP requests based on traffic percentage, we will decide to send video or not
// if traffic percentage is not defined then send video
// if traffic percentage is defined then send video based on percentage
func isVideoEnabledForAMP(adUnitConfig *adunitconfig.AdConfig) bool {
	if adUnitConfig == nil || adUnitConfig.Video == nil || adUnitConfig.Video.Enabled == nil || !*adUnitConfig.Video.Enabled {
		return false
	} else if adUnitConfig.Video.AmpTrafficPercentage == nil || rand.Intn(100) < *adUnitConfig.Video.AmpTrafficPercentage {
		return true
	}
	return false
}

func GetRequestIP(body []byte, request *http.Request) string {
	ipBytes, _, _, _ := jsonparser.Get(body, "device", "ip")
	if len(ipBytes) > 0 {
		return string(ipBytes)
	}
	return models.GetIP(request)
}

func GetRequestUserAgent(body []byte, request *http.Request) string {
	uaBytes, _, _, _ := jsonparser.Get(body, "device", "ua")
	if len(uaBytes) > 0 {
		return string(uaBytes)
	}
	return request.Header.Get("User-Agent")
}

func getProfileType(partnerConfigMap map[int]map[string]string) int {
	if profileTypeStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.ProfileTypeKey]; ok {
		ProfileType, _ := strconv.Atoi(profileTypeStr)
		return ProfileType
	}
	return 0
}

func getProfileTypePlatform(partnerConfigMap map[int]map[string]string) int {
	if profileTypePlatformStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.PLATFORM_KEY]; ok {
		if ProfileTypePlatform, ok := models.ProfileTypePlatform[profileTypePlatformStr]; ok {
			return ProfileTypePlatform
		}
	}
	return 0
}

func getAppPlatform(partnerConfigMap map[int]map[string]string) int {
	if appPlatformStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.AppPlatformKey]; ok {
		AppPlatform, _ := strconv.Atoi(appPlatformStr)
		return AppPlatform
	}
	return 0
}

func getAppIntegrationPath(partnerConfigMap map[int]map[string]string) int {
	if appIntegrationPathStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.IntegrationPathKey]; ok {
		if appIntegrationPath, ok := models.AppIntegrationPath[appIntegrationPathStr]; ok {
			return appIntegrationPath
		}
	}
	return -1
}

func getAppSubIntegrationPath(partnerConfigMap map[int]map[string]string) int {
	if appSubIntegrationPathStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.SubIntegrationPathKey]; ok {
		if appSubIntegrationPath, ok := models.AppSubIntegrationPath[appSubIntegrationPathStr]; ok {
			return appSubIntegrationPath
		}
	}
	if adserverStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.AdserverKey]; ok {
		if adserver, ok := models.AppSubIntegrationPath[adserverStr]; ok {
			return adserver
		}
	}
	return -1
}

func getAccountIdFromRawRequest(hasStoredRequest bool, storedRequest, originalRequest json.RawMessage) (string, bool, bool, []error) {
	request := originalRequest
	if hasStoredRequest {
		request = storedRequest
	}

	accountId, isAppReq, isDOOHReq, err := searchAccountId(request)
	if err != nil {
		return "", isAppReq, isDOOHReq, []error{err}
	}

	// In case the stored request did not have account data we specifically search it in the original request
	if accountId == "" && hasStoredRequest {
		accountId, _, _, err = searchAccountId(originalRequest)
		if err != nil {
			return "", isAppReq, isDOOHReq, []error{err}
		}
	}

	if accountId == "" {
		return metrics.PublisherUnknown, isAppReq, isDOOHReq, nil
	}

	return accountId, isAppReq, isDOOHReq, nil
}

func searchAccountId(request []byte) (string, bool, bool, error) {
	for _, path := range accountIdSearchPath {
		accountId, exists, err := getStringValueFromRequest(request, path.key)
		if err != nil {
			return "", path.isApp, path.isDOOH, err
		}
		if exists {
			return accountId, path.isApp, path.isDOOH, nil
		}
	}
	return "", false, false, nil
}

func getStringValueFromRequest(request []byte, key []string) (string, bool, error) {
	val, dataType, _, err := jsonparser.Get(request, key...)
	if dataType == jsonparser.NotExist {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if dataType != jsonparser.String {
		return "", true, fmt.Errorf("%s must be a string", strings.Join(key, "."))
	}
	return string(val), true, nil
}
