package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertBoolToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected int
	}{
		{
			name:     "true_value",
			input:    true,
			expected: 1,
		},
		{
			name:     "false_value",
			input:    false,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertBoolToInt(tt.input)
			assert.Equal(t, tt.expected, result, "ConvertBoolToInt(%v) should return %d", tt.input, tt.expected)
		})
	}
}
