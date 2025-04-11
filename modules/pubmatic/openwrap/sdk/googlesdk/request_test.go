package googlesdk

import (
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestGetSignalData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)

	tests := []struct {
		name     string
		input    string
		setup    func()
		expected *openrtb2.BidRequest
	}{
		{
			name:  "Empty body",
			input: "",
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.MissingSignal)
			},
			expected: nil,
		},
		{
			name:  "Invalid JSON",
			input: "{invalid-json",
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.MissingSignal)
			},
			expected: nil,
		},
		{
			name:  "Missing imp array",
			input: `{"someKey": "someValue"}`,
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.MissingSignal)
			},
			expected: nil,
		},
		{
			name:  "Empty imp array",
			input: `{"imp": []}`,
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.MissingSignal)
			},
			expected: nil,
		},
		{
			name:  "Missing ext in imp",
			input: `{"imp": [{"id": "1"}]}`,
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.MissingSignal)
			},
			expected: nil,
		},
		{
			name: "Wrong adapter ID",
			input: `{
				"imp": [{
					"ext": {
						"buyer_generated_request_data": [{
							"source_app": {"id": "wrong.adapter.id"},
							"data": "{\"id\":\"test-id\"}"
						}]
					}
				}]
			}`,
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.MissingSignal)
			},
			expected: nil,
		},
		{
			name: "Valid signal data",
			input: `{
				"imp": [{
					"ext": {
						"buyer_generated_request_data": [{
							"source_app": {"id": "com.google.ads.mediation.pubmatic.PubMaticMediationAdapter"},
							"data": "{\"id\":\"test-id\",\"app\":{\"id\":\"app-123\"}}"
						}]
					}
				}]
			}`,
			expected: &openrtb2.BidRequest{
				ID: "test-id",
				App: &openrtb2.App{
					ID: "app-123",
				},
			},
		},
		{
			name: "Invalid signal data JSON",
			input: `{
				"imp": [{
					"ext": {
						"buyer_generated_request_data": [{
							"source_app": {"id": "com.google.ads.mediation.pubmatic.PubMaticMediationAdapter"},
							"data": "{invalid-json}"
						}]
					}
				}]
			}`,
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.InvalidSignal)
			},
			expected: nil,
		},
		{
			name: "Empty signal data",
			input: `{
				"imp": [{
					"ext": {
						"buyer_generated_request_data": [{
							"source_app": {"id": "com.google.ads.mediation.pubmatic.PubMaticMediationAdapter"},
							"data": ""
						}]
					}
				}]
			}`,
			setup: func() {
				mockEngine.EXPECT().RecordSignalDataStatus("5890", "123", models.MissingSignal)
			},
			expected: nil,
		},
		{
			name: "Multiple apps with valid signal",
			input: `{
				"imp": [{
					"ext": {
						"buyer_generated_request_data": [
							{
								"source_app": {"id": "other.app"},
								"data": "{\"id\":\"wrong-id\"}"
							},
							{
								"source_app": {"id": "com.google.ads.mediation.pubmatic.PubMaticMediationAdapter"},
								"data": "{\"id\":\"test-id\",\"app\":{\"id\":\"app-123\"}}"
							}
						]
					}
				}]
			}`,
			expected: &openrtb2.BidRequest{
				ID: "test-id",
				App: &openrtb2.App{
					ID: "app-123",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			result := getSignalData([]byte(tt.input), models.RequestCtx{
				MetricsEngine: mockEngine,
			}, &wrapperData{
				PublisherId: "5890",
				ProfileId:   "123",
			})

			if tt.expected == nil {
				assert.Nil(t, result, "Expected nil result for test: %s", tt.name)
				return
			}

			assert.NotNil(t, result, "Expected non-nil result for test: %s", tt.name)
			expectedJSON, err := json.Marshal(tt.expected)
			assert.NoError(t, err, "Failed to marshal expected value for test: %s", tt.name)
			resultJSON, err := json.Marshal(result)
			assert.NoError(t, err, "Failed to marshal result for test: %s", tt.name)
			assert.JSONEq(t, string(expectedJSON), string(resultJSON), "Unexpected result for test: %s", tt.name)
		})
	}
}
func TestGetWrapperData(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *wrapperData
		expectedErr string
	}{
		{
			name:        "Empty body",
			input:       "",
			expected:    nil,
			expectedErr: "empty request body",
		},
		{
			name:        "Invalid JSON",
			input:       "{invalid-json",
			expected:    nil,
			expectedErr: "failed to get Keyval object",
		},
		{
			name:        "Missing imp array",
			input:       `{"someKey": "someValue"}`,
			expected:    nil,
			expectedErr: "failed to get Keyval object",
		},
		{
			name:        "Empty imp array",
			input:       `{"imp": []}`,
			expected:    nil,
			expectedErr: "failed to get Keyval object",
		},
		{
			name:        "Missing ext in imp",
			input:       `{"imp": [{"id": "1"}]}`,
			expected:    nil,
			expectedErr: "failed to get Keyval object",
		},
		{
			name:        "Missing Keyval in ad_unit_mapping",
			input:       `{"imp": [{"ext": {"ad_unit_mapping": {}}}]}`,
			expected:    nil,
			expectedErr: "failed to get Keyval object",
		},
		{
			name: "Valid wrapper data",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [
								{"key": "publisher_id", "value": "12345"},
								{"key": "profile_id", "value": "67890"},
								{"key": "ad_unit_id", "value": "tag-123"}
							]
						}
					}
				}]
			}`,
			expected: &wrapperData{
				PublisherId: "12345",
				ProfileId:   "67890",
				TagId:       "tag-123",
			},
			expectedErr: "",
		},
		{
			name: "Partial wrapper data",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [
								{"key": "publisher_id", "value": "12345"}
							]
						}
					}
				}]
			}`,
			expected: &wrapperData{
				PublisherId: "12345",
			},
			expectedErr: "",
		},
		{
			name: "Invalid Keyval structure",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [
								{"key": "publisher_id"}
							]
						}
					}
				}]
			}`,
			expected:    nil,
			expectedErr: "",
		},
		{
			name: "No matching keys in Keyval",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [
								{"key": "unknown_key", "value": "value"}
							]
						}
					}
				}]
			}`,
			expected:    nil,
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getWrapperData([]byte(tt.input))
			if tt.expectedErr != "" {
				assert.Nil(t, result, "Expected nil result for test: %s", tt.name)
				assert.EqualError(t, err, tt.expectedErr, "Unexpected error for test: %s", tt.name)
				return
			}

			assert.NoError(t, err, "Unexpected error for test: %s", tt.name)
			assert.Equal(t, tt.expected, result, "Unexpected result for test: %s", tt.name)
		})
	}
}
