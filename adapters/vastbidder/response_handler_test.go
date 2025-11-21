package vastbidder

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func getJSONRawMessage(obj interface{}) json.RawMessage {
	response, _ := json.Marshal(obj)
	return response[:]
}

func TestVASTTagResponseHandler_getBidResponse(t *testing.T) {
	type args struct {
		internalRequest *openrtb2.BidRequest
		externalRequest *adapters.RequestData
		response        *adapters.ResponseData
		vastTag         *openrtb_ext.ExtImpVASTBidderTag
		parser          xmlParser
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
							Ext: getJSONRawMessage(&adapters.ExtImpBidder{
								Bidder: getJSONRawMessage(&openrtb_ext.ExtImpVASTBidder{
									Tags: []*openrtb_ext.ExtImpVASTBidderTag{
										{
											TagID:    `101`,
											Duration: 15,
										},
									},
								}),
							}),
						},
					},
				},
				externalRequest: &adapters.RequestData{
					Params: &adapters.BidRequestParams{
						ImpIndex:     0,
						VASTTagIndex: 0,
					},
				},
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`<VAST version="2.0"> <Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <MediaFiles> <MediaFile><![CDATA[ad.mp4]]></MediaFile> </MediaFiles> </Linear> </Creative> </Creatives> <Extensions> <Extension type="LR-Pricing"> <Price model="CPM" currency="USD"><![CDATA[0.05]]></Price> </Extension> </Extensions> </InLine> </Ad> </VAST>`),
				},
				vastTag: &openrtb_ext.ExtImpVASTBidderTag{
					TagID:    "101",
					Duration: 15,
				},
				parser: newFastXMLParser(),
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
			handler := newResponseHandler(tt.args.internalRequest, tt.args.externalRequest, tt.args.response, tt.args.parser)
			generateRandomID = func() string {
				return `1234`
			}

			errs := handler.Validate()
			assert.Nil(t, errs)
			bidderResponse, err := handler.MakeBids()
			assert.Equal(t, tt.want.bidderResponse, bidderResponse)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

var (
	internalRequest *openrtb2.BidRequest
	externalRequest *adapters.RequestData
	response        *adapters.ResponseData
)

func init() {
	internalRequest = &openrtb2.BidRequest{
		ID: `request_id_1`,
		Imp: []openrtb2.Imp{
			{
				ID: `imp_id_1`,
				Ext: getJSONRawMessage(&adapters.ExtImpBidder{
					Bidder: getJSONRawMessage(&openrtb_ext.ExtImpVASTBidder{
						Tags: []*openrtb_ext.ExtImpVASTBidderTag{
							{
								TagID:    `101`,
								Duration: 15,
							},
						},
					}),
				}),
			},
		},
	}
	externalRequest = &adapters.RequestData{
		Params: &adapters.BidRequestParams{
			ImpIndex:     0,
			VASTTagIndex: 0,
		},
	}
	response = &adapters.ResponseData{
		StatusCode: http.StatusOK,
		Body:       []byte(`<VAST version="2.0"> <Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <MediaFiles> <MediaFile><![CDATA[ad.mp4]]></MediaFile> </MediaFiles> </Linear> </Creative> </Creatives> <Extensions> <Extension type="LR-Pricing"> <Price model="CPM" currency="USD"><![CDATA[0.05]]></Price> </Extension> </Extensions> </InLine> </Ad> </VAST>`),
	}
}

func BenchmarkFastXMLParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		handler := newResponseHandler(internalRequest, externalRequest, response, newFastXMLParser())
		handler.Validate()
		handler.MakeBids()
	}
}

func BenchmarkETreeParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		handler := newResponseHandler(internalRequest, externalRequest, response, newETreeXMLParser())
		handler.Validate()
		handler.MakeBids()
	}
}
