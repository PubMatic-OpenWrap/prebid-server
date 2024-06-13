package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestIsThrottled(t *testing.T) {
	tests := []struct {
		name            string
		percentage      *percentageValue
		randomNumbers   []int /*inclusive random number 0-99*/
		expectedLogger  bool
		expectedTracker bool
	}{
		{
			name:            "logger_tracker_100%_throttled",
			percentage:      &percentageValue{logger: 100, tracker: 100},
			randomNumbers:   []int{99},
			expectedLogger:  true,
			expectedTracker: true,
		},
		{
			name:            "both_throttled",
			percentage:      &percentageValue{logger: 50, tracker: 60},
			randomNumbers:   []int{20, 20},
			expectedLogger:  true,
			expectedTracker: true,
		},
		{
			name:            "logger_tracker_not_throttled",
			percentage:      &percentageValue{logger: 50, tracker: 0},
			randomNumbers:   []int{70, 20},
			expectedLogger:  false,
			expectedTracker: false,
		},
		{
			name:            "logger_not_throttled_tracker_throttled",
			percentage:      &percentageValue{logger: 50, tracker: 60},
			randomNumbers:   []int{70, 20},
			expectedLogger:  false,
			expectedTracker: true,
		},
		{
			name:            "logger_disabled_tracker_not_throttled",
			percentage:      &percentageValue{logger: -1, tracker: 50},
			randomNumbers:   []int{70},
			expectedLogger:  true,
			expectedTracker: false,
		},
		{
			name:            "logger_disabled_tracker_throttled",
			percentage:      &percentageValue{logger: -1, tracker: 50},
			randomNumbers:   []int{20},
			expectedLogger:  true,
			expectedTracker: true,
		},
		{
			name:            "logger_disabled_tracker_100%_throttled",
			percentage:      &percentageValue{logger: -1, tracker: 100},
			randomNumbers:   []int{99},
			expectedLogger:  true,
			expectedTracker: true,
		},
		{
			name:            "neither_throttled",
			percentage:      &percentageValue{logger: 0, tracker: 0},
			randomNumbers:   []int{20, 20},
			expectedLogger:  false,
			expectedTracker: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count := 0
			randInt = func(_ int) int {
				value := test.randomNumbers[count]
				count++
				return value
			}
			logger, tracker := test.percentage.isThrottled()
			assert.Equal(t, test.expectedLogger, logger)
			assert.Equal(t, test.expectedTracker, tracker)
		})
	}
}

func TestGetKeyParts(t *testing.T) {
	type want struct {
		expectedPubID          int
		expectedProfileID      int
		expectedLoggerPercent  int
		expectedTrackerPercent int
		expectedError          bool
	}
	tests := []struct {
		name string
		key  string
		want want
	}{
		{
			name: "ValidKey",
			key:  "123:456:50:60",
			want: want{
				expectedPubID:          123,
				expectedProfileID:      456,
				expectedLoggerPercent:  50,
				expectedTrackerPercent: 60,
				expectedError:          false,
			},
		},
		{
			name: "InvalidKeyFormat",
			key:  "invalid",
			want: want{expectedError: true},
		},
		{
			name: "InvalidPubID",
			key:  "abc:456:50:60",
			want: want{expectedError: true},
		},
		{
			name: "InvalidPubID_Negative",
			key:  "-1:456:50:60",
			want: want{expectedError: true},
		},
		{
			name: "InvalidProfileID",
			key:  "123:def:50:60",
			want: want{expectedError: true},
		},
		{
			name: "InvalidProfileID_Negative",
			key:  "123:-1:50:60",
			want: want{expectedError: true},
		},
		{
			name: "InvalidLoggerPercent",
			key:  "123:456:hij:60",
			want: want{expectedError: true},
		},
		{
			name: "InvalidLoggerPercent_Negative",
			key:  "123:456:-50:60",
			want: want{expectedError: true},
		},
		{
			name: "InvalidTrackerPercent",
			key:  "123:456:50:klm",
			want: want{expectedError: true},
		},
		{
			name: "InvalidTrackerPercent_Negative",
			key:  "123:456:50:-60",
			want: want{expectedError: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pubID, profileID, loggerPercent, trackerPercent, err := getKeyParts(test.key)

			assert.Equal(t, test.want.expectedPubID, pubID)
			assert.Equal(t, test.want.expectedProfileID, profileID)
			assert.Equal(t, test.want.expectedLoggerPercent, loggerPercent)
			assert.Equal(t, test.want.expectedTrackerPercent, trackerPercent)

			if test.want.expectedError {
				assert.NotNil(t, err)
				return
			}
		})
	}
}

func TestPubThrottlingAdd(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected pubThrottling
	}{
		{
			name:     "ValidValue",
			value:    "1:1:50:60,2:1:40:70",
			expected: pubThrottling{1: {1: {logger: 50, tracker: 60}}, 2: {1: {logger: 40, tracker: 70}}},
		},
		{
			name:     "EmptyValue",
			value:    "",
			expected: pubThrottling{},
		},
		{
			name:     "InvalidValue",
			value:    "invalid",
			expected: pubThrottling{},
		},
		{
			name:     "MissingValues",
			value:    "1:1:50:60,:1:40:70",
			expected: pubThrottling{1: {1: {logger: 50, tracker: 60}}},
		},
		{
			name:     "NegativeValues",
			value:    "-1:1:50:60,2:1:-40:70",
			expected: pubThrottling{},
		},
		{
			name:     "DisabledLoggerAndEnabledTracker",
			value:    "1:1:-1:70",
			expected: pubThrottling{1: {1: {logger: -1, tracker: 70}}},
		},
		{
			name:     "MultipleProfiles",
			value:    "1:1:50:60,1:2:40:70",
			expected: pubThrottling{1: {1: {logger: 50, tracker: 60}, 2: {logger: 40, tracker: 70}}},
		},
		{
			name:     "DuplicateEntries",
			value:    "1:1:50:60,1:1:40:70",
			expected: pubThrottling{1: {1: {logger: 40, tracker: 70}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := newPubThrottling(test.value)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name            string
		ant             pubThrottling
		source          pubThrottling
		replaceExisting bool
		expected        pubThrottling
	}{
		{
			name: "MergeWithReplaceExistingTrue",
			ant: pubThrottling{
				1: {1: {logger: 50, tracker: 60}},
			},
			source: pubThrottling{
				1: {1: {logger: 70, tracker: 80}},
				2: {2: {logger: 90, tracker: 100}},
			},
			replaceExisting: true,
			expected: pubThrottling{
				1: {1: {logger: 70, tracker: 80}},
				2: {2: {logger: 90, tracker: 100}},
			},
		},
		{
			name: "MergeWithReplaceExistingFalse",
			ant: pubThrottling{
				1: {1: {logger: 50, tracker: 60}},
			},
			source: pubThrottling{
				1: {1: {logger: 70, tracker: 80}},
				2: {2: {logger: 90, tracker: 100}},
			},
			replaceExisting: false,
			expected: pubThrottling{
				1: {1: {logger: 50, tracker: 60}},
				2: {2: {logger: 90, tracker: 100}},
			},
		},
		// Add more test cases as needed
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.ant.merge(test.source, test.replaceExisting)
			assert.Equal(t, test.expected, test.ant)
		})
	}
}

func TestUpdateAnalyticsThrottling(t *testing.T) {
	tests := []struct {
		name        string
		feature     *feature
		expectedAnt pubThrottling
	}{
		{
			name:        "NilPublisherFeature",
			feature:     &feature{},
			expectedAnt: nil,
		},
		{
			name: "WithPublisherFeature",
			feature: &feature{
				publisherFeature: map[int]map[int]models.FeatureData{
					1: {
						models.FeatureAnalyticsThrottle: {Enabled: 1, Value: "1:1:50:60"},
					},
				},
			},
			expectedAnt: pubThrottling{
				1: {1: {logger: 50, tracker: 60}},
			},
		},
		{
			name: "EmptyAntDB",
			feature: &feature{
				publisherFeature: map[int]map[int]models.FeatureData{
					1: {
						models.FeatureAnalyticsThrottle: {Enabled: 1, Value: "1:1:50:60"},
					},
				},
			},
			expectedAnt: pubThrottling{
				1: {1: {logger: 50, tracker: 60}},
			},
		},
		{
			name: "ExistingAntDB",
			feature: &feature{
				ant: analyticsThrottle{
					vault: pubThrottling{
						1: {1: {logger: 30, tracker: 40}},
					},
					db: pubThrottling{
						1: {1: {logger: 30, tracker: 40}, 2: {logger: 30, tracker: 40}},
						2: {1: {logger: 30, tracker: 40}},
					},
				},
				publisherFeature: map[int]map[int]models.FeatureData{
					1: {
						models.FeatureAnalyticsThrottle: {Enabled: 1, Value: "1:1:50:60"},
					},
				},
			},
			expectedAnt: pubThrottling{
				1: {1: {logger: 50, tracker: 60}, 2: {logger: 30, tracker: 40}},
				2: {1: {logger: 30, tracker: 40}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.feature.updateAnalyticsThrottling()
			assert.Equal(t, test.expectedAnt, test.feature.ant.db)
		})
	}
}
