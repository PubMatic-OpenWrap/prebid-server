package pubmatic

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
