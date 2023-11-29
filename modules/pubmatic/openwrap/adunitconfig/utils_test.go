package adunitconfig

import (
	"testing"

	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

var testSlotConfig = &adunitconfig.AdConfig{
	Video: &adunitconfig.Video{
		Config: &adunitconfig.VideoConfig{
			Video: openrtb2.Video{
				Placement:     1,
				Plcmt:         1,
				MinDuration:   10,
				MaxDuration:   20,
				SkipMin:       13,
				CompanionType: []adcom1.CompanionType{1, 2, 3},
			},
			ConnectionType: []int{
				10,
				20,
				30,
			},
		},
	},
}

var testDefaultconfig = &adunitconfig.AdConfig{
	Video: &adunitconfig.Video{
		Config: &adunitconfig.VideoConfig{
			Video: openrtb2.Video{
				Placement:     1,
				Plcmt:         1,
				MinDuration:   10,
				MaxDuration:   20,
				SkipMin:       13,
				CompanionType: []adcom1.CompanionType{1, 2, 3},
			},
			ConnectionType: []int{
				10,
				20,
				30,
			},
		},
	},
	Banner: &adunitconfig.Banner{
		Config: &adunitconfig.BannerConfig{
			Banner: openrtb2.Banner{
				ID: "123",
				W:  ptrutil.ToPtr[int64](100),
				H:  ptrutil.ToPtr[int64](200),
			},
		},
	},
	Floors: &openrtb_ext.PriceFloorRules{
		FloorMin:    10,
		FloorMinCur: "USD",
		Enabled:     ptrutil.ToPtr(true),
	},
}

func TestGetDefaultAllowedConnectionTypes(t *testing.T) {
	type args struct {
		adUnitConfigMap *adunitconfig.AdUnitConfig
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "adunitConfigMap_is_nil",
			args: args{
				adUnitConfigMap: nil,
			},
			want: nil,
		},
		{
			name: "adunitConfigMap_contian_non_empty_CompanionType",
			args: args{
				adUnitConfigMap: &adunitconfig.AdUnitConfig{
					Config: map[string]*adunitconfig.AdConfig{
						models.AdunitConfigDefaultKey: {
							Video: testSlotConfig.Video,
						},
					},
				},
			},
			want: []int{10, 20, 30},
		},
		{
			name: "adunitConfigMap_conatian_empty_CompanionType",
			args: args{
				adUnitConfigMap: &adunitconfig.AdUnitConfig{
					Config: map[string]*adunitconfig.AdConfig{
						models.AdunitConfigDefaultKey: {
							Video: &adunitconfig.Video{
								Config: &adunitconfig.VideoConfig{
									Video: openrtb2.Video{
										CompanionType: []adcom1.CompanionType{},
									},
									ConnectionType: []int{
										10,
										20,
										30,
									},
								},
							},
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultAllowedConnectionTypes(tt.args.adUnitConfigMap)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetFinalSlotAdUnitConfig(t *testing.T) {
	type args struct {
		slotConfig    *adunitconfig.AdConfig
		defaultConfig *adunitconfig.AdConfig
	}
	tests := []struct {
		name string
		args args
		want *adunitconfig.AdConfig
	}{
		{
			name: "both_slotConfig_and_defaultConfig_are_nil",
			args: args{
				slotConfig:    nil,
				defaultConfig: nil,
			},
			want: nil,
		},
		{
			name: "slotConfig_is_nil",
			args: args{
				slotConfig:    nil,
				defaultConfig: testDefaultconfig,
			},
			want: testDefaultconfig,
		},
		{
			name: "defaultconfig_is_nil",
			args: args{
				slotConfig:    testSlotConfig,
				defaultConfig: nil,
			},
			want: testSlotConfig,
		},
		{
			name: "both_avilable_merge_priority_to_slot",
			args: args{
				defaultConfig: testDefaultconfig,
				slotConfig:    testSlotConfig,
			},
			want: &adunitconfig.AdConfig{
				Video:  testSlotConfig.Video,
				Banner: testDefaultconfig.Banner,
				Floors: testDefaultconfig.Floors,
			},
		},
		{
			name: "Video_and_banner_is_not_avilable_in_slot_update_from_default",
			args: args{
				slotConfig: &adunitconfig.AdConfig{
					Floors: &openrtb_ext.PriceFloorRules{
						FloorMin:    10,
						FloorMinCur: "USD",
						Enabled:     ptrutil.ToPtr(true),
					},
				},
				defaultConfig: testDefaultconfig,
			},
			want: testDefaultconfig,
		},
		{
			name: "Bidfloor_is_absent_in_slot,_present_in_default,_and_default_lacks_BidFloorCur",
			args: args{
				slotConfig: &adunitconfig.AdConfig{
					BidFloor: nil,
				},
				defaultConfig: &adunitconfig.AdConfig{
					BidFloor: ptrutil.ToPtr[float64](4),
				},
			},
			want: &adunitconfig.AdConfig{
				BidFloor:    ptrutil.ToPtr[float64](4),
				BidFloorCur: ptrutil.ToPtr(models.USD),
			},
		},
		{
			name: "Bidfloor_is_absent_in_slot,_present_in_default,_and_default_also_have_BidFloorCur",
			args: args{
				slotConfig: &adunitconfig.AdConfig{
					BidFloor: nil,
				},
				defaultConfig: &adunitconfig.AdConfig{
					BidFloor:    ptrutil.ToPtr[float64](4),
					BidFloorCur: ptrutil.ToPtr("INR"),
				},
			},
			want: &adunitconfig.AdConfig{
				BidFloor:    ptrutil.ToPtr[float64](4),
				BidFloorCur: ptrutil.ToPtr("INR"),
			},
		},
		{
			name: "Bidfloor_is_present_in_slot_but_has_zero_value_and_default_have_Bidfloor_and_BidFloorCur",
			args: args{
				slotConfig: &adunitconfig.AdConfig{
					BidFloor:    ptrutil.ToPtr[float64](0.0),
					BidFloorCur: ptrutil.ToPtr("INR"),
				},
				defaultConfig: &adunitconfig.AdConfig{
					BidFloor:    ptrutil.ToPtr[float64](4.0),
					BidFloorCur: ptrutil.ToPtr("EUR"),
				},
			},
			want: &adunitconfig.AdConfig{
				BidFloor:    ptrutil.ToPtr[float64](4.0),
				BidFloorCur: ptrutil.ToPtr("EUR"),
			},
		},
		{
			name: "Bid_Floor_from_slot_config_having_only_currency._default_gets_selected",
			args: args{
				slotConfig: &adunitconfig.AdConfig{
					BidFloorCur: ptrutil.ToPtr("INR"),
				},
				defaultConfig: &adunitconfig.AdConfig{
					BidFloor:    ptrutil.ToPtr[float64](10.0),
					BidFloorCur: ptrutil.ToPtr("EUR"),
				},
			},
			want: &adunitconfig.AdConfig{
				BidFloor:    ptrutil.ToPtr[float64](10.0),
				BidFloorCur: ptrutil.ToPtr("EUR"),
			},
		},
		{
			name: "Bid_Floor,_No_bidfloorCur_in-default_config,_floor_value_from_default_gets_selected_and_default_currency_USD_gets_set",
			args: args{
				slotConfig: &adunitconfig.AdConfig{},
				defaultConfig: &adunitconfig.AdConfig{
					BidFloor: ptrutil.ToPtr[float64](10.0),
				},
			},
			want: &adunitconfig.AdConfig{
				BidFloor:    ptrutil.ToPtr[float64](10.0),
				BidFloorCur: ptrutil.ToPtr("USD"),
			},
		},
		{
			name: "slotConfig_has_bidfloor_but_not_have_BidFloorCur_set_by_default_USD",
			args: args{
				slotConfig: &adunitconfig.AdConfig{
					BidFloor: ptrutil.ToPtr[float64](4.0),
				},
				defaultConfig: &adunitconfig.AdConfig{
					BidFloor:    ptrutil.ToPtr[float64](5.0),
					BidFloorCur: ptrutil.ToPtr("EUR"),
				},
			},
			want: &adunitconfig.AdConfig{
				BidFloor:    ptrutil.ToPtr[float64](4.0),
				BidFloorCur: ptrutil.ToPtr("USD"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFinalSlotAdUnitConfig(tt.args.slotConfig, tt.args.defaultConfig)
			assert.Equal(t, tt.want, got)
		})
	}
}
