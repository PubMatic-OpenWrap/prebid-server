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
		expectedError bool
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
			expectedError: false,
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
			expectedError: true,
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
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestValidateBidVideo(t *testing.T) {
	testCases := []struct {
		name          string
		video         any
		expectedVideo any
		expectedError bool
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
			expectedError: false,
		},
		{
			name: "Invalid duration type",
			video: map[string]any{
				"duration":         "30",
				"primary_category": "sports",
			},
			expectedVideo: nil,
			expectedError: true,
		},
		{
			name: "Invalid primary category type",
			video: map[string]any{
				"duration":         30.0,
				"primary_category": 123,
			},
			expectedVideo: nil,
			expectedError: true,
		},
		{
			name:          "Invalid type (not a map)",
			video:         make(chan int),
			expectedVideo: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validatedVideo, err := validateBidVideo(tc.video)
			assert.Equal(t, tc.expectedVideo, validatedVideo)
			assert.Equal(t, tc.expectedError, err != nil)
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
			_ = resolver.setValue(tc.adapterBid, tc.value)
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
		expectedError bool
	}{
		{
			name:          "Not found dur in location",
			responseNode:  map[string]any{},
			expectedValue: nil,
			expectedError: false,
		},
		{
			name: "Found dur in location",
			responseNode: map[string]any{
				"dur": 11.0,
			},
			expectedValue: int64(11),
			expectedError: false,
		},
		{
			name: "Found dur in location but type is invalid",
			responseNode: map[string]any{
				"dur": "invalid",
			},
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.getFromORTBObject(tc.responseNode)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
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
		expectedError bool
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
			expectedError: false,
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
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "dur not found in location",
			responseNode:  map[string]any{},
			path:          "seat",
			expectedValue: nil,
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
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
		expectedError   bool
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
			expectedError: false,
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
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.setValue(tc.adapterBid, tc.value)
			assert.Equal(t, tc.expectedError, result != nil)
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
		expectedError bool
	}{
		{
			name: "Found category in location",
			responseNode: map[string]any{
				"cat": []any{"IAB-1", "IAB-2"},
			},
			path:          "cat.1",
			expectedValue: "IAB-2",
			expectedError: false,
		},
		{
			name: "Found category in location but type is invalid",
			responseNode: map[string]any{
				"cat": []any{"IAB-1", 100},
			},
			path:          "cat.1",
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Category not found in location",
			responseNode:  map[string]any{},
			path:          "seat",
			expectedValue: nil,
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.retrieveFromBidderParamLocation(tc.responseNode, tc.path)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
		})
	}
}

func TestBidVideoPrimaryCategoryGetFromORTBObject(t *testing.T) {
	resolver := &bidVideoPrimaryCategoryResolver{}
	testCases := []struct {
		name          string
		responseNode  map[string]any
		expectedValue any
		expectedError bool
	}{
		{
			name: "Found category in location",
			responseNode: map[string]any{
				"cat": []any{"IAB-1", "IAB-2"},
			},
			expectedValue: "IAB-1",
			expectedError: false,
		},
		{
			name: "Found empty category in location",
			responseNode: map[string]any{
				"cat": []any{},
			},
			expectedValue: nil,
			expectedError: false,
		},
		{
			name: "Not found category in location",
			responseNode: map[string]any{
				"field": []any{},
			},
			expectedValue: nil,
			expectedError: false,
		},
		{
			name: "Found category in location but type is invalid",
			responseNode: map[string]any{
				"cat": "invalid",
			},
			expectedValue: nil,
			expectedError: true,
		},
		{
			name: "Found category in location but first category type is invalid",
			responseNode: map[string]any{
				"cat": []any{1, 2},
			},
			expectedValue: nil,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolver.getFromORTBObject(tc.responseNode)
			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedError, err != nil)
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
		expectedError   bool
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
			expectedError: false,
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
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := resolver.setValue(tc.adapterBid, tc.value)
			assert.Equal(t, tc.expectedError, err != nil)
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
		expectedError   bool
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
			expectedError: false,
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
			expectedError: false,
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
			expectedError: false,
		},
		{
			name: "Invalid video type",
			adapterBid: map[string]any{
				"BidVideo": "invalid",
			},
			key:             "duration",
			value:           30,
			expectedAdapter: map[string]any{"BidVideo": "invalid"},
			expectedError:   true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setKeyValueInBidVideo(tc.adapterBid, tc.key, tc.value)
			assert.Equal(t, tc.expectedError, err != nil)
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
	err := ValidateStructFields(expectedFields, structType)
	if err != nil {
		t.Error(err)
	}
}
