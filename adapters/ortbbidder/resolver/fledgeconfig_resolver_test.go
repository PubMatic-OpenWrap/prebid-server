package resolver

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestFledgeConfigRetrieveFromLocation(t *testing.T) {
	resolver := &fledgeResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		path          string
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found fledgeConfig in location",
			responseNode: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"fledgeCfg": []any{
						map[string]any{
							"impid":  "imp_1",
							"bidder": "magnite",
							"config": map[string]any{
								"key": "value",
							},
						},
					},
				},
			},
			path: "ext.fledgeCfg",
			expectedValue: []*openrtb_ext.FledgeAuctionConfig{
				{
					ImpId:  "imp_1",
					Bidder: "magnite",
					Config: json.RawMessage(`{"key":"value"}`),
				},
			},
			expectedFound: true,
		},
		{
			name: "Found  invalid fledgeConfig in location",
			responseNode: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"fledgeCfg": []any{
						map[string]any{
							"impid":  1,
							"bidder": "magnite",
							"config": map[string]any{
								"key": "value",
							},
						},
					},
				},
			},
			path:          "ext.fledgeCfg",
			expectedValue: nil,
			expectedFound: false,
		},
		{
			name:          "Not found fledge config in location",
			responseNode:  map[string]any{},
			path:          "seat",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestValidateFledgeConfigs(t *testing.T) {
	testCases := []struct {
		name               string
		input              any
		expectedOutput     any
		expectedValidation bool
	}{
		{
			name: "Valid fledge configs",
			input: []any{
				map[string]any{
					"impid":   "123",
					"bidder":  "exampleBidder",
					"adapter": "exampleAdapter",
					"config": map[string]any{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			expectedOutput: []*openrtb_ext.FledgeAuctionConfig{
				{
					ImpId:   "123",
					Bidder:  "exampleBidder",
					Adapter: "exampleAdapter",
					Config:  json.RawMessage(`{"key1":"value1","key2":"value2"}`),
				},
			},
			expectedValidation: true,
		},
		{
			name: "Invalid fledge config with non-map entry",
			input: []any{
				map[string]any{
					"impid":   "123",
					"bidder":  "exampleBidder",
					"adapter": "exampleAdapter",
					"config": map[string]any{
						"key1": "value1",
						"key2": "value2",
					},
				},
				"invalidEntry",
			},
			expectedOutput:     nil,
			expectedValidation: false,
		},
		{
			name:               "nil fledge configs",
			input:              nil,
			expectedOutput:     []*openrtb_ext.FledgeAuctionConfig(nil),
			expectedValidation: false,
		},
		{
			name:               "Non-slice input",
			input:              make(chan int),
			expectedOutput:     nil,
			expectedValidation: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, valid := validateFledgeConfig(tc.input)
			assert.Equal(t, tc.expectedOutput, output)
			assert.Equal(t, tc.expectedValidation, valid)
		})
	}
}

func TestFledgeConfigSetValue(t *testing.T) {
	resolver := &fledgeResolver{}
	testCases := []struct {
		name            string
		adapterBid      map[string]any
		value           any
		expectedAdapter map[string]any
	}{
		{
			name:       "Set fledge config value",
			adapterBid: map[string]any{},
			value: []map[string]any{
				{
					"impid":   "123",
					"bidder":  "exampleBidder",
					"adapter": "exampleAdapter",
					"config": map[string]any{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			expectedAdapter: map[string]any{
				fledgeAuctionConfigKey: []map[string]any{
					{
						"impid":   "123",
						"bidder":  "exampleBidder",
						"adapter": "exampleAdapter",
						"config": map[string]any{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.adapterBid, tc.value)
			assert.Equal(t, tc.expectedAdapter, tc.adapterBid)
		})
	}
}
