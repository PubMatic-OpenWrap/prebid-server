package pangle

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyPangleResponse(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    *openrtb2.BidResponse
	}{
		{
			name: "non pangle endpoint",
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
			name: "pangle rejected drops default seatbid and sets nbr",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointPangle,
				Pangle: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID:  "test-id",
				Cur: "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ImpID: "imp-1",
							},
						},
					},
				},
			},
			expected: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
		},
		{
			name: "pangle rejected with debug keeps response ext",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointPangle,
				Debug:    true,
				Pangle: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID:  "test-id",
				Ext: json.RawMessage(`{"debug":{"httpcalls":{}},"errors":{"pubmatic":[{"code":1,"message":"no bid"}]}}`),
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ImpID: "imp-1",
							},
						},
					},
				},
			},
			expected: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
				Ext: json.RawMessage(`{"debug":{"httpcalls":{}},"errors":{"pubmatic":[{"code":1,"message":"no bid"}]}}`),
			},
		},
		{
			name: "pangle rejected keeps existing nbr and drops seatbid",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointPangle,
				Pangle: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidTechnicalError.Ptr(),
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-1",
								ImpID: "imp-1",
							},
						},
					},
				},
			},
			expected: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidTechnicalError.Ptr(),
			},
		},
		{
			name: "valid pangle response",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointPangle,
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
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPangleResponse(tt.rctx, tt.bidResponse)

			// Compare responses by marshaling to JSON
			expectedJSON, err := json.Marshal(tt.expected)
			assert.NoError(t, err)

			actualJSON, err := json.Marshal(result)
			assert.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestSetPangleResponseReject(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    bool
	}{
		{
			name: "non pangle endpoint",
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
				Endpoint: models.EndpointPangle,
				Debug:    false,
			},
			bidResponse: &openrtb2.BidResponse{
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
			expected: true,
		},
		{
			name: "nbr present with debug true",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointPangle,
				Debug:    true,
			},
			bidResponse: &openrtb2.BidResponse{
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
			expected: false,
		},
		{
			name: "empty seatbid array",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointPangle,
			},
			bidResponse: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{},
			},
			expected: true,
		},
		{
			name: "empty bid array",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointPangle,
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
				Endpoint: models.EndpointPangle,
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
			actual := SetPangleResponseReject(tt.rctx, tt.bidResponse)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
