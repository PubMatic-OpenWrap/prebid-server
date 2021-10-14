package impressions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetImpressionsA3(t *testing.T) {
	type args struct {
		podMaxDuration int
		maxAds         int
		durations      []int
	}
	type want struct {
		imps [][2]int64
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			// do not generate impressions
			name: "no_adpod_context",
			args: args{},
			want: want{
				imps: [][2]int64{},
			},
		},
		{
			// do not generate impressions
			name: "nil_durations",
			args: args{
				durations: nil,
			},
			want: want{
				imps: make([][2]int64, 0),
			},
		},
		{
			// do not generate impressions
			name: "empty_durations",
			args: args{
				durations: make([]int, 0),
			},
			want: want{
				imps: make([][2]int64, 0),
			},
		},
		{
			name: "len_of_durations_<_maxAds",
			args: args{
				podMaxDuration: 20,
				maxAds:         5,
				durations:      []int{5, 10, 15},
			},
			want: want{
				imps: [][2]int64{
					{5, 5},
					{10, 10},
					{15, 15},
					//got repeated because of current video duration impressions are less than maxads
					{5, 5},
					{10, 10},
				},
			},
		},
		{
			name: "len_of_durations_>_maxAds",
			args: args{
				podMaxDuration: 25,
				maxAds:         2,
				durations:      []int{5, 10, 15},
			},
			want: want{
				imps: [][2]int64{
					{5, 5},
					{10, 10},
					{15, 15},
					//got repeated because need to cover all video durations for creatives
				},
			},
		},
		{
			name: "durations_in_durations_>podMaxDuration",
			args: args{
				durations:      []int{5, 10, 15},
				podMaxDuration: 10,
			},
			want: want{
				imps: [][2]int64{
					// do not expect {15,15}
					{5, 5},
					{10, 10},
				},
			},
		},
		{
			name: "all_durations_in_durations_>podMaxDuration",
			args: args{
				durations:      []int{15, 20, 25},
				podMaxDuration: 10,
				maxAds:         3,
			},
			want: want{
				imps: [][2]int64{},
			},
		},
		{
			name: "valid_name",
			args: args{
				podMaxDuration: 20,
				durations:      []int{5, 10, 15},
			},
			want: want{
				imps: [][2]int64{
					{5, 5},
					{10, 10},
					{15, 15},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args
			gen := newByDurationRanges(args.durations, args.maxAds, args.podMaxDuration)
			imps := gen.Get()
			assert.Equal(t, tt.want.imps, imps)
		})
	}
}
