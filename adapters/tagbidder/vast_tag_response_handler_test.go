package tagbidder

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"

	"github.com/stretchr/testify/assert"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
)

func TestVASTTagResponseHandler_vastTagToBidderResponse(t *testing.T) {
	type args struct {
		internalRequest *openrtb.BidRequest
		externalRequest *adapters.RequestData
		response        *adapters.ResponseData
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
				internalRequest: &openrtb.BidRequest{
					ID: `request_id_1`,
					Imp: []openrtb.Imp{
						openrtb.Imp{
							ID: `imp_id_1`,
						},
					},
				},
				externalRequest: &adapters.RequestData{
					ImpIndex: 0,
				},
				response: &adapters.ResponseData{
					Body: []byte(`<VAST version="2.0"> <Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <MediaFiles> <MediaFile><![CDATA[ad.mp4]]></MediaFile> </MediaFiles> </Linear> </Creative> </Creatives> <Extensions> <Extension type="LR-Pricing"> <Price model="CPM" currency="USD"><![CDATA[0.05]]></Price> </Extension> </Extensions> </InLine> </Ad> </VAST>`),
				},
			},
			want: want{
				bidderResponse: &adapters.BidderResponse{
					Bids: []*adapters.TypedBid{
						&adapters.TypedBid{
							Bid: &openrtb.Bid{
								ID:    `1234`,
								ImpID: `imp_id_1`,
								Price: 0.05,
								AdM:   `<VAST version="2.0"> <Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <MediaFiles> <MediaFile><![CDATA[ad.mp4]]></MediaFile> </MediaFiles> </Linear> </Creative> </Creatives> <Extensions> <Extension type="LR-Pricing"> <Price model="CPM" currency="USD"><![CDATA[0.05]]></Price> </Extension> </Extensions> </InLine> </Ad> </VAST>`,
							},
							BidType: openrtb_ext.BidTypeVideo,
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
			handler := &VASTTagResponseHandler{}
			getRandomID = func() string {
				return `1234`
			}

			bidderResponse, err := handler.vastTagToBidderResponse(tt.args.internalRequest, tt.args.externalRequest, tt.args.response)
			assert.Equal(t, tt.want.bidderResponse, bidderResponse)
			assert.Equal(t, tt.want.err, err)
		})
	}
}
