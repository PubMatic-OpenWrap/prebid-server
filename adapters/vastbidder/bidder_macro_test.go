package vastbidder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/PubMatic-OpenWrap/openrtb"
)

//TestSetDefaultHeaders verifies SetDefaultHeaders
func TestSetDefaultHeaders(t *testing.T) {
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
				headers: http.Header{},
			},
		}, {
			name: "vast 4 protocol",
			args: args{
				req: &openrtb.BidRequest{
					Device: &openrtb.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb.Site{
						Page: "http://test.com/",
					},
					Imp: []openrtb.Imp{
						{
							Video: &openrtb.Video{
								Protocols: []openrtb.Protocol{
									openrtb.ProtocolVAST40,
									openrtb.ProtocolDAAST10,
								},
							},
						},
					},
				},
			},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		}, {
			name: "< vast 4",
			args: args{
				req: &openrtb.BidRequest{
					Device: &openrtb.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb.Site{
						Page: "http://test.com/",
					},
					Imp: []openrtb.Imp{
						{
							Video: &openrtb.Video{
								Protocols: []openrtb.Protocol{
									openrtb.ProtocolVAST20,
									openrtb.ProtocolDAAST10,
								},
							},
						},
					},
				},
			},
			want: want{
				headers: http.Header{
					"X-Forwarded-For": []string{"1.1.1.1"},
					"User-Agent":      []string{"user-agent"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := new(BidderMacro)
			tag.Request = tt.args.req
			if nil != tt.args.req && nil != tt.args.req.Imp && len(tt.args.req.Imp) > 0 {
				tag.Imp = &tt.args.req.Imp[0]
			}
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
			args: args{
				req: &openrtb.BidRequest{
					Device: &openrtb.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb.Site{
						Page: "http://test.com/",
					},
				},
				myBidder: newMyVastBidderMacro(map[string]string{
					"my-custom-header": "some-value",
				}),
			},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
					"My-Custom-Header":         []string{"some-value"},
				},
			},
		},
		{
			name: "override default header value",
			args: args{
				req: &openrtb.BidRequest{
					Site: &openrtb.Site{
						Page: "http://test.com/", // default header value
					},
				},
				myBidder: newMyVastBidderMacro(map[string]string{
					"X-Device-Referer": "my-custom-value",
				}),
			},
			want: want{
				headers: http.Header{
					// http://test.com/ is not expected here as value
					"X-Device-Referer": []string{"my-custom-value"},
				},
			},
		},
		{
			name: "no custom headers",
			args: args{
				req: &openrtb.BidRequest{
					Device: &openrtb.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb.Site{
						Page: "http://test.com/",
					},
				},
				myBidder: newMyVastBidderMacro(nil), // nil - no custom headers
			},
			want: want{
				headers: http.Header{ // expect default headers
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := tt.args.myBidder
			tag.(*myVastBidderMacro).Request = tt.args.req
			allHeaders := tag.getAllHeaders()
			assert.Equal(t, tt.want.headers, allHeaders)
		})
	}
}

type myVastBidderMacro struct {
	*BidderMacro
	customHeaders map[string]string
}

func newMyVastBidderMacro(customHeaders map[string]string) IBidderMacro {
	obj := &myVastBidderMacro{
		BidderMacro:   &BidderMacro{},
		customHeaders: customHeaders,
	}
	obj.IBidderMacro = obj
	return obj
}

func (tag *myVastBidderMacro) GetHeaders() http.Header {
	if nil == tag.customHeaders {
		return nil
	}
	h := http.Header{}
	for k, v := range tag.customHeaders {
		h.Set(k, v)
	}
	return h
}
