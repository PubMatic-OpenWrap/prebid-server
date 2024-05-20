package adunitconfig

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestUpdateVideoObjectWithAdunitConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)

	type args struct {
		rCtx           models.RequestCtx
		imp            openrtb2.Imp
		div            string
		connectionType *adcom1.ConnectionType
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
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
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
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{},
		},
		{
			name: "request_imp_has_Video_but_disabled_through_config_default",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr(false),
								},
							},
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
					},
				},
			},
			setup: func() {
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeVideo, "5890", "123").Times(1)
			},
			wantAdUnitCtx: models.AdUnitCtx{
				UsingDefaultConfig: true,
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(false),
					},
				},
			},
		},
		{
			name: "request_imp_has_Video_but_disabled_through_config_for_particular_slot",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr(true),
								},
							},
							"/12344/test_adunit": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr(false),
								},
							},
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					TagID: "/12344/Test_AdUnit",
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
					},
				},
			},
			setup: func() {
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeVideo, "5890", "123").Times(1)
			},
			wantAdUnitCtx: models.AdUnitCtx{
				UsingDefaultConfig: false,
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(false),
					},
				},
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(false),
					},
				},
			},
		},
		{
			name: "final_adunit_config_formed_using_both_default_and_slot._video_selected_from_slot",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.VideoConfig{
										Video: openrtb2.Video{
											Plcmt:       2,
											MinDuration: 2,
											MaxDuration: 10,
										},
									},
								},
							},
							"/12344/test_adunit": {
								Banner: &adunitconfig.Banner{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.BannerConfig{
										Banner: openrtb2.Banner{
											ID: "123",
											W:  ptrutil.ToPtr[int64](100),
											H:  ptrutil.ToPtr[int64](100),
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
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Banner: &adunitconfig.Banner{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.BannerConfig{
							Banner: openrtb2.Banner{
								ID: "123",
								W:  ptrutil.ToPtr[int64](100),
								H:  ptrutil.ToPtr[int64](100),
							},
						},
					},
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(true),
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
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.BannerConfig{
							Banner: openrtb2.Banner{
								ID: "123",
								W:  ptrutil.ToPtr[int64](100),
								H:  ptrutil.ToPtr[int64](100),
							},
						},
					},
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(true),
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
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
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
		{
			name: "AllowedConnectionTypes_updated_from_default",
			args: args{
				rCtx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.VideoConfig{
										Video: openrtb2.Video{
											CompanionType: []adcom1.CompanionType{1, 2, 3},
										},
										ConnectionType: []int{10, 20, 30},
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
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{
				MatchedSlot:              "",
				IsRegex:                  false,
				MatchedRegex:             "",
				SelectedSlotAdUnitConfig: nil,
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.VideoConfig{
							Video: openrtb2.Video{
								CompanionType: []adcom1.CompanionType{1, 2, 3},
							},
							ConnectionType: []int{10, 20, 30},
						},
					},
				},
				UsingDefaultConfig:     true,
				AllowedConnectionTypes: []int{10, 20, 30},
			},
		},
		{
			name: "Video_config_is_prsent_in_both_default_and_slot_preferance_is_given_to_slot_level",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.VideoConfig{
										Video: openrtb2.Video{
											Plcmt:       2,
											MinDuration: 2,
											MaxDuration: 10,
										},
									},
								},
							},
							"/12344/test_adunit": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.VideoConfig{
										Video: openrtb2.Video{
											MIMEs:       []string{"test"},
											Plcmt:       4,
											MinDuration: 4,
											MaxDuration: 20,
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
					Video: &openrtb2.Video{
						Plcmt:       2,
						MinDuration: 2,
						MaxDuration: 10,
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.VideoConfig{
							Video: openrtb2.Video{
								MIMEs:       []string{"test"},
								Plcmt:       4,
								MinDuration: 4,
								MaxDuration: 20,
							},
						},
					},
				},
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.VideoConfig{
							Video: openrtb2.Video{
								MIMEs:       []string{"test"},
								Plcmt:       4,
								MinDuration: 4,
								MaxDuration: 20,
							},
						},
					},
				},
				UsingDefaultConfig:     false,
				AllowedConnectionTypes: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			gotAdUnitCtx := UpdateVideoObjectWithAdunitConfig(tt.args.rCtx, tt.args.imp, tt.args.div, tt.args.connectionType)
			assert.Equal(t, tt.wantAdUnitCtx, gotAdUnitCtx)
		})
	}
}
