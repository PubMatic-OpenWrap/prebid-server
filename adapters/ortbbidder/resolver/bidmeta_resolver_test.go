package resolver

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBidMetaRetrieveFromLocation(t *testing.T) {
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
			expectedFound: true,
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
			expectedFound: false,
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
			expectedFound: false,
		},
	}
	resolver := &bidMetaResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestValidateBidMeta(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected any
		valid    bool
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
			valid: true,
		},
		{
			name: "Metadata with wrong type",
			value: map[string]any{
				bidMetaAdvertiserDomainsKey: "example.com", // should be a slice
				bidMetaAdvertiserIdKey:      "123",         // should be an float
			},
			expected: nil,
			valid:    false,
		},
		{
			name:     "Invalid type for value",
			value:    make(chan int),
			expected: nil,
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := validateBidMeta(tt.value)
			assert.Equal(t, tt.valid, valid)
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
			resolver.setValue(tc.typeBid, tc.value)
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
									"adomains": []any{"abc.com", "xyz.com"},
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.adomains",
			expectedValue: []string{"abc.com", "xyz.com"},
			expectedFound: true,
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
			expectedValue: []string(nil),
			expectedFound: false,
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
			expectedFound: false,
		},
	}
	resolver := &bidMetaAdvDomainsResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
			resolver.setValue(tc.typeBid, tc.value)
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
									"advid": 10.0,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.advid",
			expectedValue: 10,
			expectedFound: true,
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
			expectedValue: 0,
			expectedFound: false,
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
			expectedFound: false,
		},
	}
	resolver := &bidMetaAdvIDResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
			resolver.setValue(tc.typeBid, tc.value)
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
		expectedFound bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur":     "USD",
				"advname": "Acme Corp",
			},
			path:          "advname",
			expectedValue: "Acme Corp",
			expectedFound: true,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"ext": map[string]any{
					"advname": 123,
				},
			},
			path:          "ext.advname",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	resolver := &bidMetaAdvNameResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
		},
		{
			name: "Found in location but data type is other than float",
			ortbResponse: map[string]any{
				"cur":      "USD",
				"agencyid": 10,
			},
			path:          "agencyid",
			expectedValue: 0,
			expectedFound: false,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur":     "USD",
				"seatbid": []any{},
			},
			path:          "ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	resolver := &bidMetaAgencyIDResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
									"brandid": 10.0,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.brandid",
			expectedValue: 10,
			expectedFound: true,
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
			expectedValue: 0,
			expectedFound: false,
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
			expectedFound: false,
		},
	}
	resolver := &bidMetaBrandIDResolver{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: map[string]any(nil),
			expectedFound: false,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
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
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
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
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
									"networkID": 100.0,
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.networkID",
			expectedValue: 100,
			expectedFound: true,
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
			expectedValue: 0,
			expectedFound: false,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur":     "USD",
				"seatbid": []any{},
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
	}{
		{
			name: "Found in location",
			ortbResponse: map[string]any{
				"cur":          "USD",
				"primaryCatId": "testCategory",
			},
			path:          "primaryCatId",
			expectedValue: "testCategory",
			expectedFound: true,
		},
		{
			name: "Found in location but data type is other than string",
			ortbResponse: map[string]any{
				"cur":          "USD",
				"primaryCatId": 12345,
			},
			path:          "primaryCatId",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name: "Not found in location",
			ortbResponse: map[string]any{
				"cur": "USD",
			},
			path:          "nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: map[string]any(nil),
			expectedFound: false,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "seatbid.0.bid.0.ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedValue: []string(nil),
			expectedFound: false,
		},
		{
			name:          "Not found in location",
			ortbResponse:  map[string]any{},
			path:          "ext.nonexistent",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.ortbResponse, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
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
		expectedFound bool
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
			expectedFound: true,
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
			expectedFound: true,
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
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := setKeyValueInBidMeta(tt.adapterBid, tt.key, tt.value)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedBid, tt.adapterBid)
		})
	}
}
