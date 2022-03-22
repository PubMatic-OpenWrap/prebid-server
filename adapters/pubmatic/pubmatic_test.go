package pubmatic

import (
	"encoding/json"
	"testing"

	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/adapters/adapterstest"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestJsonSamples(t *testing.T) {
	bidder, buildErr := Builder(openrtb_ext.BidderPubmatic, config.Adapter{
		Endpoint: "https://hbopenbid.pubmatic.com/translator?source=prebid-server"})

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

func TestGetMapFromJSON(t *testing.T) {
	ext := json.RawMessage("{\"buyid\":\"testBuyId\"}")
	extMap := getMapFromJSON(ext)
	if extMap == nil {
		t.Errorf("it should be converted in extMap")
	}
}

func TestGetMapFromJSONWithInvalidJSON(t *testing.T) {
	ext := json.RawMessage("{\"buyid\":\"testBuyId\"}}}}")
	extMap := getMapFromJSON(ext)
	if extMap != nil {
		t.Errorf("it should be converted in extMap")
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

func TestParseImpressionObject(t *testing.T) {
	type args struct {
		imp                      *openrtb2.Imp
		extractWrapperExtFromImp bool
		extractPubIDFromImp      bool
	}
	tests := []struct {
		name                string
		args                args
		expectedWrapperExt  *pubmaticWrapperExt
		expectedPublisherId string
		wantErr             bool
		expectedBidfloor    float64
	}{
		{
			name: "imp.bidfloor empty and kadfloor set",
			args: args{
				imp: &openrtb2.Imp{
					Video: &openrtb2.Video{},
					Ext:   json.RawMessage(`{"bidder":{"kadfloor":"0.12"}}`),
				},
			},
			expectedBidfloor: 0.12,
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
			expectedBidfloor: 0.12,
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
			expectedBidfloor: 0.12,
		},
		{
			name: "imp.bidfloor set and kadfloor set, preference to kadfloor",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":"0.11"}}`),
				},
			},
			expectedBidfloor: 0.11,
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
			expectedBidfloor: 0.13,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedWrapperExt, receivedPublisherId, err := parseImpressionObject(tt.args.imp, tt.args.extractWrapperExtFromImp, tt.args.extractPubIDFromImp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expectedWrapperExt, receivedWrapperExt)
			assert.Equal(t, tt.expectedPublisherId, receivedPublisherId)
			assert.Equal(t, tt.expectedBidfloor, tt.args.imp.BidFloor)
		})
	}
}
