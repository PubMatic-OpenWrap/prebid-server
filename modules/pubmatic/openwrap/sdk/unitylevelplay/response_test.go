package unitylevelplay

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyUnityLevelPlayResponse(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    *openrtb2.BidResponse
	}{
		{
			name: "non unity levelplay endpoint",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointGoogleSDK,
			},
			bidResponse: &openrtb2.BidResponse{
				ID: "test-id",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-1",
								ImpID: "imp-1",
								Price: 1.0,
							},
						},
					},
				},
			},
			expected: &openrtb2.BidResponse{
				ID: "test-id",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-1",
								ImpID: "imp-1",
								Price: 1.0,
							},
						},
					},
				},
			},
		},
		{
			name: "unity levelplay rejected",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointUnityLevelPlay,
				UnityLevelPlay: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID: "test-id",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-1",
								ImpID: "imp-1",
								Price: 1.0,
							},
						},
					},
				},
			},
			expected: &openrtb2.BidResponse{
				ID: "test-id",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-1",
								ImpID: "imp-1",
								Price: 1.0,
							},
						},
					},
				},
			},
		},
		{
			name: "valid unity levelplay response",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointUnityLevelPlay,
			},
			bidResponse: &openrtb2.BidResponse{
				ID:  "test-id",
				Cur: "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-1",
								ImpID: "imp-1",
								Price: 1.0,
								BURL:  "http://example.com",
								Ext:   json.RawMessage(`{"test":1}`),
							},
						},
					},
				},
			},
			expected: &openrtb2.BidResponse{
				ID:    "test-id",
				BidID: "bid-1",
				Cur:   "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-1",
								ImpID: "imp-1",
								Price: 1.0,
								AdM:   `{"id":"test-id","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":1,"burl":"http://example.com","ext":{"test":1}}]}],"cur":"USD"}`,
								BURL:  "http://example.com",
								Ext:   json.RawMessage(`{"test":1}`),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyUnityLevelPlayResponse(tt.rctx, tt.bidResponse)

			// Compare responses by marshaling to JSON
			expectedJSON, err := json.Marshal(tt.expected)
			assert.NoError(t, err)

			actualJSON, err := json.Marshal(result)
			assert.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestSetUnityLevelPlayResponseReject(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    bool
	}{
		{
			name: "non unity levelplay endpoint",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointGoogleSDK,
			},
			bidResponse: &openrtb2.BidResponse{
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
			expected: false,
		},
		{
			name: "nbr present with debug false",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointUnityLevelPlay,
				Debug:   false,
			},
			bidResponse: &openrtb2.BidResponse{
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
			expected: true,
		},
		{
			name: "nbr present with debug true",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointUnityLevelPlay,
				Debug:   true,
			},
			bidResponse: &openrtb2.BidResponse{
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
			expected: false,
		},
		{
			name: "empty seatbid array",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointUnityLevelPlay,
			},
			bidResponse: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{},
			},
			expected: true,
		},
		{
			name: "empty bid array",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointUnityLevelPlay,
			},
			bidResponse: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid bid response",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointUnityLevelPlay,
			},
			bidResponse: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID: "1",
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SetUnityLevelPlayResponseReject(tt.rctx, tt.bidResponse)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
