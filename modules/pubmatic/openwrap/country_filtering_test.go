package openwrap

import (
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestShouldApplyCountryFilter(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{
			name:     "EndpointAppLovinMax",
			endpoint: models.EndpointAppLovinMax,
			want:     true,
		},
		{
			name:     "EndpointV25",
			endpoint: models.EndpointV25,
			want:     false,
		},
		{
			name:     "EndpointJson",
			endpoint: models.EndpointJson,
			want:     false,
		},
		{
			name:     "EndpointHybrid",
			endpoint: models.EndpointHybrid,
			want:     false,
		},
		{
			name:     "Empty endpoint",
			endpoint: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldApplyCountryFilter(tt.endpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCountryFilterConfig(t *testing.T) {
	tests := []struct {
		name             string
		partnerConfigMap map[int]map[string]string
		wantMode         string
		wantCodes        string
	}{
		{
			name: "Config exists",
			partnerConfigMap: map[int]map[string]string{
				models.VersionLevelConfigID: {
					models.CountryFilterModeKey: "1",
					models.CountryCodesKey:      "[\"US\",\"UK\",\"IN\"]",
				},
			},
			wantMode:  "1",
			wantCodes: "[\"US\",\"UK\",\"IN\"]",
		},
		{
			name: "Empty config",
			partnerConfigMap: map[int]map[string]string{
				models.VersionLevelConfigID: {},
			},
			wantMode:  "",
			wantCodes: "",
		},
		{
			name:             "Nil config",
			partnerConfigMap: nil,
			wantMode:         "",
			wantCodes:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, codes := getCountryFilterConfig(tt.partnerConfigMap)
			assert.Equal(t, tt.wantMode, mode)
			assert.Equal(t, tt.wantCodes, codes)
		})
	}
}

func TestIsCountryAllowed(t *testing.T) {
	tests := []struct {
		name         string
		country      string
		mode         string
		countryCodes string
		want         bool
	}{
		{
			name:         "include_mode_country_in_list",
			country:      "US",
			mode:         "1",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         true,
		},
		{
			name:         "include_mode_country_not_in_list",
			country:      "FR",
			mode:         "1",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         false,
		},
		{
			name:         "exclude_mode_country_in_list",
			country:      "US",
			mode:         "0",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         false,
		},
		{
			name:         "exclude_mode_country_not_in_list",
			country:      "FR",
			mode:         "0",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         true,
		},
		{
			name:         "empty_mode",
			country:      "US",
			mode:         "",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         true,
		},
		{
			name:         "empty_country_codes",
			country:      "US",
			mode:         "1",
			countryCodes: "",
			want:         true,
		},
		{
			name:         "empty_country",
			country:      "",
			mode:         "1",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         true,
		},
		{
			name:         "invalid_mode",
			country:      "US",
			mode:         "invalid",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         false,
		},
		{
			name:         "case_insensitive_country_match",
			country:      "us",
			mode:         "1",
			countryCodes: "[\"US\",\"UK\",\"IN\"]",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCountryAllowed(tt.country, tt.mode, tt.countryCodes)
			assert.Equal(t, tt.want, got)
		})
	}
}
