package wakanda

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestRequestSet(t *testing.T) {
	type args struct {
		request  *http.Request
		postData json.RawMessage
	}
	tests := []struct {
		name string
		args args
		want request
	}{
		{
			name: "empty_record",
			args: args{
				request:  nil,
				postData: nil,
			},
			want: request{},
		},
		{
			name: "valid_httprequest",
			args: args{
				request:  httptest.NewRequest("GET", "http://test.com/path?query=test_query", nil),
				postData: nil,
			},
			want: request{
				Method:      "GET",
				Protocol:    "HTTP/1.1",
				Host:        "test.com",
				Path:        "/path",
				QueryString: "query=test_query",
				Headers:     http.Header{},
			},
		},
		{
			name: "valid_httprequest_with_postdata",
			args: args{
				request:  httptest.NewRequest("GET", "http://test.com/path?query=test_query", nil),
				postData: json.RawMessage(`test_post_data`),
			},
			want: request{
				Method:      "GET",
				Protocol:    "HTTP/1.1",
				Host:        "test.com",
				Path:        "/path",
				QueryString: "query=test_query",
				Headers:     http.Header{},
				PostBody:    `test_post_data`,
			},
		},
		{
			name: "valid_httprequest_with_json_postdata",
			args: args{
				request: func() *http.Request {
					req := httptest.NewRequest("GET", "http://test.com/path?query=test_query", nil)
					req.Header.Set(contentType, contentTypeApplicationJSON)
					return req
				}(),
				postData: json.RawMessage(`{"test_post_data":1}`),
			},
			want: request{
				Method:      "GET",
				Protocol:    "HTTP/1.1",
				Host:        "test.com",
				Path:        "/path",
				QueryString: "query=test_query",
				Headers: http.Header{
					contentType: []string{contentTypeApplicationJSON},
				},
				PostJSON: json.RawMessage(`{"test_post_data":1}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r request
			r.set(tt.args.request, tt.args.postData)
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestResponseSetResponseRecorder(t *testing.T) {
	type args struct {
		resp *httptest.ResponseRecorder
	}
	tests := []struct {
		name string
		args args
		want response
	}{
		{
			name: "empty",
			args: args{},
			want: response{},
		},
		{
			name: "response_recorder",
			args: args{
				resp: &httptest.ResponseRecorder{
					Body: bytes.NewBuffer([]byte(`test_body`)),
					Code: 200,
					HeaderMap: http.Header{
						"x-header-1": []string{"val1"},
						"x-header-2": []string{"val2"},
					},
				},
			},
			want: response{
				StatusCode: 200,
				Headers: http.Header{
					"x-header-1": []string{"val1"},
					"x-header-2": []string{"val2"},
				},
				Response: `test_body`,
			},
		},
		{
			name: "response_recorder_application_json",
			args: args{
				resp: &httptest.ResponseRecorder{
					Body: bytes.NewBuffer([]byte(`{"test_body":1}`)),
					Code: 200,
					HeaderMap: http.Header{
						contentType:  []string{contentTypeApplicationJSON},
						"x-header-1": []string{"val1"},
					},
				},
			},
			want: response{
				StatusCode: 200,
				Headers: http.Header{
					contentType:  []string{contentTypeApplicationJSON},
					"x-header-1": []string{"val1"},
				},
				Body: json.RawMessage(`{"test_body":1}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r response
			r.setResponseRecorder(tt.args.resp)
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestResponseSetResponseWriter(t *testing.T) {
	type args struct {
		resp http.ResponseWriter
		body string
	}
	tests := []struct {
		name string
		args args
		want response
	}{
		{
			name: "empty",
			args: args{},
			want: response{},
		},
		{
			name: "response_writer",
			args: args{
				resp: &httptest.ResponseRecorder{
					HeaderMap: http.Header{
						"x-header-1": []string{"val1"},
						"x-header-2": []string{"val2"},
					},
				},
			},
			want: response{
				Headers: http.Header{
					"x-header-1": []string{"val1"},
					"x-header-2": []string{"val2"},
				},
			},
		},
		{
			name: "response_writer_with_body",
			args: args{
				resp: &httptest.ResponseRecorder{
					HeaderMap: http.Header{
						"x-header-1": []string{"val1"},
						"x-header-2": []string{"val2"},
					},
				},
				body: `test_body`,
			},
			want: response{
				Headers: http.Header{
					"x-header-1": []string{"val1"},
					"x-header-2": []string{"val2"},
				},
				Response: `test_body`,
			},
		},
		{
			name: "response_writer_with_json_body",
			args: args{
				resp: &httptest.ResponseRecorder{
					HeaderMap: http.Header{
						contentType:  []string{contentTypeApplicationJSON},
						"x-header-1": []string{"val1"},
					},
				},
				body: `{"test_body":1}`,
			},
			want: response{
				Headers: http.Header{
					contentType:  []string{contentTypeApplicationJSON},
					"x-header-1": []string{"val1"},
				},
				Body: json.RawMessage(`{"test_body":1}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r response
			r.setResponseWriter(tt.args.resp, tt.args.body)
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestNewLogRecord(t *testing.T) {
	type args struct {
		wD *DebugData
	}
	tests := []struct {
		name string
		args args
		want *logRecord
	}{
		{
			name: `sample`,
			args: args{
				wD: &DebugData{
					HTTPRequest:     httptest.NewRequest("GET", "http://test.com/path?query=test_query", nil),
					HTTPRequestBody: json.RawMessage(`test_post_data`),
					HTTPResponse: &httptest.ResponseRecorder{
						HeaderMap: http.Header{
							"x-header-1": []string{"val1"},
							"x-header-2": []string{"val2"},
						},
					},
					HTTPResponseBody: `test_body`,
					OpenRTB: &openrtb2.BidRequest{
						ID: "123",
					},
					PrebidHTTPRequest: httptest.NewRequest("GET", "http://test.com/path?query=test_query", nil),
					PrebidRequestBody: json.RawMessage(`test_body`),
					PrebidHTTPResponse: &httptest.ResponseRecorder{
						Body: bytes.NewBuffer([]byte(`test_body`)),
						Code: 200,
						HeaderMap: http.Header{
							"x-header-1": []string{"val1"},
							"x-header-2": []string{"val2"},
						},
					},
					Logger:     json.RawMessage(`{"sl":1}`),
					WinningBid: true,
				},
			},
			want: &logRecord{
				HTTPReq: request{
					Method:      "GET",
					Protocol:    "HTTP/1.1",
					Host:        "test.com",
					Path:        "/path",
					QueryString: "query=test_query",
					Headers:     http.Header{},
					PostBody:    `test_post_data`,
				},
				HTTPResp: response{
					Headers: http.Header{
						"x-header-1": []string{"val1"},
						"x-header-2": []string{"val2"},
					},
					Response: `test_body`,
				},
				OpenRTBRequest: &openrtb2.BidRequest{
					ID: "123",
				},
				PrebidReq: request{
					Method:      "GET",
					Protocol:    "HTTP/1.1",
					Host:        "test.com",
					Path:        "/path",
					QueryString: "query=test_query",
					Headers:     http.Header{},
					PostBody:    `test_body`,
				},
				PrebidResp: response{
					StatusCode: 200,
					Headers: http.Header{
						"x-header-1": []string{"val1"},
						"x-header-2": []string{"val2"},
					},
					Response: `test_body`,
				},
				Logger:     json.RawMessage(`{"sl":1}`),
				WinningBid: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLogRecord(tt.args.wD)
			assert.Equal(t, tt.want, got)
		})
	}
}
