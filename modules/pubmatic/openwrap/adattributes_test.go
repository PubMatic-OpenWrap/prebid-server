package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func TestGetSupportedAdAttributeWireIDs(t *testing.T) {
	tests := []struct {
		name        string
		os          OS
		sdkVersion  string
		adFormat    AdFormat
		expected    []int
		description string
	}{
		{
			name:        "Android 4.0.0 - below minimum",
			os:          OSAndroid,
			sdkVersion:  "4.0.0",
			adFormat:    AdFormatInterstitialDisplay,
			expected:    nil,
			description: "No support for versions below 4.1.0",
		},
		{
			name:        "Android 4.1.0 - interstitial display",
			os:          OSAndroid,
			sdkVersion:  "4.1.0",
			adFormat:    AdFormatInterstitialDisplay,
			expected:    []int{AdAttrWireEngageToClose},
			description: "Basic support for interstitial display",
		},
		{
			name:        "Android 4.3.0 - interstitial display (no matrix row after 4.2)",
			os:          OSAndroid,
			sdkVersion:  "4.3.0",
			adFormat:    AdFormatInterstitialDisplay,
			expected:    nil,
			description: "Interstitial display only supported on Android 4.1.0–4.2.0 per spec",
		},
		{
			name:        "Android 4.3.0 - rewarded video",
			os:          OSAndroid,
			sdkVersion:  "4.3.0",
			adFormat:    AdFormatRewardedVideo,
			expected:    []int{AdAttrWireEngageToClose},
			description: "Rewarded video at 4.3.0",
		},
		{
			name:        "Android 4.5.0 - banner display",
			os:          OSAndroid,
			sdkVersion:  "4.5.0",
			adFormat:    AdFormatBannerDisplay,
			expected:    []int{AdAttrWireEngageToClose},
			description: "Android 4.1.0–4.8.0 banner display",
		},
		{
			name:        "Android 4.9.0 - MREC display + video",
			os:          OSAndroid,
			sdkVersion:  "4.9.0",
			adFormat:    AdFormatMRECVideoDisplay,
			expected:    []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay},
			description: "Android 4.9–5.0 MREC: engage to close + CTA overlay",
		},
		{
			name:        "Android 4.5.0 - true double end card",
			os:          OSAndroid,
			sdkVersion:  "4.5.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard},
			description: "True double end card support",
		},
		{
			name:        "Android 4.9.0 - CTA overlay",
			os:          OSAndroid,
			sdkVersion:  "4.9.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard, AdAttrWireCTAOverlay},
			description: "CTA overlay support",
		},
		{
			name:        "Android 5.1.0 - interstitial display + video",
			os:          OSAndroid,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []int{AdAttrWireEngageToClose, AdAttrWireTrueDoubleEndCard, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus},
			description: "Interstitial display + video combination",
		},
		{
			name:        "Android 5.1.0 - MREC display + video",
			os:          OSAndroid,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatMRECVideoDisplay,
			expected:    []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus},
			description: "Android 5.1+ MREC: no true double end card on wire list",
		},
		{
			name:        "Android 5.1.0 - banner display",
			os:          OSAndroid,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatBannerDisplay,
			expected:    []int{AdAttrWireEngageToClose, AdAttrWireMRAIDAppStatus},
			description: "Android 5.1+ banner: engage to close + MRAID app status",
		},
		{
			name:        "iOS 5.1.0 - interstitial display + video",
			os:          OSiOS,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []int{AdAttrWireEngageToClose, AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus},
			description: "iOS interstitial display + video (no true double end card in matrix)",
		},
		{
			name:        "iOS 5.1.0 - MREC display + video",
			os:          OSiOS,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatMRECVideoDisplay,
			expected:    []int{AdAttrWireCTAOverlay, AdAttrWireMRAIDAppStatus},
			description: "iOS 5.1+ MREC: CTA overlay + MRAID only",
		},
		{
			name:        "iOS 5.1.0 - banner display",
			os:          OSiOS,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatBannerDisplay,
			expected:    []int{AdAttrWireMRAIDAppStatus},
			description: "iOS 5.1+ banner: MRAID app status only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSupportedAdAttributeWireIDs(tt.os, tt.sdkVersion, tt.adFormat)
			if !intSliceEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v (%s)", tt.expected, result, tt.description)
			}
		})
	}
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDetermineAdFormat(t *testing.T) {
	tests := []struct {
		name     string
		impCtx   models.ImpCtx
		expected AdFormat
	}{
		{
			name: "rewarded video (instl + rwdd)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(1)),
				Instl:             1,
				Video:             &openrtb2.Video{},
			},
			expected: AdFormatRewardedVideo,
		},
		{
			name: "rewarded flag without video — not classified as rewarded video",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(1)),
				Instl:             1,
				Video:             nil,
			},
			expected: "",
		},
		{
			name: "interstitial display + video (instl = 1 + video + banner)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(0)),
				Instl:             1,
				Video:             &openrtb2.Video{},
				IsBanner:          true,
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(320)), H: ptrutil.ToPtr(int64(50))},
			},
			expected: AdFormatInterstitialDisplayVideo,
		},
		{
			name: "interstitial display only when ad unit disabled video (still on imp until mutation)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(0)),
				Instl:             1,
				Video:             &openrtb2.Video{},
				IsBanner:          true,
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(320)), H: ptrutil.ToPtr(int64(50))},
				VideoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{Enabled: ptrutil.ToPtr(false)},
					},
				},
			},
			expected: AdFormatInterstitialDisplay,
		},
		{
			name: "interstitial display (instl = 1 + display)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(0)),
				Instl:             1,
				Video:             nil,
				IsBanner:          true,
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(320)), H: ptrutil.ToPtr(int64(50))},
			},
			expected: AdFormatInterstitialDisplay,
		},
		{
			name: "MREC video + display (instl = 0 + 300x250 + video + banner)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(0)),
				Instl:             0,
				Video:             &openrtb2.Video{},
				IsBanner:          true,
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(300)), H: ptrutil.ToPtr(int64(250))},
			},
			expected: AdFormatMRECVideoDisplay,
		},
		{
			name: "MREC banner display only when ad unit disabled video",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(0)),
				Instl:             0,
				Video:             &openrtb2.Video{},
				IsBanner:          true,
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(300)), H: ptrutil.ToPtr(int64(250))},
				VideoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{Enabled: ptrutil.ToPtr(false)},
					},
				},
			},
			expected: AdFormatBannerDisplay,
		},
		{
			name: "banner display (instl = 0 + leaderboard, no video)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(0)),
				Instl:             0,
				Video:             nil,
				IsBanner:          true,
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(320)), H: ptrutil.ToPtr(int64(50))},
			},
			expected: AdFormatBannerDisplay,
		},
		{
			name: "default (no match)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(0)),
				Instl:             0,
				Video:             nil,
				IsBanner:          false,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineAdFormat(tt.impCtx)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetermineOS(t *testing.T) {
	tests := []struct {
		name     string
		deviceOS string
		expected OS
	}{
		{
			name:     "Android",
			deviceOS: "Android",
			expected: OSAndroid,
		},
		{
			name:     "android lowercase",
			deviceOS: "android",
			expected: OSAndroid,
		},
		{
			name:     "iOS",
			deviceOS: "iOS",
			expected: OSiOS,
		},
		{
			name:     "iPhone",
			deviceOS: "iPhone",
			expected: OSiOS,
		},
		{
			name:     "iPad",
			deviceOS: "iPad",
			expected: OSiOS,
		},
		{
			name:     "unknown OS",
			deviceOS: "Windows",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineOS(tt.deviceOS)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCreateOWSDKExtension(t *testing.T) {
	tests := []struct {
		name     string
		wireIDs  []int
		expected map[string]any
	}{
		{
			name:     "no ids",
			wireIDs:  []int{},
			expected: map[string]any{},
		},
		{
			name:    "single id",
			wireIDs: []int{AdAttrWireCTAOverlay},
			expected: map[string]any{
				"adattributes": []int{AdAttrWireCTAOverlay},
			},
		},
		{
			name:    "dedupe and sort",
			wireIDs: []int{3, 1, 1, 3},
			expected: map[string]any{
				"adattributes": []int{1, 3},
			},
		},
		{
			name:    "skips non-positive ids",
			wireIDs: []int{0, -1, 4, 3},
			expected: map[string]any{
				"adattributes": []int{3, 4},
			},
		},
		{
			name:    "multiple ids sorted",
			wireIDs: []int{4, 1, 3},
			expected: map[string]any{
				"adattributes": []int{1, 3, 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateOWSDKExtension(tt.wireIDs)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected key %s not found", key)
				} else {
					if key == "adattributes" {
						expectedSlice := expectedValue.([]int)
						actualSlice, ok := actualValue.([]int)
						if !ok {
							t.Errorf("adattributes: expected []int, got %T", actualValue)
							continue
						}
						if len(expectedSlice) != len(actualSlice) {
							t.Errorf("Expected %d attributes, got %d", len(expectedSlice), len(actualSlice))
						} else {
							for i, id := range expectedSlice {
								if actualSlice[i] != id {
									t.Errorf("Expected id %d at position %d, got %d", id, i, actualSlice[i])
								}
							}
						}
					} else if actualValue != expectedValue {
						t.Errorf("Expected %v for key %s, got %v", expectedValue, key, actualValue)
					}
				}
			}
		})
	}
}
