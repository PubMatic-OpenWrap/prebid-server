package ortbbidder

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"text/template"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/stretchr/testify/assert"
)

func TestMultiRequestBuilderParseRequest(t *testing.T) {
	type args struct {
		request *openrtb2.BidRequest
	}
	type want struct {
		err        error
		rawRequest json.RawMessage
		imps       []map[string]any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "request_without_imps",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "id",
				},
			},
			want: want{
				err:        errImpMissing,
				rawRequest: nil,
				imps:       nil,
			},
		},
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
				rawRequest: json.RawMessage(`{"id":"id","imp":null}`),
				imps: []map[string]any{{
					"id": "imp_1",
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBuilder := &multiRequestBuilder{}
			err := reqBuilder.parseRequest(tt.args.request)
			assert.Equalf(t, tt.want.err, err, "mismatched error")
			assert.Equalf(t, string(tt.want.rawRequest), string(reqBuilder.rawRequest), "mismatched rawRequest")
			assert.Equalf(t, tt.want.imps, reqBuilder.imps, "mismatched imps")
		})
	}
}

func TestMultiRequestBuilderMakeRequest(t *testing.T) {
	type fields struct {
		requestBuilder multiRequestBuilder
	}
	type want struct {
		requestData []*adapters.RequestData
		errs        []error
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "nil_request",
			fields: fields{
				requestBuilder: multiRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						rawRequest: nil,
					},
				},
			},
			want: want{
				requestData: nil,
				errs:        nil,
			},
		},
		{
			name: "no_imp_object_in_builder",
			fields: fields{
				requestBuilder: multiRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						rawRequest: json.RawMessage(`{}`),
					},
				},
			},
			want: want{
				requestData: nil,
				errs:        nil,
			},
		},
		{
			name: "replace_macros_to_form_endpoint_url",
			fields: fields{
				requestBuilder: multiRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						hasMacrosInEndpoint: true,
						rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
						endpointTemplate:    template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
					},
					imps: []map[string]any{
						{
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
				},
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
				requestBuilder: multiRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						hasMacrosInEndpoint: true,
						rawRequest:          json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
						endpointTemplate:    template.Must(template.New("endpointTemplate").Option("missingkey=default").Parse(`http://{{.host}}/publisher/{{.pubid}}`)),
					},
					imps: []map[string]any{
						{
							"ext": map[string]any{},
							"id":  "imp_1",
						},
					},
				},
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
				requestBuilder: multiRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						hasMacrosInEndpoint: true,
						rawRequest:          json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
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
					imps: []map[string]any{
						{
							"ext": map[string]any{},
							"id":  "imp_1",
						},
					},
				},
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
				requestBuilder: multiRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						hasMacrosInEndpoint: true,
						rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111},"host":"imp1.host.com"}},"id":"imp_1"},{"ext":{"bidder":{"ext":{"pubid":2222},"host":"imp2.host.com"}},"id":"imp_2"}]}`),
						endpointTemplate:    template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
						requestParams: func() map[string]bidderparams.BidderParamMapper {
							hostMapper := bidderparams.BidderParamMapper{Location: "host"}
							extMapper := bidderparams.BidderParamMapper{Location: "device"}
							return map[string]bidderparams.BidderParamMapper{
								"host": hostMapper,
								"ext":  extMapper,
							}
						}(),
					},
					imps: []map[string]any{
						{
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
						{
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
			name: "one_imp_updates_request_level_param_but_another_imp_expects_original_request_param",
			fields: fields{
				requestBuilder: multiRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						hasMacrosInEndpoint: true,
						rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":1111}}},"id":"imp_1"},{"ext":{"bidder":{"ext":{"pubid":2222},"host":"imp2.host.com"}},"id":"imp_2"}]}`),
						endpointTemplate:    template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
						requestParams: func() map[string]bidderparams.BidderParamMapper {
							hostMapper := bidderparams.BidderParamMapper{Location: "host"}
							extMapper := bidderparams.BidderParamMapper{Location: "device"}
							return map[string]bidderparams.BidderParamMapper{
								"host": hostMapper,
								"ext":  extMapper,
							}
						}(),
					},
					imps: []map[string]any{
						{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext": map[string]any{
										"pubid": 1111,
									},
								},
							},
							"id": "imp_1",
						},
						{
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
			requestData, errs := tt.fields.requestBuilder.makeRequest()
			assert.Equalf(t, tt.want.requestData, requestData, "mismatched requestData")
			assert.Equalf(t, tt.want.errs, errs, "mismatched errs")
		})
	}
}
