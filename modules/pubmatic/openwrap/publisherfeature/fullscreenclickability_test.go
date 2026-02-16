package publisherfeature

import (
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/stretchr/testify/assert"
)

func TestIsUnderFSCThreshold(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	type fields struct {
		fsc fsc
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
			name: "When pubId,dspid and FSC maps are valid, pubID enabled(default) FSC return fsc with prediction algo",
			args: args{
				pubid: 5890,
				dspid: 6,
			},
			fields: fields{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: 1,
		},
		{
			name: "When pubId,dspid and FSC maps are valid, pubID disabled FSC return fsc=0",
			args: args{
				pubid: 5890,
				dspid: 6,
			},
			fields: fields{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{
						5890: {},
					},
					thresholdsPerDsp: map[int]int{6: 100}},
			},
			want: 0,
		},
		{
			name: "When pubId,dspid are not present, pubID disabled FSC return fsc=0",
			args: args{
				pubid: 58907,
				dspid: 90,
			},
			fields: fields{
				fsc: fsc{
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
				fsc:   tt.fields.fsc,
			}
			if got := fe.isUnderFSCThreshold(tt.args.pubid, tt.args.dspid); got != tt.want {
				t.Errorf("fsc.IsUnderFSCThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFscApplicable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		fsc fsc
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
			name: "Valid Case1: All Params Correct",
			args: args{
				pubId: 5890,
				seat:  "pubmatic",
				dspId: 6,
			},
			fields: fields{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: true,
		},
		{
			name: "Valid Case2: All Params Correct, seat is pubmatic alaias",
			args: args{
				pubId: 5890,
				seat:  "pubmatic2",
				dspId: 6,
			},
			fields: fields{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: true,
		},
		{
			name: "Invalid Case1: DspId is 0",
			args: args{
				pubId: 5890,
				seat:  "pubmatic",
				dspId: 0,
			},
			fields: fields{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{
						58903: {},
					},
					thresholdsPerDsp: map[int]int{6: 100},
				},
			},
			want: false,
		},
		{
			name: "Invalid Case2: different seat ",
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
				fsc:   tt.fields.fsc,
			}
			got := fe.IsFscApplicable(tt.args.pubId, tt.args.seat, tt.args.dspId)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
