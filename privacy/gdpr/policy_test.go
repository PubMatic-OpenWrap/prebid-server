package gdpr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConsent(t *testing.T) {
	testCases := []struct {
		description string
		consent     string
		expected    bool
	}{
		{
			description: "Invalid",
			consent:     "<any invalid>",
			expected:    false,
		},
		{
			description: "TCF2 Valid",
			consent:     "COzTVhaOzTVhaGvAAAENAiCIAP_AAH_AAAAAAEEUACCKAAA",
			expected:    true,
		},
	}

	for _, test := range testCases {
		result := ValidateConsent(test.consent)
		assert.Equal(t, test.expected, result, test.description)
	}
}

func TestValidateConsent(t *testing.T) {
	testCases := []struct {
		description string
		consent     string
		expectError bool
	}{
		{
			description: "Invalid",
			consent:     "<any invalid>",
			expectError: true,
		},
		{
			description: "Valid",
			consent:     "BONV8oqONXwgmADACHENAO7pqzAAppY",
			expectError: false,
		},
	}

	for _, test := range testCases {
		result := ValidateConsent(test.consent)

		if test.expectError {
			assert.Error(t, result, test.description)
		} else {
			assert.NoError(t, result, test.description)
		}
	}
}
