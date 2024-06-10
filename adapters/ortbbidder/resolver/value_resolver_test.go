package resolver

import (
	"testing"

	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestGetUsingBidderParamLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"mtype": openrtb_ext.BidType("video"),
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.mtype",
			expectedValue: openrtb_ext.BidType("video"),
			expectedFound: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"mtype": openrtb_ext.BidType("video"),
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	resolver := &valueResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.getUsingBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}
