package tagbidder

import (
	"testing"

	"github.com/PubMatic-OpenWrap/etree"

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

func TestGetBidDuration(t *testing.T) {
	type args struct {
		version     string // vast version
		creativeTag string // ad element
	}
	type want struct {
		duration float32 // seconds  (will converted from string with format as  HH:MM:SS.mmm)
		err      string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		// {name: "no vast version", want: want{err: "Invalid vast version"}, args: args{version: ""}},
		// {name: "no ad element", want: want{err: "Invalid ad element"}, args: args{adEle: nil}},
		// {name: "invalid vast version", want: want{err: "Invalid vast version"}, args: args{version: "666666"}},
		// {name: "invalid ad element", want: want{err: "Invalid ad elementt"}, args: args{version: "2.0", adEle: etree.NewElement("some element")}},
		// // multiple ad elements
		// {name: "multiple ad elements", want: want{err: "multiple ad elements"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>00:00:05/Duration> </Linear> </InLine> </Ad><Ad id="2"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>00.00.34</Duration> </Linear> </InLine> </Ad> `)}},
		// {name: "duration attrib not present", want: want{err: "duration is missing"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> </Linear> </InLine> </Ad>`)}}, // mandatory as per https://iabtechlab.com/wp-content/uploads/2016/04/VAST-2_0-FINAL.pdf
		// {name: "duration attrib present", want: want{duration: 9}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>00:00:09</Duration> </Linear> </InLine> </Ad>`)}},
		// {name: "duration attrib all upper", want: want{duration: 50}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <DURATION>00:00:50</DURATION> </Linear> </InLine> </Ad>`)}},
		// {name: "duration 00:01:08 (1 min 8 seconds = 68 seconds)", want: want{duration: 68}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>00:01:08</Duration> </Linear> </InLine> </Ad>`)}},
		{name: "duration 02:13:12 (2 hrs 13 min  12 seconds) = 7992 seconds)", want: want{duration: 7992}, args: args{version: "2.0", creativeTag: `<Creative sequence="1"> <Linear> <Duration>02:13:12</Duration> </Linear> </Creative>`}},
		// {name: "invalid duration 3:13:900 (3 hrs 13 min  900 seconds) = ?? )", want: want{err: "Invalid duration"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>3:13:900</Duration> </Linear> </InLine> </Ad>`)}},
		// {name: "invalid duration 3:13:34:44 (3 hrs 13 min  900 seconds) = ?? )", want: want{err: "Invalid duration"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>3:13:34:44</Duration> </Linear> </InLine> </Ad>`)}},
		// // 1 millsecond = 0.001 sec
		// // hence 45 seconds + 0.000038 seconds = 45.000038 seconds
		// {name: "duration = 0:0:45.038 , with milliseconds duration (0 hrs 0 min 45 seconds and 0.038 millseconds) = 45.000038 seconds )", want: want{duration: 45.000038}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>0:0:45.038</Duration> </Linear> </InLine>`)}},
		// {name: "duration = 56 (ambiguity w.r.t. HH:MM:SS.mmm format) ", want: want{err: "Invalid duration"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>56</Duration> </Linear> </InLine> </Ad>`)}},
		// {name: "duration = :56 (ambiguity w.r.t. HH:MM:SS.mmm format) ", want: want{err: "Invalid duration"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>:56</Duration> </Linear> </InLine> </Ad> `)}},
		// {name: "duration = :56: (ambiguity w.r.t. HH:MM:SS.mmm format) ", want: want{err: "Invalid duration"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>:56:</Duration> </Linear> </InLine> </Ad>`)}},
		// {name: "duration = ::56 (ambiguity w.r.t. HH:MM:SS.mmm format) ", want: want{err: "Invalid duration"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>::56</Duration> </Linear> </InLine> </Ad>`)}},
		// {name: "duration = 56.445 (ambiguity w.r.t. HH:MM:SS.mmm format) ", want: want{err: "Invalid duration"}, args: args{version: "2.0", adEle: etree.NewElement(`<Ad id="1"> <InLine> <Creatives> <Creative sequence="1"> <Linear> <Duration>56.445</Duration> </Linear> </InLine> </Ad> `)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := etree.NewDocument()
			doc.ReadFromString(tt.args.creativeTag)
			dur, err := getBidDuration(tt.args.version, doc.FindElement("./Creative"))
			assert.Equal(t, tt.want.duration, dur)
			assert.Equal(t, tt.want.err, err)
		})
	}
}
