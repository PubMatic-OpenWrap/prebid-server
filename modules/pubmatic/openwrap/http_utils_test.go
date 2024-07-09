package openwrap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestSetSecurityHeaders(t *testing.T) {
	type args struct {
		rw http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "valid security headers",
			args: args{
				&httptest.ResponseRecorder{},
			},
			want: http.Header{
				"X-Content-Type-Options":  []string{models.XContentTypeOptions},
				"X-Xss-Protection":        []string{models.XXSSProtection},
				"Content-Security-Policy": []string{models.ContentSecurityPolicy},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetSecurityHeaders(tt.args.rw)
			assert.Equal(t, tt.want, tt.args.rw.Header())
		})
	}
}

func TestSetResponseHeaders(t *testing.T) {
	type args struct {
		serverHandler http.Handler
	}
	tests := []struct {
		name            string
		args            args
		expectedStatus  int
		expectedBody    string
		expectedHeaders map[string]string
	}{
		{
			name: "valid headers",
			args: args{
				serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				}),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			expectedHeaders: map[string]string{
				"Content-Security-Policy": models.ContentSecurityPolicy,
				"X-Content-Type-Options":  models.XContentTypeOptions,
				"X-XSS-Protection":        models.XXSSProtection,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := SetResponseHeaders(tt.args.serverHandler)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}

			// Check the response headers
			for key, value := range tt.expectedHeaders {
				if rr.Header().Get(key) != value {
					t.Errorf("header %v = %v, want %v", key, rr.Header().Get(key), value)
				}
			}
		})
	}
}
