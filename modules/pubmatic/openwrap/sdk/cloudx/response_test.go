package cloudx

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v4/exchange/entities"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v4/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestApplyCloudXResponse(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    *openrtb2.BidResponse
	}{
		{
			name: "non cloudx endpoint",
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
			name: "cloudx rejected",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointCloudX,
				CloudX: struct{ Reject bool }{
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
			name: "valid cloudx response",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointCloudX,
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
			result := ApplyCloudXResponse(tt.rctx, tt.bidResponse)

			// Compare responses by marshaling to JSON
			expectedJSON, err := json.Marshal(tt.expected)
			assert.NoError(t, err)

			actualJSON, err := json.Marshal(result)
			assert.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestSetCloudXResponseReject(t *testing.T) {
	tests := []struct {
		name        string
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
		expected    bool
	}{
		{
			name: "non cloudx endpoint",
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
				Endpoint: models.EndpointCloudX,
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
				Endpoint: models.EndpointCloudX,
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
				Endpoint: models.EndpointCloudX,
			},
			bidResponse: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{},
			},
			expected: true,
		},
		{
			name: "empty bid array",
			rctx: models.RequestCtx{
				Endpoint: models.EndpointCloudX,
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
				Endpoint: models.EndpointCloudX,
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
			actual := SetCloudXResponseReject(tt.rctx, tt.bidResponse)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestUpdateBidWithTestPrice(t *testing.T) {
	type args struct {
		rctx            models.RequestCtx
		bidderResponses map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
	}
	tests := []struct {
		name                 string
		args                 args
		expectedBidResponses map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
	}{
		{
			name: "empty bidder responses",
			args: args{
				rctx: models.RequestCtx{
					Endpoint: models.EndpointCloudX,
				},
				bidderResponses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{},
			},
			expectedBidResponses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{},
		},
		{
			name: "empty seatbid",
			args: args{
				rctx: models.RequestCtx{
					Endpoint: models.EndpointCloudX,
				},
				bidderResponses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{},
					},
				},
			},
			expectedBidResponses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
				"pubmatic": {
					Bids: []*entities.PbsOrtbBid{},
				},
			},
		},
		{
			name: "update bids with test price",
			args: args{
				rctx: models.RequestCtx{
					Endpoint:      models.EndpointCloudX,
					IsTestRequest: 1,
				},
				bidderResponses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									Price: 1.0,
								},
							},
						},
					},
				},
			},
			expectedBidResponses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
				"pubmatic": {
					Bids: []*entities.PbsOrtbBid{
						{
							Bid: &openrtb2.Bid{
								Price: 99,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateBidWithTestPrice(tt.args.rctx, tt.args.bidderResponses)
			assert.Equal(t, tt.expectedBidResponses, tt.args.bidderResponses)
		})
	}
}
