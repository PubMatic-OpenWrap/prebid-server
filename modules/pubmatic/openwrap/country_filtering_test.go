package openwrap

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
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
			name:     "EndpointGoogleSDK",
			endpoint: models.EndpointGoogleSDK,
			want:     true,
		},
		{
			name:     "EndpointUnityLevelPlay",
			endpoint: models.EndpointUnityLevelPlay,
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
		randomNumber     int
	}{
		{
			name: "cache_returns_error",
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "IN"},
				PubIDStr:  "123",
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.BidderCode:       "bidderA",
						models.SERVER_SIDE_FLAG: "1",
					},
				},
			},
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria(gomock.Any()).Return(nil, errors.New("cache error"))
			},
			expectedMap:     map[string]struct{}(nil),
			expectedAllFlag: false,
			randomNumber:    100,
		},
		{
			name: "no_throttled_partners",
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "IN"},
				PubIDStr:  "123",
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.BidderCode:       "bidderA",
						models.SERVER_SIDE_FLAG: "1",
					},
				},
			},
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria(gomock.Any()).Return(map[string]struct{}{}, nil)
			},
			expectedMap:     map[string]struct{}(nil),
			expectedAllFlag: false,
			randomNumber:    100,
		},
		{
			name: "partner_throttled_and_not_in_fallback_simulate_fallback_fail",
			rCtx: models.RequestCtx{
				Endpoint:  models.EndpointV25,
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "US"},
				PubIDStr:  "456",
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.BidderCode:       "bidderA",
						models.SERVER_SIDE_FLAG: "1",
					},
					2: {
						models.BidderCode:       "bidderB",
						models.SERVER_SIDE_FLAG: "1",
					},
				},
			},
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria(gomock.Any()).Return(map[string]struct{}{"bidderA": {}}, nil)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("456", "bidderA", models.PartnerLevelThrottlingFeatureID)
				mockMetric.EXPECT().RecordCountryLevelPartnerThrottledRequests(rctx.Endpoint, "bidderA", "US")
			},
			expectedMap:     map[string]struct{}{"bidderA": {}},
			expectedAllFlag: false,
			randomNumber:    100,
		},
		{
			name: "all_partners_throttled",
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "US"},
				Endpoint:  models.EndpointV25,
				PubIDStr:  "789",
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.BidderCode:       "bidderA",
						models.SERVER_SIDE_FLAG: "1",
					},
					2: {
						models.BidderCode:       "bidderB",
						models.SERVER_SIDE_FLAG: "1",
					},
				},
			},
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria(gomock.Any()).Return(map[string]struct{}{"bidderA": {}, "bidderB": {}}, nil)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("789", "bidderA", models.PartnerLevelThrottlingFeatureID)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("789", "bidderB", models.PartnerLevelThrottlingFeatureID)
				mockMetric.EXPECT().RecordCountryLevelPartnerThrottledRequests(rctx.Endpoint, "bidderA", "US")
				mockMetric.EXPECT().RecordCountryLevelPartnerThrottledRequests(rctx.Endpoint, "bidderB", "US")
			},
			expectedMap: map[string]struct{}{
				"bidderA": {},
				"bidderB": {},
			},
			expectedAllFlag: true,
			randomNumber:    100,
		},
		{
			name: "skip_config_with_missing_bidderCode",
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "UK"},
				Endpoint:  models.EndpointV25,
				PubIDStr:  "101",
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.BidderCode:       "",
						models.SERVER_SIDE_FLAG: "1",
					},
					2: {
						models.BidderCode:       "bidderC",
						models.SERVER_SIDE_FLAG: "1",
					},
				},
			},
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria(gomock.Any()).Return(map[string]struct{}{"bidderC": {}}, nil)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("101", "bidderC", models.PartnerLevelThrottlingFeatureID)
				mockMetric.EXPECT().RecordCountryLevelPartnerThrottledRequests(rctx.Endpoint, "bidderC", "UK")
			},
			expectedMap:     map[string]struct{}{"bidderC": {}},
			expectedAllFlag: true,
			randomNumber:    100,
		},
		{
			name: "partner_throttled_and_allowed_in_fallback",
			rCtx: models.RequestCtx{
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "US"},
				Endpoint:  models.EndpointV25,
				PubIDStr:  "111",
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.BidderCode:       "bidderA",
						models.SERVER_SIDE_FLAG: "1",
					},
				},
			},
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria(gomock.Any()).Return(map[string]struct{}{"bidderA": {}}, nil)
			},
			expectedMap:     map[string]struct{}{},
			expectedAllFlag: false,
			randomNumber:    3,
		},
		{
			name: "mix_of_client_and_server_side_partners",
			rCtx: models.RequestCtx{
				Endpoint:  models.EndpointV25,
				PubIDStr:  "789",
				DeviceCtx: models.DeviceCtx{DerivedCountryCode: "US"},
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						models.PREBID_PARTNER_NAME: "ALL",
						models.BidderCode:          "ALL",
					},
					0: {
						models.PREBID_PARTNER_NAME: "pubmatic",
						models.BidderCode:          "pubmatic",
						models.SERVER_SIDE_FLAG:    "1",
					},
					1: {
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
					},
				},
			},
			cacheSetup: func() {
				mockCache.EXPECT().GetThrottlePartnersWithCriteria("US").Return(map[string]struct{}{"pubmatic": {}, "appnexus": {}}, nil)
				mockMetric.EXPECT().RecordPartnerThrottledRequests("789", "pubmatic", models.PartnerLevelThrottlingFeatureID)
				mockMetric.EXPECT().RecordCountryLevelPartnerThrottledRequests("v25", "pubmatic", "US")
				mockMetric.EXPECT().RecordPartnerThrottledRequests("789", "appnexus", models.PartnerLevelThrottlingFeatureID)
				mockMetric.EXPECT().RecordCountryLevelPartnerThrottledRequests("v25", "appnexus", "US")
			},
			expectedMap:     map[string]struct{}{"pubmatic": {}, "appnexus": {}},
			expectedAllFlag: true,
			randomNumber:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cacheSetup()

			m := &OpenWrap{
				cache:        mockCache,
				metricEngine: mockMetric,
				cfg: config.Config{
					Features: config.FeatureToggle{
						AllowPartnerLevelThrottlingPercentage: 5,
					},
				},
			}
			GetRandomNumberIn1To100 = func() int {
				return tt.randomNumber
			}
			got, gotAll := m.applyPartnerThrottling(tt.rCtx)
			assert.Equal(t, tt.expectedMap, got)
			assert.Equal(t, tt.expectedAllFlag, gotAll)
		})
	}
}
