package xmlparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseKey(t *testing.T) {
	type want struct {
		value string
		found bool
	}
	tests := []struct {
		name string
		in   string
		want want
	}{
		{
			name: "empty",
			in:   ``,
			want: want{value: "", found: false},
		},
		{
			name: "whitespaces",
			in:   `     `,
			want: want{value: ``, found: false},
		},
		{
			name: "valid",
			in:   ` key="value"  `,
			want: want{value: `key`, found: true},
		},
		{
			name: "valid_with_whitespaces",
			in:   ` key = "value"  `,
			want: want{value: `key`, found: true},
		},
		{
			name: "valid_key_start_with_underscore",
			in:   ` _key="value"  `,
			want: want{value: `_key`, found: true},
		},
		{
			name: "valid_key_contains_dash",
			in:   ` _key-1="value"  `,
			want: want{value: `_key-1`, found: true},
		},
		{
			name: "extract_key_part",
			in:   ` key"value"  `,
			want: want{value: `key`, found: true},
		},
		{
			name: "invalid_key_start_with_numeric",
			in:   ` 1key="value"  `,
			want: want{value: ``, found: false},
		},
		{
			name: "invalid_key_start_with_dash",
			in:   ` -key="value"  `,
			want: want{value: ``, found: false},
		},
		{
			name: "invalid_key_start_with_special_char",
			in:   ` #key="value"  `,
			want: want{value: ``, found: false},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si, ei, found := _parseKey([]byte(tt.in), 0, len(tt.in))
			assert.Equal(t, tt.want.found, found)
			assert.Equal(t, tt.want.value, tt.in[si:ei])
		})
	}
}

func Test_parseValue(t *testing.T) {
	type want struct {
		value string
		found bool
	}
	tests := []struct {
		name string
		in   string
		want want
	}{
		{
			name: `empty_string`,
			in:   ``,
			want: want{value: ``, found: false},
		},
		{
			name: `white_spaces`,
			in:   `    `,
			want: want{value: ``, found: false},
		},
		{
			name: `invalid_no_quotes`,
			in:   `testvalue`,
			want: want{value: ``, found: false},
		},
		{
			name: `invalid_missing_end_quote_1`,
			in:   `'testvalue`,
			want: want{value: ``, found: false},
		},
		{
			name: `invalid_missing_end_quote_2`,
			in:   `"testvalue`,
			want: want{value: ``, found: false},
		},
		{
			name: `invalid_mismatch_quote_1`,
			in:   `'testvalue"`,
			want: want{value: ``, found: false},
		},
		{
			name: `invalid_mismatch_quote_2`,
			in:   `"testvalue'`,
			want: want{value: ``, found: false},
		},
		{
			name: `valid_single_quote`,
			in:   `'testvalue'extra`,
			want: want{value: `testvalue`, found: true},
		},
		{
			name: `valid_double_quote`,
			in:   `"testvalue"extra`,
			want: want{value: `testvalue`, found: true},
		},
		{
			name: `valid_whitespace_suffix`,
			in:   `"testvalue"   `,
			want: want{value: `testvalue`, found: true},
		},
		{
			name: `valid_nested_quote_1`,
			in:   `"test'value"   `,
			want: want{value: `test'value`, found: true},
		},
		{
			name: `valid_nested_quote_2`,
			in:   `'test"value'   `,
			want: want{value: `test"value`, found: true},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si, ei, found := _parseValue([]byte(tt.in), 0, len(tt.in))
			assert.Equal(t, tt.want.found, found)
			assert.Equal(t, tt.want.value, string(tt.in[si:ei]))
		})
	}
}

func Test_parseAttributes(t *testing.T) {
	type want struct {
		key, value string
	}
	tests := []struct {
		name string
		in   string
		want []want
	}{
		{
			name: `empty`,
			in:   ``,
			want: nil,
		},
		{
			name: `whitespaces`,
			in:   `    `,
			want: nil,
		},
		{
			name: `single_attribute_valid_with_whitespaces`,
			in:   ` key = "value" `,
			want: []want{
				{key: `key`, value: `value`},
			},
		},
		{
			name: `multi_attribute_valid_with_whitespaces`,
			in:   ` key1 = "value1"  key2 = "value2" `,
			want: []want{
				{key: `key1`, value: `value1`},
				{key: `key2`, value: `value2`},
			},
		},
		{
			name: `invalid_missing_key`,
			in:   `123key="value"`,
			want: []want{},
		},
		{
			name: `invalid_missing_equals`,
			in:   `key"value"`,
			want: []want{},
		},
		{
			name: `invalid_missing_quotes`,
			in:   `key=value`,
			want: []want{},
		},
		{
			name: `invalid_key`,
			in:   ` key key = "value"`,
			want: []want{},
		},
		{
			name: `invalid_value`,
			in:   ` key = value "value"`,
			want: []want{},
		},
		{
			name: `one_valid_one_invalid`,
			in:   ` key1='value1' 2key='value2'`,
			want: []want{
				{key: `key1`, value: `value1`},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAttributes([]byte(tt.in), 0, len(tt.in))
			assert.Equal(t, len(tt.want), len(got))
			for i, attr := range got {
				assert.Equal(t, tt.want[i].key, string(tt.in[attr.key.si:attr.key.ei]))
				assert.Equal(t, tt.want[i].value, string(tt.in[attr.value.si:attr.value.ei]))
			}
		})
	}
}
