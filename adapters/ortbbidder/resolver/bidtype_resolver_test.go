package resolver

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestBidtypeResolverGetFromORTBObject(t *testing.T) {
	resolver := &bidTypeResolver{}

	t.Run("getFromORTBObject", func(t *testing.T) {
		testCases := []struct {
			name          string
			bid           map[string]any
			expectedValue any
			expectedFound bool
		}{
			{
				name: "mtype found in bid",
				bid: map[string]any{
					"mtype": 2.0,
				},
				expectedValue: openrtb_ext.BidTypeVideo,
				expectedFound: true,
			},
			{
				name: "mtype found in bid - invalid type",
				bid: map[string]any{
					"mtype": "vide0",
				},
				expectedValue: nil,
				expectedFound: false,
			},
			{
				name: "mtype found in bid - invalid value",
				bid: map[string]any{
					"mtype": 11.0,
				},
				expectedValue: nil,
				expectedFound: false,
			},
			{
				name:          "mtype not found in bid",
				bid:           map[string]any{},
				expectedValue: nil,
				expectedFound: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, found := resolver.getFromORTBObject(tc.bid)
				assert.Equal(t, tc.expectedValue, value)
				assert.Equal(t, tc.expectedFound, found)
			})
		}
	})

}

func TestBidTypeResolverRetrieveFromBidderParamLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"mtype": "video",
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.mtype",
			expectedValue: openrtb_ext.BidType("video"),
			expectedFound: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"mtype": openrtb_ext.BidType("video"),
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	resolver := &bidTypeResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestBidTypeResolverAutoDetect(t *testing.T) {
	resolver := &bidTypeResolver{}

	t.Run("autoDetect", func(t *testing.T) {
		testCases := []struct {
			name          string
			bid           map[string]any
			request       *openrtb2.BidRequest
			expectedValue any
			expectedFound bool
		}{
			{
				name: "Auto detect from imp - Video",
				bid: map[string]any{
					"adm":   "",
					"impid": "123",
				},
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{
							ID:    "123",
							Video: &openrtb2.Video{},
						},
					},
				},
				expectedValue: openrtb_ext.BidTypeVideo,
				expectedFound: true,
			},
			{
				name: "Auto detect from imp - banner",
				bid: map[string]any{
					"adm":   "",
					"impid": "123",
				},
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{
							ID:     "123",
							Banner: &openrtb2.Banner{},
						},
					},
				},
				expectedValue: openrtb_ext.BidTypeBanner,
				expectedFound: true,
			},
			{
				name: "Auto detect from imp - native",
				bid: map[string]any{
					"adm":   "",
					"impid": "123",
				},
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{
							ID:     "123",
							Native: &openrtb2.Native{},
						},
					},
				},
				expectedValue: openrtb_ext.BidTypeNative,
				expectedFound: true,
			},
			{
				name: "Auto detect from imp - multi format",
				bid: map[string]any{
					"adm":   "",
					"impid": "123",
				},
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{
							ID:     "123",
							Banner: &openrtb2.Banner{},
							Video:  &openrtb2.Video{},
						},
					},
				},
				expectedValue: openrtb_ext.BidType(""),
				expectedFound: true,
			},
			{
				name: "Auto detect with Video Adm",
				bid: map[string]any{
					"adm": "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
				},
				expectedValue: openrtb_ext.BidTypeVideo,
				expectedFound: true,
			},
			{
				name: "Auto detect with Native Adm",
				bid: map[string]any{
					"adm": "{\"native\":{\"link\":{},\"assets\":[]}}",
				},
				expectedValue: openrtb_ext.BidTypeNative,
				expectedFound: true,
			},
			{
				name: "Auto detect with Banner Adm",
				bid: map[string]any{
					"adm": "<div>Some HTML content</div>",
				},
				expectedValue: openrtb_ext.BidTypeBanner,
				expectedFound: true,
			},
			{
				name:          "Auto detect with no Adm",
				bid:           map[string]any{},
				expectedValue: nil,
				expectedFound: false,
			},
			{
				name: "Auto detect with empty Adm",
				bid: map[string]any{
					"adm": "",
				},
				expectedValue: nil,
				expectedFound: false,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, found := resolver.autoDetect(tc.request, tc.bid)
				assert.Equal(t, tc.expectedValue, value)
				assert.Equal(t, tc.expectedFound, found)
			})
		}
	})
}
func TestGetMediaTypeFromImp(t *testing.T) {
	testCases := []struct {
		name              string
		impressions       []openrtb2.Imp
		impID             string
		expectedMediaType openrtb_ext.BidType
	}{
		{
			name: "Found matching impID",
			impressions: []openrtb2.Imp{
				{ID: "imp1"},
				{ID: "imp2", Banner: &openrtb2.Banner{}},
			},
			impID:             "imp2",
			expectedMediaType: openrtb_ext.BidType("banner"),
		},
		{
			name: "ImpID not found",
			impressions: []openrtb2.Imp{
				{ID: "imp1"},
				{ID: "imp2"},
			},
			impID:             "imp3",
			expectedMediaType: openrtb_ext.BidType(""),
		},
		{
			name:              "Empty impressions slice",
			impressions:       nil,
			impID:             "imp1",
			expectedMediaType: openrtb_ext.BidType(""),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mediaType := getMediaTypeFromImp(tc.impressions, tc.impID)
			assert.Equal(t, tc.expectedMediaType, mediaType)
		})
	}
}

func TestMtypeResolverSetValue(t *testing.T) {
	resolver := &bidTypeResolver{}

	t.Run("setValue", func(t *testing.T) {
		testCases := []struct {
			name            string
			typeBid         map[string]any
			value           any
			expectedTypeBid map[string]any
		}{
			{
				name: "Set value in adapter bid",
				typeBid: map[string]any{
					"id": "123",
				},
				value: openrtb_ext.BidTypeVideo,
				expectedTypeBid: map[string]any{
					"id":      "123",
					"BidType": openrtb_ext.BidTypeVideo,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resolver.setValue(tc.typeBid, tc.value)
				assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
			})
		}
	})
}
func TestGetMediaTypeFromAdm(t *testing.T) {
	tests := []struct {
		name     string
		adm      string
		expected openrtb_ext.BidType
	}{
		{
			name:     "Video Adm",
			adm:      "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
			expected: openrtb_ext.BidTypeVideo,
		},
		{
			name:     "Native Adm",
			adm:      "{\"native\":{\"link\":{},\"assets\":[]}}",
			expected: openrtb_ext.BidTypeNative,
		},
		{
			name:     "Banner Adm",
			adm:      "<div>Some HTML content</div>",
			expected: openrtb_ext.BidTypeBanner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMediaTypeFromAdm(tt.adm)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMediaType(t *testing.T) {
	tests := []struct {
		name            string
		mtype           openrtb2.MarkupType
		expectedBidType openrtb_ext.BidType
	}{
		{
			name:            "MarkupBanner",
			mtype:           openrtb2.MarkupBanner,
			expectedBidType: openrtb_ext.BidTypeBanner,
		},
		{
			name:            "MarkupVideo",
			mtype:           openrtb2.MarkupVideo,
			expectedBidType: openrtb_ext.BidTypeVideo,
		},
		{
			name:            "MarkupAudio",
			mtype:           openrtb2.MarkupAudio,
			expectedBidType: openrtb_ext.BidTypeAudio,
		},
		{
			name:            "MarkupNative",
			mtype:           openrtb2.MarkupNative,
			expectedBidType: openrtb_ext.BidTypeNative,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToBidType(tt.mtype)
			assert.Equal(t, tt.expectedBidType, result)
		})
	}
}
