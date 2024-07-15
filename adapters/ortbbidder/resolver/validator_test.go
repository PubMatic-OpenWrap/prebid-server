package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInt(t *testing.T) {
	tests := []struct {
		input    any
		expected int
		ok       bool
	}{
		{input: 42.0, expected: 42, ok: true},
		{input: 42.9, expected: 42, ok: true},
		{input: "42", expected: 0, ok: false},
		{input: nil, expected: 0, ok: false},
	}
	for _, test := range tests {
		result, ok := validateNumber[int](test.input)
		if result != test.expected || ok != test.ok {
			t.Errorf("validateInt(%v) = (%d, %v), want (%d, %v)", test.input, result, ok, test.expected, test.ok)
		}
	}
}
func TestValidateInt64(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
		ok       bool
	}{
		{input: 42.0, expected: 42, ok: true},
		{input: 42.9, expected: 42, ok: true},
		{input: "42", expected: 0, ok: false},
		{input: nil, expected: 0, ok: false},
		{input: 42, expected: 0, ok: false},
	}
	for _, test := range tests {
		result, ok := validateNumber[int64](test.input)
		if result != test.expected || ok != test.ok {
			t.Errorf("validateInt64(%v) = (%d, %v), want (%d, %v)", test.input, result, ok, test.expected, test.ok)
		}
	}
}

func TestValidateString(t *testing.T) {
	tests := []struct {
		input    any
		expected string
		ok       bool
	}{
		{input: "hello", expected: "hello", ok: true},
		{input: "", expected: "", ok: false},
		{input: 42, expected: "", ok: false},
		{input: nil, expected: "", ok: false},
	}
	for _, test := range tests {
		result, ok := validateString(test.input)
		if result != test.expected || ok != test.ok {
			t.Errorf("validateString(%v) = (%q, %v), want (%q, %v)", test.input, result, ok, test.expected, test.ok)
		}
	}
}

func TestValidateMap(t *testing.T) {
	tests := []struct {
		input    any
		expected map[string]any
		ok       bool
	}{
		{input: map[string]any{"key": "value"}, expected: map[string]any{"key": "value"}, ok: true},
		{input: `{"key": "value"}`, expected: nil, ok: false},
		{input: nil, expected: nil, ok: false},
	}
	for _, test := range tests {
		result, ok := validateMap(test.input)
		assert.Equal(t, test.expected, result, "mismatched result")
		assert.Equal(t, test.ok, ok, "mismatched status")
	}
}

func TestValidateDataTypeSlice(t *testing.T) {
	stringTests := []struct {
		name     string
		input    any
		expected []string
		ok       bool
	}{
		{
			name:     "valid string slice",
			input:    []any{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
			ok:       true,
		},
		{
			name:     "int value in string slice",
			input:    []any{"a", 2, "c"},
			expected: []string{"a", "c"},
			ok:       true,
		},
		{
			name:     "int slice with string dataType",
			input:    []any{1, 2, 3},
			expected: []string{},
			ok:       false,
		},
		{
			name:     "invalid slice",
			input:    "not a slice",
			expected: nil,
			ok:       false,
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
			ok:       false,
		},
	}
	for _, test := range stringTests {
		result, ok := validateDataTypeSlice[string](test.input)
		assert.Equalf(t, test.expected, result, "mismatch result: %s", test.name)
		assert.Equalf(t, test.ok, ok, "mismatched status flag: %s", test.name)
	}
}
