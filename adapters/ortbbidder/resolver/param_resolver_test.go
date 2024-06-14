package resolver

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestResolveTypeBid(t *testing.T) {
	testCases := []struct {
		name            string
		bid             map[string]any
		typeBid         map[string]any
		bidderResponse  map[string]any
		location        string
		paramName       resolveType
		expectedTypeBid map[string]any
		request         *openrtb2.BidRequest
	}{
		{
			name:            "bid is nil, typeBid is nil, Response is nil",
			bid:             nil,
			typeBid:         nil,
			bidderResponse:  nil,
			location:        "",
			paramName:       "",
			expectedTypeBid: nil,
		},
		{
			name: "bid is present, typeBid is nil, Response is present",
			bid: map[string]any{
				"id": "123",
				"ext": map[string]any{
					"bidtype": openrtb_ext.BidType("video"),
				},
			},
			typeBid: nil,
			bidderResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"bidtype": openrtb_ext.BidType("video"),
								},
							},
						},
					},
				},
			},
			location:        "seatbid.0.bid.0.ext.bidtype",
			paramName:       "bidtype",
			expectedTypeBid: nil,
		},
		{
			name: "Invalid paramName",
			bid: map[string]any{
				"id": "123",
				"ext": map[string]any{
					"bidtype": openrtb_ext.BidType("video"),
				},
			},
			typeBid: map[string]any{
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
									"bidtype": openrtb_ext.BidType("video"),
								},
							},
						},
					},
				},
			},
			location:  "seatbid.0.bid.0.ext.bidtype",
			paramName: "paramName1",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(2),
				},
			},
		},
		{
			name: "Get paramName from the ortb bid object",
			bid: map[string]any{
				"id":    "123",
				"mtype": float64(2),
			},
			typeBid: map[string]any{
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
			location:  "seatbid.0.bid.0.ext.bidtype",
			paramName: "bidtype",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(2),
				},
				"BidType": openrtb_ext.BidType("video"),
			},
		},
		{
			name: "Get paramName from the bidder paramName location",
			bid: map[string]any{
				"id": "123",
				"ext": map[string]any{
					"bidtype": openrtb_ext.BidType("video"),
				},
			},
			typeBid: map[string]any{
				"Bid": map[string]any{
					"id": "123",
					"ext": map[string]any{
						"bidtype": openrtb_ext.BidType("video"),
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
									"bidtype": openrtb_ext.BidType("video"),
								},
							},
						},
					},
				},
			},
			location:  "seatbid.0.bid.0.ext.bidtype",
			paramName: "bidtype",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id": "123",
					"ext": map[string]any{
						"bidtype": openrtb_ext.BidType("video"),
					},
				},
				"BidType": openrtb_ext.BidType("video"),
			},
		},
		{
			name: "Auto detect",
			bid: map[string]any{
				"id":  "123",
				"adm": "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
			},
			typeBid: map[string]any{
				"Bid": map[string]any{
					"id":  "123",
					"adm": "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
				},
			},
			bidderResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id":  "123",
								"adm": "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
							},
						},
					},
				},
			},
			location:  "seatbid.0.bid.0.ext.bidtype",
			paramName: "bidtype",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id":  "123",
					"adm": "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
				},
				"BidType": openrtb_ext.BidType("video"),
			},
		},
		// Todo add auto detec logic test case when it is implemented
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := New(tc.request, tc.bidderResponse)
			pr.Resolve(tc.bid, tc.typeBid, tc.location, tc.paramName)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestGetUsingBidderparamNameLocation(t *testing.T) {
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
