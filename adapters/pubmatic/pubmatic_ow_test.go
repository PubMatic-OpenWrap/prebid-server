package pubmatic

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestGetAdServerTargetingForEmptyExt(t *testing.T) {
	ext := json.RawMessage(`{}`)
	targets := getTargetingKeys(ext, "pubmatic")
	// banner is the default bid type when no bidType key is present in the bid.ext
	if targets != nil && targets["hb_buyid_pubmatic"] != "" {
		t.Errorf("It should not contained AdserverTageting")
	}
}

func TestGetAdServerTargetingForValidExt(t *testing.T) {
	ext := json.RawMessage("{\"buyid\":\"testBuyId\"}")
	targets := getTargetingKeys(ext, "pubmatic")
	// banner is the default bid type when no bidType key is present in the bid.ext
	if targets == nil {
		t.Error("It should have targets")
		t.FailNow()
	}
	if targets != nil && targets["hb_buyid_pubmatic"] != "testBuyId" {
		t.Error("It should have testBuyId as targeting")
		t.FailNow()
	}
}

func TestGetAdServerTargetingForPubmaticAlias(t *testing.T) {
	ext := json.RawMessage("{\"buyid\":\"testBuyId-alias\"}")
	targets := getTargetingKeys(ext, "dummy-alias")
	// banner is the default bid type when no bidType key is present in the bid.ext
	if targets == nil {
		t.Error("It should have targets")
		t.FailNow()
	}
	if targets != nil && targets["hb_buyid_dummy-alias"] != "testBuyId-alias" {
		t.Error("It should have testBuyId as targeting")
		t.FailNow()
	}
}

func TestCopySBExtToBidExtWithBidExt(t *testing.T) {
	sbext := json.RawMessage("{\"buyid\":\"testBuyId\"}")
	bidext := json.RawMessage("{\"dspId\":\"9\"}")
	// expectedbid := json.RawMessage("{\"dspId\":\"9\",\"buyid\":\"testBuyId\"}")
	bidextnew := copySBExtToBidExt(sbext, bidext)
	if bidextnew == nil {
		t.Errorf("it should not be nil")
	}
}

func TestCopySBExtToBidExtWithNoBidExt(t *testing.T) {
	sbext := json.RawMessage("{\"buyid\":\"testBuyId\"}")
	bidext := json.RawMessage("{\"dspId\":\"9\"}")
	// expectedbid := json.RawMessage("{\"dspId\":\"9\",\"buyid\":\"testBuyId\"}")
	bidextnew := copySBExtToBidExt(sbext, bidext)
	if bidextnew == nil {
		t.Errorf("it should not be nil")
	}
}

func TestCopySBExtToBidExtWithNoSeatExt(t *testing.T) {
	bidext := json.RawMessage("{\"dspId\":\"9\"}")
	// expectedbid := json.RawMessage("{\"dspId\":\"9\",\"buyid\":\"testBuyId\"}")
	bidextnew := copySBExtToBidExt(nil, bidext)
	if bidextnew == nil {
		t.Errorf("it should not be nil")
	}
}

func TestPrepareMetaObject(t *testing.T) {
	typebanner := 0
	typevideo := 1
	typenative := 2
	typeinvalid := 233
	type args struct {
		bid    openrtb2.Bid
		bidExt *pubmaticBidExt
		seat   string
	}
	tests := []struct {
		name string
		args args
		want *openrtb_ext.ExtBidPrebidMeta
	}{
		{
			name: "Empty Meta Object and default BidType banner",
			args: args{
				bid: openrtb2.Bid{
					Cat: []string{},
				},
				bidExt: &pubmaticBidExt{},
				seat:   "",
			},
			want: &openrtb_ext.ExtBidPrebidMeta{
				MediaType: "banner",
			},
		},
		{
			name: "Valid Meta Object with Empty Seatbid.seat",
			args: args{
				bid: openrtb2.Bid{
					Cat: []string{"IAB-1", "IAB-2"},
				},
				bidExt: &pubmaticBidExt{
					DspId:        80,
					AdvertiserID: 139,
					BidType:      &typeinvalid,
				},
				seat: "",
			},
			want: &openrtb_ext.ExtBidPrebidMeta{
				NetworkID:            80,
				DemandSource:         "80",
				PrimaryCategoryID:    "IAB-1",
				SecondaryCategoryIDs: []string{"IAB-1", "IAB-2"},
				AdvertiserID:         139,
				AgencyID:             139,
				MediaType:            "banner",
			},
		},
		{
			name: "Valid Meta Object with Empty bidExt.DspId",
			args: args{
				bid: openrtb2.Bid{
					Cat: []string{"IAB-1", "IAB-2"},
				},
				bidExt: &pubmaticBidExt{
					DspId:        0,
					AdvertiserID: 139,
				},
				seat: "124",
			},
			want: &openrtb_ext.ExtBidPrebidMeta{
				NetworkID:            0,
				DemandSource:         "",
				PrimaryCategoryID:    "IAB-1",
				SecondaryCategoryIDs: []string{"IAB-1", "IAB-2"},
				AdvertiserID:         124,
				AgencyID:             124,
				MediaType:            "banner",
			},
		},
		{
			name: "Valid Meta Object with Empty Seatbid.seat and Empty bidExt.AdvertiserID",
			args: args{
				bid: openrtb2.Bid{
					Cat: []string{"IAB-1", "IAB-2"},
				},
				bidExt: &pubmaticBidExt{
					DspId:        80,
					AdvertiserID: 0,
				},
				seat: "",
			},
			want: &openrtb_ext.ExtBidPrebidMeta{
				NetworkID:            80,
				DemandSource:         "80",
				PrimaryCategoryID:    "IAB-1",
				SecondaryCategoryIDs: []string{"IAB-1", "IAB-2"},
				AdvertiserID:         0,
				AgencyID:             0,
				MediaType:            "banner",
			},
		},
		{
			name: "Valid Meta Object with Empty CategoryIds and BidType video",
			args: args{
				bid: openrtb2.Bid{
					Cat: []string{},
				},
				bidExt: &pubmaticBidExt{
					DspId:        80,
					AdvertiserID: 139,
					BidType:      &typevideo,
				},
				seat: "124",
			},
			want: &openrtb_ext.ExtBidPrebidMeta{
				NetworkID:         80,
				DemandSource:      "80",
				PrimaryCategoryID: "",
				AdvertiserID:      124,
				AgencyID:          124,
				MediaType:         "video",
			},
		},
		{
			name: "Valid Meta Object with Single CategoryId and BidType native",
			args: args{
				bid: openrtb2.Bid{
					Cat: []string{"IAB-1"},
				},
				bidExt: &pubmaticBidExt{
					DspId:        80,
					AdvertiserID: 139,
					BidType:      &typenative,
				},
				seat: "124",
			},
			want: &openrtb_ext.ExtBidPrebidMeta{
				NetworkID:            80,
				DemandSource:         "80",
				PrimaryCategoryID:    "IAB-1",
				SecondaryCategoryIDs: []string{"IAB-1"},
				AdvertiserID:         124,
				AgencyID:             124,
				MediaType:            "native",
			},
		},
		{
			name: "Valid Meta Object and BidType banner",
			args: args{
				bid: openrtb2.Bid{
					Cat: []string{"IAB-1", "IAB-2"},
				},
				bidExt: &pubmaticBidExt{
					DspId:        80,
					AdvertiserID: 139,
					BidType:      &typebanner,
				},
				seat: "124",
			},
			want: &openrtb_ext.ExtBidPrebidMeta{
				NetworkID:            80,
				DemandSource:         "80",
				PrimaryCategoryID:    "IAB-1",
				SecondaryCategoryIDs: []string{"IAB-1", "IAB-2"},
				AdvertiserID:         124,
				AgencyID:             124,
				MediaType:            "banner",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareMetaObject(tt.args.bid, tt.args.bidExt, tt.args.seat); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareMetaObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenameTransparencyParamsKey(t *testing.T) {
	type args struct {
		bidExt []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Empty bidExt",
			args: args{
				bidExt: []byte(``),
			},
			want: []byte(``),
		},
		{
			name: "Valid bidExt with params key",
			args: args{
				bidExt: []byte(`{"dsa":{"transparency":[{"domain":"params.com","params":[1,2]},{"domain":"dsaicon2.com","params":[2,3]}]}}`),
			},
			want: []byte(`{"dsa":{"transparency":[{"domain":"params.com","dsaparams":[1,2]},{"domain":"dsaicon2.com","dsaparams":[2,3]}]}}`),
		},
		{
			name: "bidExt without params key",
			args: args{
				bidExt: []byte(`{"dsa":{"transparency":[{"domain":"dsaicon1.com","dsaparams":[1,2]},{"domain":"dsaicon2.com","dsaparams":[2,3]}]}}`),
			},
			want: []byte(`{"dsa":{"transparency":[{"domain":"dsaicon1.com","dsaparams":[1,2]},{"domain":"dsaicon2.com","dsaparams":[2,3]}]}}`),
		},
		{
			name: "Empty transparency array",
			args: args{
				bidExt: []byte(`{"dsa":{"transparency":[]}}`),
			},
			want: []byte(`{"dsa":{"transparency":[]}}`),
		},
		{
			name: "bidExt with invalid transparency key",
			args: args{
				bidExt: []byte(`{"dsa":{"transparency":{"domain":"dsaicon1.com","params":[1,2]}}}`),
			},
			want: []byte(`{"dsa":{"transparency":{"domain":"dsaicon1.com","params":[1,2]}}}`),
		},
		{
			name: "Invalid bidExt structure",
			args: args{
				bidExt: []byte(`{"dsa":{"transparency":[{"domain":"dsaicon1.com","params":[1,2]},{"domain":"dsaicon2.com","params":[2,3]`), // Missing closing brackets
			},
			want: []byte(`{"dsa":{"transparency":[{"domain":"dsaicon1.com","dsaparams":[1,2]},{"domain":"dsaicon2.com","params":[2,3]`), // Missing closing brackets,
		},
		{
			name: "Invalid params key",
			args: args{
				bidExt: []byte(`{"dsa":{"transparency":[{"domain":"dsaicon1.com"}]}}`),
			},
			want: []byte(`{"dsa":{"transparency":[{"domain":"dsaicon1.com"}]}}`),
		},
		{
			name: "Missing transparency key",
			args: args{
				bidExt: []byte(`{"dsa":{}}`),
			},
			want: []byte(`{"dsa":{}}`),
		},
		{
			name: "Missing dsa key",
			args: args{
				bidExt: []byte(`{"any":value"}`),
			},
			want: []byte(`{"any":value"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renameTransparencyParamsKey(tt.args.bidExt)
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}

func TestPubmaticMakeBids(t *testing.T) {
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
			name: "rename bid.ext.dsa.transparency.params to bid.ext.dsa.transparency.dsaparams",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id":"test-request-id","seatbid":[{"seat":"958","bid":[{"id":"7706636740145184841","impid":"test-imp-id","price":0.5,"adid":"29681110","adm":"some-test-ad","adomain":["pubmatic.com"],"crid":"29681110","h":250,"w":300,"dealid":"testdeal","ext":{"dsa":{"transparency":[{"params":[1,2]}]},"dspid":6,"deal_channel":1,"prebiddealpriority":1}}],"ext":{"buyid":"testBuyId"}}],"bidid":"5778926625248726496","cur":"USD"}`),
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
							Ext:     json.RawMessage(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
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
			name: "correct bid.ext.dsa.transparency.dsaparams present in response",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id":"test-request-id","seatbid":[{"seat":"958","bid":[{"id":"7706636740145184841","impid":"test-imp-id","price":0.5,"adid":"29681110","adm":"some-test-ad","adomain":["pubmatic.com"],"crid":"29681110","h":250,"w":300,"dealid":"testdeal","ext":{"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"deal_channel":1,"prebiddealpriority":1}}],"ext":{"buyid":"testBuyId"}}],"bidid":"5778926625248726496","cur":"USD"}`),
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
							Ext:     json.RawMessage(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
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
			name: "bidExt without dsa",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id":"test-request-id","seatbid":[{"seat":"958","bid":[{"id":"7706636740145184841","impid":"test-imp-id","price":0.5,"adid":"29681110","adm":"some-test-ad","adomain":["pubmatic.com"],"crid":"29681110","h":250,"w":300,"dealid":"testdeal","ext":{"dspid":6,"deal_channel":1,"prebiddealpriority":1}}],"ext":{"buyid":"testBuyId"}}],"bidid":"5778926625248726496","cur":"USD"}`),
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
			name: "MultiBid MultiFloor request",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id": "test-request-id", "seatbid":[{"seat": "958", "bid":[{"id": "7706636740145184841", "impid": "test-imp-id_mf1", "price": 0.500000, "adid": "29681110", "adm": "some-test-ad", "adomain":["pubmatic.com"], "crid": "29681110", "h": 250, "w": 300, "dealid": "testdeal", "ext":{}}]}], "bidid": "5778926625248726496", "cur": "USD"}`),
				},
				externalRequest: &adapters.RequestData{
					BidderName: openrtb_ext.BidderPubmatic,
					Body:       []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id_mf1","banner":{"w":300,"h":250},"bidfloor":0.12,"ext":{}}],"app":{"name":"AutoScout24","bundle":"com.autoscout24","storeurl":"https://play.google.com/store/apps/details?id=com.autoscout24\u0026hl=fr"}}`),
				},
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
							Ext:     json.RawMessage(`{"mbmfv":0.120000}`),
						},
						BidType:    openrtb_ext.BidTypeBanner,
						BidVideo:   &openrtb_ext.ExtBidPrebidVideo{},
						BidTargets: map[string]string{},
						BidMeta: &openrtb_ext.ExtBidPrebidMeta{
							AdvertiserID: 958,
							AgencyID:     958,
							MediaType:    "banner",
						},
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
			want: "div123456789_mf1",
		},
		{
			name: "input string contains pattern at the start",
			args: args{
				input: "mf1_mf2_mf123456789",
			},
			want: "mf1_mf2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimSuffixWithPattern(tt.args.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_updateBidExtWithMultiFloor(t *testing.T) {
	type args struct {
		bidImpID string
		bidExt   []byte
		reqBody  []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "empty request body",
			args: args{
				bidImpID: "test-imp-id",
				reqBody:  []byte(``),
				bidExt:   []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
			},
			want: []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
		},
		{
			name: "request body with no imp",
			args: args{
				bidImpID: "test-imp-id",
				reqBody:  []byte(`{"id":"test-request-id"}`),
				bidExt:   []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
			},
			want: []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
		},
		{
			name: "request body with imp but no matching imp with bidImpID",
			args: args{
				bidImpID: "test-imp-id",
				reqBody:  []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id2","banner":{"w":300,"h":250},"bidfloor":0.12,"ext":{}}]}`),
				bidExt:   []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
			},
			want: []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
		},
		{
			name: "request body with imp and matching imp with bidImpID",
			args: args{
				bidImpID: "test-imp-id",
				reqBody:  []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id","banner":{"w":300,"h":250},"bidfloor":0.12,"ext":{}}]}`),
				bidExt:   []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
			},
			want: []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1,"mbmfv":0.120000}`),
		},
		{
			name: "request body with multiple imp and matching imp with bidImpID",
			args: args{
				bidImpID: "test-imp-id",
				reqBody:  []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id","banner":{"w":300,"h":250},"bidfloor":0.12,"ext":{}},{"id":"test-imp-id2","banner":{"w":300,"h":250},"bidfloor":0.13,"ext":{}}]}`),
				bidExt:   []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1}`),
			},
			want: []byte(`{"buyid":"testBuyId","deal_channel":1,"dsa":{"transparency":[{"dsaparams":[1,2]}]},"dspid":6,"prebiddealpriority":1,"mbmfv":0.120000}`),
		},
		{
			name: "request body with imp and matching imp with bidImpID and no bidExt",
			args: args{
				bidImpID: "test-imp-id",
				reqBody:  []byte(`{"id":"test-request-id","imp":[{"id":"test-imp-id","banner":{"w":300,"h":250},"bidfloor":0.12,"ext":{}}]}`),
				bidExt:   nil,
			},
			want: []byte(`{"mbmfv":0.120000}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateBidExtWithMultiFloor(tt.args.bidImpID, tt.args.bidExt, tt.args.reqBody)
			assert.Equal(t, tt.want, got)
		})
	}
}

//Need to write happy path test cases with nil bidExt
