package ortbbidder

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/stretchr/testify/assert"
)

func TestMakeBids(t *testing.T) {
	err := InitBidderParamsConfig("./ortbbiddertest/bidder-params", "./ortbbiddertest/bidder-response-params")
	if err != nil {
		t.Fatalf("Failed to initalise bidder config")
	}

	t.Cleanup(func() {
		g_bidderParamsConfig = nil
	})
	tests := []struct {
		name             string
		responseData     *adapters.ResponseData
		expectedResponse *adapters.BidderResponse
		request          *openrtb2.BidRequest
		requestData      *adapters.RequestData
		expectedErrors   []error
		setup            func() adapter
	}{
		{
			name:             "response data is nil",
			expectedResponse: nil,
			responseData:     nil,
			setup: func() adapter {
				return adapter{}
			},
		},
		{
			name:         "no content response data",
			responseData: &adapters.ResponseData{StatusCode: http.StatusNoContent},
			setup: func() adapter {
				return adapter{}
			},
			expectedResponse: nil,
			expectedErrors:   nil,
		},
		{
			name:         "status bad request in response data",
			responseData: &adapters.ResponseData{StatusCode: http.StatusBadRequest},
			setup: func() adapter {
				return adapter{}
			},
			expectedResponse: nil,
			expectedErrors: []error{&errortypes.BadInput{
				Message: "Unexpected status code: 400. Run with request.debug = 1 for more info",
			}},
		},
		{
			name: "status too many requests in response data",

			responseData: &adapters.ResponseData{StatusCode: http.StatusTooManyRequests},
			setup: func() adapter {
				return adapter{}
			},
			expectedResponse: nil,
			expectedErrors: []error{&errortypes.BadServerResponse{
				Message: "Unexpected status code: 429. Run with request.debug = 1 for more info",
			}},
		},
		{
			name: "malformed response data body",

			responseData: &adapters.ResponseData{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"id":1,"seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","mtype":2}]}]`),
			},
			setup: func() adapter {
				return adapter{}
			},
			expectedResponse: nil,
			expectedErrors: []error{&errortypes.FailedToUnmarshal{
				Message: "expect }, but found \x00",
			}},
		},
		{
			name: "error parsing bidder response",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: nil,
			expectedErrors:   []error{&errortypes.FailedToUnmarshal{Message: "unexpected value type: 0"}},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "invalid seatbid in response",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":1, "seatbid":"invalid"}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: nil,
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "invalid seatbid array found in response, seatbids:[invalid]"}},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "invalid seat in seatbid array",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":1, "seatbid":["invalid"]}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: nil,
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "invalid seatbid found in seatbid array, seatbid:[[invalid]]"}},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "invalid bid arrays in seatbid array",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":1,"seatbid":[{"bid":"invalid"}]}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: nil,
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "invalid bid array found in seatbid, bids:[invalid]"}},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "invalid bid in bids array",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":1,"seatbid":[{"bid":["invalid"]}]}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: nil,
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "invalid bid found in bids array, bid:[[invalid]]"}},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "failure converting to adapter response",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":1,"seatbid":[{"bid":[{"id": 1, "mtype":2}]}]}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: nil,
			expectedErrors:   []error{&errortypes.FailedToUnmarshal{Message: "cannot unmarshal ID: expects \" or n, but found 1"}},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "valid response - no bidder params",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":"1","cur":"USD","seatbid":[{"bid":[{"id":"1","mtype":2}]}]}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: &adapters.BidderResponse{
				Currency: "USD",
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:    "1",
							MType: 2,
						},
					},
				},
			},
			expectedErrors: nil,
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "valid response - bidder params present",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":"1","cur":"USD","seatbid":[{"bid":[{"id":"1","ext":{"mtype":"video"}}]}]}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: &adapters.BidderResponse{
				Currency: "USD",
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:  "1",
							Ext: json.RawMessage(`{"mtype":"video"}`),
						},
						BidType: "video",
					},
				},
			},
			expectedErrors: nil,
			setup: func() adapter {
				return adapter{
					adapterInfo: adapterInfo{
						bidderName: "owortb_testbidder",
					},
					bidderParamsConfig: g_bidderParamsConfig,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := tt.setup()
			got, err := adapter.MakeBids(tt.request, tt.requestData, tt.responseData)
			assert.Equal(t, tt.expectedResponse, got, "response mismatch")
			assert.Equal(t, tt.expectedErrors, err, "error mismatch")
		})
	}
}
