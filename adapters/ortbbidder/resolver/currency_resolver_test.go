package resolver

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
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
			request       *openrtb2.BidRequest
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
				value, found := resolver.autoDetect(tc.request, tc.bid)
				assert.Equal(t, tc.expectedValue, value)
				assert.Equal(t, tc.expectedFound, found)
			})
		}
	})

	t.Run("setValue", func(t *testing.T) {
		testCases := []struct {
			name                    string
			adapterResponse         map[string]any
			value                   any
			expectedAdapterResponse map[string]any
		}{
			{
				name: "Set value in adapter bid",
				adapterResponse: map[string]any{
					"id": "123",
				},
				value: "USD",
				expectedAdapterResponse: map[string]any{
					"id":       "123",
					"Currency": "USD",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resolver.setValue(tc.adapterResponse, tc.value)
				assert.Equal(t, tc.expectedAdapterResponse, tc.adapterResponse)
			})
		}
	})
}
