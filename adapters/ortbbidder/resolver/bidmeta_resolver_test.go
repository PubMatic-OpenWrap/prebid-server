package resolver

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestBidMetaRetrieveFromLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
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
									"metaObject": map[string]any{
										"advertiserDomains": []any{"abc.com", "xyz.com"},
										"brandId":           1.0,
									},
								},
							},
						},
					},
				},
			},
			path: "seatbid.0.bid.0.ext.metaObject",
			expectedValue: map[string]any{
				"advertiserDomains": []any{"abc.com", "xyz.com"},
				"brandId":           1.0,
			},
			expectedError: false,
		},
		{
			name: "Found invalid meta object in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"metaObject": map[string]any{
										"advertiserDomains": "abc.com",
										"brandId":           1.0,
									},
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.metaObject",
			expectedValue: nil,
			expectedError: true,
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
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	resolver := &bidMetaResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestValidateBidMeta(t *testing.T) {
	tests := []struct {
		name          string
		value         any
		expected      any
		expectedError bool
	}{
		{
			name: "Metadata with all valid fields",
			value: map[string]any{
				bidMetaSecondaryCatIdKey: []any{"music", "sports"},
				bidMetaAdvertiserIdKey:   123.0,
				bidMetaDChainKey: map[string]any{
					"field": "value",
				},
				"customField": "customValue",
			},
			expected: map[string]any{
				bidMetaSecondaryCatIdKey: []any{"music", "sports"},
				bidMetaAdvertiserIdKey:   123.0,
				bidMetaDChainKey: map[string]any{
					"field": "value",
				},
				"customField": "customValue",
			},
			expectedError: false,
		},
		{
			name: "Metadata with wrong type",
			value: map[string]any{
				bidMetaAdvertiserDomainsKey: "example.com", // should be a slice
				bidMetaAdvertiserIdKey:      "123",         // should be an float
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name:          "Invalid type for value",
			value:         make(chan int),
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateBidMeta(tt.value)
			assert.Equal(t, tt.expectedError, err != nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBidMetaSetValue(t *testing.T) {
	resolver := &bidMetaResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set value ",
			typeBid: map[string]any{
				"id": "123",
			},
			value: map[string]any{
				"any-key": "any-val",
			},
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"any-key": "any-val",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaAdvDomainsRetrieveFromLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
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
									"adomains": []any{"abc.com", "xyz.com"},
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.adomains",
			expectedValue: []string{"abc.com", "xyz.com"},
			expectedError: false,
		},
		{
			name: "Found in location but data type is invalid",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"adomains": []string{"abc.com", "xyz.com"},
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.adomains",
			expectedValue: nil,
			expectedError: true,
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
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	resolver := &bidMetaAdvDomainsResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaAdvDomainsResolverSetValue(t *testing.T) {
	resolver := &bidMetaAdvDomainsResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set value when meta object is present",
			typeBid: map[string]any{
				"id":      "123",
				"BidMeta": map[string]any{},
			},
			value: "xyz.com",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"advertiserDomains": "xyz.com",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaAdvIdRetrieveFromLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
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
									"advid": 10.0,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.advid",
			expectedValue: 10,
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than float",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"advid": 10,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.advid",
			expectedValue: nil,
			expectedError: true,
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
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	resolver := &bidMetaAdvIDResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaAdvIdResolverSetValue(t *testing.T) {
	resolver := &bidMetaAdvIDResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set value when meta object is absent",
			typeBid: map[string]any{
				"id": "123",
			},
			value: 1,
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"advertiserId": 1,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaAdvNameRetrieveFromLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur":     "USD",
				"advname": "Acme Corp",
			},
			path:          "advname",
			expectedValue: "Acme Corp",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"ext": map[string]any{
					"advname": 123,
				},
			},
			path:          "ext.advname",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	resolver := &bidMetaAdvNameResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaAdvNameResolverSetValue(t *testing.T) {
	resolver := &bidMetaAdvNameResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set value when meta object is present",
			typeBid: map[string]any{
				"id":      "123",
				"BidMeta": map[string]any{},
			},
			value: "Acme Corp",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"advertiserName": "Acme Corp",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaAgencyIDRetrieveFromLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"agencyid": 10.0,
				},
			},
			path:          "ext.agencyid",
			expectedValue: 10,
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than float",
			ortbResponse: map[string]any{
				"cur":      "USD",
				"agencyid": 10,
			},
			path:          "agencyid",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur":     "USD",
				"seatbid": []any{},
			},
			path:          "ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	resolver := &bidMetaAgencyIDResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaAgencyIDResolverSetValue(t *testing.T) {
	resolver := &bidMetaAgencyIDResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set value when meta object is absent",
			typeBid: map[string]any{
				"id": "123",
			},
			value: 1,
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"agencyId": 1,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaAgencyNameRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaAgencyNameResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"agencyName": "TestAgency",
				},
			},
			path:          "ext.agencyName",
			expectedValue: "TestAgency",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"agencyName": 12345,
				},
			},
			path:          "ext.agencyName",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaAgencyNameResolverSetValue(t *testing.T) {
	resolver := &bidMetaAgencyNameResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set agency name value",
			typeBid: map[string]any{
				"id": "123",
			},
			value: "TestAgency",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"agencyName": "TestAgency",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaBrandIDRetrieveFromLocation(t *testing.T) {
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
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
									"brandid": 10.0,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.brandid",
			expectedValue: 10,
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than float",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"brandid": 10,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.brandid",
			expectedValue: nil,
			expectedError: true,
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
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	resolver := &bidMetaBrandIDResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaBrandIDResolverSetValue(t *testing.T) {
	resolver := &bidMetaBrandIDResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set value when meta object is present",
			typeBid: map[string]any{
				"id":      "123",
				"BidMeta": map[string]any{},
			},
			value: 1,
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"brandId": 1,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaBrandNameRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaBrandNameResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"brandname": "TestBrand",
				},
			},
			path:          "ext.brandname",
			expectedValue: "TestBrand",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"brandname": 10,
				},
			},
			path:          "ext.brandname",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaBrandNameResolverSetValue(t *testing.T) {
	resolver := &bidMetaBrandNameResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set value when meta object is present",
			typeBid: map[string]any{
				"id":      "123",
				"BidMeta": map[string]any{},
			},
			value: "BrandName",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"brandName": "BrandName",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaDChainRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaDChainResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"dchain": map[string]any{
						"segment": map[string]any{
							"s": "",
							"t": 1,
						},
					},
				},
			},
			path: "ext.dchain",
			expectedValue: map[string]any{
				"segment": map[string]any{
					"s": "",
					"t": 1,
				},
			},
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than json.RawMessage",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"dchain": "invalidJSON",
				},
			},
			path:          "ext.dchain",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaDChainResolverSetValue(t *testing.T) {
	resolver := &bidMetaDChainResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           json.RawMessage
		expectedTypeBid map[string]any
	}{
		{
			name: "Set DChain value when meta object is present",
			typeBid: map[string]any{
				"id":      "123",
				"BidMeta": map[string]any{},
			},
			value: json.RawMessage(`{"segment":[{"s":"","t":1}]}`),
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"dchain": json.RawMessage(`{"segment":[{"s":"","t":1}]}`),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaDemandSourceRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaDemandSourceResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
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
							},
						},
						"ext": map[string]any{
							"demandSource": "Direct",
						},
					},
				},
			},
			path:          "seatbid.0.ext.demandSource",
			expectedValue: "Direct",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"demandSource": 100,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.demandSource",
			expectedValue: nil,
			expectedError: true,
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
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaDemandSourceResolverSetValue(t *testing.T) {
	resolver := &bidMetaDemandSourceResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set demand source",
			typeBid: map[string]any{
				"id": "123",
			},
			value: "Direct",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"demandSource": "Direct",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaMediaTypeRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaMediaTypeResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
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
									"bidder": map[string]any{
										"mediaType": "banner",
									},
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.bidder.mediaType",
			expectedValue: "banner",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"mediaType": 10,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.mediaType",
			expectedValue: nil,
			expectedError: true,
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
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaMediaTypeResolverSetValue(t *testing.T) {
	resolver := &bidMetaMediaTypeResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set media type",
			typeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"mediaType": "video",
				},
			},
			value: "banner",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"mediaType": "banner",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaNetworkIDRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaNetworkIDResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
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
									"networkID": 100.0,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.networkID",
			expectedValue: 100,
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than int",
			ortbResponse: map[string]any{
				"cur": "USD",
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"id": "123",
								"ext": map[string]any{
									"networkID": "wrongType",
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.networkID",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur":     "USD",
				"seatbid": []any{},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaNetworkIDResolverSetValue(t *testing.T) {
	resolver := &bidMetaNetworkIDResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set networkid value",
			typeBid: map[string]any{
				"id": "123",
			},
			value: 100,
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"networkId": 100,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaNetworkNameRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaNetworkNameResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur":         "USD",
				"networkName": "TestNetwork",
				"seatbid": []any{
					map[string]any{
						"bid": []any{},
					},
				},
			},
			path:          "networkName",
			expectedValue: "TestNetwork",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur":         "USD",
				"networkName": 10,
				"seatbid": []any{
					map[string]any{
						"bid": []any{},
					},
				},
			},
			path:          "networkName",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaNetworkNameResolverSetValue(t *testing.T) {
	resolver := &bidMetaNetworkNameResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set network name value",
			typeBid: map[string]any{
				"id": "123",
			},
			value: "TestNetwork",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"networkName": "TestNetwork",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaPrimaryCatIdRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaPrimaryCategoryIDResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur":          "USD",
				"primaryCatId": "testCategory",
			},
			path:          "primaryCatId",
			expectedValue: "testCategory",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur":          "USD",
				"primaryCatId": 12345,
			},
			path:          "primaryCatId",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaPrimaryCatIdResolverSetValue(t *testing.T) {
	resolver := &bidMetaPrimaryCategoryIDResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set primaryCatId value",
			typeBid: map[string]any{
				"id":      "123",
				"BidMeta": map[string]any{},
			},
			value: "testCategory",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"primaryCatId": "testCategory",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaRendererNameRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaRendererNameResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererName": "testRenderer",
				},
			},
			path:          "ext.rendererName",
			expectedValue: "testRenderer",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererName": 12345,
				},
			},
			path:          "ext.rendererName",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaRendererNameResolverSetValue(t *testing.T) {
	resolver := &bidMetaRendererNameResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set rendered name value",
			typeBid: map[string]any{
				"id": "123",
			},
			value: "testRenderer",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"rendererName": "testRenderer",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaRendererVersionRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaRendererVersionResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererVersion": "1.0.0",
				},
			},
			path:          "ext.rendererVersion",
			expectedValue: "1.0.0",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererVersion": 12345,
				},
			},
			path:          "ext.rendererVersion",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaRendererVersionResolverSetValue(t *testing.T) {
	resolver := &bidMetaRendererVersionResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           any
		expectedTypeBid map[string]any
	}{
		{
			name: "set renderer version value",
			typeBid: map[string]any{
				"id": "123",
			},
			value: "1.0.0",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"rendererVersion": "1.0.0",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaRendererDataRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaRendererDataResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererData": map[string]any{"key": "value"},
				},
			},
			path:          "ext.rendererData",
			expectedValue: map[string]any{"key": "value"},
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than json.RawMessage",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererData": 12345,
				},
			},
			path:          "ext.rendererData",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaRendererDataResolverSetValue(t *testing.T) {
	resolver := &bidMetaRendererDataResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           json.RawMessage
		expectedTypeBid map[string]any
	}{
		{
			name: "set renderer data value",
			typeBid: map[string]any{
				"id": "123",
			},
			value: json.RawMessage(`{"key":"value"}`),
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"rendererData": json.RawMessage(`{"key":"value"}`),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaRendererUrlRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaRendererUrlResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererUrl": "https://example.com/renderer",
				},
			},
			path:          "ext.rendererUrl",
			expectedValue: "https://example.com/renderer",
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"rendererUrl": 12345,
				},
			},
			path:          "ext.rendererUrl",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaRendererUrlResolverSetValue(t *testing.T) {
	resolver := &bidMetaRendererUrlResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           string
		expectedTypeBid map[string]any
	}{
		{
			name: "set renderer URL value",
			typeBid: map[string]any{
				"id": "123",
			},
			value: "https://example.com/renderer",
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"rendererUrl": "https://example.com/renderer",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestBidMetaSecCatIdsRetrieveFromLocation(t *testing.T) {
	resolver := &bidMetaSecondaryCategoryIDsResolver{}
	testCases := []struct {
		name          string
		ortbResponse  map[string]any
		path          string
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"secondaryCatIds": []any{"cat1", "cat2"},
				},
			},
			path:          "ext.secondaryCatIds",
			expectedValue: []string{"cat1", "cat2"},
			expectedError: false,
		},
		{
			name: "Found in location but data type is other than []string",
			ortbResponse: map[string]any{
				"cur": "USD",
				"ext": map[string]any{
					"secondaryCatIds": "not a slice",
				},
			},
			path:          "ext.secondaryCatIds",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "ext.nonexistent",
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidMetaSecondaryCatIdsResolverSetValue(t *testing.T) {
	resolver := &bidMetaSecondaryCategoryIDsResolver{}
	testCases := []struct {
		name            string
		typeBid         map[string]any
		value           []string
		expectedTypeBid map[string]any
	}{
		{
			name: "set secondary category IDs",
			typeBid: map[string]any{
				"id": "123",
			},
			value: []string{"cat1", "cat2"},
			expectedTypeBid: map[string]any{
				"id": "123",
				"BidMeta": map[string]any{
					"secondaryCatIds": []string{"cat1", "cat2"},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.typeBid, tc.value)
			assert.Equal(t, tc.expectedTypeBid, tc.typeBid)
		})
	}
}

func TestSetKeyValueInBidMeta(t *testing.T) {
	tests := []struct {
		name          string
		adapterBid    map[string]any
		key           string
		value         any
		expectedBid   map[string]any
		expectedError bool
	}{
		{
			name:       "Set new key-value pair when meta object is absent",
			adapterBid: map[string]any{},
			key:        "testKey",
			value:      "testValue",
			expectedBid: map[string]any{
				"BidMeta": map[string]any{
					"testKey": "testValue",
				},
			},
			expectedError: false,
		},
		{
			name: "Update existing key-value pair in meta object",
			adapterBid: map[string]any{
				"BidMeta": map[string]any{
					"existingKey": "existingValue",
				},
			},
			key:   "existingKey",
			value: "newValue",
			expectedBid: map[string]any{
				"BidMeta": map[string]any{
					"existingKey": "newValue",
				},
			},
			expectedError: false,
		},
		{
			name: "Fail to set value when meta object is invalid",
			adapterBid: map[string]any{
				"BidMeta": "",
			},
			key:   "testKey",
			value: "testValue",
			expectedBid: map[string]any{
				"BidMeta": "",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setKeyValueInBidMeta(tt.adapterBid, tt.key, tt.value)
			assert.Equal(t, tt.expectedError, err != nil)
			assert.Equal(t, tt.expectedBid, tt.adapterBid)
		})
	}
}

// TestExtBidPrebidMetaFields notifies us of any changes in the openrtb_ext.ExtBidPrebidMeta struct.
// If a new field is added in openrtb_ext.ExtBidPrebidMeta, then add the support to resolve the new field and update the test case.
// If the data type of an existing field changes then update the resolver of the respective field.
func TestExtBidPrebidMetaFields(t *testing.T) {
	// Expected field count and types
	expectedFields := map[string]reflect.Type{
		"AdapterCode":          reflect.TypeOf(""), // not expected to be set by adapter
		"AdvertiserDomains":    reflect.TypeOf([]string{}),
		"AdvertiserID":         reflect.TypeOf(0),
		"AdvertiserName":       reflect.TypeOf(""),
		"AgencyID":             reflect.TypeOf(0),
		"AgencyName":           reflect.TypeOf(""),
		"BrandID":              reflect.TypeOf(0),
		"BrandName":            reflect.TypeOf(""),
		"DChain":               reflect.TypeOf(json.RawMessage{}),
		"DemandSource":         reflect.TypeOf(""),
		"MediaType":            reflect.TypeOf(""),
		"NetworkID":            reflect.TypeOf(0),
		"NetworkName":          reflect.TypeOf(""),
		"PrimaryCategoryID":    reflect.TypeOf(""),
		"RendererName":         reflect.TypeOf(""),
		"RendererVersion":      reflect.TypeOf(""),
		"RendererData":         reflect.TypeOf(json.RawMessage{}),
		"RendererUrl":          reflect.TypeOf(""),
		"SecondaryCategoryIDs": reflect.TypeOf([]string{}),
	}
	structType := reflect.TypeOf(openrtb_ext.ExtBidPrebidMeta{})
	err := ValidateStructFields(expectedFields, structType)
	if err != nil {
		t.Error(err)
	}
}
