package resolver

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
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
		paramName       parameter
		expectedTypeBid map[string]any
		expectedErrs    []error
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
									"bidtype": "video",
								},
							},
						},
					},
				},
			},
			location:        "seatbid.0.bid.0.ext.bidtype",
			paramName:       "bidType",
			expectedTypeBid: nil,
		},
		{
			name: "Invalid paramName",
			bid: map[string]any{
				"id": "123",
				"ext": map[string]any{
					"bidtype": "video",
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
									"bidtype": "video",
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
			name: "Get param from the ortb bid object",
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
			paramName: "bidType",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(2),
				},
				"BidType": openrtb_ext.BidType("video"),
			},
		},
		{
			name: "fail to get param from the ortb bid object, fallback to get from bidder param location",
			bid: map[string]any{
				"id":    "123",
				"mtype": "a",
			},
			typeBid: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(0),
				},
			},
			bidderResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id":    "123",
								"mtype": float64(0),
								"ext": map[string]any{
									"bidtype": "banner",
								},
							},
						},
					},
				},
			},
			location:  "seatbid.0.bid.0.ext.bidtype",
			paramName: "bidType",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id":    "123",
					"mtype": float64(0),
				},
				"BidType": openrtb_ext.BidType("banner"),
			},
			expectedErrs: []error{
				util.NewWarning("failed to map response-param:[bidType] method:[standard_oRTB_param] value:[a]"),
			},
		},
		{
			name: "Get param from the bidder paramName location",
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
									"bidtype": "video",
								},
							},
						},
					},
				},
			},
			location:  "seatbid.0.bid.0.ext.bidtype",
			paramName: "bidType",
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
			name: "fail to detect from location, fallback to Auto detect",
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
				"cur":     "USD",
				"bidtype": 1,
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
			location:  "bidtype",
			paramName: "bidType",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id":  "123",
					"adm": "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
				},
				"BidType": openrtb_ext.BidType("video"),
			},
			expectedErrs: []error{
				util.NewWarning("failed to map response-param:[bidType] method:[response_param_location] value:[1]"),
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
			paramName: "bidType",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id":  "123",
					"adm": "<VAST version=\"3.0\"><Ad><Wrapper><VASTAdTagURI>",
				},
				"BidType": openrtb_ext.BidType("video"),
			},
		},
		{
			name: "Failed to Auto detect",
			bid: map[string]any{
				"id": "123",
			},
			typeBid: map[string]any{
				"Bid": map[string]any{
					"id": "123",
				},
			},
			bidderResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
							},
						},
					},
				},
			},
			location:  "seatbid.0.bid.0.ext.bidtype",
			paramName: "bidType",
			expectedTypeBid: map[string]any{
				"Bid": map[string]any{
					"id": "123",
				},
			},
			expectedErrs: []error{
				util.NewWarning("failed to map response-param:[bidType] method:[auto_detect] error:[bid.impid is missing]"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := New(tc.request, tc.bidderResponse)
			errs := pr.Resolve(tc.bid, tc.typeBid, tc.location, tc.paramName)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
			assert.Equal(t, tc.expectedErrs, errs)
		})
	}
}

func TestDefaultvalueResolver(t *testing.T) {
	tests := []struct {
		name      string
		wantValue any
		wantErr   error
	}{
		{
			name:      "test default values",
			wantValue: nil,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &paramResolver{}
			value, found := r.retrieveFromBidderParamLocation(map[string]any{}, "any.path")
			assert.Equal(t, tt.wantErr, found)
			assert.Equal(t, tt.wantValue, value)

			value, found = r.getFromORTBObject(map[string]any{})
			assert.Equal(t, tt.wantErr, found)
			assert.Equal(t, tt.wantValue, value)

			value, found = r.autoDetect(&openrtb2.BidRequest{}, map[string]any{})
			assert.Equal(t, tt.wantErr, found)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}
