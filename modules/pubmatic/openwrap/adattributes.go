package openwrap

import (
	"encoding/json"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// Wire IDs for ext.owsdk.adattributes (numeric; product spec).
const (
	AdAttrWireEngageToClose     = 1
	AdAttrWireTrueDoubleEndCard = 2
	AdAttrWireCTAOverlay        = 3
	AdAttrWireMRAIDAppStatus    = 4
	AdAttrWireCountdownOverlay  = 5
	AdAttrWireFrcClkBrowser     = 6
)

// AdFormat represents different types of ad formats
type AdFormat string

const (
	AdFormatBannerDisplay            AdFormat = "banner_display"
	AdFormatInterstitialDisplay      AdFormat = "interstitial_display"
	AdFormatInterstitialDisplayVideo AdFormat = "interstitial_display_video"
	AdFormatInterstitialVideo        AdFormat = "interstitial_video"
	AdFormatRewardedVideo            AdFormat = "rewarded_video"
	AdFormatMRECVideoDisplay         AdFormat = "mrec_video_display"
)

// OS represents the operating system
type OS string

const (
	OSAndroid OS = "android"
	OSiOS     OS = "ios"
)

const (
	MRECWidth  = 300
	MRECHeight = 250
)

// FeatureConfig defines supported ext.owsdk adattribute wire IDs for OS, SDK version, and ad format.
// MaxVersion empty means no upper bound (min inclusive only).
type FeatureConfig struct {
	OS         OS
	MinVersion string
	MaxVersion string
	AdFormat   AdFormat
	WireIDs    []int
}

// UnifiedFeatureMatrix maps OS, SDK version range, and ad format to supported adattribute wire IDs.
// Order matters: GetSupportedAdAttributeWireIDs returns the first matching row (more specific rows must appear before broader ones).
var UnifiedFeatureMatrix = []FeatureConfig{
	// Android — spec: "Supporting OpenWrap SDK versions" (interstitial display = SDK 4.1.0–4.2.0 only; banner/MREC rows per platform).
	{OS: OSAndroid, MinVersion: "4.1.0", MaxVersion: "4.2.0", AdFormat: AdFormatInterstitialDisplay, WireIDs: []int{AdAttrWireEngageToClose}},
	{OS: OSAndroid, MinVersion: "4.1.0", MaxVersion: "4.8.0", AdFormat: AdFormatBannerDisplay, WireIDs: []int{AdAttrWireEngageToClose}},

	{OS: OSAndroid, MinVersion: "4.3.0", MaxVersion: "4.3.0", AdFormat: AdFormatInterstitialDisplayVideo, WireIDs: []int{AdAttrWireEngageToClose}},
	{OS: OSAndroid, MinVersion: "4.3.0", MaxVersion: "4.3.0", AdFormat: AdFormatRewardedVideo, WireIDs: []int{AdAttrWireEngageToClose}},

	{OS: OSAndroid, MinVersion: "4.4.0", MaxVersion: "4.8.0", AdFormat: AdFormatInterstitialDisplayVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard}},
	{OS: OSAndroid, MinVersion: "4.4.0", MaxVersion: "4.8.0", AdFormat: AdFormatRewardedVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard}},

	{OS: OSAndroid, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatInterstitialDisplayVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard, AdAttrWireCTAOverlay}},
	{OS: OSAndroid, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatRewardedVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard, AdAttrWireCTAOverlay}},
	{OS: OSAndroid, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatMRECVideoDisplay, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay}},

	{OS: OSAndroid, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatInterstitialDisplayVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus}},
	{OS: OSAndroid, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatRewardedVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus}},
	{OS: OSAndroid, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatMRECVideoDisplay, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus}},
	{OS: OSAndroid, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatBannerDisplay, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireMRAIDAppStatus}},

	// iOS
	{OS: OSiOS, MinVersion: "4.1.0", MaxVersion: "4.2.0", AdFormat: AdFormatInterstitialDisplay, WireIDs: []int{AdAttrWireEngageToClose}},

	{OS: OSiOS, MinVersion: "4.3.0", MaxVersion: "4.8.0", AdFormat: AdFormatInterstitialDisplayVideo, WireIDs: []int{AdAttrWireEngageToClose}},
	{OS: OSiOS, MinVersion: "4.3.0", MaxVersion: "4.8.0", AdFormat: AdFormatRewardedVideo, WireIDs: []int{AdAttrWireEngageToClose}},

	{OS: OSiOS, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatInterstitialDisplayVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay}},
	{OS: OSiOS, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatRewardedVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay}},
	{OS: OSiOS, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatMRECVideoDisplay, WireIDs: []int{AdAttrWireCTAOverlay}},

	{OS: OSiOS, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatInterstitialDisplayVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus}},
	{OS: OSiOS, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatRewardedVideo, WireIDs: []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus}},
	{OS: OSiOS, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatMRECVideoDisplay, WireIDs: []int{AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus}},
	{OS: OSiOS, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatBannerDisplay, WireIDs: []int{AdAttrWireMRAIDAppStatus}},
}

// buildOWSDKAdAttributesMap returns server-side ext.owsdk fields (e.g. adattributes as numeric wire IDs) from device, SDK version, and format.
// Returns nil if nothing should be added.
func buildOWSDKAdAttributesMap(impCtx models.ImpCtx, deviceOS string) map[string]any {
	if deviceOS == "" {
		return nil
	}
	os := DetermineOS(deviceOS)
	if os == "" {
		return nil
	}
	sdkVersion := strings.TrimSpace(impCtx.DisplayManagerVer)
	if sdkVersion == "" {
		return nil
	}
	adFormat := DetermineAdFormat(impCtx)
	wireIDs := GetSupportedAdAttributeWireIDs(os, sdkVersion, adFormat)
	if len(wireIDs) == 0 {
		return nil
	}
	return CreateOWSDKExtension(wireIDs)
}

// mergeOWSDKAdAttributesIntoImpExt merges client ext.owsdk (e.g. ctaoverlay) with server-computed adattributes (numeric IDs) into
// the full imp.ext JSON. clientOWSDK is the incoming request's ext.owsdk before it was stripped for NewExt.
func (m *OpenWrap) mergeOWSDKAdAttributesIntoImpExt(extJSON json.RawMessage, impCtx models.ImpCtx, clientOWSDK map[string]any, deviceOS string) (json.RawMessage, error) {
	srv := buildOWSDKAdAttributesMap(impCtx, deviceOS)
	if len(clientOWSDK) == 0 && len(srv) == 0 {
		return extJSON, nil
	}
	var extMap map[string]json.RawMessage
	if len(extJSON) == 0 {
		extMap = make(map[string]json.RawMessage)
	} else if err := json.Unmarshal(extJSON, &extMap); err != nil {
		extMap = make(map[string]json.RawMessage)
	}
	owsdkOut := make(map[string]any)
	for k, v := range clientOWSDK {
		owsdkOut[k] = v
	}
	if srv != nil {
		for k, v := range srv {
			owsdkOut[k] = v
		}
	}
	if len(owsdkOut) == 0 {
		return extJSON, nil
	}
	owsdkBytes, err := json.Marshal(owsdkOut)
	if err != nil {
		return extJSON, err
	}
	extMap["owsdk"] = owsdkBytes
	return json.Marshal(extMap)
}

// GetSupportedAdAttributeWireIDs returns supported ext.owsdk adattribute wire IDs for OS, SDK version, and ad format.
func GetSupportedAdAttributeWireIDs(os OS, sdkVersion string, adFormat AdFormat) []int {
	sdkVersion = strings.TrimSpace(sdkVersion)
	if sdkVersion == "" || isVersionLessThan(sdkVersion, "4.1.0") {
		return nil
	}

	for _, config := range UnifiedFeatureMatrix {
		if config.OS == os &&
			config.AdFormat == adFormat &&
			isVersionInRange(sdkVersion, config.MinVersion, config.MaxVersion) {
			return slices.Clone(config.WireIDs)
		}
	}

	return nil
}

// isVideoEffectiveForAdFormat returns true when the impression still carries video for format / adattribute
// selection. Video may remain on imp until applyVideoAdUnitConfig runs, while ad unit config can already mark
// video disabled — align with the post-mutation request shape.
func isVideoEffectiveForAdFormat(impCtx models.ImpCtx) bool {
	if impCtx.Video == nil {
		return false
	}
	cfg := impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig
	if cfg == nil || cfg.Video == nil || cfg.Video.Enabled == nil {
		return true
	}
	return *cfg.Video.Enabled
}

// isBannerEffectiveForAdFormat is the banner counterpart of isVideoEffectiveForAdFormat.
func isBannerEffectiveForAdFormat(impCtx models.ImpCtx) bool {
	if impCtx.Banner == nil {
		return false
	}
	cfg := impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig
	if cfg == nil || cfg.Banner == nil || cfg.Banner.Enabled == nil {
		return true
	}
	return *cfg.Banner.Enabled
}

// DetermineAdFormat determines the ad format based on impression instl flag and ad unit configuration
func DetermineAdFormat(impCtx models.ImpCtx) AdFormat {
	videoOn := isVideoEffectiveForAdFormat(impCtx)
	bannerOn := isBannerEffectiveForAdFormat(impCtx)

	// Rewarded inventory: only classified as rewarded video when a video object is present (no separate rewarded-display format).
	if impCtx.IsRewardInventory != nil && *impCtx.IsRewardInventory == 1 {
		if videoOn {
			return AdFormatRewardedVideo
		}
	}

	// Check for interstitial (instl = 1)
	if impCtx.Instl == 1 {
		if videoOn && bannerOn {
			// Interstitial display + video
			return AdFormatInterstitialDisplayVideo
		}
		// check do we need to keep this logic after confirmation with preety
		if videoOn {
			return AdFormatInterstitialVideo
		}
		if bannerOn {
			return AdFormatInterstitialDisplay
		}
	}

	// Check for MREC (300x250) when instl = 0
	if bannerOn {
		if impCtx.Banner.W != nil && impCtx.Banner.H != nil {
			if *impCtx.Banner.W == MRECWidth && *impCtx.Banner.H == MRECHeight {
				// MREC logic - always return MREC video display when video is present
				if videoOn {
					return AdFormatMRECVideoDisplay
				}
			}
		}
	}
	// Non-interstitial banner (not 300x250 + video MREC): banner display for matrix / wire IDs.
	if bannerOn {
		return AdFormatBannerDisplay
	}
	return ""
}

// DetermineOS determines the OS based on device information
func DetermineOS(deviceOS string) OS {
	os := strings.ToLower(strings.TrimSpace(deviceOS))
	switch {
	case strings.Contains(os, "android"):
		return OSAndroid
	case strings.Contains(os, "ios"), strings.Contains(os, "iphone"), strings.Contains(os, "ipad"):
		return OSiOS
	default:
		return "" // Unknown OS
	}
}

//present in util.go: isIos and isAndroid
// or func isIos(os string, userAgentString string) bool {
// 	if openRTBDeviceOsIosRegex.Match([]byte(strings.ToLower(os))) || iosUARegex.Match([]byte(strings.ToLower(userAgentString))) {
// 		return true
// 	}
// 	return false
// }

// func isAndroid(os string, userAgentString string) bool {
// 	if openRTBDeviceOsAndroidRegex.Match([]byte(strings.ToLower(os))) || androidUARegex.Match([]byte(strings.ToLower(userAgentString))) {
// 		return true
// 	}
// 	return false
// }

// isVersionLessThan checks if version1 is less than version2
func isVersionLessThan(version1, version2 string) bool {
	return compareVersions(version1, version2) < 0
}

// isVersionInRange checks if version is within the specified range (inclusive)
func isVersionInRange(version, minVersion, maxVersion string) bool {
	// If maxVersion is empty, it means no upper bound
	if maxVersion == "" {
		return compareVersions(version, minVersion) >= 0
	}

	return compareVersions(version, minVersion) >= 0 && compareVersions(version, maxVersion) <= 0
}

// compareVersions compares two dot-separated numeric version strings (e.g. "5.1.0").
// Non-numeric segments are treated as 0; leading/trailing whitespace is ignored.
func compareVersions(v1, v2 string) int {
	v1, v2 = strings.TrimSpace(v1), strings.TrimSpace(v2)
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Num, v2Num int
		if i < len(v1Parts) {
			v1Num, _ = strconv.Atoi(v1Parts[i])
		}
		if i < len(v2Parts) {
			v2Num, _ = strconv.Atoi(v2Parts[i])
		}

		if v1Num < v2Num {
			return -1
		} else if v1Num > v2Num {
			return 1
		}
	}

	return 0
}

// CreateOWSDKExtension builds ext.owsdk with adattributes: sorted, deduplicated wire IDs (invalid / duplicate IDs dropped).
func CreateOWSDKExtension(wireIDs []int) map[string]any {
	owsdk := make(map[string]any)
	if len(wireIDs) == 0 {
		return owsdk
	}
	work := slices.Clone(wireIDs)
	write := 0
	for _, id := range work {
		if id > 0 {
			work[write] = id
			write++
		}
	}
	if write == 0 {
		return owsdk
	}
	work = work[:write]
	sort.Ints(work)
	j := 0
	for _, id := range work {
		if j == 0 || id != work[j-1] {
			work[j] = id
			j++
		}
	}
	work = work[:j]
	owsdk["adattributes"] = work
	return owsdk
}
