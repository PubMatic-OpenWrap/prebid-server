package pubmatic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
