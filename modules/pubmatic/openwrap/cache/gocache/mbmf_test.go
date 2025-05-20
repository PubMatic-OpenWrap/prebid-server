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

func TestCacheGetProfileAdUnitMultiFloors(t *testing.T) {
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
		setup   func()
		want    models.ProfileAdUnitMultiFloors
		wantErr error
	}{
		{
			name: "Success case",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
				metricEngine: mockEngine,
			},
			setup: func() {
				expectedFloors := models.ProfileAdUnitMultiFloors{
					1: map[string]*models.MultiFloors{
						"adunit1": {IsActive: true, Tier1: 1.0, Tier2: 0.8, Tier3: 0.6},
					},
				}
				mockDatabase.EXPECT().GetProfileAdUnitMultiFloors().Return(expectedFloors, nil)
			},
			want: models.ProfileAdUnitMultiFloors{
				1: map[string]*models.MultiFloors{
					"adunit1": {IsActive: true, Tier1: 1.0, Tier2: 0.8, Tier3: 0.6},
				},
			},
			wantErr: nil,
		},
		{
			name: "Database error",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
				metricEngine: mockEngine,
			},
			setup: func() {
				mockDatabase.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, errors.New("db error"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.ProfileAdUnitMultiFloorsQuery, "", "").Return()
			},
			want:    models.ProfileAdUnitMultiFloors{},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			c := &cache{
				cache:        tt.fields.cache,
				cfg:          tt.fields.cfg,
				db:           tt.fields.db,
				metricEngine: tt.fields.metricEngine,
			}

			got, err := c.GetProfileAdUnitMultiFloors()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
