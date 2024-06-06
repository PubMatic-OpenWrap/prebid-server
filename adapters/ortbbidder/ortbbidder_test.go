package ortbbidder

import (
	"encoding/json"
	"errors"
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
				errors: []error{newBadInputError("found nil request")},
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
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
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
						Uri:    "http://",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com",
						Body:   []byte(`{"id":"reqid","imp":[{"ext":{"bidder":{"host":"localhost.com"}},"id":"imp1","tagid":"tag1"}]}`),
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

func TestMakeRequest(t *testing.T) {
	type fields struct {
		endpointTemplate *template.Template
	}
	type args struct {
		rawRequest                json.RawMessage
		bidderParamMapper         map[string]bidderparams.BidderParamMapper
		supportSingleImpInRequest bool
	}
	type want struct {
		requestData []*adapters.RequestData
		errs        []error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name:   "nil_request",
			fields: fields{},
			args: args{
				rawRequest: nil,
			},
			want: want{
				requestData: nil,
				errs:        []error{newBadInputError("failed to unmarshal request, err:expect { or n, but found \x00")},
			},
		},
		{
			name:   "no_imp_object",
			fields: fields{},
			args: args{
				rawRequest: json.RawMessage(`{}`),
			},
			want: want{
				requestData: nil,
				errs:        []error{newBadInputError("imp object not found in request")},
			},
		},
		{
			name:   "invalid_imp_object",
			fields: fields{},
			args: args{
				rawRequest: json.RawMessage(`{"imp":["invalid"]}`),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "",
						Body:   json.RawMessage(`{"imp":["invalid"]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: []error{newBadInputError("invalid imp object found at index:0")},
			},
		},
		{
			name: "multiRequestMode_replace_macros_to_form_endpoint_url",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com/publisher/5890",
						Body:   json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "multiRequestMode_macros_value_absent_in_bidder_params",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Option("missingkey=default").Parse(`http://{{.host}}/publisher/{{.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http:///publisher/",
						Body:   json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "multiRequestMode_macro_replacement_failure",
			fields: fields{
				endpointTemplate: func() *template.Template {
					errorFunc := template.FuncMap{
						"errorFunc": func() (string, error) {
							return "", errors.New("intentional error")
						},
					}
					t := template.Must(template.New("endpointTemplate").Funcs(errorFunc).Parse(`{{errorFunc}}`))
					return t
				}(),
			},
			args: args{
				supportSingleImpInRequest: false,
				rawRequest:                json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
			},
			want: want{
				requestData: nil,
				errs: []error{newBadInputError("failed to replace macros in endpoint, err:template: endpointTemplate:1:2: " +
					"executing \"endpointTemplate\" at <errorFunc>: error calling errorFunc: intentional error")},
			},
		},
		{
			name: "multiRequestMode_first_imp_bidder_params_has_high_priority_while_replacing_macros_in_endpoint",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"host":"localhost.com"}},"id":"imp_1"},{"ext":{"bidder":{"host":"imp2.host.com"}},"id":"imp_2"}]}`),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com/publisher",
						Body:   json.RawMessage(`{"imp":[{"ext":{"bidder":{"host":"localhost.com"}},"id":"imp_1"},{"ext":{"bidder":{"host":"imp2.host.com"}},"id":"imp_2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "map_bidder_params_in_single_imp",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{}
					hostMapper.SetLocation("host")
					extMapper := bidderparams.BidderParamMapper{}
					extMapper.SetLocation("device")
					return map[string]bidderparams.BidderParamMapper{
						"host": hostMapper,
						"ext":  extMapper,
					}
				}(),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com/publisher/5890",
						Body:   json.RawMessage(`{"device":{"pubid":5890},"host":"localhost.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "multiRequestMode_map_bidder_params_in_multi_imp",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"},{"ext":{"bidder":{"tagid":"valid_tag_id"}},"id":"imp_2"}]}`),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{}
					hostMapper.SetLocation("host")
					extMapper := bidderparams.BidderParamMapper{}
					extMapper.SetLocation("device")
					tagMapper := bidderparams.BidderParamMapper{}
					tagMapper.SetLocation("imp.#.tagid")
					return map[string]bidderparams.BidderParamMapper{
						"host":  hostMapper,
						"ext":   extMapper,
						"tagid": tagMapper,
					}
				}(),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com/publisher/5890",
						Body:   json.RawMessage(`{"device":{"pubid":5890},"host":"localhost.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"},{"ext":{"bidder":{}},"id":"imp_2","tagid":"valid_tag_id"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "multiRequestMode_first_imp_bidder_param_has_high_pririty",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"},{"ext":{"bidder":{"ext":{"pubid":1111}}},"id":"imp_2"}]}`),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{}
					hostMapper.SetLocation("host")
					extMapper := bidderparams.BidderParamMapper{}
					extMapper.SetLocation("device")
					return map[string]bidderparams.BidderParamMapper{
						"host": hostMapper,
						"ext":  extMapper,
					}
				}(),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com/publisher/5890",
						Body:   json.RawMessage(`{"device":{"pubid":5890},"host":"localhost.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"},{"ext":{"bidder":{}},"id":"imp_2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "bidder_param_mapping_absent",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest:        json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
				bidderParamMapper: nil,
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com/publisher/5890",
						Body:   json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "singleRequestMode_single_imp_request",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111},"host":"imp1.host.com"}},"id":"imp_1"}]}`),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{}
					hostMapper.SetLocation("host")
					extMapper := bidderparams.BidderParamMapper{}
					extMapper.SetLocation("device")
					return map[string]bidderparams.BidderParamMapper{
						"host": hostMapper,
						"ext":  extMapper,
					}
				}(),
				supportSingleImpInRequest: true,
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://imp1.host.com/publisher/1111",
						Body:   json.RawMessage(`{"device":{"pubid":1111},"host":"imp1.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "singleRequestMode_multi_imps_request",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111},"host":"imp1.host.com"}},"id":"imp_1"},{"ext":{"bidder":{"ext":{"pubid":2222},"host":"imp2.host.com"}},"id":"imp_2"}]}`),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{}
					hostMapper.SetLocation("host")
					extMapper := bidderparams.BidderParamMapper{}
					extMapper.SetLocation("device")
					return map[string]bidderparams.BidderParamMapper{
						"host": hostMapper,
						"ext":  extMapper,
					}
				}(),
				supportSingleImpInRequest: true,
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://imp2.host.com/publisher/2222",
						Body:   json.RawMessage(`{"device":{"pubid":2222},"host":"imp2.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://imp1.host.com/publisher/1111",
						Body:   json.RawMessage(`{"device":{"pubid":1111},"host":"imp1.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "singleRequestMode_multi_imps_request_with_one_invalid_imp",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111},"host":"imp1.host.com"}},"id":"imp_1"},"invalid-imp"]}`),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{}
					hostMapper.SetLocation("host")
					extMapper := bidderparams.BidderParamMapper{}
					extMapper.SetLocation("device")
					return map[string]bidderparams.BidderParamMapper{
						"host": hostMapper,
						"ext":  extMapper,
					}
				}(),
				supportSingleImpInRequest: true,
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://imp1.host.com/publisher/1111",
						Body:   json.RawMessage(`{"device":{"pubid":1111},"host":"imp1.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: []error{newBadInputError("invalid imp object found at index:1")},
			},
		},
		{
			name: "singleRequestMode_one_imp_updates_request_level_param_but_another_imp_does_not_update",
			fields: fields{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
			},
			args: args{
				rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111}}},"id":"imp_1"},{"ext":{"bidder":{"ext":{"pubid":2222},"host":"imp2.host.com"}},"id":"imp_2"}]}`),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{}
					hostMapper.SetLocation("host")
					extMapper := bidderparams.BidderParamMapper{}
					extMapper.SetLocation("device")
					return map[string]bidderparams.BidderParamMapper{
						"host": hostMapper,
						"ext":  extMapper,
					}
				}(),
				supportSingleImpInRequest: true,
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://imp2.host.com/publisher/2222",
						Body:   json.RawMessage(`{"device":{"pubid":2222},"host":"imp2.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
					{
						Method: http.MethodPost,
						Uri:    "http:///publisher/1111",
						Body:   json.RawMessage(`{"device":{"pubid":1111},"imp":[{"ext":{"bidder":{}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
		{
			name: "singleRequestMode_macro_replacement_failure",
			fields: fields{
				endpointTemplate: func() *template.Template {
					errorFunc := template.FuncMap{
						"errorFunc": func() (string, error) {
							return "", errors.New("intentional error")
						},
					}
					t := template.Must(template.New("endpointTemplate").Funcs(errorFunc).Parse(`{{errorFunc}}`))
					return t
				}(),
			},
			args: args{
				supportSingleImpInRequest: true,
				rawRequest:                json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
			},
			want: want{
				requestData: nil,
				errs: []error{newBadInputError("failed to replace macros in endpoint, err:template: endpointTemplate:1:2: " +
					"executing \"endpointTemplate\" at <errorFunc>: error calling errorFunc: intentional error")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := adapterInfo{
				endpointTemplate: tt.fields.endpointTemplate,
			}
			requestData, errs := o.makeRequest(tt.args.rawRequest, tt.args.bidderParamMapper, tt.args.supportSingleImpInRequest)
			assert.Equalf(t, tt.want.requestData, requestData, "mismatched requestData")
			assert.Equalf(t, tt.want.errs, errs, "mismatched errs")
		})
	}
}
