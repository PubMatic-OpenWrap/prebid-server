package ortbbidder

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"text/template"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/util"
	"github.com/stretchr/testify/assert"
)

func TestSingleRequestBuilderParseRequest(t *testing.T) {
	type args struct {
		request *openrtb2.BidRequest
	}
	type want struct {
		err        error
		rawRequest json.RawMessage
		newRequest map[string]any
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
				err:        util.ErrImpMissing,
				rawRequest: json.RawMessage(`{"id":"id","imp":null}`),
				imps:       nil,
				newRequest: map[string]any{
					"id":  "id",
					"imp": nil,
				},
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
				rawRequest: json.RawMessage(`{"id":"id","imp":[{"id":"imp_1"}]}`),
				newRequest: map[string]any{
					"id": "id",
					"imp": []any{map[string]any{
						"id": "imp_1",
					}},
				},
				imps: []map[string]any{
					{
						"id": "imp_1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBuilder := &singleRequestBuilder{}
			err := reqBuilder.parseRequest(tt.args.request)
			assert.Equalf(t, tt.want.err, err, "mismatched error")
			assert.Equalf(t, string(tt.want.rawRequest), string(reqBuilder.rawRequest), "mismatched rawRequest")
			assert.Equalf(t, tt.want.imps, reqBuilder.imps, "mismatched imps")
			assert.Equalf(t, tt.want.newRequest, reqBuilder.newRequest, "mismatched newRequest")
		})
	}
}

func TestSingleRequestBuilderMakeRequest(t *testing.T) {
	type fields struct {
		requestBuilder singleRequestBuilder
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
			name: "no_imps",
			fields: fields{
				requestBuilder: singleRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						rawRequest: nil,
					},
					imps: nil,
				},
			},
			want: want{
				requestData: nil,
				errs:        []error{util.NewBadInputError("%s", util.ErrImpMissing.Error())},
			},
		},
		{
			name: "replace_macros_to_form_endpoint_url",
			fields: fields{
				requestBuilder: singleRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						hasMacrosInEndpoint: true,
						rawRequest:          json.RawMessage(`{"imp":[{"ext":{"bidder":{"ext":{"pubid":5890},"host":"localhost.com"}},"id":"imp_1"}]}`),
						endpointTemplate:    template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher/{{.ext.pubid}}`)),
					},
					newRequest: make(map[string]any),
					imps: []map[string]any{
						{
							"ext": map[string]any{
								"bidder": map[string]any{
									"ext":  map[string]any{"pubid": 5890},
									"host": "localhost.com",
								},
							},
							"id": "imp_1",
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
				requestBuilder: singleRequestBuilder{
					requestBuilderImpl: requestBuilderImpl{
						hasMacrosInEndpoint: true,
						rawRequest:          json.RawMessage(`{"imp":[{"ext":{},"id":"imp_1"}]}`),
						endpointTemplate:    template.Must(template.New("endpointTemplate").Option("missingkey=default").Parse(`http://{{.host}}/publisher/{{.pubid}}`)),
					},
					newRequest: make(map[string]any),
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
				requestBuilder: singleRequestBuilder{
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
					newRequest: make(map[string]any),
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
				errs: []error{util.NewBadInputError("failed to replace macros in endpoint, err:template: endpointTemplate:1:2: " +
					"executing \"endpointTemplate\" at <errorFunc>: error calling errorFunc: intentional error")},
			},
		},
		{
			name: "multi_imps_request",
			fields: fields{
				requestBuilder: singleRequestBuilder{
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
					newRequest: make(map[string]any),
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
						Body:   json.RawMessage(`{"device":{"pubid":1111},"host":"imp1.host.com","imp":[{"ext":{"bidder":{}},"id":"imp_1"},{"ext":{"bidder":{}},"id":"imp_2"}]}`),
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
			if tt.fields.requestBuilder.newRequest != nil {
				tt.fields.requestBuilder.newRequest[impKey] = tt.fields.requestBuilder.imps
			}
			requestData, errs := tt.fields.requestBuilder.makeRequest()
			assert.Equalf(t, tt.want.requestData, requestData, "mismatched requestData")
			assert.Equalf(t, tt.want.errs, errs, "mismatched errs")
		})
	}
}
