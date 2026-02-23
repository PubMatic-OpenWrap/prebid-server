package publisherfeature

import (
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/stretchr/testify/assert"
)

func TestIsUnderACTThreshold(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	type fields struct {
		act act
	}
	type args struct {
		pubid int
		dspid int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "When_pubId_dspid_and_ACT_maps_are_valid,_pubID_enabled(default)_ACT_return_act_with_prediction_algo",
			args: args{
				pubid: 5890,
				dspid: 6,
			},
			fields: fields{
				act: act{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: 1,
		},
		{
			name: "When_pubId_dspid_and_ACT_maps_are_valid,_pubID_disabled_ACT_return_act_with_prediction_algo",
			args: args{
				pubid: 5890,
				dspid: 6,
			},
			fields: fields{
				act: act{
					disabledPublishers: map[int]struct{}{
						5890: {},
					},
					thresholdsPerDsp: map[int]int{6: 100}},
			},
			want: 0,
		},
		{
			name: "When_pubId_dspid_are_not_present,_pubID_disabled_ACT_return_act_with_prediction_algo",
			args: args{
				pubid: 58907,
				dspid: 90,
			},
			fields: fields{
				act: act{
					disabledPublishers: map[int]struct{}{
						5890: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := feature{
				cache: mockCache,
				act:   tt.fields.act,
			}
			if got := fe.isUnderACTThreshold(tt.args.pubid, tt.args.dspid); got != tt.want {
				t.Errorf("act.IsUnderACTThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsActApplicable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		act act
	}
	type args struct {
		pubId int
		seat  string
		dspId int
	}
	tests := []struct {
		name   string
		args   args
		want   bool
		fields fields
	}{
		{
			name: "Valid_Case1:All_Params_Correct",
			args: args{
				pubId: 5890,
				seat:  "pubmatic",
				dspId: 6,
			},
			fields: fields{
				act: act{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: true,
		},
		{
			name: "Valid_Case2:All_Params_Correct_seat_is_pubmatic_alaias",
			args: args{
				pubId: 5890,
				seat:  "pubmatic2",
				dspId: 6,
			},
			fields: fields{
				act: act{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: true,
		},
		{
			name: "Invalid_Case1:_DspId_is_0",
			args: args{
				pubId: 5890,
				seat:  "pubmatic",
				dspId: 0,
			},
			fields: fields{
				act: act{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: false,
		},
		{
			name: "Invalid_Case2:_different_seat ",
			args: args{
				pubId: 5890,
				seat:  "appnexus",
				dspId: 6,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := feature{
				cache: mockCache,
				act:   tt.fields.act,
			}
			got := fe.IsActApplicable(tt.args.pubId, tt.args.seat, tt.args.dspId)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
