package gocache

import (
	"errors"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database/mock"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_cache_GetAppSubIntegrationPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)

	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.Cache
		db    database.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]int
		wantErr bool
		setup   func()
	}{
		{
			name: "Valid Data present in DB, return same",
			want: map[string]int{
				"app_sub_int_1": 1,
				"app_sub_int_2": 2,
			},
			setup: func() {
				mockDatabase.EXPECT().GetAppSubIntegrationPaths().Return(map[string]int{
					"app_sub_int_1": 1,
					"app_sub_int_2": 2,
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
			want: map[string]int{},
			setup: func() {
				mockDatabase.EXPECT().GetAppSubIntegrationPaths().Return(map[string]int{}, errors.New("QUERY FAILD"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.AppSubIntegrationPathMapQuery, "", "").Return()
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			wantErr: true,
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			c := &cache{
				cache:        tt.fields.cache,
				cfg:          tt.fields.cfg,
				db:           tt.fields.db,
				metricEngine: mockEngine,
			}
			got, err := c.GetAppSubIntegrationPaths()
			if (err != nil) != tt.wantErr {
				t.Errorf("cache.GetAppSubIntegrationPaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
