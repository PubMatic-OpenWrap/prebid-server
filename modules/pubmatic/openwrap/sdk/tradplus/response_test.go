package tradplus

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyTradPlusResponse(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    *openrtb2.BidResponse
	}{
		{
			name: "non tradplus endpoint",
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
			name: "tradplus rejected returns nbr",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointTradPlus,
				TradPlus: struct{ Reject bool }{
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
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
		},
		{
			name: "no bid returns nbr instead of blank default bid",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointTradPlus,
				TradPlus: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID: "test-id",
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ID:    "default-bid",
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
			name: "empty seatbid returns nbr",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointTradPlus,
				TradPlus: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID: "test-id",
			},
			expected: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
			},
		},
		{
			name: "debug no bid keeps response ext",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointTradPlus,
				Debug:    true,
				TradPlus: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID:  "test-id",
				Ext: json.RawMessage(`{"debug":{"httpcalls":{}},"errors":{"pubmatic":[]},"owlogger":"keep-me"}`),
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{ID: "default-bid", ImpID: "imp-1"},
						},
					},
				},
			},
			expected: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
				Ext: json.RawMessage(`{"debug":{"httpcalls":{}},"errors":{"pubmatic":[]},"owlogger":"keep-me"}`),
			},
		},
		{
			name: "existing nbr is preserved and blank bid is dropped",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointTradPlus,
				TradPlus: struct{ Reject bool }{
					Reject: true,
				},
			},
			bidResponse: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{ID: "default-bid", ImpID: "imp-1"},
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
			name: "valid tradplus response",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointTradPlus,
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
			result := ApplyTradPlusResponse(tt.rctx, tt.bidResponse)

			// Compare responses by marshaling to JSON
			expectedJSON, err := json.Marshal(tt.expected)
			assert.NoError(t, err)

			actualJSON, err := json.Marshal(result)
			assert.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestApplyTradPlusResponse_PreservesBidExp(t *testing.T) {
	rctx := models.RequestCtx{
		Endpoint: models.EndpointTradPlus,
		ImpBidCtx: map[string]models.ImpCtx{
			"imp-9": {Exp: 600},
		},
	}
	br := &openrtb2.BidResponse{
		ID:  "resp",
		Cur: "USD",
		SeatBid: []openrtb2.SeatBid{{
			Bid: []openrtb2.Bid{{
				ID:    "bid-1",
				ImpID: "imp-9",
				Exp:   120,
				Price: 1.0,
			}},
		}},
	}

	out := ApplyTradPlusResponse(rctx, br)
	assert.Equal(t, int64(120), out.SeatBid[0].Bid[0].Exp)

	var decoded openrtb2.BidResponse
	assert.NoError(t, json.Unmarshal([]byte(out.SeatBid[0].Bid[0].AdM), &decoded))
	assert.Equal(t, int64(120), decoded.SeatBid[0].Bid[0].Exp)
}

func TestSetTradPlusResponseReject(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    bool
	}{
		{
			name: "non tradplus endpoint",
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
				Endpoint: models.EndpointTradPlus,
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
				Endpoint: models.EndpointTradPlus,
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
				Endpoint: models.EndpointTradPlus,
			},
			bidResponse: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{},
			},
			expected: true,
		},
		{
			name: "empty bid array",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointTradPlus,
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
				Endpoint: models.EndpointTradPlus,
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
			actual := SetTradPlusResponseReject(tt.rctx, tt.bidResponse)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
