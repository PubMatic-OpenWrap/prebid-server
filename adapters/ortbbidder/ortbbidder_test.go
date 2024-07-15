package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"text/template"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	InitBidderParamsConfig("../../static/bidder-params", "../../static/bidder-response-params")
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
			name: "bidder_with_requestType",
			args: args{
				bidderName: "ortbbidder",
				config: config.Adapter{
					ExtraAdapterInfo: `{"requestType":"single"}`,
				},
				server: config.Server{},
			},
			want: want{
				bidder: &adapter{
					adapterInfo: adapterInfo{
						extraInfo: extraAdapterInfo{
							RequestType: "single",
						},
						Adapter: config.Adapter{
							ExtraAdapterInfo: `{"requestType":"single"}`,
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
			name: "bidder_without_requestType",
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
				bidderCfg: bidderparams.NewBidderConfig(),
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
				adapterInfo: adapterInfo{config.Adapter{Endpoint: "http://test_bidder.com"}, extraAdapterInfo{RequestType: "single"}, "testbidder", nil},
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
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestType: "single"}, "testbidder", template}
				}(),
				bidderCfg: bidderparams.NewBidderConfig(),
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
						ImpIDs: []string{"imp1"},
					},
				},
				errors: nil,
			},
		},
		{
			name: "multi_requestType_to_form_requestdata",
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
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestType: "multi"}, "testbidder", template}
				}(),
				bidderCfg: bidderparams.NewBidderConfig(),
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
						ImpIDs: []string{"imp1"},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://test_bidder.com",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
						ImpIDs: []string{"imp2"},
					},
				},
			},
		},
		{
			name: "multi_requestType_validate_endpoint_macro",
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
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestType: "multi"}, "testbidder", template}
				}(),
				bidderCfg: bidderparams.NewBidderConfig(),
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
						ImpIDs: []string{"imp1"},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://",
						Body:   []byte(`{"id":"reqid","imp":[{"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
						ImpIDs: []string{"imp2"},
					},
				},
			},
		},
		{
			name: "single_requestType_to_form_requestdata",
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
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestType: "single"}, "testbidder", template}
				}(),
				bidderCfg: bidderparams.NewBidderConfig(),
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
						ImpIDs: []string{"imp1", "imp2"},
					},
				},
			},
		},
		{
			name: "single_requestType_validate_endpoint_macros",
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
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestType: ""}, "testbidder", template}
				}(),
				bidderCfg: bidderparams.NewBidderConfig(),
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
						ImpIDs: []string{"imp1", "imp2"},
					},
				},
			},
		},
		{
			name: "multi_requestType_add_request_params_in_request",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "reqid",
					Imp: []openrtb2.Imp{
						{ID: "imp1", TagID: "tag1", Ext: json.RawMessage(`{"bidder": {"host": "localhost.com"}}`)},
						{ID: "imp2", TagID: "tag2", Ext: json.RawMessage(`{"bidder": {"zone": "testZone"}}`)},
					},
				},
				adapterInfo: func() adapterInfo {
					endpoint := "http://{{.host}}"
					template, _ := template.New("endpointTemplate").Parse(endpoint)
					return adapterInfo{config.Adapter{Endpoint: endpoint}, extraAdapterInfo{RequestType: "multi"}, "testbidder", template}
				}(),
				bidderCfg: func() *bidderparams.BidderConfig {
					cfg := bidderparams.NewBidderConfig()
					cfg.BidderConfigMap["testbidder"] = &bidderparams.Config{
						RequestParams: map[string]bidderparams.BidderParamMapper{
							"host": {Location: "server.host"},
							"zone": {Location: "ext.zone"},
						},
					}

					return cfg
				}(),
			},
			want: want{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://localhost.com",
						Body:   []byte(`{"id":"reqid","imp":[{"ext":{"bidder":{}},"id":"imp1","tagid":"tag1"}],"server":{"host":"localhost.com"}}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						}, ImpIDs: []string{"imp1"},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://",
						Body:   []byte(`{"ext":{"zone":"testZone"},"id":"reqid","imp":[{"ext":{"bidder":{}},"id":"imp2","tagid":"tag2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
						ImpIDs: []string{"imp2"},
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
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name:         "no content response data",
			responseData: &adapters.ResponseData{StatusCode: http.StatusNoContent},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
			expectedResponse: nil,
			expectedErrors:   nil,
		},
		{
			name:         "status bad request in response data",
			responseData: &adapters.ResponseData{StatusCode: http.StatusBadRequest},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
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
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
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
				Body:       []byte(`{"id":1,"seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1","bidtype":2}]}]`),
			},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
			expectedResponse: nil,
			expectedErrors: []error{&errortypes.BadServerResponse{
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
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "unexpected value type: 0"}},
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
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "invalid seatbid found in seatbid array, seatbid:[invalid]"}},
			setup: func() adapter {
				return adapter{
					bidderParamsConfig: &bidderparams.BidderConfig{},
				}
			},
		},
		{
			name: "invalid bid arrays in seatbid",
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
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "invalid bid found in bids array, bid:[invalid]"}},
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
			expectedErrors:   []error{&errortypes.BadServerResponse{Message: "cannot unmarshal ID: expects \" or n, but found 1"}},
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
						BidType: "video",
					},
				},
			},
			expectedErrors: nil,
			setup: func() adapter {
				bc := bidderparams.NewBidderConfig()
				bc.BidderConfigMap["owortb_testbidder"] = &bidderparams.Config{
					ResponseParams: map[string]bidderparams.BidderParamMapper{},
				}
				return adapter{
					bidderParamsConfig: bc,
					adapterInfo: adapterInfo{
						bidderName: "owortb_testbidder",
					},
				}
			},
		},
		{
			name: "valid response - bidder params present",
			responseData: &adapters.ResponseData{
				Body:       []byte(`{"id":"1","cur":"","seatbid":[{"bid":[{"id":"1","ext":{"bidtype":"video"}}]}],"ext":{"currency":"USD"}}`),
				StatusCode: http.StatusOK,
			},
			expectedResponse: &adapters.BidderResponse{
				Currency: "",
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:  "1",
							Ext: json.RawMessage(`{"bidtype":"video"}`),
						},
						BidType: "video",
					},
				},
			},
			expectedErrors: nil,
			setup: func() adapter {
				bc := bidderparams.NewBidderConfig()
				bc.BidderConfigMap["owortb_testbidder"] = &bidderparams.Config{
					ResponseParams: map[string]bidderparams.BidderParamMapper{
						"bidType":  {Location: "seatbid.#.bid.#.ext.bidtype"},
						"currency": {Location: "ext.currency"},
					},
				}
				return adapter{
					adapterInfo: adapterInfo{
						bidderName: "owortb_testbidder",
					},
					bidderParamsConfig: bc,
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
