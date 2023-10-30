package gocache

import (
	"errors"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mock"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestGetFSCDisabledPublishers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.DBCache
		db    database.Database
	}
	tests := []struct {
		name    string
		want    map[int]struct{}
		wantErr bool
		setup   func()
		fields  fields
	}{
		{
			name: "Valid Data present in DB, return same",
			want: map[int]struct{}{
				5890: {},
				5891: {},
			},
			setup: func() {
				mockDatabase.EXPECT().GetFSCDisabledPublishers().Return(map[int]struct{}{
					5890: {},
					5891: {},
				}, nil)
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					CacheDefaultExpiry: 1000,
				},
			},
			wantErr: false,
		},
		{
			name: "Error In DB, Set Empty",
			want: map[int]struct{}{},
			setup: func() {
				mockDatabase.EXPECT().GetFSCDisabledPublishers().Return(map[int]struct{}{}, errors.New("QUERY FAILED"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.AllFscDisabledPublishersQuery, "", "").Return()
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					CacheDefaultExpiry: 1000,
				},
			},
			wantErr: true,
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			c := &cache{
				cache:        tt.fields.cache,
				cfg:          tt.fields.cfg,
				db:           tt.fields.db,
				metricEngine: mockEngine,
			}
			got, err := c.GetFSCDisabledPublishers()
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetFSCDisabledPublishers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestGetFSCThresholdPerDSP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.DBCache
		db    database.Database
	}
	tests := []struct {
		name    string
		want    map[int]int
		wantErr bool
		setup   func()
		fields  fields
	}{
		{
			name: "Valid Data present in DB, return same",
			want: map[int]int{
				6: 100,
				5: 45,
			},
			setup: func() {
				mockDatabase.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{
					6: 100,
					5: 45,
				}, nil)
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					CacheDefaultExpiry: 1000,
				},
			},
			wantErr: false,
		},
		{
			name: "Error In DB, Set Empty",
			want: map[int]int{},
			setup: func() {
				mockDatabase.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{}, errors.New("QUERY FAILD"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.AllDspFscPcntQuery, "", "").Return()
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					CacheDefaultExpiry: 1000,
				},
			},
			wantErr: true,
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			c := &cache{
				cache:        tt.fields.cache,
				cfg:          tt.fields.cfg,
				db:           tt.fields.db,
				metricEngine: mockEngine,
			}
			got, err := c.GetFSCThresholdPerDSP()
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetFSCThresholdPerDSP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)

		})
	}
}
