package ortbbidder

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/stretchr/testify/assert"
)

func TestNewResponseBuilder(t *testing.T) {
	testCases := []struct {
		name           string
		responseParams map[string]bidderparams.BidderParamMapper
		expected       *responseBuilder
	}{
		{
			name:           "With nil responseParams",
			responseParams: nil,
			expected: &responseBuilder{
				responseParams: make(map[string]bidderparams.BidderParamMapper),
			},
		},
		{
			name: "With non-nil responseParams",
			responseParams: map[string]bidderparams.BidderParamMapper{
				"test": {},
			},
			expected: &responseBuilder{
				responseParams: map[string]bidderparams.BidderParamMapper{
					"test": {},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := newResponseBuilder(tc.responseParams)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseResponse(t *testing.T) {
	testCases := []struct {
		name          string
		responseBytes json.RawMessage
		expectedError error
	}{
		{
			name:          "Valid response",
			responseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1", "ext":{"mtype":"video"}}]}]}`),
			expectedError: nil,
		},
		{
			name:          "Invalid response",
			responseBytes: []byte(`{"id":"bid-resp-id","cur":"USD","seatbid":[{"seat":"test_bidder","bid":[{"id":"bid-1", "ext":{"mtype":"video"}}]}`), // missing closing bracket
			expectedError: errors.New("expect ] in the end, but found \x00"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rb := &responseBuilder{}
			err := rb.parseResponse(tc.responseBytes)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
