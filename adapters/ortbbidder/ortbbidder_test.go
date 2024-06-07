package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"text/template"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/adapterstest"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestMakeRequests(t *testing.T) {
	type args struct {
		request     *openrtb2.BidRequest
		requestInfo *adapters.ExtraRequestInfo
		adapterInfo adapterInfo
		bidderCfg   *bidderparams.BidderConfig
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
				bidderCfg: &bidderparams.BidderConfig{},
			},
			want: want{
				errors: []error{newBadInputError(errImpMissing.Error())},
			},
		},
		{
			name: "bidderParamsConfig_is_nil",
			args: args{
				request: &openrtb2.BidRequest{
					ID:  "reqid",
					Imp: []openrtb2.Imp{{ID: "imp1", TagID: "tag1"}},
				},
				adapterInfo: adapterInfo{config.Adapter{Endpoint: "http://test_bidder.com"}, extraAdapterInfo{RequestMode: "single"}, "testbidder", nil},
				bidderCfg:   nil,
			},
			want: want{
				errors: []error{newBadInputError("found nil bidderParamsConfig")},
			},
		},
		{
			name: "bidderParamsConfig_not_contains_bidder_param_data",
			args: args{
				request: &openrtb2.BidRequest{
					ID:  "reqid",
					Imp: []openrtb2.Imp{{ID: "imp1", TagID: "tag1"}},
				},
				adapterInfo: func() adapterInfo {
					endpoint := "http://test_bidder.com"
					template, _ := template.New("endpointTemplate").Parse(endpoint)
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestMode: "single"}, "testbidder", template}
				}(),
				bidderCfg: &bidderparams.BidderConfig{},
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp1","tagid":"tag1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errors: nil,
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
				adapterInfo: func() adapterInfo {
					endpoint := "http://test_bidder.com"
					template, _ := template.New("endpointTemplate").Parse(endpoint)
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestMode: "single"}, "testbidder", template}
				}(),
				bidderCfg: &bidderparams.BidderConfig{},
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp1","tagid":"tag1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
			},
		},
		{
			name: "single_requestmode_validate_endpoint_macro",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "reqid",
					Imp: []openrtb2.Imp{
						{ID: "imp1", TagID: "tag1", Ext: json.RawMessage(`{"bidder": {"host": "localhost.com"}}`)},
						{ID: "imp2", TagID: "tag2"},
					},
				},
				adapterInfo: func() adapterInfo {
					endpoint := "http://{{.host}}"
					template, _ := template.New("endpointTemplate").Parse(endpoint)
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestMode: "single"}, "testbidder", template}
				}(),
				bidderCfg: &bidderparams.BidderConfig{},
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com",
						Body:   []byte(`{"id":"reqid","imp":[{"ext":{"bidder":{"host":"localhost.com"}},"id":"imp1","tagid":"tag1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
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
				adapterInfo: func() adapterInfo {
					endpoint := "http://test_bidder.com"
					template, _ := template.New("endpointTemplate").Parse(endpoint)
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestMode: ""}, "testbidder", template}
				}(),
				bidderCfg: &bidderparams.BidderConfig{},
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp1","tagid":"tag1"},{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
			},
		},
		{
			name: "multi_requestmode_validate_endpoint_macros",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "reqid",
					Imp: []openrtb2.Imp{
						{ID: "imp1", TagID: "tag1", Ext: json.RawMessage(`{"bidder": {"host": "localhost.com"}}`)},
						{ID: "imp2", TagID: "tag2"},
					},
				},
				adapterInfo: func() adapterInfo {
					endpoint := "http://{{.host}}"
					template, _ := template.New("endpointTemplate").Parse(endpoint)
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestMode: ""}, "testbidder", template}
				}(),
				bidderCfg: &bidderparams.BidderConfig{},
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com",
						Body:   []byte(`{"id":"reqid","imp":[{"ext":{"bidder":{"host":"localhost.com"}},"id":"imp1","tagid":"tag1"},{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &adapter{adapterInfo: tt.args.adapterInfo, bidderParamsConfig: tt.args.bidderCfg}
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
	oldMapper := g_bidderParamsConfig
	defer func() {
		g_bidderParamsConfig = oldMapper
	}()
	g_bidderParamsConfig = &bidderparams.BidderConfig{}
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
	oldMapper := g_bidderParamsConfig
	defer func() {
		g_bidderParamsConfig = oldMapper
	}()
	g_bidderParamsConfig = &bidderparams.BidderConfig{}
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

func TestBuilder(t *testing.T) {
	InitBidderParamsConfig("../../static/bidder-params")
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
				err:    fmt.Errorf("failed to parse extra_info: expect { or n, but found i"),
			},
		},
		{
			name: "fails_to_parse_template_endpoint",
			args: args{
				bidderName: "ortbbidder",
				config: config.Adapter{
					ExtraAdapterInfo: "{}",
					Endpoint:         "http://{{.Host}",
				},
				server: config.Server{},
			},
			want: want{
				bidder: nil,
				err:    fmt.Errorf("failed to parse endpoint url template: template: endpointTemplate:1: bad character U+007D '}'"),
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
						endpointTemplate: func() *template.Template {
							template, _ := template.New("endpointTemplate").Option("missingkey=zero").Parse("")
							return template
						}(),
					},
					bidderParamsConfig: g_bidderParamsConfig,
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
						endpointTemplate: func() *template.Template {
							template, _ := template.New("endpointTemplate").Option("missingkey=zero").Parse("")
							return template
						}(),
					},
					bidderParamsConfig: g_bidderParamsConfig,
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

func TestInitBidderParamsConfig(t *testing.T) {
	tests := []struct {
		name    string
		dirPath string
		wantErr bool
	}{
		{
			name:    "test_InitBiddersConfigMap_success",
			dirPath: "../../static/bidder-params/",
			wantErr: false,
		},
		{
			name:    "test_InitBiddersConfigMap_failure",
			dirPath: "/invalid_directory/",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitBidderParamsConfig(tt.dirPath)
			assert.Equal(t, err != nil, tt.wantErr, "mismatched error")
		})
	}
}

func TestIsORTBBidder(t *testing.T) {
	type args struct {
		bidderName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ortb_bidder",
			args: args{
				bidderName: "owortb_magnite",
			},
			want: true,
		},
		{
			name: "non_ortb_bidder",
			args: args{
				bidderName: "magnite",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isORTBBidder(tt.args.bidderName)
			assert.Equal(t, tt.want, got, "mismatched output of isORTBBidder")
		})
	}
}
