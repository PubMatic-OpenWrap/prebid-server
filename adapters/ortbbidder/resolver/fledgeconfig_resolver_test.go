package resolver

import (
	"testing"

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
			expectedValue: []map[string]any{
				{
					"impid":  "imp_1",
					"bidder": "magnite",
					"config": map[string]any{
						"key": "value",
					},
				},
			},
			expectedFound: true,
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

func TestValidateFledgeConfig(t *testing.T) {
	testCases := []struct {
		name               string
		input              any
		expectedOutput     map[string]any
		expectedValidation bool
	}{
		{
			name: "Valid fledge config with all valid keys",
			input: map[string]any{
				"impid":   "123",
				"bidder":  "exampleBidder",
				"adapter": "exampleAdapter",
				"config": map[string]any{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expectedOutput: map[string]any{
				"impid":   "123",
				"bidder":  "exampleBidder",
				"adapter": "exampleAdapter",
				"config": map[string]any{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expectedValidation: true,
		},
		{
			name: "Invalid fledge config with non-string impid",
			input: map[string]any{
				"impid":   123,
				"bidder":  "exampleBidder",
				"adapter": "exampleAdapter",
				"config": map[string]any{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expectedOutput: map[string]any{
				"bidder":  "exampleBidder",
				"adapter": "exampleAdapter",
				"config": map[string]any{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expectedValidation: true,
		},
		{
			name: "Invalid fledge config with non-map config",
			input: map[string]any{
				"impid":   "123",
				"bidder":  "exampleBidder",
				"adapter": "exampleAdapter",
				"config":  "invalidConfig",
			},
			expectedOutput: map[string]any{
				"impid":   "123",
				"bidder":  "exampleBidder",
				"adapter": "exampleAdapter",
			},
			expectedValidation: true,
		},
		{
			name: "Invalid fledge config with unknown keys",
			input: map[string]any{
				"impid":      "123",
				"bidder":     "exampleBidder",
				"adapter":    "exampleAdapter",
				"config":     map[string]any{"key1": "value1"},
				"unknownKey": "unknownValue",
			},
			expectedOutput: map[string]any{
				"impid":      "123",
				"bidder":     "exampleBidder",
				"adapter":    "exampleAdapter",
				"config":     map[string]any{"key1": "value1"},
				"unknownKey": "unknownValue",
			},
			expectedValidation: true,
		},
		{
			name: "Empty fledge config",
			input: map[string]any{
				"impid":   "",
				"bidder":  "",
				"adapter": "",
				"config":  map[string]any{},
			},
			expectedOutput: map[string]any{
				"impid":   "",
				"bidder":  "",
				"adapter": "",
				"config":  map[string]any{},
			},
			expectedValidation: true,
		},
		{
			name:               "Non-map input",
			input:              "invalid",
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

func TestValidateFledgeConfigs(t *testing.T) {
	testCases := []struct {
		name               string
		input              any
		expectedOutput     []map[string]any
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
			expectedOutput: []map[string]any{
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
			expectedOutput: []map[string]any{
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
			expectedValidation: true,
		},
		{
			name:               "Empty fledge configs",
			input:              []any{},
			expectedOutput:     []map[string]any{},
			expectedValidation: false,
		},
		{
			name:               "Non-slice input",
			input:              "invalidInput",
			expectedOutput:     nil,
			expectedValidation: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, valid := validateFledgeConfigs(tc.input)
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
