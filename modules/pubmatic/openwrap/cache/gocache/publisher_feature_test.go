package gocache

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_cache_GetPublisherFeatureMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	type fields struct {
		cache        *gocache.Cache
		cfg          config.Cache
		db           database.Database
		metricEngine metrics.MetricsEngine
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[int]map[int]models.FeatureData
		setup   func()
		wantErr bool
	}{
		{
			name: "Valid Data present in DB, return same",
			want: map[int]map[int]models.FeatureData{
				5890: {
					models.FeatureFSC: {
						Enabled: 0,
					},
					models.FeatureTBF: {
						Enabled: 1,
						Value:   `{"1234": 100}`,
					},
					models.FeatureAMPMultiFormat: {
						Enabled: 1,
					},
				},
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherFeatureMap().Return(map[int]map[int]models.FeatureData{
					5890: {
						models.FeatureFSC: {
							Enabled: 0,
						},
						models.FeatureTBF: {
							Enabled: 1,
							Value:   `{"1234": 100}`,
						},
						models.FeatureAMPMultiFormat: {
							Enabled: 1,
						},
					},
				}, nil)
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			wantErr: false,
		},
		{
			name: "Error In DB, Set Empty",
			want: map[int]map[int]models.FeatureData{},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherFeatureMap().Return(map[int]map[int]models.FeatureData{}, errors.New("QUERY FAILD"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.PublisherFeatureMapQuery, "", "").Return()
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
				metricEngine: mockEngine,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			c := &cache{
				cache:        tt.fields.cache,
				cfg:          tt.fields.cfg,
				db:           tt.fields.db,
				metricEngine: tt.fields.metricEngine,
			}
			got, err := c.GetPublisherFeatureMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("cache.GetPublisherFeatureMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
