package adunitconfig

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestUpdateNativeObjectWithAdunitConfig(t *testing.T) {
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
			name: "adunitConfig_is_nil",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig:  nil,
					PubIDStr:      "5890",
					ProfileIDStr:  "123",
				},
				imp: openrtb2.Imp{
					Native: &openrtb2.Native{
						Request: "Native_Reuqest",
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{},
		},
		{
			name: "adunitConfig_is_empty",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig:  &adunitconfig.AdUnitConfig{},
					PubIDStr:      "5890",
					ProfileIDStr:  "123",
				},
				imp: openrtb2.Imp{
					Native: &openrtb2.Native{
						Request: "Native_Reuqest",
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{},
		},
		{
			name: "request_imp_has_Native_but_disabled_through_config_default",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Native: &adunitconfig.Native{
									Enabled: ptrutil.ToPtr(false),
								},
							},
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					Native: &openrtb2.Native{
						Request: "Native_Reuqest",
					},
				},
			},
			setup: func() {
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeNative, "5890", "123").Times(1)
			},
			wantAdUnitCtx: models.AdUnitCtx{
				UsingDefaultConfig: true,
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Native: &adunitconfig.Native{
						Enabled: ptrutil.ToPtr(false),
					},
				},
			},
		},
		{
			name: "request_imp_has_Native_but_disabled_through_config_for_particular_slot",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Native: &adunitconfig.Native{
									Enabled: ptrutil.ToPtr(true),
								},
							},
							"/12344/test_adunit": {
								Native: &adunitconfig.Native{
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
					Native: &openrtb2.Native{
						Request: "Native_Reuqest",
					},
				},
			},
			setup: func() {
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeNative, "5890", "123").Times(1)
			},
			wantAdUnitCtx: models.AdUnitCtx{
				UsingDefaultConfig: false,
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Native: &adunitconfig.Native{
						Enabled: ptrutil.ToPtr(false),
					},
				},
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Native: &adunitconfig.Native{
						Enabled: ptrutil.ToPtr(false),
					},
				},
			},
		},
		{
			name: "final_adunit_config_formed_using_both_default_and_slot._native_selected_from_slot",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"/12344/test_adunit": {
								Native: &adunitconfig.Native{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.NativeConfig{
										Native: openrtb2.Native{
											Request: "Native_Reuqest",
										},
									},
								},
							},
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
						},
					},
					PubIDStr:     "5890",
					ProfileIDStr: "123",
				},
				imp: openrtb2.Imp{
					TagID: "/12344/Test_AdUnit",
					Native: &openrtb2.Native{
						Request: "Native_Reuqest",
					},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Native: &adunitconfig.Native{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.NativeConfig{
							Native: openrtb2.Native{
								Request: "Native_Reuqest",
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
					Native: &adunitconfig.Native{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.NativeConfig{
							Native: openrtb2.Native{
								Request: "Native_Reuqest",
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
				UsingDefaultConfig:     false,
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
					Native: &openrtb2.Native{
						Request: "Native_Reuqest",
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
			name: "native_config_is_prsent_in_both_default_and_slot_preferance_is_given_to_slot_level",
			args: args{
				rCtx: models.RequestCtx{
					MetricsEngine: mockEngine,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						ConfigPattern: "_AU_",
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								Native: &adunitconfig.Native{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.NativeConfig{
										Native: openrtb2.Native{
											Request: "Native_Reuqest_Default",
											Ver:     "1.2",
										},
									},
								},
							},
							"/12344/test_adunit": {
								Native: &adunitconfig.Native{
									Enabled: ptrutil.ToPtr(true),
									Config: &adunitconfig.NativeConfig{
										Native: openrtb2.Native{
											Request: "Native_Reuqest",
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
					Video: &openrtb2.Video{},
				},
			},
			wantAdUnitCtx: models.AdUnitCtx{
				MatchedSlot:  "/12344/Test_AdUnit",
				IsRegex:      false,
				MatchedRegex: "",
				SelectedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Native: &adunitconfig.Native{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.NativeConfig{
							Native: openrtb2.Native{
								Request: "Native_Reuqest",
							},
						},
					},
				},
				AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
					Native: &adunitconfig.Native{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.NativeConfig{
							Native: openrtb2.Native{
								Request: "Native_Reuqest",
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
			gotAdUnitCtx := UpdateNativeObjectWithAdunitConfig(tt.args.rCtx, tt.args.imp, tt.args.div)
			assert.Equal(t, tt.wantAdUnitCtx, gotAdUnitCtx)
		})
	}
}
