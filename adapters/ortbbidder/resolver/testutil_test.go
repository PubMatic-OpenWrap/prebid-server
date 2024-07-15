package resolver

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func validateStructFields(t *testing.T, expectedFields map[string]reflect.Type, structType reflect.Type) {
	fieldCount := structType.NumField()

	// Check if the number of fields matches the expected count
	if fieldCount != len(expectedFields) {
		t.Errorf("Expected %d fields, but got %d fields", len(expectedFields), fieldCount)
	}

	// Check if the field types match the expected types
	for i := 0; i < fieldCount; i++ {
		field := structType.Field(i)
		expectedType, ok := expectedFields[field.Name]
		if !ok {
			t.Errorf("Unexpected field: %s", field.Name)
		}
		if field.Type != expectedType {
			t.Errorf("Field %s: expected type %v, but got %v", field.Name, expectedType, field.Type)
		}
	}
}

func TestTypedBidFields(t *testing.T) {
	expectedFields := map[string]reflect.Type{
		"Bid":          reflect.TypeOf(&openrtb2.Bid{}),
		"BidMeta":      reflect.TypeOf(&openrtb_ext.ExtBidPrebidMeta{}),
		"BidType":      reflect.TypeOf(openrtb_ext.BidTypeBanner),
		"BidVideo":     reflect.TypeOf(&openrtb_ext.ExtBidPrebidVideo{}),
		"BidTargets":   reflect.TypeOf(map[string]string{}),
		"DealPriority": reflect.TypeOf(0),
		"Seat":         reflect.TypeOf(openrtb_ext.BidderName("")),
	}

	structType := reflect.TypeOf(adapters.TypedBid{})
	validateStructFields(t, expectedFields, structType)
}

func TestBidderResponseFields(t *testing.T) {
	expectedFields := map[string]reflect.Type{
		"Currency":             reflect.TypeOf(""),
		"Bids":                 reflect.TypeOf([]*adapters.TypedBid{nil}),
		"FledgeAuctionConfigs": reflect.TypeOf([]*openrtb_ext.FledgeAuctionConfig{}),
	}
	structType := reflect.TypeOf(adapters.BidderResponse{})
	validateStructFields(t, expectedFields, structType)
}
