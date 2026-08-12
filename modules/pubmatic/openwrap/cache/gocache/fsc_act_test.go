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

func TestCache_GetFSCAndACTThresholdsPerDSP(t *testing.T) {
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
		name     string
		fields   fields
		wantFsc  map[int]int
		wantAct  map[int]int
		setup    func()
		wantErr  bool
		wantWrap string
	}{
		{
			name:    "Valid Data present in DB, return same",
			wantFsc: map[int]int{6: 70, 7: 50},
			wantAct: map[int]int{6: 80, 7: 60},
			setup: func() {
				mockDatabase.EXPECT().GetFSCAndACTThresholdsPerDSP().
					Return(map[int]int{6: 70, 7: 50}, map[int]int{6: 80, 7: 60}, nil)
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
			name:    "Error In DB, record metric and return wrapped error",
			wantFsc: nil,
			wantAct: nil,
			setup: func() {
				mockDatabase.EXPECT().GetFSCAndACTThresholdsPerDSP().
					Return(nil, nil, errors.New("QUERY FAILED"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.AllDspFscAndActPcntQuery, "", "").Return()
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
				metricEngine: mockEngine,
			},
			wantErr:  true,
			wantWrap: "ErrorFscActDspUpdate",
		},
		{
			name:    "Empty maps from DB, return empty",
			wantFsc: map[int]int{},
			wantAct: map[int]int{},
			setup: func() {
				mockDatabase.EXPECT().GetFSCAndACTThresholdsPerDSP().
					Return(map[int]int{}, map[int]int{}, nil)
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
			gotFsc, gotAct, err := c.GetFSCAndACTThresholdsPerDSP()
			if (err != nil) != tt.wantErr {
				t.Errorf("cache.GetFSCAndACTThresholdsPerDSP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantWrap != "" && err != nil {
				assert.Contains(t, err.Error(), tt.wantWrap, tt.name)
			}
			assert.Equal(t, tt.wantFsc, gotFsc, tt.name)
			assert.Equal(t, tt.wantAct, gotAct, tt.name)
		})
	}
}
