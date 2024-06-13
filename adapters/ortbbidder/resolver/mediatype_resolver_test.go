package resolver

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestMtypeResolver(t *testing.T) {
	resolver := &mtypeResolver{}

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

	t.Run("setValue", func(t *testing.T) {
		testCases := []struct {
			name        string
			adapterBid  map[string]any
			value       any
			expectedBid map[string]any
		}{
			{
				name: "Set value in adapter bid",
				adapterBid: map[string]any{
					"id": "123",
				},
				value: openrtb_ext.BidTypeVideo,
				expectedBid: map[string]any{
					"id":      "123",
					"BidType": openrtb_ext.BidTypeVideo,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resolver.setValue(tc.adapterBid, tc.value)
				assert.Equal(t, tc.expectedBid, tc.adapterBid)
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

func Test_mtypeResolver_autoDetect(t *testing.T) {
	type fields struct {
		valueResolver valueResolver
	}
	type args struct {
		request *openrtb2.BidRequest
		bid     map[string]any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   any
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mtypeResolver{
				valueResolver: tt.fields.valueResolver,
			}
			got, got1 := r.autoDetect(tt.args.request, tt.args.bid)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mtypeResolver.autoDetect() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("mtypeResolver.autoDetect() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
