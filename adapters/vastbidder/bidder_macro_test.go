package vastbidder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/PubMatic-OpenWrap/openrtb"
)

func TestGetDefaultHeaders(t *testing.T) {
	type args struct {
		req *openrtb.BidRequest
	}
	type want struct {
		headers http.Header
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "check all default headers",
			args: args{req: &openrtb.BidRequest{
				Device: &openrtb.Device{
					IP:       "1.1.1.1",
					UA:       "user-agent",
					Language: "en",
				},
				Site: &openrtb.Site{
					Page: "http://test.com/",
				},
			}},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		},
		{
			name: "nil bid request",
			args: args{req: nil},
			want: want{
				headers: nil,
			},
		},
		{
			name: "no headers set",
			args: args{req: &openrtb.BidRequest{}},
			want: want{
				headers: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := new(BidderMacro)
			tag.Request = tt.args.req
			setDefaultHeaders(tag)
			assert.Equal(t, tt.want.headers, tag.impReqHeaders)
		})
	}
}

//TestGetAllHeaders verifies default and custom headers are returned
func TestGetAllHeaders(t *testing.T) {
	type args struct {
		req      *openrtb.BidRequest
		myBidder IBidderMacro
	}
	type want struct {
		headers http.Header
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Default and custom headers check",
			args: args{req: &openrtb.BidRequest{
				Device: &openrtb.Device{
					IP:       "1.1.1.1",
					UA:       "user-agent",
					Language: "en",
				},
				Site: &openrtb.Site{
					Page: "http://test.com/",
				},
			}, myBidder: NewMyVastBidderMacro()},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
					"my-custom-header":         []string{"some-value"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := tt.args.myBidder
			tag.(*MyVastBidderMacro).Request = tt.args.req
			allHeaders := tag.getAllHeaders()
			assert.Equal(t, tt.want.headers, allHeaders)
		})
	}
}

type MyVastBidderMacro struct {
	*BidderMacro
}

func NewMyVastBidderMacro() IBidderMacro {
	return &MyVastBidderMacro{
		BidderMacro: &BidderMacro{},
	}
}

func (tag *MyVastBidderMacro) GetHeaders() http.Header {
	h := http.Header{}
	h.Set("my-custom-header", "some-value")
	return h
}
