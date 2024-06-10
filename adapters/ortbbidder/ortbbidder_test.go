package ortbbidder

// import (
// 	"encoding/json"
// 	"net/http"
// 	"testing"

// 	"github.com/prebid/openrtb/v20/openrtb2"
// 	"github.com/prebid/prebid-server/v2/adapters"
// 	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
// 	"github.com/prebid/prebid-server/v2/errortypes"
// 	"github.com/stretchr/testify/assert"
// )

// func TestInitBidderParamsConfig(t *testing.T) {
// 	tests := []struct {
// 		name                  string
// 		requestParamsDirPath  string
// 		responseParamsDirPath string
// 		wantErr               bool
// 	}{
// 		{
// 			name:                  "test_InitBiddersConfigMap_success",
// 			requestParamsDirPath:  "../../static/bidder-params/",
// 			responseParamsDirPath: "../../static/bidder-response-params/",
// 			wantErr:               false,
// 		},
// 		{
// 			name:                 "test_InitBiddersConfigMap_failure",
// 			requestParamsDirPath: "/invalid_directory/",
// 			wantErr:              true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := InitBidderParamsConfig(tt.requestParamsDirPath, tt.responseParamsDirPath)
// 			assert.Equal(t, err != nil, tt.wantErr, "mismatched error")
// 		})
// 	}
// }

// func TestIsORTBBidder(t *testing.T) {
// 	type args struct {
// 		bidderName string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{
// 		{
// 			name: "ortb_bidder",
// 			args: args{
// 				bidderName: "owortb_magnite",
// 			},
// 			want: true,
// 		},
// 		{
// 			name: "non_ortb_bidder",
// 			args: args{
// 				bidderName: "magnite",
// 			},
// 			want: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := isORTBBidder(tt.args.bidderName)
// 			assert.Equal(t, tt.want, got, "mismatched output of isORTBBidder")
// 		})
// 	}
// }

// func TestMakeBids(t *testing.T) {

// 	err := InitBidderParamsConfig("./ortbbiddertest/bidder-params", "./ortbbiddertest/bidder-response-params")
// 	if err != nil {
// 		t.Fatalf("Failed to initalise bidder config")
// 	}

// 	t.Cleanup(func() {
// 		g_bidderParamsConfig = nil
// 	})

// 	type args struct {
// 		request      *openrtb2.BidRequest
// 		requestData  *adapters.RequestData
// 		responseData *adapters.ResponseData
// 		setup        func() adapter
// 	}
// 	type want struct {
// 		response *adapters.BidderResponse
// 		errors   []error
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want want
// 	}{
// 		{
// 			name: "response data is nil",
// 			args: args{
// 				responseData: nil,
// 				setup: func() adapter {
// 					return adapter{}
// 				},
// 			},
// 			want: want{
// 				response: nil,
// 				errors:   nil,
// 			},
// 		},
// 		{
// 			name: "no content response data",
// 			args: args{
// 				responseData: &adapters.ResponseData{StatusCode: http.StatusNoContent},
// 				setup: func() adapter {
// 					return adapter{}
// 				},
// 			},
// 			want: want{
// 				response: nil,
// 				errors:   nil,
// 			},
// 		},
// 		{
// 			name: "status bad request in response data",
// 			args: args{
// 				responseData: &adapters.ResponseData{StatusCode: http.StatusBadRequest},
// 				setup: func() adapter {
// 					return adapter{}
// 				},
// 			},
// 			want: want{
// 				response: nil,
// 				errors: []error{&errortypes.BadInput{
// 					Message: "Unexpected status code: 400. Run with request.debug = 1 for more info",
// 				}},
// 			},
// 		},
// 		// {
// 		// 	name: "status too many requests in response data",
// 		// 	args: args{
// 		// 		responseData: &adapters.ResponseData{StatusCode: http.StatusTooManyRequests},
// 		// 		setup: func() adapter {
// 		// 			return adapter{}
// 		// 		},
// 		// 	},
// 		// 	want: want{
// 		// 		response: nil,
// 		// 		errors: []error{&errortypes.BadServerResponse{
// 		// 			Message: "Unexpected status code: 429. Run with request.debug = 1 for more info",
// 		// 		}},
// 		// 	},
// 		// },
// 		// {
// 		// 	name: "malformed response data body",
// 		// 	args: args{
// 		// 		responseData: &adapters.ResponseData{
// 		// 			StatusCode: http.StatusOK,
// 		// 			Body:       []byte(`{"id":1,"seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","mtype":2}]}]`),
// 		// 		},
// 		// 		setup: func() adapter {
// 		// 			return adapter{}
// 		// 		},
// 		// 	},
// 		// 	want: want{
// 		// 		response: nil,
// 		// 		errors: []error{&errortypes.FailedToUnmarshal{
// 		// 			Message: "expect }, but found \x00",
// 		// 		}},
// 		// 	},
// 		// },
// 		// {
// 		// 	name: "valid response no bidder params",
// 		// 	args: args{
// 		// 		responseData: &adapters.ResponseData{
// 		// 			StatusCode: http.StatusOK,
// 		// 			Body:       []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","mtype":2}]}]}`),
// 		// 		},
// 		// 		setup: func() adapter {
// 		// 			return adapter{
// 		// 				bidderParamsConfig: &bidderparams.BidderConfig{},
// 		// 			}
// 		// 		},
// 		// 	},
// 		// 	want: want{
// 		// 		response: &adapters.BidderResponse{
// 		// 			Currency: "USD",
// 		// 			Bids: []*adapters.TypedBid{
// 		// 				{
// 		// 					Bid: &openrtb2.Bid{
// 		// 						ID:    "bid-1",
// 		// 						MType: 2,
// 		// 					},
// 		// 				},
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 		// {
// 		// 	name: "valid response bidder params present - get from ortb field",
// 		// 	args: args{
// 		// 		responseData: &adapters.ResponseData{
// 		// 			StatusCode: http.StatusOK,
// 		// 			Body:       []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","mtype":2}]}]}`),
// 		// 		},
// 		// 		setup: func() adapter {
// 		// 			return adapter{
// 		// 				bidderParamsConfig: g_bidderParamsConfig,
// 		// 				adapterInfo: adapterInfo{
// 		// 					bidderName: "owortb_testbidder",
// 		// 				},
// 		// 			}
// 		// 		},
// 		// 	},
// 		// 	want: want{
// 		// 		response: &adapters.BidderResponse{
// 		// 			Currency: "USD",
// 		// 			Bids: []*adapters.TypedBid{
// 		// 				{
// 		// 					Bid: &openrtb2.Bid{
// 		// 						ID:    "bid-1",
// 		// 						MType: 2,
// 		// 					},
// 		// 					BidType: "video",
// 		// 				},
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 		// {
// 		// 	name: "valid response bidder params present - get from bidder param location",
// 		// 	args: args{
// 		// 		responseData: &adapters.ResponseData{
// 		// 			StatusCode: http.StatusOK,
// 		// 			Body:       []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","ext":{"mtype":"video"}}]}]}`),
// 		// 		},
// 		// 		setup: func() adapter {
// 		// 			return adapter{
// 		// 				bidderParamsConfig: g_bidderParamsConfig,
// 		// 				adapterInfo: adapterInfo{
// 		// 					bidderName: "owortb_testbidder",
// 		// 				},
// 		// 				paramProcessor: NewParamProcessor(),
// 		// 			}
// 		// 		},
// 		// 	},
// 		// 	want: want{
// 		// 		response: &adapters.BidderResponse{
// 		// 			Currency: "USD",
// 		// 			Bids: []*adapters.TypedBid{
// 		// 				{
// 		// 					Bid: &openrtb2.Bid{
// 		// 						ID:  "bid-1",
// 		// 						Ext: json.RawMessage(`{"mtype":"video"}`),
// 		// 					},
// 		// 					BidType: "video",
// 		// 				},
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			adapter := tt.args.setup()
// 			response, errs := adapter.MakeBids(tt.args.request, tt.args.requestData, tt.args.responseData)
// 			assert.Equalf(t, tt.want.response, response, "mismatched response")
// 			assert.Equal(t, errs, tt.want.errors, "mismatched errors")
// 			// for i, err := range errs {
// 			// 	assert.Contains(t, err.Error(), tt.want.errors[i].Error(), "Unexpected error message")
// 			// }
// 		})
// 	}
// }

// func TestSetResponseParams(t *testing.T) {
// 	// Initialize the adapter
// 	a := &adapter{}

// 	// Define the test cases
// 	tests := []struct {
// 		name           string
// 		body           json.RawMessage
// 		responseParams map[string]bidderparams.BidderParamMapper
// 		wantErr        string
// 		wantBytes      []byte
// 	}{
// 		{
// 			name:           "invalid JSON",
// 			body:           json.RawMessage(`{"cur": "USD", "seatbid": [{"bid": [}]}`), // missing closing brackets
// 			responseParams: map[string]bidderparams.BidderParamMapper{},
// 			wantErr:        "unexpected value type: 0",
// 			wantBytes:      nil,
// 		},
// 		{
// 			name:           "seatbid type assertion failure",
// 			body:           json.RawMessage(`{"cur": "USD", "seatbid": "invalid"}`), // seatbid is not an array
// 			responseParams: map[string]bidderparams.BidderParamMapper{},
// 			wantErr:        "error:[invalid_seatbid_found_in_responsebody], seatbid:[invalid]",
// 			wantBytes:      nil,
// 		},
// 		{
// 			name:           "seatbid object invalid ",
// 			body:           json.RawMessage(`{"cur": "USD", "seatbid": ["invalid"]}`), // seatbid is not an array
// 			responseParams: map[string]bidderparams.BidderParamMapper{},
// 			wantErr:        "error:[invalid_seatbid_found_in_seatbids_list], seatbid:[[invalid]]",
// 			wantBytes:      nil,
// 		},
// 		{
// 			name:           "bids type assertion failure",
// 			body:           json.RawMessage(`{"cur": "USD", "seatbid": [{"bid": "invalid"}]}`), // bid is not an object
// 			responseParams: map[string]bidderparams.BidderParamMapper{},
// 			wantErr:        "error:[invalid_bid_found_in_seatbid], bid:[invalid]",
// 			wantBytes:      nil,
// 		},
// 		{
// 			name:           "bids type assertion failure",
// 			body:           json.RawMessage(`{"cur": "USD", "seatbid": [{"bid": ["invalid"]}]}`), // bid is not an object
// 			responseParams: map[string]bidderparams.BidderParamMapper{},
// 			wantErr:        "error:[invalid_bid_found_in_bids_list], bid:[[invalid]]",
// 			wantBytes:      nil,
// 		},
// 		{
// 			name:           "valid response, no bidder param",
// 			body:           json.RawMessage(`{"cur": "USD", "seatbid": [{"bid": [{"id":"123"}]}]}`), // bid is not an object
// 			responseParams: map[string]bidderparams.BidderParamMapper{},
// 			wantErr:        "",
// 			wantBytes:      []byte(`{"Bids":[{"Bid":{"id":"123"}}],"Currency":"USD"}`),
// 		},
// 		// Add more test cases as needed
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotBytes, err := a.setResponseParams(tt.body, tt.responseParams)
// 			if tt.wantErr != "" {
// 				assert.EqualError(t, err, tt.wantErr)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.Equal(t, tt.wantBytes, gotBytes)
// 		})
// 	}
// }

// func BenchmarkSetResponseParams(b *testing.B) {
// 	// Initialize an instance of adapter
// 	adapter := adapter{}

// 	// Prepare the input parameters
// 	bidderResponseBody := []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1", "ext":{"mtype":"video"}}]}]}`) // replace with actual JSON response
// 	responseParams := map[string]bidderparams.BidderParamMapper{
// 		"mtype": {
// 			Path: "seatbid.#.bid.#.ext.mtype",
// 		},
// 		"currency": {
// 			Path: "cur",
// 		},
// 	}
// 	processor := &resolver.ParamResolver{}
// 	// Run the benchmark
// 	for i := 0; i < b.N; i++ {
// 		adapter.setResponseParams(bidderResponseBody, responseParams, processor)
// 	}
// }

// func TestSetResponseParams(t *testing.T) {
// 	// Initialize an instance of adapter
// 	adapter := adapter{
// 		processor: resolver.NewParamProcessor(),
// 	}

// 	// Prepare the input parameters
// 	bidderResponseBody := []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1", "ext":{"mtype":"video"}}]}]}`) // replace with actual JSON response
// 	responseParams := map[string]bidderparams.BidderParamMapper{
// 		"mtype": {
// 			Path: "seatbid.#.bid.#.ext.mtype",
// 		},
// 		"currency": {
// 			Path: "cur",
// 		},
// 	}

// 	_, err := adapter.setResponseParams(bidderResponseBody, responseParams)
// 	fmt.Println(err)
// }

// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	  105494	     11169 ns/op	    4599 B/op	      94 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.983s

// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	  104610	     11217 ns/op	    4599 B/op	      94 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.572s

// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	   98408	     11577 ns/op	    4599 B/op	      94 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.542s

// Function arguments
// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	  128482	      9009 ns/op	    3926 B/op	      85 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.990s

// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	  127963	      9854 ns/op	    3926 B/op	      85 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.636s

// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	  123307	     10853 ns/op	    3926 B/op	      85 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.730s

// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	  126006	      9230 ns/op	    3926 B/op	      85 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.545s

// goos: darwin
// goarch: amd64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkSetResponseParams-12    	  130268	      8953 ns/op	    3926 B/op	      85 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/v2/adapters/ortbbidder	1.990s
