package resolver

import (
	"testing"

	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	testCases := []struct {
		name           string
		sourceNode     map[string]any
		targetNode     map[string]any
		bidderResponse map[string]any
		location       string
		param          string
		expectedNode   map[string]any
	}{
		{
			name:           "SourceNode is nil, TargetNode is nil, Response is nil",
			sourceNode:     nil,
			targetNode:     nil,
			bidderResponse: nil,
			location:       "",
			param:          "",
			expectedNode:   nil,
		},
		{
			name: "SourceNode is present, TargetNode is nil, Response is present",
			sourceNode: map[string]any{
				"id": "123",
				"ext": map[string]any{
					"mtype": openrtb_ext.BidType("video"),
				},
			},
			targetNode: nil,
			bidderResponse: map[string]any{
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
			location:     "seatbid.0.bid.0.ext.mtype",
			param:        "mtype",
			expectedNode: nil,
		},
		{
			name: "Invalid param",
			sourceNode: map[string]any{
				"id": "123",
				"ext": map[string]any{
					"mtype": openrtb_ext.BidType("video"),
				},
			},
			targetNode: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(2),
				},
			},
			bidderResponse: map[string]any{
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
			location: "seatbid.0.bid.0.ext.mtype",
			param:    "param1",
			expectedNode: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(2),
				},
			},
		},
		{
			name: "Get param from the ortb bid object",
			sourceNode: map[string]any{
				"id":    "123",
				"mtype": float64(2),
			},
			targetNode: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(2),
				},
			},
			bidderResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id":    "123",
								"mtype": float64(2),
							},
						},
					},
				},
			},
			location: "seatbid.0.bid.0.ext.mtype",
			param:    "mtype",
			expectedNode: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(2),
				},
				"BidType": openrtb_ext.BidType("video"),
			},
		},
		{
			name: "Get param from the bidder param location",
			sourceNode: map[string]any{
				"id": "123",
				"ext": map[string]any{
					"mtype": openrtb_ext.BidType("video"),
				},
			},
			targetNode: map[string]any{
				"Bid": map[string]any{
					"id": "123",
					"ext": map[string]any{
						"mtype": openrtb_ext.BidType("video"),
					},
				},
			},
			bidderResponse: map[string]any{
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
			location: "seatbid.0.bid.0.ext.mtype",
			param:    "mtype",
			expectedNode: map[string]any{
				"Bid": map[string]any{
					"id": "123",
					"ext": map[string]any{
						"mtype": openrtb_ext.BidType("video"),
					},
				},
				"BidType": openrtb_ext.BidType("video"),
			},
		},
		// Todo add auto detec logic test case when it is implemented
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := New(tc.bidderResponse)
			pr.Resolve(tc.sourceNode, tc.targetNode, tc.location, tc.param)
			assert.Equal(t, tc.expectedNode, tc.targetNode)
		})
	}
}
