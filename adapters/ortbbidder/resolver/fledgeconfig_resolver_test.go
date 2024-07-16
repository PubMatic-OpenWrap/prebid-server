package resolver

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
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
		expectedError error
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
			expectedError: nil,
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
			expectedError: util.NewWarning("failed to map response-param:[fledgeAuctionConfig] value:[[map[bidder:magnite config:map[key:value] impid:1]]]"),
		},
		{
			name:          "Not found fledge config in location",
			responseNode:  map[string]any{},
			path:          "seat",
			expectedValue: nil,
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateFledgeConfigs(t *testing.T) {
	testCases := []struct {
		name           string
		input          any
		expectedOutput any
		expectedError  bool
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
			expectedError: false,
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
			expectedOutput: nil,
			expectedError:  true,
		},
		{
			name:           "nil fledge configs",
			input:          nil,
			expectedOutput: []*openrtb_ext.FledgeAuctionConfig(nil),
			expectedError:  false,
		},
		{
			name:           "Non-slice input",
			input:          make(chan int),
			expectedOutput: nil,
			expectedError:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := validateFledgeConfig(tc.input)
			assert.Equal(t, tc.expectedOutput, output)
			assert.Equal(t, tc.expectedError, err != nil)
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
