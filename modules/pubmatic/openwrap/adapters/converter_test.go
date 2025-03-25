package adapters

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestFixBidderParams(t *testing.T) {

	type Args struct {
		AdapterName string          `json:"adapterName"`
		RequestJSON json.RawMessage `json:"requestJSON"`
	}
	type Want struct {
		ExpectedJSON  json.RawMessage `json:"expectedJSON"`
		ExpectedError string          `json:"error"`
	}
	type test struct {
		Name string `json:"name"`
		Args Args   `json:"args"`
		Want Want   `json:"want"`
	}

	var tests []test
	//reading test cases from file
	readTestCasesFromFile(t, `./tests/hybrid_bidders.json`, &tests)

	//prerequisite
	validator := getPrebidBidderParamsValidator(t, `../../../../static/bidder-params`)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			//resolving alias only
			bidderCode := ResolveOWBidder(tt.Args.AdapterName)

			//FixBidderParams fixing bidder parameters
			result, err := FixBidderParams("req-id", tt.Args.AdapterName, bidderCode, tt.Args.RequestJSON)

			//Verify error check
			if len(tt.Want.ExpectedError) > 0 {
				if assert.Error(t, err) {
					assert.Equal(t, err.Error(), tt.Want.ExpectedError)
				}
			} else {
				assert.NoError(t, err)
			}

			//verify json
			AssertJSON(t, tt.Want.ExpectedJSON, result)

			//validate schema for resultant string
			err = validator.Validate(openrtb_ext.BidderName(bidderCode), result)
			assert.NoError(t, err)
		})
	}
}

func TestFixBidderParamsS2S(t *testing.T) {

	type Args struct {
		AdapterName string          `json:"adapterName"`
		RequestJSON json.RawMessage `json:"requestJSON"`
	}
	type Want struct {
		ExpectedJSON  json.RawMessage `json:"expectedJSON"`
		ExpectedError string          `json:"error"`
	}
	type test struct {
		Name string `json:"name"`
		Args Args   `json:"args"`
		Want Want   `json:"want"`
	}

	var tests []test
	//reading test cases from file
	readTestCasesFromFile(t, `./tests/s2s_bidders.json`, &tests)

	//prerequisite
	validator := getPrebidBidderParamsValidator(t, `../../../../static/bidder-params`)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			//resolving alias only
			bidderCode := ResolveOWBidder(tt.Args.AdapterName)

			//FixBidderParams fixing bidder parameters
			result, err := FixBidderParams("req-id", tt.Args.AdapterName, bidderCode, tt.Args.RequestJSON)

			//Verify error check
			if len(tt.Want.ExpectedError) > 0 {
				if assert.Error(t, err) {
					assert.Equal(t, err.Error(), tt.Want.ExpectedError)
				}
			} else {
				assert.NoError(t, err)
			}

			//verify json
			AssertJSON(t, tt.Want.ExpectedJSON, result)

			//validate schema for resultant string
			err = validator.Validate(openrtb_ext.BidderName(bidderCode), result)
			assert.NoError(t, err)
		})
	}
}
