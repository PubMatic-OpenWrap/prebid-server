package pubmatic

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/adapterstest"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestJsonSamples(t *testing.T) {
	bidder, buildErr := Builder(openrtb_ext.BidderPubmatic, config.Adapter{
		Endpoint: "https://hbopenbid.pubmatic.com/translator?source=prebid-server"}, config.Server{ExternalUrl: "http://hosturl.com", GvlID: 1, DataCenter: "2"})

	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}

	adapterstest.RunJSONBidderTest(t, "pubmatictest", bidder)
}

func TestGetBidTypeVideo(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 1
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeVideo {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeVideo, actualBidTypeValue)
	}
}

func TestGetBidTypeForMissingBidTypeExt(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	actualBidTypeValue := getBidType(pubmaticExt)
	// banner is the default bid type when no bidType key is present in the bid.ext
	if actualBidTypeValue != "banner" {
		t.Errorf("Expected Bid Type value was: banner, actual value is: %v", actualBidTypeValue)
	}
}

func TestGetBidTypeBanner(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 0
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeBanner {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeBanner, actualBidTypeValue)
	}
}

func TestGetBidTypeNative(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 2
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeNative {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeNative, actualBidTypeValue)
	}
}

func TestGetBidTypeForUnsupportedCode(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 99
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeBanner {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeBanner, actualBidTypeValue)
	}
}

func TestParseImpressionObject(t *testing.T) {
	type args struct {
		imp                      *openrtb2.Imp
		extractWrapperExtFromImp bool
		extractPubIDFromImp      bool
		displayManager           string
		displayManagerVer        string
	}
	type want struct {
		bidfloor          float64
		impExt            json.RawMessage
		displayManager    string
		displayManagerVer string
	}
	tests := []struct {
		name                string
		args                args
		expectedWrapperExt  *pubmaticWrapperExt
		expectedPublisherId string
		want                want
		wantErr             bool
	}{
		{
			name: "imp.bidfloor empty and kadfloor set",
			args: args{
				imp: &openrtb2.Imp{
					Video: &openrtb2.Video{},
					Ext:   json.RawMessage(`{"bidder":{"kadfloor":"0.12"}}`),
				},
			},
			want: want{
				bidfloor: 0.12,
				impExt:   json.RawMessage(nil),
			},
		},
		{
			name: "imp.bidfloor set and kadfloor empty",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{}}`),
				},
			},
			want: want{
				bidfloor: 0.12,
				impExt:   json.RawMessage(nil),
			},
		},
		{
			name: "imp.bidfloor set and kadfloor invalid",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":"aaa"}}`),
				},
			},
			want: want{
				bidfloor: 0.12,
				impExt:   json.RawMessage(nil),
			},
		},
		{
			name: "imp.bidfloor set and kadfloor set, higher imp.bidfloor",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":"0.11"}}`),
				},
			},
			want: want{
				bidfloor: 0.12,
				impExt:   json.RawMessage(nil),
			},
		},
		{
			name: "imp.bidfloor set and kadfloor set, higher kadfloor",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":"0.13"}}`),
				},
			},
			want: want{
				bidfloor: 0.13,
				impExt:   json.RawMessage(nil),
			},
		},
		{
			name: "kadfloor string set with whitespace",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":" \t  0.13  "}}`),
				},
			},
			want: want{
				bidfloor: 0.13,
				impExt:   json.RawMessage(nil),
			},
		},
		{
			name: "bidViewability Object is set in imp.ext.prebid.pubmatic, pass to imp.ext",
			args: args{
				imp: &openrtb2.Imp{
					Video: &openrtb2.Video{},
					Ext:   json.RawMessage(`{"bidder":{"bidViewability":{"adSizes":{"728x90":{"createdAt":1679993940011,"rendered":20,"totalViewTime":424413,"viewed":17}},"adUnit":{"createdAt":1679993940011,"rendered":25,"totalViewTime":424413,"viewed":17}}}}`),
				},
			},
			want: want{
				impExt: json.RawMessage(`{"bidViewability":{"adSizes":{"728x90":{"createdAt":1679993940011,"rendered":20,"totalViewTime":424413,"viewed":17}},"adUnit":{"createdAt":1679993940011,"rendered":25,"totalViewTime":424413,"viewed":17}}}`),
			},
		},
		{
			name: "Populate imp.displaymanager and imp.displaymanagerver if both are empty in imp",
			args: args{
				imp: &openrtb2.Imp{
					Video: &openrtb2.Video{},
					Ext:   json.RawMessage(`{"bidder":{"kadfloor":"0.12"}}`),
				},
				displayManager:    "prebid-mobile",
				displayManagerVer: "1.0.0",
			},
			want: want{
				bidfloor:          0.12,
				impExt:            json.RawMessage(nil),
				displayManager:    "prebid-mobile",
				displayManagerVer: "1.0.0",
			},
		},
		{
			name: "do not populate imp.displaymanager and imp.displaymanagerver in imp if only displaymanager or displaymanagerver is present in args",
			args: args{
				imp: &openrtb2.Imp{
					Video:             &openrtb2.Video{},
					Ext:               json.RawMessage(`{"bidder":{"kadfloor":"0.12"}}`),
					DisplayManagerVer: "1.0.0",
				},
				displayManager:    "prebid-mobile",
				displayManagerVer: "1.0.0",
			},
			want: want{
				bidfloor:          0.12,
				impExt:            json.RawMessage(nil),
				displayManagerVer: "1.0.0",
			},
		},
		{
			name: "do not populate imp.displaymanager and imp.displaymanagerver if already present in imp",
			args: args{
				imp: &openrtb2.Imp{
					Video:             &openrtb2.Video{},
					Ext:               json.RawMessage(`{"bidder":{"kadfloor":"0.12"}}`),
					DisplayManager:    "prebid-mobile",
					DisplayManagerVer: "1.0.0",
				},
				displayManager:    "prebid-android",
				displayManagerVer: "2.0.0",
			},
			want: want{
				bidfloor:          0.12,
				impExt:            json.RawMessage(nil),
				displayManager:    "prebid-mobile",
				displayManagerVer: "1.0.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedWrapperExt, receivedPublisherId, _, err := parseImpressionObject(tt.args.imp, tt.args.extractWrapperExtFromImp, tt.args.extractPubIDFromImp, tt.args.displayManager, tt.args.displayManagerVer)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expectedWrapperExt, receivedWrapperExt)
			assert.Equal(t, tt.expectedPublisherId, receivedPublisherId)
			assert.Equal(t, tt.want.bidfloor, tt.args.imp.BidFloor)
			assert.Equal(t, tt.want.impExt, tt.args.imp.Ext)
			assert.Equal(t, tt.want.displayManager, tt.args.imp.DisplayManager)
			assert.Equal(t, tt.want.displayManagerVer, tt.args.imp.DisplayManagerVer)
		})
	}
}

func TestExtractPubmaticExtFromRequest(t *testing.T) {
	type args struct {
		request *openrtb2.BidRequest
	}
	tests := []struct {
		name           string
		args           args
		expectedReqExt extRequestAdServer
		expectedCookie []string
		wantErr        bool
	}{
		{
			name: "nil request",
			args: args{
				request: nil,
			},
			wantErr: false,
		},
		{
			name: "nil req.ext",
			args: args{
				request: &openrtb2.BidRequest{Ext: nil},
			},
			wantErr: false,
		},
		{
			name: "Pubmatic wrapper ext missing/empty (empty bidderparms)",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{}}}`),
				},
			},
			expectedReqExt: extRequestAdServer{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						BidderParams: json.RawMessage("{}"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Only Pubmatic wrapper ext present",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":{"profile":123,"version":456}}}}`),
				},
			},
			expectedReqExt: extRequestAdServer{
				Wrapper: &pubmaticWrapperExt{ProfileID: 123, VersionID: 456},
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						BidderParams: json.RawMessage(`{"wrapper":{"profile":123,"version":456}}`),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Pubmatic wrapper ext",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":"}}}`),
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Pubmatic acat ext",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"acat":[" drg \t","dlu","ssr"],"wrapper":{"profile":123,"version":456}}}}`),
				},
			},
			expectedReqExt: extRequestAdServer{
				Wrapper: &pubmaticWrapperExt{ProfileID: 123, VersionID: 456},
				Acat:    []string{"drg", "dlu", "ssr"},
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						BidderParams: json.RawMessage(`{"acat":[" drg \t","dlu","ssr"],"wrapper":{"profile":123,"version":456}}`),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Pubmatic acat ext",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"acat":[1,3,4],"wrapper":{"profile":123,"version":456}}}}`),
				},
			},
			expectedReqExt: extRequestAdServer{
				Wrapper: &pubmaticWrapperExt{ProfileID: 123, VersionID: 456},
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						BidderParams: json.RawMessage(`{"acat":[1,3,4],"wrapper":{"profile":123,"version":456}}`),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Pubmatic marketplace ext",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":["groupm"]}}},"bidderparams":{"wrapper":{"profile":123,"version":456}}}}`),
				},
			},
			expectedReqExt: extRequestAdServer{
				Marketplace: &marketplaceReqExt{AllowedBidders: []string{"pubmatic", "groupm"}},
				Wrapper:     &pubmaticWrapperExt{ProfileID: 123, VersionID: 456},
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						BidderParams:         json.RawMessage(`{"wrapper":{"profile":123,"version":456}}`),
						AlternateBidderCodes: &openrtb_ext.ExtAlternateBidderCodes{Enabled: true, Bidders: map[string]openrtb_ext.ExtAdapterAlternateBidderCodes{"pubmatic": {Enabled: true, AllowedBidderCodes: []string{"groupm"}}}},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReqExt, gotCookie, err := extractPubmaticExtFromRequest(tt.args.request)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expectedReqExt, gotReqExt)
			assert.Equal(t, tt.expectedCookie, gotCookie)
		})
	}
}

func TestPubmaticAdapter_MakeRequests(t *testing.T) {
	type fields struct {
		URI string
	}
	type args struct {
		request *openrtb2.BidRequest
		reqInfo *adapters.ExtraRequestInfo
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		expectedReqData []*adapters.RequestData
		wantErr         bool
	}{
		// Happy paths covered by TestJsonSamples()
		// Covering only error scenarios here
		{
			name: "invalid bidderparams",
			args: args{
				request: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":"123"}}}`)},
			},
			wantErr: true,
		},
		{
			name: "request with multi floor",
			fields: fields{
				URI: "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{"bidder":{"floors":[1.2,1.3,1.4]}}`),
						},
					},
				},
			},
			expectedReqData: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":1.2}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr","publisher":{}},"ext":{"prebid":{}}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
					},
					ImpIDs: []string{"test-imp-id_mf1"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf2","banner":{"w":300,"h":250},"bidfloor":1.3}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr","publisher":{}},"ext":{"prebid":{}}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
					},
					ImpIDs: []string{"test-imp-id_mf2"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf3","banner":{"w":300,"h":250},"bidfloor":1.4}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr","publisher":{}},"ext":{"prebid":{}}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
					},
					ImpIDs: []string{"test-imp-id_mf3"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PubmaticAdapter{
				URI: tt.fields.URI,
			}
			gotReqData, gotErr := a.MakeRequests(tt.args.request, tt.args.reqInfo)
			assert.Equal(t, tt.wantErr, len(gotErr) != 0)
			assert.Equal(t, tt.expectedReqData, gotReqData)
		})
	}
}

func TestPubmaticAdapter_MakeBids(t *testing.T) {
	type fields struct {
		URI string
	}
	type args struct {
		internalRequest *openrtb2.BidRequest
		externalRequest *adapters.RequestData
		response        *adapters.ResponseData
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  []error
		wantResp *adapters.BidderResponse
	}{
		{
			name: "happy path, valid response with all bid params",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id": "test-request-id", "seatbid":[{"seat": "958", "bid":[{"id": "7706636740145184841", "impid": "test-imp-id", "price": 0.500000, "adid": "29681110", "adm": "some-test-ad", "adomain":["pubmatic.com"], "crid": "29681110", "h": 250, "w": 300, "dealid": "testdeal", "ext":{"dspid": 6, "deal_channel": 1, "prebiddealpriority": 1}}], "ext": {"buyid": "testBuyId"}}], "bidid": "5778926625248726496", "cur": "USD"}`),
				},
				externalRequest: &adapters.RequestData{BidderName: openrtb_ext.BidderPubmatic},
			},
			wantErr: nil,
			wantResp: &adapters.BidderResponse{
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:      "7706636740145184841",
							ImpID:   "test-imp-id",
							Price:   0.500000,
							AdID:    "29681110",
							AdM:     "some-test-ad",
							ADomain: []string{"pubmatic.com"},
							CrID:    "29681110",
							H:       250,
							W:       300,
							DealID:  "testdeal",
							Ext:     json.RawMessage(`{"buyid":"testBuyId","deal_channel":1,"dspid":6,"prebiddealpriority":1}`),
						},
						DealPriority: 1,
						BidType:      openrtb_ext.BidTypeBanner,
						BidVideo:     &openrtb_ext.ExtBidPrebidVideo{},
						BidTargets:   map[string]string{"hb_buyid_pubmatic": "testBuyId"},
						BidMeta: &openrtb_ext.ExtBidPrebidMeta{
							AdvertiserID: 958,
							AgencyID:     958,
							NetworkID:    6,
							DemandSource: "6",
							MediaType:    "banner",
						},
					},
				},
				Currency: "USD",
			},
		},
		{
			name: "ignore invalid prebiddealpriority",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id": "test-request-id", "seatbid":[{"seat": "958", "bid":[{"id": "7706636740145184841", "impid": "test-imp-id", "price": 0.500000, "adid": "29681110", "adm": "some-test-ad", "adomain":["pubmatic.com"], "crid": "29681110", "h": 250, "w": 300, "dealid": "testdeal", "ext":{"dspid": 6, "deal_channel": 1, "prebiddealpriority": -1}}]}], "bidid": "5778926625248726496", "cur": "USD"}`),
				},
				externalRequest: &adapters.RequestData{BidderName: openrtb_ext.BidderPubmatic},
			},
			wantErr: nil,
			wantResp: &adapters.BidderResponse{
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:      "7706636740145184841",
							ImpID:   "test-imp-id",
							Price:   0.500000,
							AdID:    "29681110",
							AdM:     "some-test-ad",
							ADomain: []string{"pubmatic.com"},
							CrID:    "29681110",
							H:       250,
							W:       300,
							DealID:  "testdeal",
							Ext:     json.RawMessage(`{"dspid": 6, "deal_channel": 1, "prebiddealpriority": -1}`),
						},
						BidType:    openrtb_ext.BidTypeBanner,
						BidVideo:   &openrtb_ext.ExtBidPrebidVideo{},
						BidTargets: map[string]string{},
						BidMeta: &openrtb_ext.ExtBidPrebidMeta{
							AdvertiserID: 958,
							AgencyID:     958,
							NetworkID:    6,
							DemandSource: "6",
							MediaType:    "banner",
						},
					},
				},
				Currency: "USD",
			},
		},
		{
			name: "BidExt Nil cases",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id": "test-request-id", "seatbid":[{"seat": "958", "bid":[{"id": "7706636740145184841", "impid": "test-imp-id", "price": 0.500000, "adid": "29681110", "adm": "some-test-ad", "adomain":["pubmatic.com"], "crid": "29681110", "h": 250, "w": 300, "dealid": "testdeal", "ext":null}]}], "bidid": "5778926625248726496", "cur": "USD"}`),
				},
				externalRequest: &adapters.RequestData{BidderName: openrtb_ext.BidderPubmatic},
			},
			wantErr: nil,
			wantResp: &adapters.BidderResponse{
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:      "7706636740145184841",
							ImpID:   "test-imp-id",
							Price:   0.500000,
							AdID:    "29681110",
							AdM:     "some-test-ad",
							ADomain: []string{"pubmatic.com"},
							CrID:    "29681110",
							H:       250,
							W:       300,
							DealID:  "testdeal",
							Ext:     json.RawMessage(`null`),
						},
						BidType:    openrtb_ext.BidTypeBanner,
						BidVideo:   &openrtb_ext.ExtBidPrebidVideo{},
						BidTargets: map[string]string{},
					},
				},
				Currency: "USD",
			},
		},
		{
			name: "response impid has suffix(mf1)",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id": "test-request-id", "seatbid":[{"seat": "958", "bid":[{"id": "7706636740145184841", "impid": "test-imp-id_mf1", "price": 0.500000, "adid": "29681110", "adm": "some-test-ad", "adomain":["pubmatic.com"], "crid": "29681110", "h": 250, "w": 300, "dealid": "testdeal", "ext":null}]}], "bidid": "5778926625248726496", "cur": "USD"}`),
				},
				externalRequest: &adapters.RequestData{BidderName: openrtb_ext.BidderPubmatic},
			},
			wantErr: nil,
			wantResp: &adapters.BidderResponse{
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:      "7706636740145184841",
							ImpID:   "test-imp-id",
							Price:   0.500000,
							AdID:    "29681110",
							AdM:     "some-test-ad",
							ADomain: []string{"pubmatic.com"},
							CrID:    "29681110",
							H:       250,
							W:       300,
							DealID:  "testdeal",
							Ext:     json.RawMessage(`null`),
						},
						BidType:    openrtb_ext.BidTypeBanner,
						BidVideo:   &openrtb_ext.ExtBidPrebidVideo{},
						BidTargets: map[string]string{},
					},
				},
				Currency: "USD",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PubmaticAdapter{
				URI: tt.fields.URI,
			}
			gotResp, gotErr := a.MakeBids(tt.args.internalRequest, tt.args.externalRequest, tt.args.response)
			assert.Equal(t, tt.wantErr, gotErr, gotErr)
			assert.Equal(t, tt.wantResp, gotResp)
		})
	}
}

func Test_getAlternateBidderCodesFromRequest(t *testing.T) {
	type args struct {
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "request.ext nil",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: nil},
			},
			want: nil,
		},
		{
			name: "alternatebiddercodes not present in request.ext",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{}}`)},
			},
			want: nil,
		},
		{
			name: "alternatebiddercodes feature disabled",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":false,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":["groupm"]}}}}}`)},
			},
			want: []string{"pubmatic"},
		},
		{
			name: "alternatebiddercodes disabled at bidder level",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":false,"allowedbiddercodes":["groupm"]}}}}}`)},
			},
			want: []string{"pubmatic"},
		},
		{
			name: "alternatebiddercodes list not defined",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true}}}}}`)},
			},
			want: []string{"all"},
		},
		{
			name: "wildcard in alternatebiddercodes list",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":["*"]}}}}}`)},
			},
			want: []string{"all"},
		},
		{
			name: "empty alternatebiddercodes list",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":[]}}}}}`)},
			},
			want: []string{"pubmatic"},
		},
		{
			name: "only groupm in alternatebiddercodes allowed",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":["groupm"]}}}}}`)},
			},
			want: []string{"pubmatic", "groupm"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqExt *openrtb_ext.ExtRequest
			if len(tt.args.bidRequest.Ext) > 0 {
				err := json.Unmarshal(tt.args.bidRequest.Ext, &reqExt)
				if err != nil {
					t.Errorf("getAlternateBidderCodesFromRequest() = %v", err)
				}
			}

			got := getAlternateBidderCodesFromRequestExt(reqExt)
			assert.ElementsMatch(t, got, tt.want, tt.name)
		})
	}
}

func TestPopulateFirstPartyDataImpAttributes(t *testing.T) {
	type args struct {
		data      json.RawMessage
		impExtMap map[string]interface{}
	}
	tests := []struct {
		name           string
		args           args
		expectedImpExt map[string]interface{}
	}{
		{
			name: "Only Targeting present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"sport":["rugby","cricket"]}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "sport=rugby,cricket",
			},
		},
		{
			name: "Targeting and adserver object present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"adserver": {"name": "gam","adslot": "/1111/home"},"pbadslot": "/2222/home","sport":["rugby","cricket"]}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"dfp_ad_unit_code": "/1111/home",
				"key_val":          "sport=rugby,cricket",
			},
		},
		{
			name: "Targeting and pbadslot key present in imp.ext.data ",
			args: args{
				data:      json.RawMessage(`{"pbadslot": "/2222/home","sport":["rugby","cricket"]}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"dfp_ad_unit_code": "/2222/home",
				"key_val":          "sport=rugby,cricket",
			},
		},
		{
			name: "Targeting and Invalid Adserver object in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"adserver": "invalid","sport":["rugby","cricket"]}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "sport=rugby,cricket",
			},
		},
		{
			name: "key_val already present in imp.ext.data",
			args: args{
				data: json.RawMessage(`{"sport":["rugby","cricket"]}`),
				impExtMap: map[string]interface{}{
					"key_val": "k1=v1|k2=v2",
				},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "k1=v1|k2=v2|sport=rugby,cricket",
			},
		},
		{
			name: "int data present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"age": 25}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "age=25",
			},
		},
		{
			name: "float data present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"floor": 0.15}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "floor=0.15",
			},
		},
		{
			name: "bool data present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"k1": true}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "k1=true",
			},
		},
		{
			name: "imp.ext.data is not present",
			args: args{
				data:      nil,
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{},
		},
		{
			name: "string with spaces present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"  category  ": "   cinema  "}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "category=cinema",
			},
		},
		{
			name: "string array with spaces present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"  country\t": ["  India", "\tChina  "]}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "country=India,China",
			},
		},
		{
			name: "Invalid data present in imp.ext.data",
			args: args{
				data:      json.RawMessage(`{"country": [1, "India"],"category":"movies"}`),
				impExtMap: map[string]interface{}{},
			},
			expectedImpExt: map[string]interface{}{
				"key_val": "category=movies",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			populateFirstPartyDataImpAttributes(tt.args.data, tt.args.impExtMap)
			assert.Equal(t, tt.expectedImpExt, tt.args.impExtMap)
		})
	}
}

func TestPopulateFirstPartyDataImpAttributesForMultipleAttributes(t *testing.T) {
	impExtMap := map[string]interface{}{
		"key_val": "k1=v1|k2=v2",
	}
	data := json.RawMessage(`{"sport":["rugby","cricket"],"pageType":"article","age":30,"floor":1.25}`)
	expectedKeyValArr := []string{"age=30", "floor=1.25", "k1=v1", "k2=v2", "pageType=article", "sport=rugby,cricket"}

	populateFirstPartyDataImpAttributes(data, impExtMap)

	//read dctr value and split on "|" for comparison
	actualKeyValArr := strings.Split(impExtMap[dctrKeyName].(string), "|")
	sort.Strings(actualKeyValArr)
	assert.Equal(t, expectedKeyValArr, actualKeyValArr)
}

func TestGetStringArray(t *testing.T) {
	tests := []struct {
		name   string
		input  []interface{}
		output []string
	}{
		{
			name:   "Valid String Array",
			input:  append(make([]interface{}, 0), "hello", "world"),
			output: []string{"hello", "world"},
		},
		{
			name:   "Invalid String Array",
			input:  append(make([]interface{}, 0), 1, 2),
			output: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringArray(tt.input)
			assert.Equal(t, tt.output, got)
		})
	}
}

func TestGetMapFromJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  json.RawMessage
		output map[string]interface{}
	}{
		{
			name:  "Valid JSON",
			input: json.RawMessage(`{"buyid":"testBuyId"}`),
			output: map[string]interface{}{
				"buyid": "testBuyId",
			},
		},
		{
			name:   "Invalid JSON",
			input:  json.RawMessage(`{"buyid":}`),
			output: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMapFromJSON(tt.input)
			assert.Equal(t, tt.output, got)
		})
	}
}

func TestPubmaticAdapter_buildMultiFloorRequests(t *testing.T) {
	type fields struct {
		URI        string
		bidderName string
	}
	type args struct {
		request      *openrtb2.BidRequest
		impFloorsMap map[string][]float64
		cookies      []string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantRequestData []*adapters.RequestData
		wantError       []error
	}{
		{
			name: "request with single imp and single floor",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
					},
				},
				impFloorsMap: map[string][]float64{
					"test-imp-id": {1.2},
				},
				cookies: []string{"test-cookie"},
			},
			wantRequestData: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":1.2,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf1"},
				},
			},
			wantError: []error{},
		},
		{
			name: "request with single imp and two floors",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
					},
				},
				impFloorsMap: map[string][]float64{
					"test-imp-id": {1.2, 1.3},
				},
				cookies: []string{"test-cookie"},
			},
			wantRequestData: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":1.2,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf1"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf2","banner":{"w":300,"h":250},"bidfloor":1.3,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf2"},
				},
			},
			wantError: []error{},
		},
		{
			name: "request with single imp and max multi floors(3)",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
					},
				},
				impFloorsMap: map[string][]float64{
					"test-imp-id": {1.2, 1.3, 1.4},
				},
				cookies: []string{"test-cookie"},
			},
			wantRequestData: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":1.2,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf1"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf2","banner":{"w":300,"h":250},"bidfloor":1.3,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf2"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf3","banner":{"w":300,"h":250},"bidfloor":1.4,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf3"},
				},
			},
			wantError: []error{},
		},
		{
			name: "request with multiple imp and single floor",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
						{
							ID:       "test-imp-id2",
							BidFloor: 0.13,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
					},
				},
				impFloorsMap: map[string][]float64{
					"test-imp-id":  {1.2},
					"test-imp-id2": {1.3},
				},
				cookies: []string{"test-cookie"},
			},
			wantRequestData: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":1.2,"ext":{}},{"id":"test-imp-id2_mf1","banner":{"w":300,"h":250},"bidfloor":1.3,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf1", "test-imp-id2_mf1"},
				},
			},
			wantError: []error{},
		},
		{
			name: "request with multiple imp and 3 floors (max) for only one imp",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
						{
							ID:       "test-imp-id2",
							BidFloor: 0.34,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
					},
				},
				impFloorsMap: map[string][]float64{
					"test-imp-id": {1.2, 1.3, 1.4},
				},
				cookies: []string{"test-cookie"},
			},
			wantRequestData: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":1.2,"ext":{}},{"id":"test-imp-id2","banner":{"w":300,"h":250},"bidfloor":0.34,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf1", "test-imp-id2"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf2","banner":{"w":300,"h":250},"bidfloor":1.3,"ext":{}},{"id":"test-imp-id2","banner":{"w":300,"h":250},"bidfloor":0.34,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf2", "test-imp-id2"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf3","banner":{"w":300,"h":250},"bidfloor":1.4,"ext":{}},{"id":"test-imp-id2","banner":{"w":300,"h":250},"bidfloor":0.34,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf3", "test-imp-id2"},
				},
			},
			wantError: []error{},
		},
		{
			name: "request with multiple imp with 3 floors for one imp and 2 floors for another imp",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
						{
							ID:       "test-imp-id2",
							BidFloor: 0.34,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
					},
				},
				impFloorsMap: map[string][]float64{
					"test-imp-id":  {1.2, 1.3, 1.4},
					"test-imp-id2": {1.2, 1.3},
				},
				cookies: []string{"test-cookie"},
			},
			wantRequestData: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":1.2,"ext":{}},{"id":"test-imp-id2_mf1","banner":{"w":300,"h":250},"bidfloor":1.2,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf1", "test-imp-id2_mf1"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf2","banner":{"w":300,"h":250},"bidfloor":1.3,"ext":{}},{"id":"test-imp-id2_mf2","banner":{"w":300,"h":250},"bidfloor":1.3,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf2", "test-imp-id2_mf2"},
				},
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf3","banner":{"w":300,"h":250},"bidfloor":1.4,"ext":{}},{"id":"test-imp-id2","banner":{"w":300,"h":250},"bidfloor":0.34,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id_mf3", "test-imp-id2"},
				},
			},
			wantError: []error{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PubmaticAdapter{
				URI:        tt.fields.URI,
				bidderName: tt.fields.bidderName,
			}
			gotRequestData, gotError := a.buildMultiFloorRequests(tt.args.request, tt.args.impFloorsMap, tt.args.cookies)
			assert.Equal(t, tt.wantRequestData, gotRequestData)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestPubmaticAdapter_buildAdapterRequest(t *testing.T) {
	type fields struct {
		URI        string
		bidderName string
	}
	type args struct {
		request *openrtb2.BidRequest
		cookies []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*adapters.RequestData
		wantErr bool
	}{
		{
			name: "failed to marshal request",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{`),
				},
				cookies: []string{"test-cookie"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "request with single imp",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
							Ext: json.RawMessage(`{}`),
						},
					},
				},
				cookies: []string{"test-cookie"},
			},
			want: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id","banner":{"w":300,"h":250},"bidfloor":0.12,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id"},
				},
			},
			wantErr: false,
		},
		{
			name: "request with multiple imp",
			fields: fields{
				URI:        "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
				bidderName: "pubmatic",
			},
			args: args{
				request: &openrtb2.BidRequest{
					ID: "test-request-id",
					App: &openrtb2.App{
						Name:     "AutoScout24",
						Bundle:   "com.autoscout24",
						StoreURL: "https://play.google.com/store/apps/details?id=com.autoscout24&hl=fr",
					},
					Imp: []openrtb2.Imp{
						{
							ID:       "test-imp-id",
							BidFloor: 0.12,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
						},
						{
							ID:       "test-imp-id2",
							BidFloor: 0.34,
							Banner: &openrtb2.Banner{
								W: ptrutil.ToPtr[int64](300),
								H: ptrutil.ToPtr[int64](250),
							},
						},
					},
				},
				cookies: []string{"test-cookie"},
			},
			want: []*adapters.RequestData{
				{
					Method: "POST",
					Uri:    "https://hbopenbid.pubmatic.com/translator?source=prebid-server",
					Body:   []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id","banner":{"w":300,"h":250},"bidfloor":0.12},{"id":"test-imp-id2","banner":{"w":300,"h":250},"bidfloor":0.34}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
					Headers: http.Header{
						"Content-Type": []string{"application/json;charset=utf-8"},
						"Accept":       []string{"application/json"},
						"Cookie":       []string{"test-cookie"},
					},
					ImpIDs: []string{"test-imp-id", "test-imp-id2"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PubmaticAdapter{
				URI:        tt.fields.URI,
				bidderName: tt.fields.bidderName,
			}
			got, err := a.buildAdapterRequest(tt.args.request, tt.args.cookies)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildAdapterRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrimSuffixWithPattern(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "input string is empty",
			args: args{
				input: "",
			},
			want: "",
		},
		{
			name: "input string does not contain pattern",
			args: args{
				input: "div123456789",
			},
			want: "div123456789",
		},
		{
			name: "input string contains pattern",
			args: args{
				input: "div123456789_mf1",
			},
			want: "div123456789",
		},
		{
			name: "input string contains pattern at the end",
			args: args{
				input: "div123456789_mf1_mf2",
			},
			want: "div123456789",
		},
		{
			name: "input string contains pattern at the start",
			args: args{
				input: "mf1_mf2_div123456789",
			},
			want: "mf1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimSuffixWithPattern(tt.args.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDisplayManagerAndVer(t *testing.T) {
	type args struct {
		app *openrtb2.App
	}
	type want struct {
		displayManager    string
		displayManagerVer string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "request app object is not nil but app.ext has no source and version",
			args: args{

				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{}`),
				},
			},
			want: want{
				displayManager:    "",
				displayManagerVer: "",
			},
		},
		{
			name: "request app object is not nil and app.ext has source and version",
			args: args{

				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{"source":"prebid-mobile","version":"1.0.0"}`),
				},
			},
			want: want{
				displayManager:    "prebid-mobile",
				displayManagerVer: "1.0.0",
			},
		},
		{
			name: "request app object is not nil and app.ext.prebid has source and version",
			args: args{
				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{"prebid":{"source":"prebid-mobile","version":"1.0.0"}}`),
				},
			},
			want: want{
				displayManager:    "prebid-mobile",
				displayManagerVer: "1.0.0",
			},
		},
		{
			name: "request app object is not nil and app.ext has only version",
			args: args{
				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{"version":"1.0.0"}`),
				},
			},
			want: want{
				displayManager:    "",
				displayManagerVer: "",
			},
		},
		{
			name: "request app object is not nil and app.ext has only source",
			args: args{
				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{"source":"prebid-mobile"}`),
				},
			},
			want: want{
				displayManager:    "",
				displayManagerVer: "",
			},
		},
		{
			name: "request app object is not nil and app.ext have empty source but version is present",
			args: args{
				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{"source":"", "version":"1.0.0"}`),
				},
			},
			want: want{
				displayManager:    "",
				displayManagerVer: "",
			},
		},
		{
			name: "request app object is not nil and app.ext have empty version but source is present",
			args: args{
				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{"source":"prebid-mobile", "version":""}`),
				},
			},
			want: want{
				displayManager:    "",
				displayManagerVer: "",
			},
		},
		{
			name: "request app object is not nil and both app.ext and app.ext.prebid have source and version",
			args: args{
				app: &openrtb2.App{
					Name: "AutoScout24",
					Ext:  json.RawMessage(`{"source":"prebid-mobile-android","version":"2.0.0","prebid":{"source":"prebid-mobile","version":"1.0.0"}}`),
				},
			},
			want: want{
				displayManager:    "prebid-mobile",
				displayManagerVer: "1.0.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayManager, displayManagerVer := getDisplayManagerAndVer(tt.args.app)
			assert.Equal(t, tt.want.displayManager, displayManager)
			assert.Equal(t, tt.want.displayManagerVer, displayManagerVer)
		})
	}
}
