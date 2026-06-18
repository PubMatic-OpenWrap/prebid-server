package openwrap

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

// Helper function to validate OWSDK presence and content (adattributes = numeric wire IDs).
func validateOWSDK(t *testing.T, owsdkRaw json.RawMessage, expectedIDs []int) {
	var owsdk map[string]any
	if err := json.Unmarshal(owsdkRaw, &owsdk); err != nil {
		t.Fatalf("Failed to unmarshal OWSDK extension: %v", err)
	}

	if adAttrs, ok := owsdk["adattributes"]; ok {
		adAttrsSlice, ok := adAttrs.([]interface{})
		if !ok {
			t.Errorf("Expected adattributes to be a slice")
			return
		}
		got := make([]int, 0, len(adAttrsSlice))
		for _, v := range adAttrsSlice {
			switch n := v.(type) {
			case float64:
				got = append(got, int(n))
			case int:
				got = append(got, n)
			case int64:
				got = append(got, int(n))
			default:
				t.Fatalf("unexpected adattributes element type %T", v)
			}
		}
		assert.Equal(t, expectedIDs, got, "adattributes wire IDs mismatch")
	} else {
		t.Errorf("Expected adattributes in OWSDK extension")
	}
}

func TestMergeOWSDKAdAttributesIntoImpExt_Integration(t *testing.T) {
	ow := &OpenWrap{}

	tests := []struct {
		name            string
		bidRequest      *openrtb2.BidRequest
		imp             *openrtb2.Imp
		impCtx          models.ImpCtx
		expectedAttrIDs []int
		expectOWSDK     bool
	}{
		{
			name: "Android_5.1.0 - interstitial_video + display",
			bidRequest: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					OS: "Android",
				},
			},
			imp: &openrtb2.Imp{
				ID:     "test_imp_1",
				Instl:  1,
				Video:  &openrtb2.Video{MinDuration: 10, MaxDuration: 10},
				Banner: &openrtb2.Banner{W: ptrutil.ToPtr(int64(300)), H: ptrutil.ToPtr(int64(250))},
				Ext:    json.RawMessage(`{"prebid":{}}`),
			},
			impCtx: models.ImpCtx{
				DisplayManagerVer: "5.1.0",
				Instl:             1,
				Video:             &openrtb2.Video{MinDuration: 10, MaxDuration: 10},
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(300)), H: ptrutil.ToPtr(int64(250))},
			},
			expectedAttrIDs: []int{1, 2, 3, 4},
			expectOWSDK:     true,
		},
		{
			name: "iOS_4.9.0 - MREC_display + video",
			bidRequest: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					OS: "iOS",
				},
			},
			imp: &openrtb2.Imp{
				ID:     "test_imp_2",
				Instl:  0,
				Video:  &openrtb2.Video{MinDuration: 10, MaxDuration: 10},
				Banner: &openrtb2.Banner{W: ptrutil.ToPtr(int64(300)), H: ptrutil.ToPtr(int64(250))},
				Ext:    json.RawMessage(`{"prebid":{}}`),
			},
			impCtx: models.ImpCtx{
				DisplayManagerVer: "4.9.0",
				Instl:             0,
				Video:             &openrtb2.Video{MinDuration: 10, MaxDuration: 10},
				Banner:            &openrtb2.Banner{W: ptrutil.ToPtr(int64(300)), H: ptrutil.ToPtr(int64(250))},
			},
			expectedAttrIDs: []int{3},
			expectOWSDK:     true,
		},
		{
			name: "Android_4.0.0 - below_minimum_version",
			bidRequest: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					OS: "Android",
				},
			},
			imp: &openrtb2.Imp{
				ID:    "test_imp_3",
				Instl: 1,
				Video: &openrtb2.Video{},
				Ext:   json.RawMessage(`{"prebid":{}}`),
			},
			impCtx: models.ImpCtx{
				DisplayManagerVer: "4.0.0",
				Instl:             1,
				Video:             &openrtb2.Video{},
			},
			expectedAttrIDs: nil,
			expectOWSDK:     false,
		},
		{
			name: "Unknown_OS",
			bidRequest: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					OS: "Windows",
				},
			},
			imp: &openrtb2.Imp{
				ID:    "test_imp_4",
				Instl: 1,
				Video: &openrtb2.Video{},
				Ext:   json.RawMessage(`{"prebid":{}}`),
			},
			impCtx: models.ImpCtx{
				DisplayManagerVer: "5.1.0",
				Instl:             1,
				Video:             &openrtb2.Video{},
			},
			expectedAttrIDs: nil,
			expectOWSDK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := tt.imp.Ext
			if len(ext) == 0 {
				ext = json.RawMessage("{}")
			}
			deviceOS := ""
			if tt.bidRequest != nil && tt.bidRequest.Device != nil {
				deviceOS = tt.bidRequest.Device.OS
			}
			out, err := ow.mergeOWSDKAdAttributesIntoImpExt(ext, tt.impCtx, nil, deviceOS)
			if err != nil {
				t.Fatalf("mergeOWSDKAdAttributesIntoImpExt: %v", err)
			}
			tt.imp.Ext = out

			// Parse the impression extension
			var extMap map[string]json.RawMessage
			if err := json.Unmarshal(tt.imp.Ext, &extMap); err != nil {
				t.Fatalf("Failed to unmarshal imp extension: %v", err)
			}

			// Check if OWSDK extension exists
			owsdkRaw, exists := extMap["owsdk"]
			if tt.expectOWSDK {
				assert.True(t, exists, "Expected OWSDK extension to be present")
				validateOWSDK(t, owsdkRaw, tt.expectedAttrIDs)
			} else {
				assert.False(t, exists, "Expected OWSDK extension to be absent")
			}
		})
	}
}
