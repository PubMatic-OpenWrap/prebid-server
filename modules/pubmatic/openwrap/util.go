package openwrap

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/usersync"
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
)

const test = "_test"

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
}

// GetDevicePlatform determines the device from which request has been generated
func GetDevicePlatform(httpReqUAHeader string, bidRequest *openrtb2.BidRequest, platform string) models.DevicePlatform {
	userAgentString := httpReqUAHeader
	if userAgentString == "" && bidRequest.Device != nil && len(bidRequest.Device.UA) != 0 {
		userAgentString = bidRequest.Device.UA
	}

	switch platform {
	case models.PLATFORM_AMP:
		return models.DevicePlatformMobileWeb

	case models.PLATFORM_APP:
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

	case models.PLATFORM_DISPLAY:
		//Its web; now determine mobile or desktop
		var deviceType int
		if bidRequest.Device != nil && bidRequest.Device.DeviceType != 0 {
			deviceType = int(bidRequest.Device.DeviceType)
		}
		if isMobile(deviceType, userAgentString) {
			return models.DevicePlatformMobileWeb
		}
		return models.DevicePlatformDesktop

	case models.PLATFORM_VIDEO:
		var deviceType int
		if bidRequest.Device != nil && bidRequest.Device.DeviceType != 0 {
			deviceType = int(bidRequest.Device.DeviceType)
		}
		if deviceType == models.DeviceTypeConnectedTv || deviceType == models.DeviceTypeSetTopBox {
			return models.DevicePlatformConnectedTv
		}

		if bidRequest.Site != nil {
			//Its web; now determine mobile or desktop
			if isMobile(int(bidRequest.Device.DeviceType), userAgentString) {
				return models.DevicePlatformMobileWeb
			}
			return models.DevicePlatformDesktop
		}

		if bidRequest.App != nil {
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

func isMobile(deviceType int, userAgentString string) bool {
	if deviceType == models.DeviceTypeMobile || deviceType == models.DeviceTypeTablet ||
		deviceType == models.DeviceTypePhone {
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
			pageURL, err := url.Parse(source)
			if err == nil && pageURL != nil {
				origin = pageURL.Host
			}

		}
	} else if bidRequest.App != nil {
		source = bidRequest.App.Bundle
		origin = source
	}
	return source, origin
}

// getHostName Generates server name from node and pod name in K8S  environment
func getHostName() string {
	var (
		nodeName string
		podName  string
	)

	if nodeName, _ = os.LookupEnv(models.ENV_VAR_NODE_NAME); nodeName == "" {
		nodeName = models.DEFAULT_NODENAME
	} else {
		nodeName = strings.Split(nodeName, ".")[0]
	}

	if podName, _ = os.LookupEnv(models.ENV_VAR_POD_NAME); podName == "" {
		podName = models.DEFAULT_PODNAME
	} else {
		podName = strings.TrimPrefix(podName, "ssheaderbidding-")
	}

	serverName := nodeName + ":" + podName

	return serverName
}

// parseUIDCookies returns the parsed-cookie if uidCookie is nil else returns new cookie object
func parseUIDCookies(uidCookie *http.Cookie) *usersync.Cookie {

	if uidCookie != nil {
		return usersync.ParseCookie(uidCookie)
	}
	return usersync.NewCookie()
}

// RecordPublisherPartnerNoCookieStats parse request cookies and records the stats if cookie is not found for partner
func RecordPublisherPartnerNoCookieStats(rctx models.RequestCtx) {

	cookie := parseUIDCookies(rctx.UidCookie)
	for _, partnerConfig := range rctx.PartnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] == "0" {
			continue
		}

		partnerName := partnerConfig[models.PREBID_PARTNER_NAME]
		syncer := models.SyncerMap[adapters.ResolveOWBidder(partnerName)]
		if syncer != nil {
			uid, _, _ := cookie.GetUID(syncer.Key())
			if uid != "" {
				continue
			}
		}
		rctx.MetricsEngine.RecordPublisherPartnerNoCookieStats(rctx.PubIDStr, partnerConfig[models.BidderCode])
	}
}

// getPubmaticErrorCode is temporary function which returns the pubmatic specific error code for standardNBR code
func getPubmaticErrorCode(standardNBR int) int {
	switch standardNBR {
	case nbr.InvalidPublisherID:
		return 604 // ErrMissingPublisherID

	case nbr.InvalidRequest:
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

	}

	return -1
}