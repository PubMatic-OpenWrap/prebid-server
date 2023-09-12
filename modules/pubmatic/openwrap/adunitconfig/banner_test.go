package adunitconfig

import (
	"reflect"
	"testing"

	mock_metrics "github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func TestUpdateBannerObjectWithAdunitConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)

	type args struct {
		rCtx models.RequestCtx
		imp  openrtb2.Imp
		div  string
	}
	tests := []struct {
		name          string
		args          args
		setup         func()
		wantAdUnitCtx models.AdUnitCtx
	}{
		{
			name: "AdunitConfig_is_nil",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig:  nil,
					PubIDStr:      "5890",
					ProfileIDStr:  "123",
				},
				imp: openrtb2.Imp{
					Banner: &openrtb2.Banner{
						W:  ptrutil.ToPtr[int64](100),
						H:  ptrutil.ToPtr[int64](200),
						ID: "123",
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{},
		},
		{
			name: "AdunitConfig_is_empty",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig:  &adunitconfig.AdUnitConfig{},
					PubIDStr:      "5890",
					ProfileIDStr:  "123",
				},
				imp: openrtb2.Imp{
					Banner: &openrtb2.Banner{
						W:  ptrutil.ToPtr[int64](100),
						H:  ptrutil.ToPtr[int64](200),
						ID: "123",
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{},
		},
		{
			name: "request_imp_has_Banner_but_disabled_through_config_default",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Banner: &adunitconfig.Banner{
									Enabled: ptrutil.ToPtr[bool](false),
								},
							},
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					Banner: &openrtb2.Banner{
						W:  ptrutil.ToPtr[int64](100),
						H:  ptrutil.ToPtr[int64](200),
						ID: "123",
					},
				},
			},
			setup: func() {
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeBanner, "5890", "123").Times(1)
			},
			wantAdUnitCtx: models.AdUnitCtx{
				UsingDefaultConfig: true,
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Banner: &adunitconfig.Banner{
						Enabled: ptrutil.ToPtr[bool](false),
					},
				},
			},
		},
		{
			name: "request_imp_has_Banner_but_disabled_through_config_for_particular_slot",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Banner: &adunitconfig.Banner{
									Enabled: ptrutil.ToPtr[bool](true),
								},
							},
							"/12344/test_adunit": {
								Banner: &adunitconfig.Banner{
									Enabled: ptrutil.ToPtr[bool](false),
								},
							},
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					TagID: "/12344/Test_AdUnit",
					Banner: &openrtb2.Banner{
						W:  ptrutil.ToPtr[int64](100),
						H:  ptrutil.ToPtr[int64](200),
						ID: "123",
					},
				},
			},
			setup: func() {
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeBanner, "5890", "123").Times(1)
			},
			wantAdUnitCtx: models.AdUnitCtx{
				UsingDefaultConfig: false,
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Banner: &adunitconfig.Banner{
						Enabled: ptrutil.ToPtr[bool](false),
					},
				},
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Banner: &adunitconfig.Banner{
						Enabled: ptrutil.ToPtr[bool](false),
					},
				},
			},
		},
		{
			name: "final_adunit_config_formed_using_both_default_and_slot._banner_selected_from_slot",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Banner: &adunitconfig.Banner{
									Enabled: ptrutil.ToPtr[bool](true),
									Config: &adunitconfig.BannerConfig{
										Banner: openrtb2.Banner{
											W:  ptrutil.ToPtr[int64](100),
											H:  ptrutil.ToPtr[int64](200),
											ID: "123",
										},
									},
								},
							},
							"/12344/test_adunit": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr[bool](true),
									Config: &adunitconfig.VideoConfig{
										Video: openrtb2.Video{
											Plcmt:       2,
											MinDuration: 2,
											MaxDuration: 10,
										},
									},
								},
							},
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					TagID: "/12344/Test_AdUnit",
					Banner: &openrtb2.Banner{
						W:  ptrutil.ToPtr[int64](100),
						H:  ptrutil.ToPtr[int64](200),
						ID: "123",
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Banner: &adunitconfig.Banner{
						Enabled: ptrutil.ToPtr[bool](true),
						Config: &adunitconfig.BannerConfig{
							Banner: openrtb2.Banner{
								ID: "123",
								W:  ptrutil.ToPtr[int64](100),
								H:  ptrutil.ToPtr[int64](200),
							},
						},
					},
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr[bool](true),
						Config: &adunitconfig.VideoConfig{
							Video: openrtb2.Video{
								Plcmt:       2,
								MinDuration: 2,
								MaxDuration: 10,
							},
						},
					},
				},
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Banner: &adunitconfig.Banner{
						Enabled: ptrutil.ToPtr[bool](true),
						Config: &adunitconfig.BannerConfig{
							Banner: openrtb2.Banner{
								ID: "123",
								W:  ptrutil.ToPtr[int64](100),
								H:  ptrutil.ToPtr[int64](200),
							},
						},
					},
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr[bool](true),
						Config: &adunitconfig.VideoConfig{
							Video: openrtb2.Video{
								Plcmt:       2,
								MinDuration: 2,
								MaxDuration: 10,
							},
						},
					},
				},
				UsingDefaultConfig:     true,
				AllowedConnectionTypes: nil,
			},
		},
		{
			name: "both_slot_and_default_config_are_nil",
			args: args{
				rCtx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default":            nil,
							"/12344/test_adunit": nil,
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					TagID: "/12344/Test_AdUnit",
					Banner: &openrtb2.Banner{
						ID: "123",
						W:  ptrutil.ToPtr[int64](100),
						H:  ptrutil.ToPtr[int64](200),
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{
				MatchedSlot:              "/12344/Test_AdUnit",
				IsRegex:                  false,
				MatchedRegex:             "",
				SelectedSlotAdUnitConfig: nil,
				AppliedSlotAdUnitConfig:  nil,
				UsingDefaultConfig:       false,
				AllowedConnectionTypes:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if gotAdUnitCtx := UpdateBannerObjectWithAdunitConfig(tt.args.rCtx, tt.args.imp, tt.args.div); !reflect.DeepEqual(gotAdUnitCtx, tt.wantAdUnitCtx) {
				t.Errorf("UpdateBannerObjectWithAdunitConfig() = %v, want %v", gotAdUnitCtx, tt.wantAdUnitCtx)
			}
		})
	}
}
