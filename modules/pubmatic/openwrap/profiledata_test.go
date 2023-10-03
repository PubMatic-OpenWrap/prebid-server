package openwrap

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestOpenWrap_getProfileData(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCache := mock_cache.NewMockCache(ctrl)
	defer ctrl.Finish()

	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx       models.RequestCtx
		bidRequest openrtb2.BidRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    map[int]map[string]string
		wantErr bool
	}{
		{
			name: "get_profile_data_for_test_mode_platform_is_APP",
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						HBTimeout: 100,
					},
				},
				cache: mockCache,
			},
			args: args{
				rCtx: models.RequestCtx{
					DisplayID:     1,
					IsTestRequest: 2,
				},
				bidRequest: openrtb2.BidRequest{
					App: &openrtb2.App{},
				},
			},
			want: map[int]map[string]string{
				1: {
					models.PARTNER_ID:          models.PUBMATIC_PARTNER_ID_STRING,
					models.PREBID_PARTNER_NAME: string(openrtb_ext.BidderPubmatic),
					models.BidderCode:          string(openrtb_ext.BidderPubmatic),
					models.SERVER_SIDE_FLAG:    models.PUBMATIC_SS_FLAG,
					models.KEY_GEN_PATTERN:     models.ADUNIT_SIZE_KGP,
					models.TIMEOUT:             "100",
				},
				-1: {
					models.PLATFORM_KEY:     models.PLATFORM_APP,
					models.DisplayVersionID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "get_profile_data_for_test_mode_platform_is_other_than_APP",
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						HBTimeout: 100,
					},
				},
				cache: mockCache,
			},
			args: args{
				rCtx: models.RequestCtx{
					DisplayID:     1,
					IsTestRequest: 2,
				},
				bidRequest: openrtb2.BidRequest{
					App: nil,
				},
			},
			want: map[int]map[string]string{
				1: {
					models.PARTNER_ID:          models.PUBMATIC_PARTNER_ID_STRING,
					models.PREBID_PARTNER_NAME: string(openrtb_ext.BidderPubmatic),
					models.BidderCode:          string(openrtb_ext.BidderPubmatic),
					models.SERVER_SIDE_FLAG:    models.PUBMATIC_SS_FLAG,
					models.KEY_GEN_PATTERN:     models.ADUNIT_SIZE_KGP,
					models.TIMEOUT:             "100",
				},
				-1: {
					models.PLATFORM_KEY:     "",
					models.DisplayVersionID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "get_profile_data_for_other_than_test_mode",
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						HBTimeout: 100,
					},
				},
				cache: mockCache,
			},
			args: args{
				rCtx: models.RequestCtx{
					IsTestRequest: 1,
					PubID:         5890,
					ProfileID:     123,
					DisplayID:     1,
					Endpoint:      models.PLATFORM_APP,
				},
				bidRequest: openrtb2.BidRequest{
					App: nil,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(5890, 123, 1, models.PLATFORM_APP).Return(
					map[int]map[string]string{
						1: {
							models.PARTNER_ID:          "2",
							models.PREBID_PARTNER_NAME: "appnexus",
							models.BidderCode:          "appnexus",
							models.SERVER_SIDE_FLAG:    "1",
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
							models.TIMEOUT:             "200",
						},
						-1: {
							models.PLATFORM_KEY:     models.PLATFORM_APP,
							models.DisplayVersionID: "1",
						},
					}, nil)
			},
			want: map[int]map[string]string{
				1: {
					models.PARTNER_ID:          "2",
					models.PREBID_PARTNER_NAME: "appnexus",
					models.BidderCode:          "appnexus",
					models.SERVER_SIDE_FLAG:    "1",
					models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
					models.TIMEOUT:             "200",
				},
				-1: {
					models.PLATFORM_KEY:     models.PLATFORM_APP,
					models.DisplayVersionID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "get_profile_data_for_other_than_test_mode_and_cache_returned_error",
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						HBTimeout: 100,
					},
				},
				cache: mockCache,
			},
			args: args{
				rCtx: models.RequestCtx{
					IsTestRequest: 1,
					PubID:         5890,
					ProfileID:     123,
					DisplayID:     1,
					Endpoint:      models.PLATFORM_APP,
				},
				bidRequest: openrtb2.BidRequest{
					App: nil,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(5890, 123, 1, models.PLATFORM_APP).Return(
					nil, fmt.Errorf("error GetPartnerConfigMap"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			m := OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			got, err := m.getProfileData(tt.args.rCtx, tt.args.bidRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenWrap.getProfileData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getTestModePartnerConfigMap(t *testing.T) {
	type args struct {
		platform       string
		timeout        int64
		displayVersion int
	}
	tests := []struct {
		name string
		args args
		want map[int]map[string]string
	}{
		{
			name: "get_test_mode_partnerConfigMap",
			args: args{
				platform:       "in-app",
				timeout:        200,
				displayVersion: 2,
			},
			want: map[int]map[string]string{
				1: {
					models.PARTNER_ID:          "1",
					models.PREBID_PARTNER_NAME: "pubmatic",
					models.BidderCode:          "pubmatic",
					models.SERVER_SIDE_FLAG:    "1",
					models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
					models.TIMEOUT:             "200",
				},
				-1: {
					models.PLATFORM_KEY:     "in-app",
					models.DisplayVersionID: "2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTestModePartnerConfigMap(tt.args.platform, tt.args.timeout, tt.args.displayVersion)
			assert.Equal(t, tt.want, got)
		})
	}
}
