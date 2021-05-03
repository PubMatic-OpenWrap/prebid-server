package vastbidder

import (
	"net/http"
	"testing"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

//TestMakeRequests verifies
// 1. default and custom headers are set
func TestMakeRequests(t *testing.T) {

	type args struct {
		customHeaders map[string]string
		req           *openrtb.BidRequest
	}
	type want struct {
		impIDReqHeaderMap map[string]http.Header
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "multi_impression_req",
			args: args{
				customHeaders: map[string]string{
					"my-custom-header": "custom-value",
				},
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
						{ // vast 2.0
							ID: "vast_2_0_imp_req",
							Video: &openrtb.Video{
								Protocols: []openrtb.Protocol{
									openrtb.ProtocolVAST20,
								},
							},
							Ext: []byte(`{"bidder" :{}}`),
						},
						{
							ID: "vast_4_0_imp_req",
							Video: &openrtb.Video{ // vast 4.0
								Protocols: []openrtb.Protocol{
									openrtb.ProtocolVAST40,
								},
							},
							Ext: []byte(`{"bidder" :{}}`),
						},
						{
							ID: "vast_2_0_4_0_wrapper_imp_req",
							Video: &openrtb.Video{ // vast 2 and 4.0 wrapper
								Protocols: []openrtb.Protocol{
									openrtb.ProtocolVAST40Wrapper,
									openrtb.ProtocolVAST20,
								},
							},
							Ext: []byte(`{"bidder" :{}}`),
						},
						{
							ID: "other_non_vast_protocol",
							Video: &openrtb.Video{ // DAAST 1.0
								Protocols: []openrtb.Protocol{
									openrtb.ProtocolDAAST10,
								},
							},
							Ext: []byte(`{"bidder" :{}}`),
						},
						{

							ID: "no_protocol_field_set",
							Video: &openrtb.Video{ // vast 2 and 4.0 wrapper
								Protocols: []openrtb.Protocol{},
							},
							Ext: []byte(`{"bidder" :{}}`),
						},
					},
				},
			},
			want: want{
				impIDReqHeaderMap: map[string]http.Header{
					"vast_2_0_imp_req": {
						"X-Forwarded-For":  []string{"1.1.1.1"},
						"User-Agent":       []string{"user-agent"},
						"My-Custom-Header": []string{"custom-value"},
					},
					"vast_4_0_imp_req": {
						"X-Device-Ip":              []string{"1.1.1.1"},
						"X-Device-User-Agent":      []string{"user-agent"},
						"X-Device-Referer":         []string{"http://test.com/"},
						"X-Device-Accept-Language": []string{"en"},
						"My-Custom-Header":         []string{"custom-value"},
					},
					"vast_2_0_4_0_wrapper_imp_req": {
						"X-Device-Ip":              []string{"1.1.1.1"},
						"X-Forwarded-For":          []string{"1.1.1.1"},
						"X-Device-User-Agent":      []string{"user-agent"},
						"User-Agent":               []string{"user-agent"},
						"X-Device-Referer":         []string{"http://test.com/"},
						"X-Device-Accept-Language": []string{"en"},
						"My-Custom-Header":         []string{"custom-value"},
					},
					"other_non_vast_protocol": {
						"My-Custom-Header": []string{"custom-value"},
					}, // no default headers expected
					"no_protocol_field_set": { // set all default headers
						"X-Device-Ip":              []string{"1.1.1.1"},
						"X-Forwarded-For":          []string{"1.1.1.1"},
						"X-Device-User-Agent":      []string{"user-agent"},
						"User-Agent":               []string{"user-agent"},
						"X-Device-Referer":         []string{"http://test.com/"},
						"X-Device-Accept-Language": []string{"en"},
						"My-Custom-Header":         []string{"custom-value"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bidderName := openrtb_ext.BidderName("myVastBidderMacro")
			RegisterNewBidderMacro(bidderName, func() IBidderMacro {
				return newMyVastBidderMacro(tt.args.customHeaders)
			})
			bidder := NewTagBidder(bidderName, config.Adapter{})
			reqData, err := bidder.MakeRequests(tt.args.req, nil)
			assert.Nil(t, err)
			for _, req := range reqData {
				impID := tt.args.req.Imp[req.Params.ImpIndex].ID
				expectedHeaders := tt.want.impIDReqHeaderMap[impID]
				assert.Equal(t, expectedHeaders, req.Headers, "test for - "+impID)
			}
		})
	}
}
