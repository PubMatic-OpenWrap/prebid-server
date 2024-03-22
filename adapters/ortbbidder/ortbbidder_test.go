package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/adapters/adapterstest"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestPrepareRequestData(t *testing.T) {
	type args struct {
		request  *openrtb2.BidRequest
		endpoint string
	}
	type want struct {
		requestData *adapters.RequestData
		err         error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Valid_Request",
			args: args{
				request: &openrtb2.BidRequest{
					ID:  "123",
					Imp: []openrtb2.Imp{{ID: "imp1"}},
				},
				endpoint: "https://example.com",
			},
			want: want{
				requestData: &adapters.RequestData{
					Method: http.MethodPost,
					Uri:    "https://example.com",
					Body:   []byte(`{"id":"123","imp":[{"id":"imp1"}]}`),
				},
				err: nil,
			},
		},
		{
			name: "Nil_Request",
			args: args{
				request:  nil,
				endpoint: "https://example.com",
			},
			want: want{
				requestData: nil,
				err:         fmt.Errorf("found nil request"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqData, err := prepareRequestData(tt.args.request, tt.args.endpoint)
			assert.Equal(t, reqData, tt.want.requestData, "mismatched requestData")
			assert.Equal(t, err, tt.want.err, "mismatched error")
		})
	}
}

func TestBuilder(t *testing.T) {
	originalAdapter := ortbAdapter
	defer func() {
		ortbAdapter = originalAdapter
	}()
	type args struct {
		bidderName openrtb_ext.BidderName
		config     config.Adapter
		server     config.Server
	}
	tests := []struct {
		name    string
		args    args
		want    adapters.Bidder
		wantErr error
		setup   func()
	}{
		{
			name:    "ortbBidder_is_nil",
			args:    args{},
			want:    nil,
			wantErr: fmt.Errorf("oRTB bidder is not initialised"),
			setup: func() {
				ortbAdapter = nil
			},
		},
		{
			name:    "ortbBidder_is_not_nil",
			args:    args{},
			want:    &oRTBAdapter{},
			wantErr: nil,
			setup: func() {
				ortbAdapter = &oRTBAdapter{}
			},
		},
	}
	for _, tt := range tests {
		tt.setup()
		got, err := Builder(tt.args.bidderName, tt.args.config, tt.args.server)
		assert.Equal(t, got, tt.want, "mismatched adapter for %v", tt.name)
		assert.Equal(t, err, tt.wantErr, "mismatched error for %v", tt.name)
	}
}

func TestMakeRequests(t *testing.T) {
	type args struct {
		request     *openrtb2.BidRequest
		requestInfo *adapters.ExtraRequestInfo
	}
	type want struct {
		requestData []*adapters.RequestData
		errors      []error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "request_is_nil",
			args: args{},
			want: want{
				errors: []error{fmt.Errorf("Found either nil request or nil requestInfo")},
			},
		},
		{
			name: "requestInfo_is_nil",
			args: args{},
			want: want{
				errors: []error{fmt.Errorf("Found either nil request or nil requestInfo")},
			},
		},
		{
			name: "bidderInfo_absent",
			args: args{
				request: &openrtb2.BidRequest{},
				requestInfo: &adapters.ExtraRequestInfo{
					BidderCoreName: "xyz",
				},
			},
			want: want{
				errors: []error{fmt.Errorf("bidder-info not found for bidder-[xyz]")},
			},
		},
		{
			name: "multi_requestmode_to_form_requestdata",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "reqid",
					Imp: []openrtb2.Imp{
						{ID: "imp1", TagID: "tag1"},
						{ID: "imp2", TagID: "tag2"},
					},
				},
				requestInfo: &adapters.ExtraRequestInfo{
					BidderCoreName: openrtb_ext.BidderName("ortb_test_multi_requestmode"),
				},
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp1","tagid":"tag1"},{"id":"imp2","tagid":"tag2"}]}`),
					},
				},
			},
		},
		{
			name: "single_requestmode_to_form_requestdata",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "reqid",
					Imp: []openrtb2.Imp{
						{ID: "imp1", TagID: "tag1"},
						{ID: "imp2", TagID: "tag2"},
					},
				},
				requestInfo: &adapters.ExtraRequestInfo{
					BidderCoreName: openrtb_ext.BidderName("ortb_test_single_requestmode"),
				},
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp1","tagid":"tag1"}]}`),
					},
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp2","tagid":"tag2"}]}`),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &oRTBAdapter{
				BidderInfo: getBidderInfos(),
			}
			requestData, errors := adapter.MakeRequests(tt.args.request, tt.args.requestInfo)
			assert.Equalf(t, tt.want.requestData, requestData, "mismatched requestData")
			assert.Equalf(t, tt.want.errors, errors, "mismatched errors")
		})
	}
}

func TestMakeBids(t *testing.T) {
	type args struct {
		request      *openrtb2.BidRequest
		requestData  *adapters.RequestData
		responseData *adapters.ResponseData
	}
	type want struct {
		response *adapters.BidderResponse
		errors   []error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "responseData_is_nil",
			args: args{
				responseData: nil,
			},
			want: want{
				response: nil,
				errors:   nil,
			},
		},
		{
			name: "StatusNoContent_in_responseData",
			args: args{
				responseData: &adapters.ResponseData{StatusCode: http.StatusNoContent},
			},
			want: want{
				response: nil,
				errors:   nil,
			},
		},
		{
			name: "StatusBadRequest_in_responseData",
			args: args{
				responseData: &adapters.ResponseData{StatusCode: http.StatusBadRequest},
			},
			want: want{
				response: nil,
				errors: []error{&errortypes.BadInput{
					Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", http.StatusBadRequest),
				}},
			},
		},
		{
			name: "getMediaTypeForBid_returns_error",
			args: args{
				responseData: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id":"bid-resp-id","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","mtype":2},{"id":"bid-2","mtype":5}]}]}`),
				},
			},
			want: want{
				response: &adapters.BidderResponse{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "bid-1",
								MType: 2,
							},
							BidType: "video",
						},
					},
				},
				errors: []error{fmt.Errorf("Failed to parse bid mType for bidID \"bid-2\"")},
			},
		},
		{
			name: "valid_response",
			args: args{
				responseData: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id":"bid-resp-id","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","mtype":2}]}]}`),
				},
			},
			want: want{
				response: &adapters.BidderResponse{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "bid-1",
								MType: 2,
							},
							BidType: "video",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &oRTBAdapter{
				BidderInfo: getBidderInfos(),
			}
			response, errs := adapter.MakeBids(tt.args.request, tt.args.requestData, tt.args.responseData)
			assert.Equalf(t, tt.want.response, response, "mismatched response")
			assert.Equalf(t, tt.want.errors, errs, "mismatched errors")
		})
	}
}

func TestGetMediaTypeForBid(t *testing.T) {
	type args struct {
		bid openrtb2.Bid
	}
	type want struct {
		bidType openrtb_ext.BidType
		err     error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Valid_Banner_Bid",
			args: args{
				bid: openrtb2.Bid{ID: "1", MType: openrtb2.MarkupBanner},
			},
			want: want{
				bidType: openrtb_ext.BidTypeBanner,
				err:     nil,
			},
		},
		{
			name: "Valid_Video_Bid",
			args: args{
				bid: openrtb2.Bid{ID: "2", MType: openrtb2.MarkupVideo},
			},
			want: want{
				bidType: openrtb_ext.BidTypeVideo,
				err:     nil,
			},
		},
		{
			name: "Valid_Audio_Bid",
			args: args{
				bid: openrtb2.Bid{ID: "3", MType: openrtb2.MarkupAudio},
			},
			want: want{
				bidType: openrtb_ext.BidTypeAudio,
				err:     nil,
			},
		},
		{
			name: "Valid_Native_Bid",
			args: args{
				bid: openrtb2.Bid{ID: "4", MType: openrtb2.MarkupNative},
			},
			want: want{
				bidType: openrtb_ext.BidTypeNative,
				err:     nil,
			},
		},
		{
			name: "Invalid_Bid_Type",
			args: args{
				bid: openrtb2.Bid{ID: "5", MType: 123},
			},
			want: want{
				bidType: "",
				err:     fmt.Errorf("Failed to parse bid mType for bidID \"5\""),
			},
		},
		{
			name: "bidExt.prebid.type_has_high_priority",
			args: args{
				bid: openrtb2.Bid{ID: "5", MType: openrtb2.MarkupVideo, Ext: json.RawMessage(`{"prebid":{"type":"video"}}`)},
			},
			want: want{
				bidType: "video",
				err:     nil,
			},
		},
		{
			name: "bidExt.prebid_is_missing_fallback_to_bid.mtype",
			args: args{
				bid: openrtb2.Bid{ID: "5", MType: openrtb2.MarkupVideo, Ext: json.RawMessage(`{}`)},
			},
			want: want{
				bidType: "video",
				err:     nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bidType, err := getMediaTypeForBid(tt.args.bid)
			assert.Equal(t, tt.want.bidType, bidType, "mismatched bidType")
			assert.Equal(t, tt.want.err, err, "mismatched error")
		})
	}
}

func TestJsonSamplesForTestBidder(t *testing.T) {
	originalAdapter := ortbAdapter
	defer func() {
		ortbAdapter = originalAdapter
	}()
	ortbAdapter = &oRTBAdapter{
		BidderInfo: getBidderInfos(),
	}
	bidder, buildErr := Builder("", config.Adapter{}, config.Server{})
	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}
	adapterstest.RunJSONBidderTest(t, "ortb_test_single_requestmode", bidder)
	adapterstest.RunJSONBidderTest(t, "ortb_test_multi_requestmode", bidder)
}

func getBidderInfos() config.BidderInfos {
	return config.BidderInfos{
		"ortb_test_single_requestmode": config.BidderInfo{
			Endpoint: "http://test_bidder.com",
			OpenWrap: config.OpenWrap{
				RequestMode: "single",
			},
		},
		"ortb_test_multi_requestmode": config.BidderInfo{
			Endpoint: "http://test_bidder.com",
			OpenWrap: config.OpenWrap{
				RequestMode: "multi",
			},
		},
	}
}

func TestInitORTBAdapter(t *testing.T) {
	t.Run("init_ortb_adapter", func(t *testing.T) {
		InitORTBAdapter(config.BidderInfos{})
		assert.NotNilf(t, ortbAdapter, "ortbAdapter should not be nil")
	})
}
