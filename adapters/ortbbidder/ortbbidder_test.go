package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/adapterstest"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestMakeRequests(t *testing.T) {
	type args struct {
		request      *openrtb2.BidRequest
		requestInfo  *adapters.ExtraRequestInfo
		adapterInfo  adapterInfo
		bidderCfgMap *biddersConfigMap
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
			args: args{
				bidderCfgMap: &biddersConfigMap{
					biddersConfig: make(map[string]*bidderConfig),
				},
			},
			want: want{
				errors: []error{fmt.Errorf("Found either nil request or nil requestInfo")},
			},
		},
		{
			name: "requestInfo_is_nil",
			args: args{
				bidderCfgMap: &biddersConfigMap{
					biddersConfig: make(map[string]*bidderConfig),
				},
			},
			want: want{
				errors: []error{fmt.Errorf("Found either nil request or nil requestInfo")},
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
				adapterInfo: adapterInfo{config.Adapter{Endpoint: "http://test_bidder.com"}, extraAdapterInfo{RequestMode: ""}, "testbidder"},
				bidderCfgMap: &biddersConfigMap{
					biddersConfig: make(map[string]*bidderConfig),
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
				adapterInfo: adapterInfo{config.Adapter{Endpoint: "http://test_bidder.com"}, extraAdapterInfo{RequestMode: "single"}, "testbidder"},
				bidderCfgMap: &biddersConfigMap{
					biddersConfig: make(map[string]*bidderConfig),
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
		{
			name: "biddersConfigMap_is_nil",
			args: args{
				request: &openrtb2.BidRequest{
					ID:  "reqid",
					Imp: []openrtb2.Imp{{ID: "imp1", TagID: "tag1"}},
				},
				requestInfo: &adapters.ExtraRequestInfo{
					BidderCoreName: openrtb_ext.BidderName("ortb_test_single_requestmode"),
				},
				adapterInfo:  adapterInfo{config.Adapter{Endpoint: "http://test_bidder.com"}, extraAdapterInfo{RequestMode: "single"}, "testbidder"},
				bidderCfgMap: nil,
			},
			want: want{
				errors: []error{fmt.Errorf("Found nil biddersConfigMap")},
			},
		},
		{
			name: "biddersConfig_is_nil",
			args: args{
				request: &openrtb2.BidRequest{
					ID:  "reqid",
					Imp: []openrtb2.Imp{{ID: "imp1", TagID: "tag1"}},
				},
				requestInfo: &adapters.ExtraRequestInfo{
					BidderCoreName: openrtb_ext.BidderName("ortb_test_single_requestmode"),
				},
				adapterInfo: adapterInfo{config.Adapter{Endpoint: "http://test_bidder.com"}, extraAdapterInfo{RequestMode: "single"}, "testbidder"},
				bidderCfgMap: &biddersConfigMap{
					biddersConfig: nil,
				},
			},
			want: want{
				errors: []error{fmt.Errorf("Found nil biddersConfigMap")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &adapter{adapterInfo: tt.args.adapterInfo, biddersConfigMap: tt.args.bidderCfgMap}
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
			adapter := &adapter{}
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
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid_banner_bid",
			args: args{
				bid: openrtb2.Bid{ID: "1", MType: openrtb2.MarkupBanner},
			},
			want: want{
				bidType: openrtb_ext.BidTypeBanner,
			},
		},
		{
			name: "valid_video_bid",
			args: args{
				bid: openrtb2.Bid{ID: "2", MType: openrtb2.MarkupVideo},
			},
			want: want{
				bidType: openrtb_ext.BidTypeVideo,
			},
		},
		{
			name: "valid_audio_bid",
			args: args{
				bid: openrtb2.Bid{ID: "3", MType: openrtb2.MarkupAudio},
			},
			want: want{
				bidType: openrtb_ext.BidTypeAudio,
			},
		},
		{
			name: "valid_native_bid",
			args: args{
				bid: openrtb2.Bid{ID: "4", MType: openrtb2.MarkupNative},
			},
			want: want{
				bidType: openrtb_ext.BidTypeNative,
			},
		},
		{
			name: "invalid_bid_type",
			args: args{
				bid: openrtb2.Bid{ID: "5", MType: 123},
			},
			want: want{
				bidType: "",
			},
		},
		{
			name: "bid.MType_has_high_priority",
			args: args{
				bid: openrtb2.Bid{ID: "5", MType: openrtb2.MarkupVideo, Ext: json.RawMessage(`{"prebid":{"type":"video"}}`)},
			},
			want: want{
				bidType: "video",
			},
		},
		{
			name: "bid.ext.prebid.type_is_absent",
			args: args{
				bid: openrtb2.Bid{ID: "5", Ext: json.RawMessage(`{"prebid":{}}`)},
			},
			want: want{
				bidType: "",
			},
		},
		{
			name: "bid.ext.prebid.type_json_unmarshal_fails",
			args: args{
				bid: openrtb2.Bid{ID: "5", Ext: json.RawMessage(`{"prebid":{invalid-json}}`)},
			},
			want: want{
				bidType: "",
			},
		},
		{
			name: "bid.ext.prebid.type_is_valid",
			args: args{
				bid: openrtb2.Bid{ID: "5", Ext: json.RawMessage(`{"prebid":{"type":"banner"}}`)},
			},
			want: want{
				bidType: "banner",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bidType := getMediaTypeForBid(tt.args.bid)
			assert.Equal(t, tt.want.bidType, bidType, "mismatched bidType")
		})
	}
}

func TestJsonSamplesForSingleRequestMode(t *testing.T) {
	oldMapper := g_biddersConfigMap
	defer func() {
		g_biddersConfigMap = oldMapper
	}()
	g_biddersConfigMap = &biddersConfigMap{biddersConfig: make(map[string]*bidderConfig)}
	bidder, buildErr := Builder("owgeneric_single_requestmode",
		config.Adapter{
			Endpoint:         "http://test_bidder.com",
			ExtraAdapterInfo: `{"requestMode":"single"}`,
		}, config.Server{})
	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}
	adapterstest.RunJSONBidderTest(t, "ortbbiddertest/owortb_generic_single_requestmode", bidder)
}

func TestJsonSamplesForMultiRequestMode(t *testing.T) {
	oldMapper := g_biddersConfigMap
	defer func() {
		g_biddersConfigMap = oldMapper
	}()
	g_biddersConfigMap = &biddersConfigMap{biddersConfig: make(map[string]*bidderConfig)}
	bidder, buildErr := Builder("owgeneric_multi_requestmode",
		config.Adapter{
			Endpoint:         "http://test_bidder.com",
			ExtraAdapterInfo: ``,
		}, config.Server{})
	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}
	adapterstest.RunJSONBidderTest(t, "ortbbiddertest/owortb_generic_multi_requestmode", bidder)
}

func TestPrepareRequestData(t *testing.T) {
	type fields struct {
		Adapter config.Adapter
	}
	type args struct {
		request *openrtb2.BidRequest
		mapper  map[string]bidderProperty
	}
	type want struct {
		requestData *adapters.RequestData
		err         error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "valid_request",
			fields: fields{
				Adapter: config.Adapter{Endpoint: "https://example.com"},
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID:  "123",
					Imp: []openrtb2.Imp{{ID: "imp1"}},
				},
				mapper: make(map[string]bidderProperty),
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
			name: "nil_request",
			fields: fields{
				Adapter: config.Adapter{Endpoint: "https://example.com"},
			},
			args: args{
				request: nil,
				mapper:  make(map[string]bidderProperty),
			},
			want: want{
				requestData: nil,
				err:         fmt.Errorf("found nil request"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := adapterInfo{
				Adapter: tt.fields.Adapter,
			}
			got, err := o.prepareRequestData(tt.args.request, tt.args.mapper)
			assert.Equal(t, tt.want.requestData, got, "mismatched requestData")
			assert.Equal(t, tt.want.err, err, "mismatched error")
		})
	}
}

func TestBuilder(t *testing.T) {
	InitBiddersConfigMap("../../static/bidder-params")
	type args struct {
		bidderName openrtb_ext.BidderName
		config     config.Adapter
		server     config.Server
	}
	type want struct {
		err    error
		bidder adapters.Bidder
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "fails_to_parse_extra_info",
			args: args{
				bidderName: "ortbbidder",
				config: config.Adapter{
					ExtraAdapterInfo: "invalid-string",
				},
				server: config.Server{},
			},
			want: want{
				bidder: nil,
				err:    fmt.Errorf("Failed to parse extra_info for bidder:[ortbbidder] err:[invalid character 'i' looking for beginning of value]"),
			},
		},
		{
			name: "bidder_with_requestMode",
			args: args{
				bidderName: "ortbbidder",
				config: config.Adapter{
					ExtraAdapterInfo: `{"requestMode":"single"}`,
				},
				server: config.Server{},
			},
			want: want{
				bidder: &adapter{
					adapterInfo: adapterInfo{
						extraInfo: extraAdapterInfo{
							RequestMode: "single",
						},
						Adapter: config.Adapter{
							ExtraAdapterInfo: `{"requestMode":"single"}`,
						},
						bidderName: "ortbbidder",
					},
					biddersConfigMap: &biddersConfigMap{
						biddersConfig: map[string]*bidderConfig{
							"owortb_testbidder": {
								requestProperties: map[string]bidderProperty{
									"adunitID": {location: []string{"ext", "adunit", "id"}},
								},
							},
						},
					},
				},
				err: nil,
			},
		},
		{
			name: "bidder_without_requestMode",
			args: args{
				bidderName: "ortbbidder",
				config: config.Adapter{
					ExtraAdapterInfo: "",
				},
				server: config.Server{},
			},
			want: want{
				bidder: &adapter{
					adapterInfo: adapterInfo{
						Adapter: config.Adapter{
							ExtraAdapterInfo: ``,
						},
						bidderName: "ortbbidder",
					},
					biddersConfigMap: &biddersConfigMap{
						biddersConfig: map[string]*bidderConfig{
							"owortb_testbidder": {
								requestProperties: map[string]bidderProperty{
									"adunitID": {location: []string{"ext", "adunit", "id"}},
								},
							},
						},
					},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Builder(tt.args.bidderName, tt.args.config, tt.args.server)
			assert.Equal(t, tt.want.bidder, got, "mismatched bidder")
			assert.Equal(t, tt.want.err, err, "mismatched error")
		})
	}
}
