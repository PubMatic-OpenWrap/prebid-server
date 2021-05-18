package gamoshi

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/openrtb_ext"
)

// This file actually intends to test static/bidder-params/gamoshi.json
//
// These also validate the format of the external API: request.imp[i].ext.gamoshi

// TestValidParams makes sure that the Gamoshi schema accepts all imp.ext fields which we intend to support.
func TestValidParams(t *testing.T) {
	validator, err := openrtb_ext.NewBidderParamsValidator("../../static/bidder-params")
	if err != nil {
		t.Fatalf("Failed to fetch the json-schemas. %v", err)
	}

	for _, validParam := range validParams {
		if err := validator.Validate(openrtb_ext.BidderGamoshi, json.RawMessage(validParam)); err != nil {
			t.Errorf("Schema rejected Gamoshi params: %s", validParam)
		}
	}
}

// TestInvalidParams makes sure that the Gamoshi schema rejects all the imp.ext fields we don't support.
func TestInvalidParams(t *testing.T) {
	validator, err := openrtb_ext.NewBidderParamsValidator("../../static/bidder-params")
	if err != nil {
		t.Fatalf("Failed to fetch the json-schemas. %v", err)
	}

	for _, invalidParam := range invalidParams {
		if err := validator.Validate(openrtb_ext.BidderGamoshi, json.RawMessage(invalidParam)); err == nil {
			t.Errorf("Schema allowed unexpected params: %s", invalidParam)
		}
	}
}

var validParams = []string{
	`{"supplyPartnerId": "123"}`,
	`{"supplyPartnerId": "test", "headerbidding": false}`,
}

var invalidParams = []string{
	`{"supplyPartnerId": 100}`,
	`{"supplyPartnerId": false}`,
	`{"supplyPartnerId": true}`,
	`{"supplyPartnerId": 123, "headerbidding": true}`,
	``,
	`null`,
	`true`,
	`9`,
	`1.2`,
	`[]`,
	`{}`,
}
