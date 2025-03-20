package ortbbidder

import (
	"errors"
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/resolver"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v3/errortypes"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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
			expectedError: &errortypes.BadServerResponse{
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
			expectedError:    &errortypes.BadServerResponse{Message: "chan int is unsupported type"},
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
		isDebugEnabled      bool
		responseParams      map[string]bidderparams.BidderParamMapper
		expectedError       []error
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
			expectedError: []error{&errortypes.BadServerResponse{Message: "expect ] in the end, but found \x00"}},
		},
		{
			name:                "Invalid seatbid object in response",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":"invalid"}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: []error{&errortypes.BadServerResponse{Message: "invalid seatbid array found in response, seatbids:[invalid]"}},
		},
		{
			name:                "Invalid seatbid is seatbid arrays",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":["invalid"]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: []error{&errortypes.BadServerResponse{Message: "invalid seatbid found in seatbid array, seatbid:[invalid]"}},
		},
		{
			name:                "Invalid bid in seatbid",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":"invalid"}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: []error{&errortypes.BadServerResponse{Message: "invalid bid array found in seatbid, bids:[invalid]"}},
		},
		{
			name:                "Invalid bid in bids array",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":["invalid"]}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"Currency": {
					Location: "cur",
				},
			},
			expectedError: []error{&errortypes.BadServerResponse{Message: "invalid bid found in bids array, bid:[invalid]"}},
		},
		{
			name:                "Valid bidder respone, no bidder params, debug is disabled in request",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"123"}]}]}`),
			responseParams:      map[string]bidderparams.BidderParamMapper{},
			expectedError:       []error{util.NewWarning("Potential issue encountered while setting the response parameter [bidType]")},
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
			name:                "Valid bidder respone, no bidder params, debug is enabled in request",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"123"}]}]}`),
			responseParams:      map[string]bidderparams.BidderParamMapper{},
			isDebugEnabled:      true,
			expectedError:       []error{util.NewWarning("invalid value sent by bidder at [bid.impid] for [bid.ext.prebid.type]")},
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
			name:                "Valid bidder respone, with single bidder param - bidType",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"123","ext": {"bidtype": "video"}}]}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"currency": {
					Location: "cur",
				},
				"bidType": {
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
						"BidType": openrtb_ext.BidType("video"),
					},
				},
			},
		},
		{
			name:                "failed to set the adapter-response level param - fledgeConfig",
			bidderResponseBytes: []byte(`{"fledge":"","id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"123","ext": {"bidtype": "video"}}]}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"currency": {
					Location: "cur",
				},
				"fledgeAuctionConfig": {
					Location: "fledge",
				},
			},
			expectedError: []error{
				util.NewWarning("Potential issue encountered while setting the response parameter [fledgeAuctionConfig]"),
				util.NewWarning("Potential issue encountered while setting the response parameter [bidType]"),
			},
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
					},
				},
			},
		},
		{
			name: "Valid bidder respone, with multiple bidder params",
			bidderResponseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","ext":{"dp":2},"bid":[{"id":"123","cat":["music"],"ext":{"bidtype":"video","advertiserId":"5"` +
				`,"networkId":5,"duration":10,"meta_object":{"advertiserDomains":["xyz.com"],"mediaType":"video"}}}]}]}`),
			responseParams: map[string]bidderparams.BidderParamMapper{
				"currency":            {Location: "cur"},
				"bidType":             {Location: "seatbid.#.bid.#.ext.bidtype"},
				"bidDealPriority":     {Location: "seatbid.#.ext.dp"},
				"bidVideoDuration":    {Location: "seatbid.#.bid.#.ext.duration"},
				"bidMeta":             {Location: "seatbid.#.bid.#.ext.meta_object"},
				"bidMetaAdvertiserId": {Location: "seatbid.#.bid.#.ext.advertiserId"},
				"bidMetaNetworkId":    {Location: "seatbid.#.bid.#.ext.networkId"},
			},
			expectedError: []error{util.NewWarning("Potential issue encountered while setting the response parameter [bidMetaAdvertiserId]")},
			expectedResponse: map[string]any{
				"Currency": "USD",
				"Bids": []any{
					map[string]any{
						"Bid": map[string]any{
							"id":  "123",
							"cat": []any{"music"},
							"ext": map[string]any{
								"bidtype":      "video",
								"advertiserId": "5",
								"networkId":    5.0,
								"duration":     10.0,
								"meta_object": map[string]any{
									"advertiserDomains": []any{"xyz.com"},
									"mediaType":         "video",
								},
							},
						},
						"BidType": openrtb_ext.BidType("video"),
						"BidVideo": map[string]any{
							"primary_category": "music",
							"duration":         int64(10),
						},
						"DealPriority": 2,
						"BidMeta": map[string]any{
							"advertiserDomains": []any{"xyz.com"},
							"mediaType":         "video",
							"networkId":         int(5),
						},
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
				isDebugEnabled: tc.isDebugEnabled,
			}
			err := rb.setPrebidBidderResponse(tc.bidderResponseBytes)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResponse, rb.adapterRespone, "mismatched adapterRespone")
		})
	}
}

// TestTypedBidFields notifies us of any changes in the adapters.TypedBid struct.
// If a new field is added in adapters.TypedBid, then add the support to resolve the new field and update the test case.
// If the data type of an existing field changes then update the resolver of the respective field.
func TestTypedBidFields(t *testing.T) {
	expectedFields := map[string]reflect.Type{
		"Bid":          reflect.TypeOf(&openrtb2.Bid{}),
		"BidMeta":      reflect.TypeOf(&openrtb_ext.ExtBidPrebidMeta{}),
		"BidType":      reflect.TypeOf(openrtb_ext.BidTypeBanner),
		"BidVideo":     reflect.TypeOf(&openrtb_ext.ExtBidPrebidVideo{}),
		"BidTargets":   reflect.TypeOf(map[string]string{}),
		"DealPriority": reflect.TypeOf(0),
		"Seat":         reflect.TypeOf(openrtb_ext.BidderName("")),
	}

	structType := reflect.TypeOf(adapters.TypedBid{})
	err := resolver.ValidateStructFields(expectedFields, structType)
	if err != nil {
		t.Error(err)
	}
}

// TestBidderResponseFields notifies us of any changes in the adapters.BidderResponse struct.
// If a new field is added in adapters.BidderResponse, then add the support to resolve the new field and update the test case.
// If the data type of an existing field changes then update the resolver of the respective field.
func TestBidderResponseFields(t *testing.T) {
	expectedFields := map[string]reflect.Type{
		"Currency":             reflect.TypeOf(""),
		"Bids":                 reflect.TypeOf([]*adapters.TypedBid{nil}),
		"FledgeAuctionConfigs": reflect.TypeOf([]*openrtb_ext.FledgeAuctionConfig{}),
		"FastXMLMetrics":       reflect.TypeOf(&openrtb_ext.FastXMLMetrics{}),
	}
	structType := reflect.TypeOf(adapters.BidderResponse{})
	err := resolver.ValidateStructFields(expectedFields, structType)
	if err != nil {
		t.Error(err)
	}
}

func TestCollectWarningMessages(t *testing.T) {
	type args struct {
		errs           []error
		resolverErrors []error
		parameter      string
		isDebugEnabled bool
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "No resolver errors",
			args: args{
				errs:           []error{},
				resolverErrors: []error{},
				parameter:      "param1",
				isDebugEnabled: false,
			},
			want: []error{},
		},
		{
			name: "Resolver errors with warnings and debugging enabled",
			args: args{
				errs: []error{},
				resolverErrors: []error{
					resolver.NewValidationFailedError("Warning 1"),
					resolver.NewValidationFailedError("Warning 2"),
				},
				parameter:      "param2",
				isDebugEnabled: true,
			},
			want: []error{
				util.NewWarning("Warning 1"),
				util.NewWarning("Warning 2"),
			},
		},
		{
			name: "Resolver errors with warnings and debugging disabled",
			args: args{
				errs: []error{},
				resolverErrors: []error{
					resolver.NewValidationFailedError("Warning 1"),
				},
				parameter:      "param3",
				isDebugEnabled: false,
			},
			want: []error{
				util.NewWarning("Potential issue encountered while setting the response parameter [param3]"),
			},
		},
		{
			name: "Resolver errors without warnings",
			args: args{
				errs: []error{},
				resolverErrors: []error{
					errors.New("Non-warning error"),
				},
				parameter:      "param4",
				isDebugEnabled: false,
			},
			want: []error{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectWarningMessages(tt.args.errs, tt.args.resolverErrors, tt.args.parameter, tt.args.isDebugEnabled)
			assert.Equal(t, tt.want, got)
		})
	}
}
