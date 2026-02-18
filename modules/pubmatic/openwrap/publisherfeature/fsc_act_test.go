package publisherfeature

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeature_updateFscAndActConfigMapsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	type fields struct {
		publisherFeature map[int]map[int]models.FeatureData
		fsc              fsc
		act              act
	}
	type wantMaps struct {
		fscDisabled   map[int]struct{}
		fscThresholds map[int]int
		actDisabled   map[int]struct{}
		actThresholds map[int]int
	}
	tests := []struct {
		name    string
		fields  fields
		setup   func()
		wantErr bool
		want    wantMaps
	}{
		{
			name: "Cache returns valid FSC and ACT thresholds and disabled publishers updated from publisherFeature map",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						models.FeatureFSC: models.FeatureData{Enabled: 0},
						models.FeatureACT: models.FeatureData{Enabled: 0},
					},
				},
				fsc: fsc{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
				act: act{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCAndACTThresholdsPerDSP().Return(map[int]int{6: 70}, map[int]int{6: 80}, nil)
			},
			want: wantMaps{
				fscDisabled:   map[int]struct{}{5890: {}},
				fscThresholds: map[int]int{6: 70},
				actDisabled:   map[int]struct{}{5890: {}},
				actThresholds: map[int]int{6: 80},
			},
			wantErr: false,
		},
		{
			name: "Cache returns DB error",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{},
				fsc:              fsc{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
				act:              act{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCAndACTThresholdsPerDSP().Return(nil, nil, errors.New("QUERY FAILED"))
			},
			want:    wantMaps{fscDisabled: map[int]struct{}{}, fscThresholds: map[int]int{}, actDisabled: map[int]struct{}{}, actThresholds: map[int]int{}},
			wantErr: true,
		},
		{
			name: "publisherFeature map is empty and cache returns valid thresholds",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{},
				fsc:              fsc{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
				act:              act{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCAndACTThresholdsPerDSP().Return(map[int]int{6: 70}, map[int]int{6: 80}, nil)
			},
			want: wantMaps{
				fscDisabled:   map[int]struct{}{},
				fscThresholds: map[int]int{6: 70},
				actDisabled:   map[int]struct{}{},
				actThresholds: map[int]int{6: 80},
			},
			wantErr: false,
		},
		{
			name: "cache returns nil thresholds and publisherFeature has disabled FSC and ACT",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						models.FeatureFSC: models.FeatureData{Enabled: 0},
						models.FeatureACT: models.FeatureData{Enabled: 0},
					},
				},
				fsc: fsc{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
				act: act{disabledPublishers: make(map[int]struct{}), thresholdsPerDsp: make(map[int]int)},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCAndACTThresholdsPerDSP().Return(nil, nil, nil)
			},
			want: wantMaps{
				fscDisabled:   map[int]struct{}{5890: {}},
				fscThresholds: map[int]int{},
				actDisabled:   map[int]struct{}{5890: {}},
				actThresholds: map[int]int{},
			},
			wantErr: false,
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
				act:              tt.fields.act,
			}
			err := fe.updateFscAndActConfigMapsFromCache()
			if (err != nil) != tt.wantErr {
				t.Errorf("updateFscAndActConfigMapsFromCache() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want.fscThresholds, fe.fsc.thresholdsPerDsp)
			assert.Equal(t, tt.want.fscDisabled, fe.fsc.disabledPublishers)
			assert.Equal(t, tt.want.actThresholds, fe.act.thresholdsPerDsp)
			assert.Equal(t, tt.want.actDisabled, fe.act.disabledPublishers)
		})
	}
}
