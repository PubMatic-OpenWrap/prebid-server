package ortbbidder

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestNewResponseBuilder(t *testing.T) {
	testCases := []struct {
		name           string
		request        *openrtb2.BidRequest
		responseParams map[string]bidderparams.BidderParamMapper
		expected       *responseBuilder
	}{
		{
			name: "With non-nil responseParams",
			responseParams: map[string]bidderparams.BidderParamMapper{
				"test": {},
			},
			request: &openrtb2.BidRequest{},
			expected: &responseBuilder{
				responseParams: map[string]bidderparams.BidderParamMapper{
					"test": {},
				},
				request: &openrtb2.BidRequest{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := newResponseBuilder(tc.responseParams, tc.request)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildAdapterResponse(t *testing.T) {
	testCases := []struct {
		name             string
		adapterResponse  map[string]any
		expectedResponse *adapters.BidderResponse
		expectedError    error
	}{
		{
			name: "Valid adapter response",
			adapterResponse: map[string]any{
				"Currency": "USD",
				"Bids": []any{
					map[string]any{
						"Bid": map[string]any{
							"id":    "123",
							"mtype": 2,
						},
						"BidType": "video",
					},
				},
			},
			expectedResponse: &adapters.BidderResponse{
				Currency: "USD",
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:    "123",
							MType: 2,
						},
						BidType: "video",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Invalid adapter response - conversion failed",
			adapterResponse: map[string]any{
				"Currency": "USD",
				"Bids": map[string]any{
					"Bid": map[string]any{
						"id":    "123",
						"mtype": "video",
					},
					"BidType": "video",
				},
			},
			expectedResponse: nil,
			expectedError: &errortypes.FailedToUnmarshal{
				Message: "cannot unmarshal adapters.BidderResponse.Bids: decode slice: expect [ or n, but found {",
			},
		},
		{
			name: "Invalid adapter response - marshal failed",
			adapterResponse: map[string]any{
				"Currency": 123, // Invalid type
				"Bids": map[string]any{
					"Bid": map[string]any{
						"id": "123",
					},
					"BidType": make(chan int),
				},
			},
			expectedResponse: nil,
			expectedError:    &errortypes.FailedToMarshal{Message: "chan int is unsupported type"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rb := &responseBuilder{
				adapterRespone: tc.adapterResponse,
			}
			actualResponse, err := rb.buildAdapterResponse()
			assert.Equal(t, tc.expectedError, err, "error mismatch")
			assert.Equal(t, tc.expectedResponse, actualResponse, "response mismatch")
		})
	}
}

func TestSetPrebidBidderResponse(t *testing.T) {
	testCases := []struct {
		name                string
		bidderResponse      map[string]any
		bidderResponseBytes []byte
		responseParams      map[string]bidderparams.BidderParamMapper
		expectedError       error
		expectedResponse    map[string]any
	}{
		{
			name:                "Invalid bidder response, unmarshal failure",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1", "ext":{"mtype":"video"}}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: &errortypes.FailedToUnmarshal{Message: "expect ] in the end, but found \x00"},
		},
		{
			name:                "Invalid seatbid object in response",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":"invalid"}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid seatbid array found in response, seatbids:[invalid]"},
		},
		{
			name:                "Invalid seatbid is seatbid arrays",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":["invalid"]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid seatbid found in seatbid array, seatbid:[invalid]"},
		},
		{
			name:                "Invalid bid in seatbid",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":"invalid"}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid bid array found in seatbid, bids:[invalid]"},
		},
		{
			name:                "Invalid bid in bids array",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":["invalid"]}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid bid found in bids array, bid:[invalid]"},
		},
		{
			name:                "Valid bidder respone, no bidder params",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"123"}]}]}`),
			responseParams:      map[string]bidderparams.BidderParamMapper{},
			expectedError:       nil,
			expectedResponse: map[string]any{
				"Currency": "USD",
				"Bids": []any{
					map[string]any{
						"Bid": map[string]any{
							"id": "123",
						},
					},
				},
			},
		},
		{
			name:                "Valid bidder respone, no bidder params - bidtype populated",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"123","mtype": 2}]}]}`),
			responseParams:      map[string]bidderparams.BidderParamMapper{},
			expectedError:       nil,
			expectedResponse: map[string]any{
				"Currency": "USD",
				"Bids": []any{
					map[string]any{
						"Bid": map[string]any{
							"id":    "123",
							"mtype": float64(2),
						},
						"BidType": openrtb_ext.BidType("video"),
					},
				},
			},
		},
		{
			name:                "Valid bidder respone, with bidder params",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"123","ext": {"bidtype": "video"}}]}]}`),
			bidderResponse: map[string]any{
				"cur": "USD",
				seatBidKey: []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"bidtype": "video",
								},
							},
						},
					},
				},
			},
			responseParams: map[string]bidderparams.BidderParamMapper{
				"currency": {
					Location: "cur",
				},
				"bidtype": {
					Location: "seatbid.#.bid.#.ext.bidtype",
				},
			},
			expectedError: nil,
			expectedResponse: map[string]any{
				"Currency": "USD",
				"Bids": []any{
					map[string]any{
						"Bid": map[string]any{
							"id": "123",
							"ext": map[string]any{
								"bidtype": "video",
							},
						},
						"BidType": "video",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rb := &responseBuilder{
				bidderResponse: tc.bidderResponse,
				responseParams: tc.responseParams,
			}
			err := rb.setPrebidBidderResponse(tc.bidderResponseBytes)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResponse, rb.adapterRespone)
		})
	}
}
