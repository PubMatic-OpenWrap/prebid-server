package vastbidder

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestVASTTagResponseHandler_getBidResponse(t *testing.T) {
	type args struct {
		internalRequest *openrtb2.BidRequest
		externalRequest *adapters.RequestData
		response        *adapters.ResponseData
		vastTag         *openrtb_ext.ExtImpVASTBidderTag
	}
	type want struct {
		bidderResponse *adapters.BidderResponse
		err            []error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: `InlinePricingNode`,
			args: args{
				internalRequest: &openrtb2.BidRequest{
					ID: `request_id_1`,
					Imp: []openrtb2.Imp{
						{
							ID: `imp_id_1`,
						},
					},
				},
				externalRequest: &adapters.RequestData{
					Params: &adapters.BidRequestParams{
						ImpIndex: 0,
					},
				},
				response: &adapters.ResponseData{
					Body: []byte(`<VAST version="2.0"> <Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <MediaFiles> <MediaFile><![CDATA[ad.mp4]]></MediaFile> </MediaFiles> </Linear> </Creative> </Creatives> <Extensions> <Extension type="LR-Pricing"> <Price model="CPM" currency="USD"><![CDATA[0.05]]></Price> </Extension> </Extensions> </InLine> </Ad> </VAST>`),
				},
				vastTag: &openrtb_ext.ExtImpVASTBidderTag{
					TagID:    "101",
					Duration: 15,
				},
			},
			want: want{
				bidderResponse: &adapters.BidderResponse{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    `1234`,
								ImpID: `imp_id_1`,
								Price: 0.05,
								AdM:   `<VAST version="2.0"> <Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <MediaFiles> <MediaFile><![CDATA[ad.mp4]]></MediaFile> </MediaFiles> </Linear> </Creative> </Creatives> <Extensions> <Extension type="LR-Pricing"> <Price model="CPM" currency="USD"><![CDATA[0.05]]></Price> </Extension> </Extensions> </InLine> </Ad> </VAST>`,
								CrID:  "cr_1234",
								Ext:   json.RawMessage(`{"prebid":{"type":"video","video":{"duration":15,"primary_category":"","vasttagid":"101"}}}`),
							},
							BidType: openrtb_ext.BidTypeVideo,
							BidVideo: &openrtb_ext.ExtBidPrebidVideo{
								VASTTagID: "101",
								Duration:  15,
							},
						},
					},
					Currency: `USD`,
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewVASTTagResponseHandler()
			GetRandomID = func() string {
				return `1234`
			}
			handler.VASTTag = tt.args.vastTag

			bidderResponse, err := handler.getBidResponse(tt.args.internalRequest, tt.args.externalRequest, tt.args.response)
			assert.Equal(t, tt.want.bidderResponse, bidderResponse)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

var (
	fastXMLResponseHandler *responseHandler
	etreeResponseHandler   *responseHandler
)

func init() {
	vastTagResponseHandler = NewVASTTagResponseHandler()
	vastTagResponseHandler.VASTTag = &openrtb_ext.ExtImpVASTBidderTag{TagID: "101", Duration: 15}
	internalRequest := &openrtb2.BidRequest{ID: `request_id_1`, Imp: []openrtb2.Imp{{ID: `imp_id_1`}}}
	externalRequest := &adapters.RequestData{Params: &adapters.BidRequestParams{ImpIndex: 0}}
	response := &adapters.ResponseData{
		Body: []byte(`<VAST version="2.0"> <Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <MediaFiles> <MediaFile><![CDATA[ad.mp4]]></MediaFile> </MediaFiles> </Linear> </Creative> </Creatives> <Extensions> <Extension type="LR-Pricing"> <Price model="CPM" currency="USD"><![CDATA[0.05]]></Price> </Extension> </Extensions> </InLine> </Ad> </VAST>`),
	}

	fastXMLResponseHandler = newResponseHandler(internalRequest, externalRequest, response)
	//fastXMLResponseHandler.vastTag
}

func BenchmarkGetBidResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fastXMLResponseHandler.Validate()
		fastXMLResponseHandler.MakeBids()
	}
}

func BenchmarkVASTTagToBidderResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		etreeResponseHandler.Validate()
		etreeResponseHandler.MakeBids()
	}
}
