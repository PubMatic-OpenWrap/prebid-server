package googlesdk

import (
	"encoding/json"
	"testing"

	nativeResponse "github.com/prebid/openrtb/v20/native1/response"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestSetGoogleSDKResponseReject(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		want        bool
	}{
		{
			name: "NBR present with debug false",
			rctx: models.RequestCtx{Debug: false},
			bidResponse: &openrtb2.BidResponse{
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
			want: true,
		},
		{
			name: "NBR present with debug true",
			rctx: models.RequestCtx{Debug: true},
			bidResponse: &openrtb2.BidResponse{
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
			want: false,
		},
		{
			name:        "Empty bid response",
			rctx:        models.RequestCtx{Debug: false},
			bidResponse: &openrtb2.BidResponse{},
			want:        true,
		},
		{
			name: "Valid bid response",
			rctx: models.RequestCtx{Debug: false},
			bidResponse: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{{ID: "1"}},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetGoogleSDKResponseReject(tt.rctx, tt.bidResponse)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDeclaredAd(t *testing.T) {
	tests := []struct {
		name string
		rctx models.RequestCtx
		bid  openrtb2.Bid
		want models.DeclaredAd
	}{
		{
			name: "Banner ad",
			rctx: models.RequestCtx{
				Trackers: map[string]models.OWTracker{
					"1": {BidType: "banner"},
				},
			},
			bid: openrtb2.Bid{
				ID:  "1",
				AdM: "<a href='http://example.com'>Click</a>",
			},
			want: models.DeclaredAd{
				HTMLSnippet:     "<a href='http://example.com'>Click</a>",
				ClickThroughURL: []string{"http://example.com"},
			},
		},
		{
			name: "Banner ad with click_urls array",
			rctx: models.RequestCtx{
				Trackers: map[string]models.OWTracker{
					"2": {BidType: "banner"},
				},
			},
			bid: openrtb2.Bid{
				ID:  "2",
				AdM: `{"click_urls":["http://array-url.com"]}`,
			},
			want: models.DeclaredAd{
				HTMLSnippet:     `{"click_urls":["http://array-url.com"]}`,
				ClickThroughURL: []string{"http://array-url.com"},
			},
		},
		{
			name: "Video ad",
			rctx: models.RequestCtx{
				Trackers: map[string]models.OWTracker{
					"1": {BidType: "video"},
				},
			},
			bid: openrtb2.Bid{
				ID:  "1",
				AdM: "<VAST><Ad><InLine><Creatives><Creative><Linear><VideoClicks><ClickThrough>http://example.com</ClickThrough></VideoClicks></Linear></Creative></Creatives></InLine></Ad></VAST>",
			},
			want: models.DeclaredAd{
				VideoVastXML:    "<VAST><Ad><InLine><Creatives><Creative><Linear><VideoClicks><ClickThrough>http://example.com</ClickThrough></VideoClicks></Linear></Creative></Creatives></InLine></Ad></VAST>",
				ClickThroughURL: []string{"http://example.com"},
			},
		},
		{
			name: "Native ad",
			rctx: models.RequestCtx{
				Trackers: map[string]models.OWTracker{
					"1": {BidType: "native"},
				},
			},
			bid: openrtb2.Bid{
				ID:  "1",
				AdM: `{"link":{"url":"http://example.com"}}`,
			},
			want: models.DeclaredAd{
				NativeResponse: &nativeResponse.Response{
					Link: nativeResponse.Link{
						URL: "http://example.com",
					},
				},
				ClickThroughURL: []string{"http://example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDeclaredAd(tt.rctx, tt.bid)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetSDKRenderedAdID(t *testing.T) {
	tests := []struct {
		name     string
		app      *openrtb2.App
		endpoint string
		want     string
	}{
		{
			name:     "Non GoogleSDK endpoint",
			app:      &openrtb2.App{},
			endpoint: "other",
			want:     "",
		},
		{
			name: "Valid SDK ID in installed_sdk.id",
			app: &openrtb2.App{
				Ext: json.RawMessage(`{"installed_sdk":{"id":"test-sdk-id"}}`),
			},
			endpoint: models.EndpointGoogleSDK,
			want:     "test-sdk-id",
		},
		{
			name: "Valid SDK ID in installed_sdk array",
			app: &openrtb2.App{
				Ext: json.RawMessage(`{"installed_sdk":[{"id":"test-sdk-id"}]}`),
			},
			endpoint: models.EndpointGoogleSDK,
			want:     "test-sdk-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetSDKRenderedAdID(tt.app, tt.endpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBannerClickThroughURL(t *testing.T) {
	tests := []struct {
		name     string
		creative string
		want     []string
	}{
		{
			name:     "Empty creative",
			creative: "",
			want:     []string{},
		},
		{
			name:     "JSON creative with click_urls array",
			creative: `{"click_urls":["http://example.com"]}`,
			want:     []string{"http://example.com"},
		},
		{
			name:     "HTML creative with anchor tag",
			creative: `<a href="http://example.com">Click</a>`,
			want:     []string{"http://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBannerClickThroughURL(tt.creative)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractClickURLFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		creative string
		want     string
	}{
		{
			name:     "No click_urls",
			creative: `{"other":"value"}`,
			want:     "",
		},
		{
			name:     "Click URLs array",
			creative: `{"click_urls":["http://example.com"]}`,
			want:     "http://example.com",
		},
		{
			name:     "Click URL string",
			creative: `{"click_urls":"http://example.com"}`,
			want:     "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractClickURLFromJSON(tt.creative)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractClickURLFromHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "Valid anchor tag",
			html: `<a href="http://example.com">Click</a>`,
			want: "http://example.com",
		},
		{
			name: "No anchor tag",
			html: `<div>No link here</div>`,
			want: "",
		},
		{
			name: "Invalid HTML",
			html: "Not HTML",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractClickURLFromHTML(tt.html)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestApplyGoogleSDKResponse(t *testing.T) {
	tests := []struct {
		name    string
		rctx    models.RequestCtx
		bidResp *openrtb2.BidResponse
		wantID  string
		wantNBR bool
	}{
		{
			name:    "Non GoogleSDK endpoint returns input",
			rctx:    models.RequestCtx{Endpoint: "other"},
			bidResp: &openrtb2.BidResponse{ID: "test-non-gsdk"},
			wantID:  "test-non-gsdk",
			wantNBR: false,
		},
		{
			name:    "GoogleSDK endpoint, debug true, empty SeatBid",
			rctx:    models.RequestCtx{Endpoint: models.EndpointGoogleSDK, Debug: true},
			bidResp: &openrtb2.BidResponse{ID: "test-debug-empty"},
			wantID:  "test-debug-empty",
			wantNBR: false,
		},
		{
			name:    "GoogleSDK endpoint, debug true, NBR present",
			rctx:    models.RequestCtx{Endpoint: models.EndpointGoogleSDK, Debug: true},
			bidResp: &openrtb2.BidResponse{ID: "test-debug-nbr", NBR: openrtb3.NoBidUnknownError.Ptr()},
			wantID:  "test-debug-nbr",
			wantNBR: true,
		},
		{
			name:    "GoogleSDK endpoint, reject true sets NBR",
			rctx:    models.RequestCtx{Endpoint: models.EndpointGoogleSDK, GoogleSDK: models.GoogleSDK{Reject: true}, StartTime: 0},
			bidResp: &openrtb2.BidResponse{ID: "test-reject", NBR: openrtb3.NoBidUnknownError.Ptr()},
			wantID:  "test-reject",
			wantNBR: true,
		},
		{
			name:    "GoogleSDK endpoint, customizeBid path",
			rctx:    models.RequestCtx{Endpoint: models.EndpointGoogleSDK},
			bidResp: &openrtb2.BidResponse{ID: "test-customok", Cur: "USD", SeatBid: []openrtb2.SeatBid{{Bid: []openrtb2.Bid{{ID: "bid1"}}}}},
			wantID:  "test-customok",
			wantNBR: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyGoogleSDKResponse(tt.rctx, tt.bidResp)
			assert.Equal(t, tt.wantID, got.ID)
			if tt.wantNBR {
				assert.NotNil(t, got.NBR)
			}
		})
	}
}

func TestCustomizeBid(t *testing.T) {
	type want struct {
		bids   []openrtb2.Bid
		wantOK bool
	}

	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		want        want
	}{
		{
			name:        "marshal error returns nil,false",
			rctx:        models.RequestCtx{},
			bidResponse: &openrtb2.BidResponse{}, // Empty SeatBid triggers error path
			want: want{
				bids:   nil,
				wantOK: false,
			},
		},
		{
			name: "empty bid returns nil,false",
			rctx: models.RequestCtx{},
			bidResponse: &openrtb2.BidResponse{
				ID: "id",
				SeatBid: []openrtb2.SeatBid{
					{Bid: []openrtb2.Bid{}},
				},
			},
			want: want{
				bids:   nil,
				wantOK: false,
			},
		},
		{
			name: "empty seatbid returns nil,false",
			rctx: models.RequestCtx{},
			bidResponse: &openrtb2.BidResponse{
				ID: "id",
			},
			want: want{
				bids:   nil,
				wantOK: false,
			},
		},
		{
			name: "happy path",
			rctx: models.RequestCtx{
				GoogleSDK: models.GoogleSDK{
					SDKRenderedAdID: "sdkrenderedaid",
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID: "id",
				SeatBid: []openrtb2.SeatBid{
					{Bid: []openrtb2.Bid{{ID: "bidid"}}},
				},
			},
			want: want{
				bids: []openrtb2.Bid{
					{
						ID:  "bidid",
						Ext: json.RawMessage(`{"sdk_rendered_ad":{"id":"sdkrenderedaid","rendering_data":"{\"id\":\"id\",\"seatbid\":[{\"bid\":[{\"id\":\"bidid\",\"impid\":\"\",\"price\":0}]}]}","declared_ad":{}}}`),
						AdM: "",
					},
				},
				wantOK: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bids, ok := customizeBid(tt.rctx, tt.bidResponse)
			assert.Equal(t, tt.want.bids, bids)
			assert.Equal(t, tt.want.wantOK, ok)
		})
	}
}
