package sdkutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyPath(t *testing.T) {
	tests := []struct {
		name      string
		source    []byte
		target    []byte
		path      []string
		expected  []byte
		expectErr bool
	}{
		{
			name:      "Nil source",
			source:    nil,
			target:    []byte(`{"key":"value"}`),
			path:      []string{"key"},
			expected:  []byte(`{"key":"value"}`),
			expectErr: false,
		},
		{
			name:      "Nil target",
			source:    []byte(`{"key":"value"}`),
			target:    nil,
			path:      []string{"key"},
			expected:  []byte(`{"key":"value"}`),
			expectErr: false,
		},
		{
			name:      "Copy string value",
			source:    []byte(`{"key":"value"}`),
			target:    []byte(`{"other_key":"other_value"}`),
			path:      []string{"key"},
			expected:  []byte(`{"other_key":"other_value","key":"value"}`),
			expectErr: false,
		},
		{
			name:      "Copy number value",
			source:    []byte(`{"key":123}`),
			target:    []byte(`{}`),
			path:      []string{"key"},
			expected:  []byte(`{"key":123}`),
			expectErr: false,
		},
		{
			name:      "Copy boolean value",
			source:    []byte(`{"key":true}`),
			target:    []byte(`{}`),
			path:      []string{"key"},
			expected:  []byte(`{"key":true}`),
			expectErr: false,
		},
		{
			name:      "Skip empty string",
			source:    []byte(`{"key":""}`),
			target:    []byte(`{}`),
			path:      []string{"key"},
			expected:  []byte(`{}`),
			expectErr: false,
		},
		{
			name:      "Skip empty array",
			source:    []byte(`{"key":[]}`),
			target:    []byte(`{}`),
			path:      []string{"key"},
			expected:  []byte(`{}`),
			expectErr: false,
		},
		{
			name:      "Skip empty object",
			source:    []byte(`{"key":{}}`),
			target:    []byte(`{}`),
			path:      []string{"key"},
			expected:  []byte(`{}`),
			expectErr: false,
		},
		{
			name:      "Copy non-empty array",
			source:    []byte(`{"key":[1,2,3]}`),
			target:    []byte(`{}`),
			path:      []string{"key"},
			expected:  []byte(`{"key":[1,2,3]}`),
			expectErr: false,
		},
		{
			name:      "Copy non-empty object",
			source:    []byte(`{"key":{"nested":"value"}}`),
			target:    []byte(`{}`),
			path:      []string{"key"},
			expected:  []byte(`{"key":{"nested":"value"}}`),
			expectErr: false,
		},
		{
			name:      "Invalid path",
			source:    []byte(`{"key":"value"}`),
			target:    []byte(`{}`),
			path:      []string{"invalid"},
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "Empty value in source but valid value in target",
			source:    []byte(`{"key":""}`),
			target:    []byte(`{"key":"existing"}`),
			path:      []string{"key"},
			expected:  []byte(`{"key":"existing"}`),
			expectErr: false,
		},
		{
			name:      "Empty value in source but valid object in target",
			source:    []byte(`{"key":{}}`),
			target:    []byte(`{"key":{"nested":{"nested_key":"nested_value"}}}`),
			path:      []string{"key"},
			expected:  []byte(`{"key":{"nested":{"nested_key":"nested_value"}}}`),
			expectErr: false,
		},
		{
			name:      "Invalid path with target non empty",
			source:    []byte(`{"key":"value"}`),
			target:    []byte(`{"key":"existing"}`),
			path:      []string{"invalid"},
			expected:  []byte(`{"key":"existing"}`),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CopyPath(tt.source, tt.target, tt.path...)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, string(tt.expected), string(result))
			}
		})
	}
}
