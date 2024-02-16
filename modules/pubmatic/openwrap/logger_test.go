package openwrap

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func Test_getIncomingSlots(t *testing.T) {
	type args struct {
		imp openrtb2.Imp
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
						W: 300,
						H: 250,
					},
				},
			},
			want: []string{"300x250v"},
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
						W: 300,
						H: 250,
					},
				},
			},
			want: []string{"300x250", "400x300", "300x250v"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIncomingSlots(tt.args.imp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getIncomingSlots() = %v, want %v", got, tt.want)
			}
		})
	}
}
