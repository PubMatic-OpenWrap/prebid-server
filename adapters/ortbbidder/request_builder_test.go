package ortbbidder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"text/template"

	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestBuilder(t *testing.T) {
	type args struct {
		requestType      string
		endpoint         string
		endpointTemplate *template.Template
		requestParams    map[string]bidderparams.BidderParamMapper
	}
	tests := []struct {
		name string
		args args
		want requestBuilder
	}{
		{
			name: "singlerequestType",
			args: args{
				requestType: "single",
				endpoint:    "http://localhost/publisher",
			},
			want: &singleRequestBuilder{
				requestBuilderImpl: requestBuilderImpl{
					endpoint: "http://localhost/publisher",
				},
			},
		},
		{
			name: "defaultrequestType",
			args: args{
				requestType: "",
				endpoint:    "http://localhost/publisher",
			},
			want: &singleRequestBuilder{
				requestBuilderImpl: requestBuilderImpl{
					endpoint: "http://localhost/publisher",
				},
			},
		},
		{
			name: "multirequestType",
			args: args{
				requestType: "multi",
				endpoint:    "http://{{.host}}/publisher",
			},
			want: &multiRequestBuilder{
				requestBuilderImpl: requestBuilderImpl{
					endpoint:            "http://{{.host}}/publisher",
					hasMacrosInEndpoint: true,
				},
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newRequestBuilder(tt.args.requestType, tt.args.endpoint, tt.args.endpointTemplate, tt.args.requestParams)
			assert.Equalf(t, tt.want, got, "mismathed requestbuilder")
		})
	}
}

func TestGetEndpoint(t *testing.T) {
	type fields struct {
		endpoint            string
		hasMacrosInEndpoint bool
		endpointTemplate    *template.Template
	}
	type args struct {
		bidderParams map[string]any
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
				endpointTemplate:    nil,
			},
			args: args{
				bidderParams: map[string]any{},
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
				endpointTemplate:    template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher`)),
			},
			args: args{
				bidderParams: map[string]any{},
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
				endpointTemplate:    template.Must(template.New("endpointTemplate").Parse(`http://{{.host}}/publisher`)),
			},
			args: args{
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
			args: args{
				bidderParams: map[string]any{},
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
			reqBuilder := &requestBuilderImpl{
				endpoint:            tt.fields.endpoint,
				hasMacrosInEndpoint: tt.fields.hasMacrosInEndpoint,
				endpointTemplate:    tt.fields.endpointTemplate,
			}
			endpoint, err := reqBuilder.getEndpoint(tt.args.bidderParams)
			assert.Equalf(t, tt.want.endpoint, endpoint, "mismatched endpoint")
			assert.Equalf(t, tt.want.err, err, "mismatched error")
		})
	}
}

func TestCloneRequest(t *testing.T) {
	type args struct {
		request json.RawMessage
	}
	type want struct {
		requestNode map[string]any
		err         error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "clone_request",
			args: args{
				request: json.RawMessage(`{"id":"reqId","imps":[{"id":"impId"}]}`),
			},
			want: want{
				requestNode: map[string]any{
					"id": "reqId",
					"imps": []any{
						map[string]any{
							"id": "impId",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requstNode, err := cloneRequest(tt.args.request)
			assert.Equal(t, tt.want.requestNode, requstNode, "mismatched requestnode")
			assert.Equal(t, tt.want.err, err, "mismatched error")
		})
	}
}

func TestAppendRequestData(t *testing.T) {
	type args struct {
		requestData []*adapters.RequestData
		request     map[string]any
		uri         string
	}
	type want struct {
		reqData []*adapters.RequestData
		err     error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "append_request_data_to_nil_object",
			args: args{
				requestData: nil,
				request: map[string]any{
					"id": "reqId",
				},
				uri: "http://endpoint.com",
			},
			want: want{
				reqData: []*adapters.RequestData{{
					Method: http.MethodPost,
					Uri:    "http://endpoint.com",
					Body:   []byte(`{"id":"reqId"}`),
					Headers: http.Header{
						"Content-Type": {"application/json;charset=utf-8"},
						"Accept":       {"application/json"},
					},
				}},
			},
		},
		{
			name: "append_request_data_to_non_empty_object",
			args: args{
				requestData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://endpoint.com",
						Body:   []byte(`{"id":"req_1"}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					},
				},
				request: map[string]any{
					"id": "req_2",
				},
				uri: "http://endpoint.com",
			},
			want: want{
				reqData: []*adapters.RequestData{
					{
						Method: http.MethodPost,
						Uri:    "http://endpoint.com",
						Body:   []byte(`{"id":"req_1"}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					}, {
						Method: http.MethodPost,
						Uri:    "http://endpoint.com",
						Body:   []byte(`{"id":"req_2"}`),
						Headers: http.Header{
							"Content-Type": {"application/json;charset=utf-8"},
							"Accept":       {"application/json"},
						},
					}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appendRequestData(tt.args.requestData, tt.args.request, tt.args.uri)
			assert.Equal(t, tt.want.reqData, got, "mismatched request-data")
			assert.Equal(t, tt.want.err, err, "mismatched error")
		})
	}
}
