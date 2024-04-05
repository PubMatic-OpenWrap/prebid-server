package publisherfeature

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeature_updateFscConfigMapsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	type fields struct {
		publisherFeature map[int]map[int]models.FeatureData
		fsc              fsc
	}
	type wantMaps struct {
		disabledPublishers map[int]struct{}
		thresholdsPerDsp   map[int]int
	}
	tests := []struct {
		name    string
		fields  fields
		setup   func()
		wantErr bool
		want    wantMaps
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				publisherFeature: nil,
				fsc: fsc{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
			},
			wantErr: false,
			want: wantMaps{
				disabledPublishers: map[int]struct{}{},
				thresholdsPerDsp:   map[int]int{},
			},
		},
		{
			name: "Cache returns valid thresholdsPerDsp and disabled publishers updated from publisherFeature map",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						1: models.FeatureData{
							Enabled: 0,
						},
					},
				},
				fsc: fsc{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{6: 70}, nil)
			},
			want: wantMaps{
				thresholdsPerDsp: map[int]int{6: 70},
				disabledPublishers: map[int]struct{}{
					5890: {},
				},
			},
			wantErr: false,
		},
		{
			name: "Cache returns DB error for thresholdsPerDsp",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{},
				fsc: fsc{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{}, errors.New("QUERY FAILED"))
			},
			wantErr: true,
			want: wantMaps{
				disabledPublishers: map[int]struct{}{},
				thresholdsPerDsp:   map[int]int{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			fe := feature{
				cache:            mockCache,
				publisherFeature: tt.fields.publisherFeature,
				fsc:              tt.fields.fsc,
			}
			if err := fe.updateFscConfigMapsFromCache(); (err != nil) != tt.wantErr {
				t.Errorf("feature.updateFscConfigMapsFromCache() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want.thresholdsPerDsp, fe.fsc.thresholdsPerDsp)
			assert.Equal(t, tt.want.disabledPublishers, fe.fsc.disabledPublishers)
		})
	}
}

func TestPredictFscValue(t *testing.T) {
	type args struct {
		percentage int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "getting from predict output",
			args: args{
				percentage: 100,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := predictFscValue(tt.args.percentage); got != tt.want {
				t.Errorf("predictFscValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
