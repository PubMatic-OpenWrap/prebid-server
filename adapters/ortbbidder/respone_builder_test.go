package ortbbidder

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/stretchr/testify/assert"
)

func TestNewResponseBuilder(t *testing.T) {
	testCases := []struct {
		name           string
		responseParams map[string]bidderparams.BidderParamMapper
		expected       *responseBuilder
	}{
		{
			name:           "With nil responseParams",
			responseParams: nil,
			expected: &responseBuilder{
				responseParams: make(map[string]bidderparams.BidderParamMapper),
			},
		},
		{
			name: "With non-nil responseParams",
			responseParams: map[string]bidderparams.BidderParamMapper{
				"test": {},
			},
			expected: &responseBuilder{
				responseParams: map[string]bidderparams.BidderParamMapper{
					"test": {},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := newResponseBuilder(tc.responseParams)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseResponse(t *testing.T) {
	testCases := []struct {
		name          string
		responseBytes json.RawMessage
		expectedError error
	}{
		{
			name:          "Valid response",
			responseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1", "ext":{"mtype":"video"}}]}]}`),
			expectedError: nil,
		},
		{
			name:          "Invalid response",
			responseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1", "ext":{"mtype":"video"}}]}`), // missing closing bracket
			expectedError: errors.New("expect ] in the end, but found \x00"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rb := &responseBuilder{}
			err := rb.parseResponse(tc.responseBytes)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertToAdapterResponse(t *testing.T) {
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
			actualResponse, err := rb.convertToAdapterResponse()
			assert.Equal(t, tc.expectedError, err, "error mismatch")
			assert.Equal(t, tc.expectedResponse, actualResponse, "response mismatch")
		})
	}
}

func TestBuildResponse(t *testing.T) {
	testCases := []struct {
		name             string
		bidderResponse   map[string]any
		responseParams   map[string]bidderparams.BidderParamMapper
		expectedError    error
		expectedResponse map[string]any
	}{
		{
			name: "Invalid seatbid object",
			bidderResponse: map[string]any{
				"cur":      "USD",
				seatBidKey: map[string]any{},
			},
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Path: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid seatbid array found in response, seatbids:[map[]]"},
		},
		{
			name: "Invalid seatbid object",
			bidderResponse: map[string]any{
				"cur": "USD",
				seatBidKey: []any{
					"invalid",
				},
			},
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Path: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid seatbid found in seatbid array, seatbid:[[invalid]]"},
		},
		{
			name: "Invalid bid object in seatbid",
			bidderResponse: map[string]any{
				"cur": "USD",
				seatBidKey: []any{
					map[string]any{
						"bid": "invalid",
					},
				},
			},
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Path: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid bid array found in seatbid, bids:[invalid]"},
		},
		{
			name: "Invalid bid object in bids",
			bidderResponse: map[string]any{
				"cur": "USD",
				seatBidKey: []any{
					map[string]any{
						"bid": []any{
							"invalid",
						},
					},
				},
			},
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Path: "cur",
				},
			},
			expectedError: &errortypes.BadServerResponse{Message: "invalid bid found in bids array, bid:[[invalid]]"},
		},
		{
			name: "Valid bidder respone, no bidder params",
			bidderResponse: map[string]any{
				"cur": "USD",
				seatBidKey: []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id":    "123",
								"mtype": 2,
							},
						},
					},
				},
			},
			responseParams: map[string]bidderparams.BidderParamMapper{},
			expectedError:  nil,
			expectedResponse: map[string]any{
				"Currency": "USD",
				"Bids": []any{
					map[string]any{
						"Bid": map[string]any{
							"id":    "123",
							"mtype": 2,
						},
					},
				},
			},
		},
		{
			name: "Valid bidder respone, with bidder params",
			bidderResponse: map[string]any{
				"cur": "USD",
				seatBidKey: []any{
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
			responseParams: map[string]bidderparams.BidderParamMapper{
				"currency": {
					Path: "cur",
				},
				"mtype": {
					Path: "seatbid.#.bid.#.ext.mtype",
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
								"mtype": "video",
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
			err := rb.buildResponse()
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResponse, rb.adapterRespone)
		})
	}
}
