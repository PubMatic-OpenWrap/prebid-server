package resolver

// import (
// 	"testing"

// 	"github.com/prebid/prebid-server/v2/openrtb_ext"
// 	"github.com/stretchr/testify/assert"
// )

// func TestParamProcessorResolveParam(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		param        string
// 		node         map[string]any
// 		targetNode   map[string]any
// 		responseNode map[string]any
// 		location     string
// 		expected     map[string]any
// 	}{
// 		{
// 			name:  "Get mytype from the bid object",
// 			param: "mtype",
// 			node:  nil,
// 			targetNode: map[string]any{
// 				"Bid": map[string]any{
// 					"id":    "123",
// 					"mtype": 2,
// 				},
// 			},
// 			responseNode: map[string]any{
// 				"cur": "USD",
// 				"seatbid": []any{
// 					map[string]any{
// 						"bid": []any{
// 							map[string]any{
// 								"id":    "123",
// 								"mtype": 2,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			location: "seatbid.0.bid.0.mtype",
// 			expected: map[string]any{
// 				"Bid": map[string]any{
// 					"id":    "123",
// 					"mtype": 2,
// 				},
// 				"BidType": openrtb_ext.BidTypeVideo,
// 			},
// 		},
// 		{
// 			name:  "Get mytype from the bid object",
// 			param: "mtype",
// 			node: map[string]any{
// 				"mtype": float64(2),
// 			},
// 			targetNode: map[string]any{
// 				"Bid": map[string]any{
// 					"id":    "123",
// 					"mtype": 2,
// 				},
// 			},
// 			responseNode: map[string]any{
// 				"cur": "USD",
// 				"seatbid": []any{
// 					map[string]any{
// 						"bid": []any{
// 							map[string]any{
// 								"id":    "123",
// 								"mtype": 2,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			location: "seatbid.0.bid.0.mtype",
// 			expected: map[string]any{
// 				"Bid": map[string]any{
// 					"id":    "123",
// 					"mtype": 2,
// 				},
// 				"BidType": openrtb_ext.BidTypeVideo,
// 			},
// 		},
// 		{
// 			name:  "Get mytype from the bidder param location",
// 			param: "mtype",
// 			node: map[string]any{
// 				"id": "123",
// 				"ext": map[string]any{
// 					"mtype": "video",
// 				},
// 			},
// 			targetNode: map[string]any{
// 				"Bid": map[string]any{
// 					"id": "123",
// 				},
// 			},
// 			responseNode: map[string]any{
// 				"cur": "USD",
// 				"seatbid": []any{
// 					map[string]any{
// 						"bid": []any{
// 							map[string]any{
// 								"id": "123",
// 								"ext": map[string]any{
// 									"mtype": "video",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			location: "seatbid.0.bid.0.ext.mtype",
// 			expected: map[string]any{
// 				"Bid": map[string]any{
// 					"id": "123",
// 					"ext": map[string]any{
// 						"mtype": "video",
// 					},
// 				},
// 				"BidType": openrtb_ext.BidTypeVideo,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			pp := NewParamProcessor()
// 			pp.ResolveParam(tt.node, tt.targetNode, tt.responseNode, tt.location, tt.param)
// 			assert.Equal(t, tt.expected, tt.targetNode)
// 		})
// 	}
// }
