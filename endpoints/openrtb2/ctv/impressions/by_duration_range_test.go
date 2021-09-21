package impressions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type args struct {
	podMinDuration     int
	podMaxDuration     int
	minAds             int
	maxAds             int
	durationRangeInSec []int
	durationMatching   string
}
type want struct {
	imps [][2]int64
}

var EXACT_MATCH string = "exact"
var ROUND_UP string = "roundup"

var impressionsTestsA3 = []struct {
	scenario string
	args     args
	want     want
}{
	// {scenario: "no_adpod_context"},
	{scenario: "nil_durationRangeInSec",
		args: args{durationRangeInSec: nil},
		want: want{imps: make([][2]int64, 0)}}, // do not generate impressions

	{scenario: "empty_durationRangeInSec",
		args: args{durationRangeInSec: make([]int, 0)},
		want: want{imps: make([][2]int64, 0)}}, // do not generate impressions

	// {scenario: "invalid_durationRangeInSec"},
	// {scenario: "nil_durationMatching"},
	// {scenario: "empty_durationMatching"},
	// {scenario: "invalid_durationMatching"},
	// {scenario: "len_of_durationRangeInSec_<_maxAds"},
	// {scenario: "len_of_durationRangeInSec_>_maxAds"},
	{scenario: "durations_in_durationRangeInSec_>podMaxDuration",
		args: args{
			durationRangeInSec: []int{5, 10, 15},
			durationMatching:   EXACT_MATCH,
			podMaxDuration:     10,
		}, want: want{
			imps: [][2]int64{ // do not expect {15,15}
				{5, 5},
				{10, 10},
			},
		},
	},
	{scenario: "durations_in_durationRangeInSec_<podMinDuration",
		args: args{
			podMinDuration:     10,
			durationRangeInSec: []int{5, 10, 15},
			durationMatching:   EXACT_MATCH,
			podMaxDuration:     20,
		}, want: want{
			imps: [][2]int64{ // do not expect {15,15}
				{10, 10},
				{15, 15},
			},
		},
	},
	{scenario: "valid_scenario",
		args: args{
			podMaxDuration:     20,
			durationRangeInSec: []int{5, 10, 15},
			durationMatching:   EXACT_MATCH,
		}, want: want{
			imps: [][2]int64{
				{5, 5},
				{10, 10},
				{15, 15},
			},
		},
	},
}

func TestGetImpressionsA3(t *testing.T) {
	for _, test := range impressionsTestsA3 {
		t.Run(test.scenario, func(t *testing.T) {
			args := test.args
			gen := newByDurationRanges(args.durationRangeInSec, args.podMinDuration, args.podMaxDuration)
			imps := gen.Get()
			assert.Equalf(t, test.want.imps, imps, "Expected '%v' but got '%v'", test.want.imps, imps)
		})
	}
}
