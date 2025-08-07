package sdkutils

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
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
			expected:  []byte(`{}`),
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
			}
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestAddSize300x600ForInterstitialBanner(t *testing.T) {
	tests := []struct {
		name     string
		imp      openrtb2.Imp
		expected openrtb2.Imp
	}{
		{
			name: "Nil Banner",
			imp: openrtb2.Imp{
				ID: "test_imp",
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
			},
		},
		{
			name: "Banner with W/H set to 320x480, no 300x600",
			imp: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](320),
					H: ptrutil.ToPtr[int64](480),
				},
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](320),
					H: ptrutil.ToPtr[int64](480),
					Format: []openrtb2.Format{
						{W: 300, H: 600},
					},
				},
			},
		},
		{
			name: "Banner with W/H set to 300x600 already",
			imp: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](300),
					H: ptrutil.ToPtr[int64](600),
				},
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](300),
					H: ptrutil.ToPtr[int64](600),
				},
			},
		},
		{
			name: "Banner with Format containing 320x480, no 300x600",
			imp: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 480},
						{W: 320, H: 50},
					},
				},
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 480},
						{W: 320, H: 50},
						{W: 300, H: 600},
					},
				},
			},
		},
		{
			name: "Banner with Format containing both 320x480 and 300x600",
			imp: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 480},
						{W: 300, H: 600},
						{W: 320, H: 50},
					},
				},
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 480},
						{W: 300, H: 600},
						{W: 320, H: 50},
					},
				},
			},
		},
		{
			name: "Banner with neither 320x480 nor 300x600 sizes",
			imp: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 50},
						{W: 728, H: 90},
					},
				},
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 50},
						{W: 728, H: 90},
					},
				},
			},
		},
		{
			name: "Banner with W/H set to different size and Format containing 320x480",
			imp: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](728),
					H: ptrutil.ToPtr[int64](90),
					Format: []openrtb2.Format{
						{W: 320, H: 480},
					},
				},
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](728),
					H: ptrutil.ToPtr[int64](90),
					Format: []openrtb2.Format{
						{W: 320, H: 480},
						{W: 300, H: 600},
					},
				},
			},
		},
		{
			name: "Banner with nil W/H but Format containing 320x480",
			imp: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 480},
					},
				},
			},
			expected: openrtb2.Imp{
				ID: "test_imp",
				Banner: &openrtb2.Banner{
					Format: []openrtb2.Format{
						{W: 320, H: 480},
						{W: 300, H: 600},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddSize300x600ForInterstitialBanner(&tt.imp)
			assert.Equal(t, tt.expected, tt.imp, "Banner formats should match expected")
		})
	}
}
