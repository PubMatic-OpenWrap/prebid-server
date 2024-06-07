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
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/stretchr/testify/assert"
)

func TestParseRequest(t *testing.T) {
	type args struct {
		request *openrtb2.BidRequest
	}
	type want struct {
		err         error
		rawRequest  json.RawMessage
		requestNode map[string]any
		imps        []any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		// {
		// 	name: "request_is_nil",
		// 	args: args{
		// 		request: nil,
		// 	},
		// 	want: want{
		// 		err:        errImpMissing,
		// 		rawRequest: json.RawMessage(`null`),
		// 	},
		// },
		{
			name: "request_is_valid",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "id",
					Imp: []openrtb2.Imp{
						{
							ID: "imp_1",
						},
					},
				},
			},
			want: want{
				err:        nil,
				rawRequest: json.RawMessage(`{"id":"id","imp":[{"id":"imp_1"}]}`),
				requestNode: map[string]any{
					"id": "id",
					"imp": []any{map[string]any{
						"id": "imp_1",
					}},
				},
				imps: []any{map[string]any{
					"id": "imp_1",
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBuilder := &requestBuilder{}
			err := reqBuilder.parseRequest(tt.args.request)
			assert.Equalf(t, tt.want.err, err, "mismatched error")
			assert.Equalf(t, string(tt.want.rawRequest), string(reqBuilder.rawRequest), "mismatched rawRequest")
			assert.Equalf(t, tt.want.requestNode, reqBuilder.requestNode, "mismatched requestNode")
			assert.Equalf(t, tt.want.imps, reqBuilder.imps, "mismatched imps")
		})
	}
}

func TestBuildEndpoint(t *testing.T) {
	type fields struct {
		endpoint            string
		hasMacrosInEndpoint bool
	}
	type args struct {
		endpointTemplate *template.Template
		bidderParams     map[string]any
	}
	type want struct {
		err      error
		endpoint string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "macros_present_but_hasMacrosInEndpoint_is_false",
			fields: fields{
				endpoint:            "http://{{.host}}/publisher",
				hasMacrosInEndpoint: false,
			},
			args: args{
				endpointTemplate: nil,
				bidderParams:     map[string]any{},
			},
			want: want{
				endpoint: "http://{{.host}}/publisher",
			},
		},
		{
			name: "macros_present_and_bidder_params_not_present",
			fields: fields{
				endpoint:            "http://{{.host}}/publisher",
				hasMacrosInEndpoint: true,
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher`)),
				bidderParams:     map[string]any{},
			},
			want: want{
				endpoint: "http:///publisher",
			},
		},
		{
			name: "macros_present_and_bidder_params_present",
			fields: fields{
				endpoint:            "http://{{.host}}/publisher",
				hasMacrosInEndpoint: true,
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher`)),
				bidderParams: map[string]any{
					"host": "localhost",
				},
			},
			want: want{
				endpoint: "http://localhost/publisher",
			},
		},
		{
			name: "resolveMacros_returns_error",
			fields: fields{
				endpoint:            "http://{{.errorFunc}}/publisher",
				hasMacrosInEndpoint: true,
			},
			args: args{
				bidderParams: map[string]any{},
				endpointTemplate: func() *template.Template {
					errorFunc := template.FuncMap{
						"errorFunc": func() (string, error) {
							return "", errors.New("intentional error")
						},
					}
					template := template.Must(template.New("endpointTemplate").Funcs(errorFunc).Parse(`{{errorFunc}}`))
					return template
				}(),
			},
			want: want{
				endpoint: "",
				err: fmt.Errorf("failed to replace macros in endpoint, err:template: endpointTemplate:1:2: " +
					"executing \"endpointTemplate\" at <errorFunc>: error calling errorFunc: intentional error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBuilder := &requestBuilder{
				endpoint:            tt.fields.endpoint,
				hasMacrosInEndpoint: tt.fields.hasMacrosInEndpoint,
			}
			endpoint, err := reqBuilder.buildEndpoint(tt.args.endpointTemplate, tt.args.bidderParams)
			assert.Equalf(t, tt.want.endpoint, endpoint, "mismatched endpoint")
			assert.Equalf(t, tt.want.err, err, "mismatched error")
		})
	}
}

func TestNewRequestBuilder(t *testing.T) {
	type args struct {
		requestMode string
		endpoint    string
	}
	tests := []struct {
		name string
		args args
		want requestModeBuilder
	}{
		{
			name: "singleRequestMode",
			args: args{
				requestMode: requestModeSingle,
				endpoint:    "http://localhost/publisher",
			},
			want: &singleRequestModeBuilder{
				&requestBuilder{
					endpoint: "http://localhost/publisher",
				},
			},
		},

		{
			name: "multiRequestMode",
			args: args{
				requestMode: requestModeSingle,
				endpoint:    "http://{{.host}}/publisher",
			},
			want: &singleRequestModeBuilder{
				&requestBuilder{
					endpoint:            "http://{{.host}}/publisher",
					hasMacrosInEndpoint: true,
				},
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newRequestBuilder(tt.args.requestMode, tt.args.endpoint)
			assert.Equalf(t, tt.want, got, "mismacthed requestbuilder")
		})
	}
}

func Test_singleRequestModeBuilder_makeRequest(t *testing.T) {
	type fields struct {
		requestBuilder *requestBuilder
	}
	type args struct {
		endpointTemplate  *template.Template
		bidderParamMapper map[string]bidderparams.BidderParamMapper
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
			name: "nil_request",
			fields: fields{
				requestBuilder: &requestBuilder{
					rawRequest: nil,
				},
			},
			args: args{},
			want: want{
				requestData: nil,
				errs:        []error{newBadInputError("failed to empty the imp key in request")},
			},
		},
		{
			name: "no_imp_object_in_builder",
			fields: fields{
				requestBuilder: &requestBuilder{
					rawRequest: json.RawMessage(`{}`),
				},
			},
			args: args{
				endpointTemplate: nil,
			},
			want: want{
				requestData: nil,
				errs:        nil,
			},
		},
		{
			name: "invalid_imp_object",
			fields: fields{
				requestBuilder: &requestBuilder{
					rawRequest: json.RawMessage(`{"imp":["invalid"]}`),
					imps:       []any{"invalid"},
				},
			},
			args: args{},
			want: want{
				requestData: nil,
				errs:        []error{newBadInputError("invalid imp object found at index:0")},
			},
		},
		{
			name: "replace_macros_to_form_endpoint_url",
			fields: fields{
				requestBuilder: &requestBuilder{
					rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
					imps: []any{
						map[string]any{
							"id": "imp_1",
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 5890,
									},
									"host": "localhost.com",
								},
							},
						},
					},
					requestNode: map[string]any{
						"imp": []any{
							map[string]any{
								"ext": map[string]any{
									"bidder": map[string]any{
										"ext": map[string]any{
											"pubid": 5890,
										},
										"host": "localhost.com",
									},
								},
								"id": "imp_1",
							},
						},
					},
					hasMacrosInEndpoint: true,
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
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
			name: "macros_value_absent_in_bidder_params",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
					requestNode: map[string]any{
						"imp": []any{
							map[string]any{
								"ext": map[string]any{},
								"id":  "imp_1",
							},
						},
					},
					imps: []any{
						map[string]any{
							"ext": map[string]any{},
							"id":  "imp_1",
						},
					},
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Option("missingkey=default").Parse(`http://{{.host}}/publisher/{{.pubid}}`)),
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
			name: "buildEndpoint_returns_error",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
					requestNode: map[string]any{
						"imp": []any{
							map[string]any{
								"ext": map[string]any{},
								"id":  "imp_1",
							},
						},
					},
					imps: []any{
						map[string]any{
							"ext": map[string]any{},
							"id":  "imp_1",
						},
					},
				},
			},
			args: args{
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
			want: want{
				requestData: nil,
				errs: []error{newBadInputError("failed to replace macros in endpoint, err:template: endpointTemplate:1:2: " +
					"executing \"endpointTemplate\" at <errorFunc>: error calling errorFunc: intentional error")},
			},
		},
		{
			name: "multi_imps_request",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111},"host":"imp1.host.com"}},"id":"imp_1"},{"ext":{"bidder":{"ext":{"pubid":2222},"host":"imp2.host.com"}},"id":"imp_2"}]}`),
					requestNode: map[string]any{
						"imp": []any{
							map[string]any{
								"ext": map[string]any{
									"bidder": map[string]any{
										"ext": map[string]any{
											"pubid": 1111,
										},
										"host": "imp1.host.com",
									},
								},
								"id": "imp_1",
							},
							map[string]any{
								"ext": map[string]any{
									"bidder": map[string]any{
										"ext": map[string]any{
											"pubid": 2222,
										},
										"host": "imp2.host.com",
									},
								},
								"id": "imp_2",
							},
						},
					},
					imps: []any{
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 1111,
									},
									"host": "imp1.host.com",
								},
							},
							"id": "imp_1",
						},
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 2222,
									},
									"host": "imp2.host.com",
								},
							},
							"id": "imp_2",
						},
					},
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{Location: "host"}
					extMapper := bidderparams.BidderParamMapper{Location: "device"}
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
						Uri:    "http://imp1.host.com/publisher/1111",
						Body:   json.RawMessage(`{"device":{"pubid":1111},"host":"imp1.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://imp2.host.com/publisher/2222",
						Body:   json.RawMessage(`{"device":{"pubid":2222},"host":"imp2.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_2"}]}`),
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
			name: "multi_imps_request_with_one_invalid_imp",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111},"host":"imp1.host.com"}},"id":"imp_1"},"invalid-imp"]}`),
					requestNode: map[string]any{
						"imp": []any{
							map[string]any{
								"ext": map[string]any{
									"bidder": map[string]any{
										"ext": map[string]any{
											"pubid": 1111,
										},
										"host": "imp1.host.com",
									},
								},
								"id": "imp_1",
							},
							"invalid",
						},
					},
					imps: []any{
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 1111,
									},
									"host": "imp1.host.com",
								},
							},
							"id": "imp_1",
						},
						"invalid",
					},
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{Location: "host"}
					extMapper := bidderparams.BidderParamMapper{Location: "device"}
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
			name: "one_imp_updates_request_level_param_but_another_imp_does_not_update",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111}}},"id":"imp_1"},{"ext":{"bidder":{"ext":{"pubid":2222},"host":"imp2.host.com"}},"id":"imp_2"}]}`),
					requestNode: map[string]any{
						"imp": []any{
							map[string]any{
								"ext": map[string]any{
									"bidder": map[string]any{
										"ext": map[string]any{
											"pubid": 1111,
										},
									},
								},
								"id": "imp_1",
							},
							map[string]any{
								"ext": map[string]any{
									"bidder": map[string]any{
										"ext": map[string]any{
											"pubid": 2222,
										},
										"host": "imp2.host.com",
									},
								},
								"id": "imp_2",
							},
						},
					},
					imps: []any{
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 1111,
									},
								},
							},
							"id": "imp_1",
						},
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 2222,
									},
									"host": "imp2.host.com",
								},
							},
							"id": "imp_2",
						},
					},
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{Location: "host"}
					extMapper := bidderparams.BidderParamMapper{Location: "device"}
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
						Uri:    "http:///publisher/1111",
						Body:   json.RawMessage(`{"device":{"pubid":1111},"imp":[{"ext":{"bidder":{}},"id":"imp_1"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
					{
						Method: http.MethodPost,
						Uri:    "http://imp2.host.com/publisher/2222",
						Body:   json.RawMessage(`{"device":{"pubid":2222},"host":"imp2.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_2"}]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sreq := &singleRequestModeBuilder{
				requestBuilder: tt.fields.requestBuilder,
			}
			requestData, errs := sreq.makeRequest(tt.args.endpointTemplate, tt.args.bidderParamMapper)
			assert.Equalf(t, tt.want.requestData, requestData, "mismatched requestData")
			assert.Equalf(t, tt.want.errs, errs, "mismatched errs")
		})
	}
}

func Test_multiRequestModeBuilder_makeRequest(t *testing.T) {
	type fields struct {
		requestBuilder *requestBuilder
	}
	type args struct {
		endpointTemplate  *template.Template
		bidderParamMapper map[string]bidderparams.BidderParamMapper
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
			name: "no_imp_object_in_builder",
			fields: fields{
				requestBuilder: &requestBuilder{
					rawRequest: json.RawMessage(`{}`),
				},
			},
			args: args{
				endpointTemplate: nil,
			},
			want: want{
				requestData: nil,
				errs:        nil,
			},
		},
		{
			name: "invalid_imp_object",
			fields: fields{
				requestBuilder: &requestBuilder{
					rawRequest: json.RawMessage(`{"imp":["invalid"]}`),
					imps:       []any{"invalid"},
				},
			},
			args: args{},
			want: want{
				requestData: nil,
				errs:        []error{newBadInputError("invalid imp object found at index:0")},
			},
		},
		{
			name: "replace_macros_to_form_endpoint_url",
			fields: fields{
				requestBuilder: &requestBuilder{
					rawRequest: json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
					imps: []any{
						map[string]any{
							"id": "imp_1",
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 5890,
									},
									"host": "localhost.com",
								},
							},
						},
					},
					requestNode:         map[string]any{},
					hasMacrosInEndpoint: true,
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
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
			name: "buildEndpoint_returns_error",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
					requestNode:         map[string]any{},
					imps: []any{
						map[string]any{
							"ext": map[string]any{},
							"id":  "imp_1",
						},
					},
				},
			},
			args: args{
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
			want: want{
				requestData: nil,
				errs: []error{newBadInputError("failed to replace macros in endpoint, err:template: endpointTemplate:1:2: " +
					"executing \"endpointTemplate\" at <errorFunc>: error calling errorFunc: intentional error")},
			},
		},
		{
			name: "map_bidder_params_in_multi_imp",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"},{"ext":{"bidder":{"tagid":"valid_tag_id"}},"id":"imp_2"}]}`),
					requestNode:         map[string]any{},
					imps: []any{
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 5890,
									},
									"host": "localhost.com",
								},
							},
							"id": "imp_1",
						},
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"tagid": "valid_tag_id",
								},
							},
							"id": "imp_2",
						},
					},
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{Location: "host"}
					extMapper := bidderparams.BidderParamMapper{Location: "device"}
					tagMapper := bidderparams.BidderParamMapper{Location: "imp.#.tagid"}
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
			name: "multi_imps_request_with_one_invalid_imp",
			fields: fields{
				requestBuilder: &requestBuilder{
					hasMacrosInEndpoint: true,
					rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111},"host":"imp1.host.com"}},"id":"imp_1"},"invalid-imp"]}`),
					requestNode:         map[string]any{},
					imps: []any{
						map[string]any{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 1111,
									},
									"host": "imp1.host.com",
								},
							},
							"id": "imp_1",
						},
						"invalid",
					},
				},
			},
			args: args{
				endpointTemplate: template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
				bidderParamMapper: func() map[string]bidderparams.BidderParamMapper {
					hostMapper := bidderparams.BidderParamMapper{Location: "host"}
					extMapper := bidderparams.BidderParamMapper{Location: "device"}
					tagMapper := bidderparams.BidderParamMapper{Location: "imp.#.tagid"}
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
						Uri:    "http://imp1.host.com/publisher/1111",
						Body:   json.RawMessage(`{"device":{"pubid":1111},"host":"imp1.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"},"invalid"]}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				errs: []error{newBadInputError("invalid imp object found at index:1")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &multiRequestModeBuilder{
				requestBuilder: tt.fields.requestBuilder,
			}
			if builder.requestNode != nil {
				builder.requestNode[impKey] = builder.imps
			}
			requestData, errs := builder.makeRequest(tt.args.endpointTemplate, tt.args.bidderParamMapper)
			assert.Equalf(t, tt.want.requestData, requestData, "mismatched requestData")
			assert.Equalf(t, tt.want.errs, errs, "mismatched errs")
		})
	}
}
