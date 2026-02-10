package publisherfeature

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeature_updateActConfigMapsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	type fields struct {
		publisherFeature map[int]map[int]models.FeatureData
		act              act
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
			name: "Cache_returns_valid_thresholdsPerDsp_and_disabled_publishers_updated_from_publisherFeature_map",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						17: models.FeatureData{
							Enabled: 0,
						},
					},
				},
				act: act{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
			},
			setup: func() {
				mockCache.EXPECT().GetACTThresholdPerDSP().Return(map[int]int{6: 70}, nil)
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
			name: "Cache_returns_DB_error_for_thresholdsPerDsp",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{},
				act: act{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
			},
			setup: func() {
				mockCache.EXPECT().GetACTThresholdPerDSP().Return(map[int]int{}, errors.New("QUERY FAILED"))
			},
			wantErr: true,
			want: wantMaps{
				disabledPublishers: map[int]struct{}{},
				thresholdsPerDsp:   map[int]int{},
			},
		},
		{
			name: "publisherFeature_map_is_empty_and_cache_returns_valid_thresholdsPerDsp",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{},
				act: act{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
			},
			setup: func() {
				mockCache.EXPECT().GetACTThresholdPerDSP().Return(map[int]int{
					6: 70,
				}, nil)
			},
			wantErr: false,
			want: wantMaps{
				disabledPublishers: map[int]struct{}{},
				thresholdsPerDsp: map[int]int{
					6: 70,
				},
			},
		},
		{
			name: "cache_returns_nil_thresholdsPerDsp_and_publisherFeature_map_is_not_empty",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						17: models.FeatureData{
							Enabled: 0,
						},
					},
				},
				act: act{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
			},
			setup: func() {
				mockCache.EXPECT().GetACTThresholdPerDSP().Return(nil, nil)
			},
			wantErr: false,
			want: wantMaps{
				disabledPublishers: map[int]struct{}{
					5890: {},
				},
				thresholdsPerDsp: map[int]int{},
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
				act:              tt.fields.act,
			}
			if err := fe.updateActConfigMapsFromCache(); (err != nil) != tt.wantErr {
				t.Errorf("feature.updateActConfigMapsFromCache() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want.thresholdsPerDsp, fe.act.thresholdsPerDsp)
			assert.Equal(t, tt.want.disabledPublishers, fe.act.disabledPublishers)
		})
	}
}

func TestPredictActValue(t *testing.T) {
	type args struct {
		percentage int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "getting_from_predict_output",
			args: args{
				percentage: 100,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := predictActValue(tt.args.percentage); got != tt.want {
				t.Errorf("predictActValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
