package googlesdk

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/feature"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestGetFlexSlotSizes(t *testing.T) {
	type args struct {
		banner   *openrtb2.Banner
		features feature.Features
	}
	tests := []struct {
		name string
		args args
		want []openrtb2.Format
	}{
		{
			name: "nil banner",
			args: args{
				banner:   nil,
				features: feature.Features{},
			},
			want: nil,
		},
		{
			name: "nil banner.Ext",
			args: args{
				banner:   &openrtb2.Banner{},
				features: feature.Features{},
			},
			want: nil,
		},
		{
			name: "nil features",
			args: args{
				banner:   &openrtb2.Banner{Ext: json.RawMessage(`{}`)},
				features: nil,
			},
			want: nil,
		},
		{
			name: "missing google sdk feature",
			args: args{
				banner:   &openrtb2.Banner{Ext: json.RawMessage(`{}`)},
				features: feature.Features{},
			},
			want: nil,
		},
		{
			name: "empty google sdk feature",
			args: args{
				banner:   &openrtb2.Banner{Ext: json.RawMessage(`{}`)},
				features: feature.Features{feature.FeatureNameGoogleSDK: []feature.Feature{}},
			},
			want: nil,
		},
		{
			name: "invalid banner.Ext",
			args: args{
				banner:   &openrtb2.Banner{Ext: json.RawMessage(`invalid`)},
				features: feature.Features{feature.FeatureNameGoogleSDK: []feature.Feature{{}}},
			},
			want: nil,
		},
		{
			name: "nil Flexslot in banner.Ext",
			args: func() args {
				ext := openrtb_ext.ExtImpBanner{}
				extBytes, _ := json.Marshal(ext)
				return args{
					banner:   &openrtb2.Banner{Ext: extBytes},
					features: feature.Features{feature.FeatureNameGoogleSDK: []feature.Feature{{}}},
				}
			}(),
			want: nil,
		},
		{
			name: "no FeatureFlexSlot in features",
			args: func() args {
				ext := openrtb_ext.ExtImpBanner{Flexslot: &openrtb_ext.FlexSlot{Wmin: 10, Wmax: 100, Hmin: 10, Hmax: 100}}
				extBytes, _ := json.Marshal(ext)
				return args{
					banner:   &openrtb2.Banner{Ext: extBytes},
					features: feature.Features{feature.FeatureNameGoogleSDK: []feature.Feature{{Name: "other"}}},
				}
			}(),
			want: nil,
		},
		{
			name: "FeatureFlexSlot with non-[]string Data",
			args: func() args {
				ext := openrtb_ext.ExtImpBanner{Flexslot: &openrtb_ext.FlexSlot{Wmin: 10, Wmax: 100, Hmin: 10, Hmax: 100}}
				extBytes, _ := json.Marshal(ext)
				return args{
					banner: &openrtb2.Banner{Ext: extBytes},
					features: feature.Features{
						feature.FeatureNameGoogleSDK: []feature.Feature{
							{Name: feature.FeatureFlexSlot, Data: 123},
						},
					},
				}
			}(),
			want: nil,
		},
		{
			name: "FeatureFlexSlot with empty []string Data",
			args: func() args {
				ext := openrtb_ext.ExtImpBanner{Flexslot: &openrtb_ext.FlexSlot{Wmin: 10, Wmax: 100, Hmin: 10, Hmax: 100}}
				extBytes, _ := json.Marshal(ext)
				return args{
					banner: &openrtb2.Banner{Ext: extBytes},
					features: feature.Features{
						feature.FeatureNameGoogleSDK: []feature.Feature{
							{Name: feature.FeatureFlexSlot, Data: []string{}},
						},
					},
				}
			}(),
			want: nil,
		},
		{
			name: "FeatureFlexSlot with valid and invalid sizes",
			args: func() args {
				ext := openrtb_ext.ExtImpBanner{Flexslot: &openrtb_ext.FlexSlot{Wmin: 10, Wmax: 100, Hmin: 10, Hmax: 100}}
				extBytes, _ := json.Marshal(ext)
				return args{
					banner: &openrtb2.Banner{Ext: extBytes},
					features: feature.Features{
						feature.FeatureNameGoogleSDK: []feature.Feature{
							{Name: feature.FeatureFlexSlot, Data: []string{"20x30", "invalid", "200x300", "50x50"}},
						},
					},
				}
			}(),
			want: []openrtb2.Format{
				{W: 20, H: 30},
				{W: 50, H: 50},
			},
		},
		{
			name: "FeatureFlexSlot with sizes outside allowed range",
			args: func() args {
				ext := openrtb_ext.ExtImpBanner{Flexslot: &openrtb_ext.FlexSlot{Wmin: 10, Wmax: 20, Hmin: 10, Hmax: 20}}
				extBytes, _ := json.Marshal(ext)
				return args{
					banner: &openrtb2.Banner{Ext: extBytes},
					features: feature.Features{
						feature.FeatureNameGoogleSDK: []feature.Feature{
							{Name: feature.FeatureFlexSlot, Data: []string{"30x40", "50x60"}},
						},
					},
				}
			}(),
			want: nil,
		},
		{
			name: "invalid flexslot range",
			args: func() args {
				ext := openrtb_ext.ExtImpBanner{Flexslot: &openrtb_ext.FlexSlot{Wmin: 100, Wmax: 10, Hmin: 100, Hmax: 10}}
				extBytes, _ := json.Marshal(ext)
				return args{
					banner: &openrtb2.Banner{Ext: extBytes},
					features: feature.Features{
						feature.FeatureNameGoogleSDK: []feature.Feature{
							{Name: feature.FeatureFlexSlot, Data: []string{"20x30", "50x50"}},
						},
					},
				}
			}(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFlexSlotSizes(tt.args.banner, tt.args.features)
			if tt.want == nil {
				assert.Empty(t, got)
				return
			}
			sort.Slice(got, func(i, j int) bool {
				return got[i].W < got[j].W || (got[i].W == got[j].W && got[i].H < got[j].H)
			})

			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].W < tt.want[j].W || (tt.want[i].W == tt.want[j].W && tt.want[i].H < tt.want[j].H)
			})

			assert.Equal(t, tt.want, got, "GetFlexSlotSizes() = %v, want %v", got, tt.want)
		})
	}
}
func TestSetFlexSlotSizes(t *testing.T) {
	type args struct {
		banner *openrtb2.Banner
		rCtx   models.RequestCtx
	}
	tests := []struct {
		name         string
		args         args
		expected     []openrtb2.Format
		original     []openrtb2.Format
		shouldChange bool
	}{
		{
			name: "nil banner",
			args: args{
				banner: nil,
				rCtx:   models.RequestCtx{},
			},
			expected:     nil,
			shouldChange: false,
		},
		{
			name: "nil FlexSlot in rCtx.GoogleSDK",
			args: args{
				banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 10, H: 20}}},
				rCtx:   models.RequestCtx{},
			},
			expected:     []openrtb2.Format{{W: 10, H: 20}},
			shouldChange: false,
		},
		{
			name: "empty FlexSlot in rCtx.GoogleSDK",
			args: args{
				banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 10, H: 20}}},
				rCtx:   models.RequestCtx{GoogleSDK: models.GoogleSDK{FlexSlot: []openrtb2.Format{}}},
			},
			expected:     []openrtb2.Format{{W: 10, H: 20}},
			shouldChange: false,
		},
		{
			name: "FlexSlot with new sizes",
			args: args{
				banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 10, H: 20}}},
				rCtx: models.RequestCtx{
					GoogleSDK: models.GoogleSDK{
						FlexSlot: []openrtb2.Format{{W: 30, H: 40}, {W: 50, H: 60}},
					},
				},
			},
			expected:     []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}, {W: 50, H: 60}},
			shouldChange: true,
		},
		{
			name: "FlexSlot with duplicate and new sizes",
			args: args{
				banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}}},
				rCtx: models.RequestCtx{
					GoogleSDK: models.GoogleSDK{
						FlexSlot: []openrtb2.Format{{W: 30, H: 40}, {W: 50, H: 60}},
					},
				},
			},
			expected:     []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}, {W: 50, H: 60}},
			shouldChange: true,
		},
		{
			name: "FlexSlot with all duplicates",
			args: args{
				banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}}},
				rCtx: models.RequestCtx{
					GoogleSDK: models.GoogleSDK{
						FlexSlot: []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}},
					},
				},
			},
			expected:     []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}},
			shouldChange: false,
		},
		{
			name: "flexslot but nil banner format",
			args: args{
				banner: &openrtb2.Banner{Format: nil},
				rCtx: models.RequestCtx{
					GoogleSDK: models.GoogleSDK{
						FlexSlot: []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}},
					},
				},
			},
			expected:     []openrtb2.Format{{W: 10, H: 20}, {W: 30, H: 40}},
			shouldChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy original for comparison if needed
			var orig []openrtb2.Format
			if tt.args.banner != nil {
				orig = make([]openrtb2.Format, len(tt.args.banner.Format))
				copy(orig, tt.args.banner.Format)
			}
			SetFlexSlotSizes(tt.args.banner, tt.args.rCtx)
			if tt.args.banner == nil {
				assert.Nil(t, tt.args.banner)
				return
			}
			// Sort for comparison
			got := tt.args.banner.Format
			sort.Slice(got, func(i, j int) bool {
				return got[i].W < got[j].W || (got[i].W == got[j].W && got[i].H < got[j].H)
			})
			sort.Slice(tt.expected, func(i, j int) bool {
				return tt.expected[i].W < tt.expected[j].W || (tt.expected[i].W == tt.expected[j].W && tt.expected[i].H < tt.expected[j].H)
			})
			assert.Equal(t, tt.expected, got)
			if !tt.shouldChange && tt.args.banner != nil {
				assert.Equal(t, orig, tt.args.banner.Format)
			}
		})
	}
}
