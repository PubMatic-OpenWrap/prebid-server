package openwrap

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
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

func TestAapplyPartnerThrottling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCache(ctrl)
	mockMetric := mock_metrics.NewMockMetricsEngine(ctrl)

	tests := []struct {
		name             string
		cacheSetup       func()
		rCtx             models.RequestCtx
		partnerConfigMap map[int]map[string]string
		expectedMap      map[string]struct{}
		expectedAllFlag  bool
	}{
		{
			name: "cache_returns_error",
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria("IN", models.PartnerLevelThrottlingCriteria, models.PartnerLevelThrottlingCriteriaValue).Return(nil, errors.New("cache error"))
			},
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "IN"},
				PubIDStr:  "123",
			},
			partnerConfigMap: map[int]map[string]string{
				1: {models.BidderCode: "bidderA"},
			},
			expectedMap:     map[string]struct{}{},
			expectedAllFlag: false,
		},
		{
			name: "no_throttled_partners",
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria("IN", models.PartnerLevelThrottlingCriteria, models.PartnerLevelThrottlingCriteriaValue).Return([]string{}, nil)
			},
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "IN"},
				PubIDStr:  "123",
			},
			partnerConfigMap: map[int]map[string]string{
				1: {models.BidderCode: "bidderA"},
			},
			expectedMap:     map[string]struct{}{},
			expectedAllFlag: false,
		},
		{
			name: "partner_throttled_and_not_in_fallback_simulate_fallback_fail",
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria("US", models.PartnerLevelThrottlingCriteria, models.PartnerLevelThrottlingCriteriaValue).Return([]string{"bidderA"}, nil)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("456", "bidderA")
			},
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "US"},
				PubIDStr:  "456",
			},
			partnerConfigMap: map[int]map[string]string{
				1: {models.BidderCode: "bidderA"},
				2: {models.BidderCode: "bidderB"},
			},
			expectedMap:     map[string]struct{}{"bidderA": {}},
			expectedAllFlag: false,
		},
		{
			name: "all_partners_throttled",
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria("US", models.PartnerLevelThrottlingCriteria, models.PartnerLevelThrottlingCriteriaValue).Return([]string{"bidderA", "bidderB"}, nil)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("789", "bidderA")
				mockMetric.EXPECT().RecordPartnerThrottledRequests("789", "bidderB")
			},
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "US"},
				PubIDStr:  "789",
			},
			partnerConfigMap: map[int]map[string]string{
				1: {models.BidderCode: "bidderA"},
				2: {models.BidderCode: "bidderB"},
			},
			expectedMap: map[string]struct{}{
				"bidderA": {},
				"bidderB": {},
			},
			expectedAllFlag: true,
		},
		{
			name: "skip_config_with_missing_bidderCode",
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria("UK", models.PartnerLevelThrottlingCriteria, models.PartnerLevelThrottlingCriteriaValue).Return([]string{"bidderC"}, nil)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("101", "bidderC")
			},
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "UK"},
				PubIDStr:  "101",
			},
			partnerConfigMap: map[int]map[string]string{
				1: {models.BidderCode: ""},
				2: {models.BidderCode: "bidderC"},
			},
			expectedMap:     map[string]struct{}{"bidderC": {}},
			expectedAllFlag: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cacheSetup()

			m := &OpenWrap{
				cache:        mockCache,
				metricEngine: mockMetric,
			}

			got, gotAll := m.applyPartnerThrottling(tt.rCtx, tt.partnerConfigMap)
			assert.Equal(t, tt.expectedMap, got)
			assert.Equal(t, tt.expectedAllFlag, gotAll)
		})
	}
}
