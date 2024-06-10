package resolver

import (
	"testing"

	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestMtypeResolver(t *testing.T) {
	resolver := &mtypeResolver{}

	t.Run("getFromORTBObject", func(t *testing.T) {
		testCases := []struct {
			name          string
			bid           map[string]any
			expectedValue any
			expectedFound bool
		}{
			{
				name: "mtype found in bid",
				bid: map[string]any{
					"mtype": 2.0,
				},
				expectedValue: openrtb_ext.BidTypeVideo,
				expectedFound: true,
			},
			{
				name:          "mtype not found in bid",
				bid:           map[string]any{},
				expectedValue: nil,
				expectedFound: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, found := resolver.getFromORTBObject(tc.bid)
				assert.Equal(t, tc.expectedValue, value)
				assert.Equal(t, tc.expectedFound, found)
			})
		}
	})

	t.Run("autoDetect", func(t *testing.T) {
		testCases := []struct {
			name          string
			bid           map[string]any
			expectedValue any
			expectedFound bool
		}{
			{
				name: "Auto detect with bid",
				bid: map[string]any{
					"id": "123",
				},
				expectedValue: nil,
				expectedFound: false, // The function always returns false
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, found := resolver.autoDetect(tc.bid)
				assert.Equal(t, tc.expectedValue, value)
				assert.Equal(t, tc.expectedFound, found)
			})
		}
	})

	t.Run("setValue", func(t *testing.T) {
		testCases := []struct {
			name        string
			adapterBid  map[string]any
			value       any
			expectedBid map[string]any
		}{
			{
				name: "Set value in adapter bid",
				adapterBid: map[string]any{
					"id": "123",
				},
				value: openrtb_ext.BidTypeVideo,
				expectedBid: map[string]any{
					"id":      "123",
					"BidType": openrtb_ext.BidTypeVideo,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resolver.setValue(tc.adapterBid, tc.value)
				assert.Equal(t, tc.expectedBid, tc.adapterBid)
			})
		}
	})
}
