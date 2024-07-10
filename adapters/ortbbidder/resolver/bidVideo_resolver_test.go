package resolver

import (
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestBidVideoRetrieveFromLocation(t *testing.T) {
	resolver := &bidVideoResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		path          string
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found bidVideo in location",
			responseNode: map[string]any{
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"duration": 100.0,
								"ext": map[string]any{
									"video": map[string]any{
										"duration":         11.0,
										"primary_category": "sport",
										"extra_key":        "extra_value",
									},
								},
							},
						},
					},
				},
			},
			path: "seatbid.0.bid.0.ext.video",
			expectedValue: map[string]any{
				"duration":         11.0,
				"primary_category": "sport",
				"extra_key":        "extra_value",
			},
			expectedFound: true,
		},
		{
			name: "bidVideo found but few fields are invalid",
			responseNode: map[string]any{
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"duration": 100.0,
								"ext": map[string]any{
									"video": map[string]any{
										"duration":         "11", // invalid
										"primary_category": "sport",
										"extra_key":        "extra_value",
									},
								},
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.ext.video",
			expectedValue: nil,
			expectedFound: false,
		},
		{
			name: "bidVideo not found in location",
			responseNode: map[string]any{
				"seatbid": []any{
					map[string]any{},
				},
			},
			path:          "seatbid.0.bid.0.ext.video",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestValidateBidVideo(t *testing.T) {
	testCases := []struct {
		name            string
		video           any
		expectedVideo   any
		expectedIsValid bool
	}{
		{
			name: "Valid video map",
			video: map[string]any{
				"duration":         30.0,
				"primary_category": "sports",
				"extra_key":        "extra_value",
			},
			expectedVideo: map[string]any{
				"duration":         30.0,
				"primary_category": "sports",
				"extra_key":        "extra_value",
			},
			expectedIsValid: true,
		},
		{
			name: "Invalid duration type",
			video: map[string]any{
				"duration":         "30",
				"primary_category": "sports",
			},
			expectedVideo:   nil,
			expectedIsValid: false,
		},
		{
			name: "Invalid primary category type",
			video: map[string]any{
				"duration":         30.0,
				"primary_category": 123,
			},
			expectedVideo:   nil,
			expectedIsValid: false,
		},
		{
			name:            "Invalid type (not a map)",
			video:           make(chan int),
			expectedVideo:   nil,
			expectedIsValid: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validatedVideo, isValid := validateBidVideo(tc.video)
			assert.Equal(t, tc.expectedVideo, validatedVideo)
			assert.Equal(t, tc.expectedIsValid, isValid)
		})
	}
}

func TestBidVideoSetValue(t *testing.T) {
	resolver := &bidVideoResolver{}
	testCases := []struct {
		name            string
		adapterBid      map[string]any
		value           any
		expectedAdapter map[string]any
	}{
		{
			name:       "Set bidVideo value",
			adapterBid: map[string]any{},
			value: map[string]any{
				"duration":         30,
				"primary_category": "IAB-1",
			},
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					"duration":         30,
					"primary_category": "IAB-1",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver.setValue(tc.adapterBid, tc.value)
			assert.Equal(t, tc.expectedAdapter, tc.adapterBid)
		})
	}
}

func TestBidVideoDurationGetFromORTBObject(t *testing.T) {
	resolver := &bidVideoDurationResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found dur in location",
			responseNode: map[string]any{
				"dur": 11.0,
			},
			expectedValue: int64(11),
			expectedFound: true,
		},
		{
			name: "Found dur in location but type is invalid",
			responseNode: map[string]any{
				"dur": "invalid",
			},
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.getFromORTBObject(tc.responseNode)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestBidVideoDurarionRetrieveFromLocation(t *testing.T) {
	resolver := &bidVideoDurationResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		path          string
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found dur in location",
			responseNode: map[string]any{
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"duration": 100.0,
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.duration",
			expectedValue: int64(100),
			expectedFound: true,
		},
		{
			name: "Found dur in location but type is invalid",
			responseNode: map[string]any{
				"seatbid": []any{
					map[string]any{
						"bid": []any{
							map[string]any{
								"duration": 100,
							},
						},
					},
				},
			},
			path:          "seatbid.0.bid.0.duration",
			expectedValue: int64(0),
			expectedFound: false,
		},
		{
			name:          "dur not found in location",
			responseNode:  map[string]any{},
			path:          "seat",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestSetValueBidVideoDuration(t *testing.T) {
	resolver := &bidVideoDurationResolver{}
	testCases := []struct {
		name            string
		adapterBid      map[string]any
		value           any
		expectedAdapter map[string]any
		expectedResult  bool
	}{
		{
			name:       "Set video duration when video is absent",
			adapterBid: map[string]any{},
			value:      10,
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					bidVideoDurationKey: 10,
				},
			},
			expectedResult: true,
		},
		{
			name: "Set videoduration when video is present",
			adapterBid: map[string]any{
				"BidVideo": map[string]any{
					bidVideoPrimaryCategoryKey: "IAB-1",
				},
			},
			value: 10,
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					"duration":                 10,
					bidVideoPrimaryCategoryKey: "IAB-1",
				},
			},
			expectedResult: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.setValue(tc.adapterBid, tc.value)
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedAdapter, tc.adapterBid)
		})
	}
}

func TestBidVideoPrimaryCategoryRetrieveFromLocation(t *testing.T) {
	resolver := &bidVideoPrimaryCategoryResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		path          string
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found category in location",
			responseNode: map[string]any{
				"cat": []any{"IAB-1", "IAB-2"},
			},
			path:          "cat.1",
			expectedValue: "IAB-2",
			expectedFound: true,
		},
		{
			name: "Found category in location but type is invalid",
			responseNode: map[string]any{
				"cat": []any{"IAB-1", 100},
			},
			path:          "cat.1",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "Category not found in location",
			responseNode:  map[string]any{},
			path:          "seat",
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestBidVideoPrimaryCategoryGetFromORTBObject(t *testing.T) {
	resolver := &bidVideoPrimaryCategoryResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		expectedValue any
		expectedFound bool
	}{
		{
			name: "Found category in location",
			responseNode: map[string]any{
				"cat": []any{"IAB-1", "IAB-2"},
			},
			expectedValue: "IAB-1",
			expectedFound: true,
		},
		{
			name: "Found category in location but type is invalid",
			responseNode: map[string]any{
				"cat": "invalid",
			},
			expectedValue: nil,
			expectedFound: false,
		},
		{
			name: "Found category in location but first category type is invalid",
			responseNode: map[string]any{
				"cat": []any{1, 2},
			},
			expectedValue: nil,
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := resolver.getFromORTBObject(tc.responseNode)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestSetValuePrimaryCategory(t *testing.T) {
	resolver := &bidVideoPrimaryCategoryResolver{}
	testCases := []struct {
		name            string
		adapterBid      map[string]any
		value           any
		expectedAdapter map[string]any
		expectedResult  bool
	}{
		{
			name:       "Set video key-value when video is absent",
			adapterBid: map[string]any{},
			value:      "IAB-1",
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					bidVideoPrimaryCategoryKey: "IAB-1",
				},
			},
			expectedResult: true,
		},
		{
			name: "Set video key-value when video is present",
			adapterBid: map[string]any{
				"BidVideo": map[string]any{
					"duration": 30,
				},
			},
			value: "IAB-1",
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					"duration":                 30,
					bidVideoPrimaryCategoryKey: "IAB-1",
				},
			},
			expectedResult: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.setValue(tc.adapterBid, tc.value)
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedAdapter, tc.adapterBid)
		})
	}
}

func TestSetKeyValueInBidVideo(t *testing.T) {
	testCases := []struct {
		name            string
		adapterBid      map[string]any
		key             string
		value           any
		expectedAdapter map[string]any
		expectedResult  bool
	}{
		{
			name:       "Set video key-value when video is absent",
			adapterBid: map[string]any{},
			key:        "duration",
			value:      30,
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					"duration": 30,
				},
			},
			expectedResult: true,
		},
		{
			name: "Set video key-value when video is present",
			adapterBid: map[string]any{
				"BidVideo": map[string]any{},
			},
			key:   "duration",
			value: 30,
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					"duration": 30,
				},
			},
			expectedResult: true,
		},
		{
			name: "Override existing video key-value",
			adapterBid: map[string]any{
				"BidVideo": map[string]any{
					"duration": 15,
				},
			},
			key:   "duration",
			value: 30,
			expectedAdapter: map[string]any{
				"BidVideo": map[string]any{
					"duration": 30,
				},
			},
			expectedResult: true,
		},
		{
			name: "Invalid video type",
			adapterBid: map[string]any{
				"BidVideo": "invalid",
			},
			key:             "duration",
			value:           30,
			expectedAdapter: map[string]any{"BidVideo": "invalid"},
			expectedResult:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := setKeyValueInBidVideo(tc.adapterBid, tc.key, tc.value)
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedAdapter, tc.adapterBid)
		})
	}
}

// TestExtBidPrebidVideo notifies us of any changes in the openrtb_ext.ExtBidPrebidVideo struct.
// If a new field is added in openrtb_ext.ExtBidPrebidVideo, then add the support to resolve the new field and update the test case.
// If the data type of an existing field changes then update the resolver of the respective field.
func TestExtBidPrebidVideoFields(t *testing.T) {
	// Expected field count and types
	expectedFields := map[string]reflect.Type{
		"Duration":        reflect.TypeOf(0),
		"PrimaryCategory": reflect.TypeOf(""),
		"VASTTagID":       reflect.TypeOf(""), // not expected to be set by adapter
	}

	structType := reflect.TypeOf(openrtb_ext.ExtBidPrebidVideo{})
	fieldCount := structType.NumField()

	// Check if the number of fields matches the expected count
	if fieldCount != len(expectedFields) {
		t.Errorf("Expected %d fields, but got %d fields", len(expectedFields), fieldCount)
	}

	// Check if the field types match the expected types
	for i := 0; i < fieldCount; i++ {
		field := structType.Field(i)
		expectedType, ok := expectedFields[field.Name]
		if !ok {
			t.Errorf("Unexpected field: %s", field.Name)
		}
		if field.Type != expectedType {
			t.Errorf("Field %s: expected type %v, but got %v", field.Name, expectedType, field.Type)
		}
	}
}
