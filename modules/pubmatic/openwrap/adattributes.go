package openwrap

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// AdAttribute represents the supported ad attributes
type AdAttribute string

const (
	EngageToCloseInstl    AdAttribute = "eng_to_close_instl"
	EngageToCloseInstlRwd AdAttribute = "eng_to_close_instl_rwd"
	TrueDoubleEndCard     AdAttribute = "true_dbl_endcard"
	CTAOverlay            AdAttribute = "cta_overlay"
	MRAIDAppStatus        AdAttribute = "mraid_app_status"
	CountdownOverlay      AdAttribute = "cntdwn_overlay"
)

// AdFormat represents different types of ad formats
type AdFormat string

const (
	AdFormatInterstitialDisplay      AdFormat = "interstitial_display"
	AdFormatInterstitialDisplayVideo AdFormat = "interstitial_display_video"
	AdFormatRewardedVideo            AdFormat = "rewarded_video"
	AdFormatMRECVideoDisplay         AdFormat = "mrec_video_display"
)

// OS represents the operating system
type OS string

const (
	OSAndroid OS = "android"
	OSiOS     OS = "ios"
)

// FeatureConfig defines supported attributes for OS, version, and ad format combinations
type FeatureConfig struct {
	OS         OS
	MinVersion string
	MaxVersion string
	AdFormat   AdFormat
	Attributes []AdAttribute
}

// UnifiedFeatureMatrix maps OS, version, and ad format combinations to supported attributes
var UnifiedFeatureMatrix = []FeatureConfig{
	// Android configurations - Based on specification table
	{OS: OSAndroid, MinVersion: "4.1.0", MaxVersion: "4.20.0", AdFormat: AdFormatInterstitialDisplay, Attributes: []AdAttribute{EngageToCloseInstl}},

	{OS: OSAndroid, MinVersion: "4.3.0", MaxVersion: "4.3.0", AdFormat: AdFormatInterstitialDisplayVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd}},
	{OS: OSAndroid, MinVersion: "4.3.0", MaxVersion: "4.3.0", AdFormat: AdFormatRewardedVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd}},

	{OS: OSAndroid, MinVersion: "4.4.0", MaxVersion: "4.8.0", AdFormat: AdFormatInterstitialDisplayVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard}},
	{OS: OSAndroid, MinVersion: "4.4.0", MaxVersion: "4.8.0", AdFormat: AdFormatRewardedVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard}},

	{OS: OSAndroid, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatInterstitialDisplayVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay}},
	{OS: OSAndroid, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatRewardedVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay}},
	{OS: OSAndroid, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatMRECVideoDisplay, Attributes: []AdAttribute{CTAOverlay}},

	{OS: OSAndroid, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatInterstitialDisplayVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay, MRAIDAppStatus}},
	{OS: OSAndroid, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatRewardedVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay, MRAIDAppStatus}},
	{OS: OSAndroid, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatMRECVideoDisplay, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay, MRAIDAppStatus}},

	// iOS configurations - Based on specification table
	{OS: OSiOS, MinVersion: "4.1.0", MaxVersion: "4.2.0", AdFormat: AdFormatInterstitialDisplay, Attributes: []AdAttribute{EngageToCloseInstl}},

	{OS: OSiOS, MinVersion: "4.3.0", MaxVersion: "4.8.0", AdFormat: AdFormatInterstitialDisplayVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd}},
	{OS: OSiOS, MinVersion: "4.3.0", MaxVersion: "4.8.0", AdFormat: AdFormatRewardedVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd}},

	{OS: OSiOS, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatInterstitialDisplayVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay}},
	{OS: OSiOS, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatRewardedVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay}},
	{OS: OSiOS, MinVersion: "4.9.0", MaxVersion: "5.0.0", AdFormat: AdFormatMRECVideoDisplay, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay}},

	{OS: OSiOS, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatInterstitialDisplayVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay, MRAIDAppStatus}},
	{OS: OSiOS, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatRewardedVideo, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay, MRAIDAppStatus}},
	{OS: OSiOS, MinVersion: "5.1.0", MaxVersion: "", AdFormat: AdFormatMRECVideoDisplay, Attributes: []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay, MRAIDAppStatus}},
}

// buildOWSDKAdAttributesMap returns server-side ext.owsdk fields (e.g. adattributes) from device, SDK version, and format.
// Returns nil if nothing should be added.
func buildOWSDKAdAttributesMap(bidRequest *openrtb2.BidRequest, impCtx models.ImpCtx) map[string]any {
	if bidRequest.Device == nil {
		return nil
	}
	os := DetermineOS(bidRequest.Device.OS)
	if os == "" {
		return nil
	}
	sdkVersion := impCtx.DisplayManagerVer
	if sdkVersion == "" {
		return nil
	}
	adFormat := DetermineAdFormat(impCtx)
	supportedAttributes := GetSupportedAdAttributes(os, sdkVersion, adFormat)
	if len(supportedAttributes) == 0 {
		return nil
	}
	return CreateOWSDKExtension(supportedAttributes)
}

// mergeOWSDKAdAttributesIntoImpExt merges client ext.owsdk (e.g. ctaoverlay) with server-computed adattributes into
// the full imp.ext JSON. clientOWSDK is the incoming request's ext.owsdk before it was stripped for NewExt.
func (m *OpenWrap) mergeOWSDKAdAttributesIntoImpExt(extJSON json.RawMessage, bidRequest *openrtb2.BidRequest, impCtx models.ImpCtx, clientOWSDK map[string]any) (json.RawMessage, error) {
	srv := buildOWSDKAdAttributesMap(bidRequest, impCtx)
	if len(clientOWSDK) == 0 && (srv == nil || len(srv) == 0) {
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

// addAdAttributesToOWSDK adds ad attributes to the OWSDK extension based on OS, SDK version, and ad format
func (m *OpenWrap) addAdAttributesToOWSDK(bidRequest *openrtb2.BidRequest, imp *openrtb2.Imp, impCtx models.ImpCtx) {
	if imp.Ext == nil {
		imp.Ext = json.RawMessage("{}")
	}
	out, err := m.mergeOWSDKAdAttributesIntoImpExt(imp.Ext, bidRequest, impCtx, nil)
	if err != nil {
		glog.Errorf("[owsdk_merge_imp_ext][ImpID]: %s [Error]: %s", imp.ID, err.Error())
		return
	}
	imp.Ext = out
}

// GetSupportedAdAttributes returns list of supported ad attributes based on OS, SDK version, and ad format
func GetSupportedAdAttributes(os OS, sdkVersion string, adFormat AdFormat) []AdAttribute {
	// Check if SDK version is below minimum supported version
	if isVersionLessThan(sdkVersion, "4.1.0") {
		return nil
	}

	// Find matching configuration in unified matrix
	for _, config := range UnifiedFeatureMatrix {
		if config.OS == os &&
			config.AdFormat == adFormat &&
			isVersionInRange(sdkVersion, config.MinVersion, config.MaxVersion) {
			return config.Attributes
		}
	}

	// Return empty slice if no matching configuration found
	return []AdAttribute{}
}

// DetermineAdFormat determines the ad format based on impression instl flag and ad unit configuration
func DetermineAdFormat(impCtx models.ImpCtx) AdFormat {
	// Check if it's rewarded inventory (instl + rwdd)
	if impCtx.IsRewardInventory != nil && *impCtx.IsRewardInventory == 1 {
		if impCtx.Video != nil {
			return AdFormatRewardedVideo
		}
	}

	// Check for interstitial (instl = 1)
	if impCtx.Instl == 1 {
		if impCtx.Video != nil && impCtx.Banner != nil {
			// Interstitial display + video
			return AdFormatInterstitialDisplayVideo
		}
		// Interstitial display only
		return AdFormatInterstitialDisplay
	}

	// Check for MREC (300x250) when instl = 0
	if impCtx.Banner != nil {
		if impCtx.Banner.W != nil && impCtx.Banner.H != nil {
			if *impCtx.Banner.W == 300 && *impCtx.Banner.H == 250 {
				// MREC logic - always return MREC video display when video is present
				if impCtx.Video != nil {
					return AdFormatMRECVideoDisplay
				}
			}
		}
	}
	return ""
}

// DetermineOS determines the OS based on device information
func DetermineOS(deviceOS string) OS {
	os := strings.ToLower(deviceOS)
	switch {
	case strings.Contains(os, "android"):
		return OSAndroid
	case strings.Contains(os, "ios"), strings.Contains(os, "iphone"), strings.Contains(os, "ipad"):
		return OSiOS
	default:
		return "" // Unknown OS
	}
}

// Helper functions

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

// compareVersions compares two version strings
func compareVersions(v1, v2 string) int {
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

// containsAttribute checks if a slice contains a specific attribute
func containsAttribute(attributes []AdAttribute, attribute AdAttribute) bool {
	for _, attr := range attributes {
		if attr == attribute {
			return true
		}
	}
	return false
}

// CreateOWSDKExtension creates the OWSDK extension with ad attributes
func CreateOWSDKExtension(attributes []AdAttribute) map[string]any {
	owsdk := make(map[string]any)

	if len(attributes) > 0 {
		// Convert AdAttribute slice to string slice
		attrStrings := make([]string, len(attributes))
		for i, attr := range attributes {
			attrStrings[i] = string(attr)
		}
		owsdk["adattributes"] = attrStrings
	}

	return owsdk
}
