package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyResolver(t *testing.T) {
	resolver := &currencyResolver{}

	t.Run("getFromORTBObject", func(t *testing.T) {
		testCases := []struct {
			name          string
			ortbResponse  map[string]any
			expectedValue any
			expectedFound bool
		}{
			{
				name: "Currency found in ORTB response",
				ortbResponse: map[string]any{
					"cur": "USD",
				},
				expectedValue: "USD",
				expectedFound: true,
			},
			{
				name:          "Currency not found in ORTB response",
				ortbResponse:  map[string]any{},
				expectedValue: nil,
				expectedFound: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, found := resolver.getFromORTBObject(tc.ortbResponse)
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
				expectedFound: false,
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
				value: "USD",
				expectedBid: map[string]any{
					"id":       "123",
					"Currency": "USD",
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
