package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestGetIncomingSlots(t *testing.T) {
	type args struct {
		imp            openrtb2.Imp
		videoAdUnitCtx models.AdUnitCtx
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "only_native_slot",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Native: &openrtb2.Native{
						Request: `{"ver":"1.2"}`,
					},
				},
			},
			want: []string{"1x1"},
		},
		{
			name: "native_with_other_slots_then_do_not_consider_native",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Native: &openrtb2.Native{
						Request: `{"ver":"1.2"}`,
					},
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
					},
				},
			},
			want: []string{"300x250"},
		},
		{
			name: "only_banner_slot",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
					},
				},
			},
			want: []string{"300x250"},
		},
		{
			name: "banner_with_format",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
						Format: []openrtb2.Format{
							{
								W: 400,
								H: 300,
							},
						},
					},
				},
			},
			want: []string{"300x250", "400x300"},
		},
		{
			name: "only_video_slot",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
					},
				},
			},
			want: []string{"300x250"},
		},
		{
			name: "all_slots",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Native: &openrtb2.Native{
						Request: `{"ver":"1.2"}`,
					},
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
						Format: []openrtb2.Format{
							{
								W: 400,
								H: 300,
							},
						},
					},
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
					},
				},
			},
			want: []string{"300x250", "400x300"},
		},
		{
			name: "duplicate_slot",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
						Format: []openrtb2.Format{
							{
								W: 300,
								H: 250,
							},
						},
					},
				},
			},
			want: []string{"300x250"},
		},
		{
			name: "video sizes from adunit config, sizes not present in request",
			args: args{
				imp: openrtb2.Imp{
					ID:    "1",
					Video: &openrtb2.Video{},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(true),
							Config: &adunitconfig.VideoConfig{
								Video: openrtb2.Video{
									W: ptrutil.ToPtr(int64(640)),
									H: ptrutil.ToPtr(int64(480)),
								},
							},
						},
					},
				},
			},
			want: []string{"640x480"},
		},
		{
			name: "video sizes from request, sizes present in adunit and request",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr(int64(380)),
						H: ptrutil.ToPtr(int64(120)),
					},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(true),
							Config: &adunitconfig.VideoConfig{
								Video: openrtb2.Video{
									W: ptrutil.ToPtr(int64(640)),
									H: ptrutil.ToPtr(int64(480)),
								},
							},
						},
					},
				},
			},
			want: []string{"380x120"},
		},
		{
			name: "video object presnt but sizes not provided",
			args: args{
				imp: openrtb2.Imp{
					ID:    "1",
					Video: &openrtb2.Video{},
				},
				videoAdUnitCtx: models.AdUnitCtx{},
			},
			want: []string{"0x0"},
		},
		{
			name: "No sizes as video slot disabled from adunit",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr(int64(380)),
						H: ptrutil.ToPtr(int64(120)),
					},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(false),
							Config: &adunitconfig.VideoConfig{
								Video: openrtb2.Video{
									W: ptrutil.ToPtr(int64(640)),
									H: ptrutil.ToPtr(int64(480)),
								},
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "slot from adunit, enabled is not specified",
			args: args{
				imp: openrtb2.Imp{
					ID: "1",
					Video: &openrtb2.Video{
						W: nil,
						H: nil,
					},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Config: &adunitconfig.VideoConfig{
								Video: openrtb2.Video{
									W: ptrutil.ToPtr(int64(640)),
									H: ptrutil.ToPtr(int64(480)),
								},
							},
						},
					},
				},
			},
			want: []string{"640x480"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots := models.GetIncomingSlots(tt.args.imp, tt.args.videoAdUnitCtx)
			assert.ElementsMatch(t, tt.want, slots, "mismatched slots")
		})
	}
}
func TestUpdateIncomingSlotsWithFormat(t *testing.T) {
	tests := []struct {
		name          string
		incomingSlots []string
		format        []openrtb2.Format
		want          []string
	}{
		{
			name:          "no_format_returns_incomingSlots",
			incomingSlots: []string{"300x250", "728x90"},
			format:        nil,
			want:          []string{"300x250", "728x90"},
		},
		{
			name:          "empty_format_returns_incomingSlots",
			incomingSlots: []string{"300x250"},
			format:        []openrtb2.Format{},
			want:          []string{"300x250"},
		},
		{
			name:          "add_new_format_sizes",
			incomingSlots: []string{"300x250"},
			format: []openrtb2.Format{
				{W: 728, H: 90},
				{W: 160, H: 600},
			},
			want: []string{"300x250", "728x90", "160x600"},
		},
		{
			name:          "format_with_duplicate_sizes",
			incomingSlots: []string{"300x250"},
			format: []openrtb2.Format{
				{W: 300, H: 250},
				{W: 728, H: 90},
			},
			want: []string{"300x250", "728x90"},
		},
		{
			name:          "empty_incomingSlots_only_format",
			incomingSlots: []string{},
			format: []openrtb2.Format{
				{W: 320, H: 50},
			},
			want: []string{"320x50"},
		},
		{
			name:          "both_empty_returns_empty",
			incomingSlots: []string{},
			format:        []openrtb2.Format{},
			want:          []string{},
		},
		{
			name:          "format_with_zero_values",
			incomingSlots: []string{"300x250"},
			format: []openrtb2.Format{
				{W: 0, H: 0},
			},
			want: []string{"300x250", "0x0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateIncomingSlotsWithFormat(tt.incomingSlots, tt.format)
			assert.ElementsMatch(t, tt.want, got, "unexpected slots")
		})
	}
}
