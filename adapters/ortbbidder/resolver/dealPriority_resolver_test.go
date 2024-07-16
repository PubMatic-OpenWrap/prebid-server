package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBidDealPriorityFromLocation(t *testing.T) {
	resolver := &bidDealPriorityResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found dealPriority in location",
			responseNode: map[string]any{
				"cur": "USD",
				"dp":  10.0,
			},
			path:          "dp",
			expectedValue: 10,
			expectedError: false,
		},
		{
			name: "Found invalid dealPriority in location",
			responseNode: map[string]any{
				"cur": "USD",
				"dp":  "invalid",
			},
			path:          "dp",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found dealPriority in location",
			responseNode:  map[string]any{},
			path:          "seat",
			expectedValue: nil,
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidDealPriorityResolverSetValue(t *testing.T) {
	resolver := &bidDealPriorityResolver{}
	testCases := []struct {
		name            string
		adapterBid      map[string]any
		value           any
		expectedAdapter map[string]any
	}{
		{
			name:       "Set deal priority value",
			adapterBid: map[string]any{},
			value:      10,
			expectedAdapter: map[string]any{
				bidDealPriorityKey: 10,
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
