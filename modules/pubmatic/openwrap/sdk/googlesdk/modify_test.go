package googlesdk

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestGetSignalData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *openrtb2.BidRequest
	}{
		{
			name:     "Empty body",
			input:    "",
			expected: nil,
		},
		{
			name:     "Invalid JSON",
			input:    "{invalid-json",
			expected: nil,
		},
		{
			name:     "Missing imp array",
			input:    `{"someKey": "someValue"}`,
			expected: nil,
		},
		{
			name:     "Empty imp array",
			input:    `{"imp": []}`,
			expected: nil,
		},
		{
			name:     "Missing ext in imp",
			input:    `{"imp": [{"id": "1"}]}`,
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
			result := getSignalData([]byte(tt.input))
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
		name     string
		input    string
		expected *wrapperData
		wantErr  bool
	}{
		{
			name:     "Empty body",
			input:    "",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid JSON",
			input:    "{invalid-json",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Missing imp array",
			input:    `{"someKey": "someValue"}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Missing ad_unit_mapping",
			input:    `{"imp": [{"ext": {}}]}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Missing Keyval",
			input:    `{"imp": [{"ext": {"ad_unit_mapping": {}}}]}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Keyval not an array",
			input:    `{"imp": [{"ext": {"ad_unit_mapping": {"Keyval": {}}}}]}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Empty Keyval array",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": []
						}
					}
				}]
			}`,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "Invalid Keyval array element",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [123]
						}
					}
				}]
			}`,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "Empty values in Keyval",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [{
								"publisher_id": "",
								"profile_id": "",
								"ad_unit_id": ""
							}]
						}
					}
				}]
			}`,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "Complete valid data",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [{
								"publisher_id": "5890",
								"profile_id": "2345",
								"ad_unit_id": "/tag/id"
							}]
						}
					}
				}]
			}`,
			expected: &wrapperData{
				PublisherId: "5890",
				ProfileId:   "2345",
				TagId:       "/tag/id",
			},
			wantErr: false,
		},
		{
			name: "Multiple Keyval elements with data spread",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [
								{
									"publisher_id": "5890",
									"profile_id": "2345"
								},
								{
									"ad_unit_id": "/tag/id"
								}
							]
						}
					}
				}]
			}`,
			expected: &wrapperData{
				PublisherId: "5890",
				ProfileId:   "2345",
				TagId:       "/tag/id",
			},
			wantErr: false,
		},
		{
			name: "Partial data with only publisher_id",
			input: `{
				"imp": [{
					"ext": {
						"ad_unit_mapping": {
							"Keyval": [{
								"publisher_id": "5890"
							}]
						}
					}
				}]
			}`,
			expected: &wrapperData{
				PublisherId: "5890",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getWrapperData([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test: %s", tt.name)
				assert.Nil(t, result, "Expected nil result for test: %s", tt.name)
				return
			}

			if tt.expected == nil {
				assert.Nil(t, result, "Expected nil result for test: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for test: %s", tt.name)
				assert.Equal(t, tt.expected, result, "Unexpected result for test: %s", tt.name)
			}
		})
	}
}
