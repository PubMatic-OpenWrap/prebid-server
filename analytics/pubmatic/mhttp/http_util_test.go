package mhttp

import (
	"testing"

	// mock_pubmatic "github.com/prebid/prebid-server/analytics/pubmatic/mock"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpCall(t *testing.T) {
	type args struct {
		url      string
		postdata string
	}
	type want struct {
		method string
		err    bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "test GET method",
			args: args{
				url:      "http://t.pubmatic.com",
				postdata: "",
			},
			want: want{
				method: "GET",
			},
		},
		{
			name: "test POST method",
			args: args{
				url:      "http://t.pubmatic.com/wl",
				postdata: "any-data",
			},
			want: want{
				method: "POST",
			},
		},
		{
			name: "test invalid url",
			args: args{
				url:      "http://invalid-url param=12;",
				postdata: "any-data",
			},
			want: want{
				method: "POST",
				err:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, err := NewHttpCall(tt.args.url, tt.args.postdata)
			assert.NotNil(t, hc, tt.name)
			assert.Nil(t, err, tt.name)
			if !tt.want.err {
				assert.Equal(t, tt.want.method, hc.request.Method, tt.name)
			}
			assert.Equal(t, tt.want.err, hc.err != nil, tt.name)
		})
	}
}

func TestAddHeadersAndCookies(t *testing.T) {
	type args struct {
		headerKey, headerValue string
		cookieKey, cookieValue string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "test add headers and cookies",
			args: args{
				headerKey:   "header-key",
				headerValue: "header-val",
				cookieKey:   "cookie-key",
				cookieValue: "cookie-val",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc, _ := NewHttpCall("t.pubmatic.com", "")
			hc.AddHeader(tt.args.headerKey, tt.args.headerValue)
			hc.AddCookie(tt.args.cookieKey, tt.args.cookieValue)
			assert.Equal(t, tt.args.headerValue, hc.request.Header.Get(tt.args.headerKey), tt.name)
			cookie, _ := hc.request.Cookie(tt.args.cookieKey)
			assert.Equal(t, tt.args.cookieValue, cookie.Value, tt.name)
		})
	}
}

func TestNewMultiHttpContextAndAddHttpCall(t *testing.T) {
	mhc := NewMultiHttpContext()
	assert.NotNil(t, mhc)
	assert.Equal(t, mhc.hccount, 0)
	maxHttpCalls = 1
	mhc.AddHttpCall(&HttpCall{})
	mhc.AddHttpCall(&HttpCall{})
	mhc.AddHttpCall(&HttpCall{})
	assert.Equal(t, mhc.hccount, 1)
}

func TestInit(t *testing.T) {
	type args struct {
		maxClients                            int32
		maxConnections, maxCalls, respTimeout int
	}
	type want struct {
		maxHttpClients                   int32
		maxHttpConnections, maxHttpCalls int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid values in accepted range",
			args: args{
				maxClients:     1,
				maxConnections: 2,
				maxCalls:       3,
				respTimeout:    10,
			},
			want: want{
				maxHttpClients:     1,
				maxHttpConnections: 2,
				maxHttpCalls:       3,
			},
		},
		{
			name: "values not in the accepted range",
			args: args{
				maxClients:     11111,
				maxConnections: 2000,
				maxCalls:       300,
				respTimeout:    10000,
			},
			want: want{
				maxHttpClients:     10240,
				maxHttpConnections: 1024,
				maxHttpCalls:       200,
			},
		},
	}
	for _, tt := range tests {
		Init(tt.args.maxClients, tt.args.maxConnections, tt.args.maxCalls, tt.args.respTimeout)
		assert.Equal(t, tt.want.maxHttpClients, maxHttpClients, tt.name)
		assert.Equal(t, tt.want.maxHttpConnections, maxHttpConnections, tt.name)
		assert.Equal(t, tt.want.maxHttpCalls, maxHttpCalls, tt.name)
	}
}
