package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func TestGetSupportedAdAttributes(t *testing.T) {
	tests := []struct {
		name        string
		os          OS
		sdkVersion  string
		adFormat    AdFormat
		expected    []AdAttribute
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
			expected:    []AdAttribute{EngageToCloseInstl},
			description: "Basic support for interstitial display",
		},
		{
			name:        "Android 4.3.0 - rewarded video",
			os:          OSAndroid,
			sdkVersion:  "4.3.0",
			adFormat:    AdFormatRewardedVideo,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd},
			description: "Support for rewarded video added",
		},
		{
			name:        "Android 4.5.0 - true double end card",
			os:          OSAndroid,
			sdkVersion:  "4.5.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard},
			description: "True double end card support added",
		},
		{
			name:        "Android 4.9.0 - CTA overlay",
			os:          OSAndroid,
			sdkVersion:  "4.9.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay},
			description: "CTA overlay support added",
		},
		{
			name:        "Android 5.1.0 - interstitial display + video",
			os:          OSAndroid,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay, MRAIDAppStatus},
			description: "Interstitial display + video combination",
		},
		{
			name:        "Android 5.1.0 - MREC display + video",
			os:          OSAndroid,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatMRECVideoDisplay,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay, MRAIDAppStatus},
			description: "MREC display + video combination",
		},
		{
			name:        "Android 5.1.0 - full support",
			os:          OSAndroid,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay, MRAIDAppStatus},
			description: "Full feature support",
		},
		{
			name:        "Android 5.1.0 - MREC display + video",
			os:          OSAndroid,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatMRECVideoDisplay,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, TrueDoubleEndCard, CTAOverlay, MRAIDAppStatus},
			description: "MREC display + video combination",
		},
		{
			name:        "iOS 5.1.0 - interstitial display + video",
			os:          OSiOS,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatInterstitialDisplayVideo,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay, MRAIDAppStatus},
			description: "iOS interstitial display + video combination (no true double end card)",
		},
		{
			name:        "iOS 5.1.0 - MREC display + video",
			os:          OSiOS,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatMRECVideoDisplay,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay, MRAIDAppStatus},
			description: "iOS MREC display + video combination",
		},
		{
			name:        "iOS 5.1.0 - MREC display + video",
			os:          OSiOS,
			sdkVersion:  "5.1.0",
			adFormat:    AdFormatMRECVideoDisplay,
			expected:    []AdAttribute{EngageToCloseInstl, EngageToCloseInstlRwd, CTAOverlay, MRAIDAppStatus},
			description: "iOS MREC display + video should include all features",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSupportedAdAttributes(tt.os, tt.sdkVersion, tt.adFormat)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d attributes, got %d", len(tt.expected), len(result))
				return
			}

			for i, attr := range result {
				if i >= len(tt.expected) || attr != tt.expected[i] {
					t.Errorf("Expected attribute at position %d to be %v, got %v", i, tt.expected[i], attr)
				}
			}
		})
	}
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
			name: "rewarded display (instl + rwdd)",
			impCtx: models.ImpCtx{
				IsRewardInventory: ptrutil.ToPtr(int8(1)),
				Instl:             1,
				Video:             nil,
			},
			expected: AdFormatInterstitialDisplay,
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
		name       string
		attributes []AdAttribute
		expected   map[string]any
	}{
		{
			name:       "no attributes",
			attributes: []AdAttribute{},
			expected:   map[string]any{},
		},
		{
			name:       "single attribute",
			attributes: []AdAttribute{CTAOverlay},
			expected: map[string]any{
				"adattributes": []string{"cta_overlay"},
			},
		},
		{
			name:       "multiple attributes with mraid status",
			attributes: []AdAttribute{EngageToCloseInstl, CTAOverlay, MRAIDAppStatus},
			expected: map[string]any{
				"adattributes": []string{"eng_to_close_instl", "cta_overlay", "mraid_app_status"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateOWSDKExtension(tt.attributes)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected key %s not found", key)
				} else {
					// Compare string slices for adattributes
					if key == "adattributes" {
						expectedSlice := expectedValue.([]string)
						actualSlice := actualValue.([]string)
						if len(expectedSlice) != len(actualSlice) {
							t.Errorf("Expected %d attributes, got %d", len(expectedSlice), len(actualSlice))
						} else {
							for i, attr := range expectedSlice {
								if actualSlice[i] != attr {
									t.Errorf("Expected attribute %s at position %d, got %s", attr, i, actualSlice[i])
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
