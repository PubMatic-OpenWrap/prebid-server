package gocache

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database/mock"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adpodconfig"
	"github.com/stretchr/testify/assert"
)

func TestCachePopulateCacheWithAdpodConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)

	type fields struct {
		cache *gocache.Cache
		cfg   config.Cache
		db    database.Database
	}
	type args struct {
		pubID          int
		profileID      int
		displayVersion int
	}
	type want struct {
		err         error
		adpodConfig *adpodconfig.AdpodConfig
		cacheEntry  bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   want
	}{
		{
			name: "Adpod config is added in cache",
			fields: fields{
				cache: gocache.New(10, 10),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func() {
				mockDatabase.EXPECT().GetAdpodConfig(testPubID, testProfileID, testVersionID).Return(&adpodconfig.AdpodConfig{
					Dynamic: []adpodconfig.Dynamic{
						{
							MinDuration: 10,
							MaxDuration: 20,
							PodDur:      60,
							MaxSeq:      3,
						},
					},
				}, nil)
			},
			want: want{
				err:        nil,
				cacheEntry: true,
				adpodConfig: &adpodconfig.AdpodConfig{
					Dynamic: []adpodconfig.Dynamic{
						{
							MinDuration: 10,
							MaxDuration: 20,
							PodDur:      60,
							MaxSeq:      3,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
				db:    tt.fields.db,
			}
			err := c.populateCacheWithAdpodConfig(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			assert.Equal(t, tt.want.err, err)
			cacheKey := key(PubAdpodConfig, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			adpodconfig, found := c.Get(cacheKey)
			if tt.want.cacheEntry {
				assert.True(t, found)
				assert.Equal(t, tt.want.adpodConfig, adpodconfig)
			} else {
				assert.False(t, found)
				assert.Nil(t, adpodconfig)
			}
		})
	}
}

func TestCacheGetAdpodConfigs(t *testing.T) {
	type fields struct {
		cache *gocache.Cache
		cfg   config.Cache
	}
	type args struct {
		pubID          int
		profileID      int
		displayVersion int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(ctrl *gomock.Controller, cache *gocache.Cache) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine)
		want    *adpodconfig.AdpodConfig
		wantErr bool
	}{
		{
			name: "Adpod config retrived from cache",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func(ctrl *gomock.Controller, cache *gocache.Cache) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockDatabase.EXPECT().GetAdpodConfig(testPubID, testProfileID, testVersionID).Return(&adpodconfig.AdpodConfig{
					Dynamic: []adpodconfig.Dynamic{
						{
							MinDuration: 10,
							MaxDuration: 20,
							PodDur:      60,
							MaxSeq:      3,
						},
					},
				}, nil)
				return mockDatabase, mockEngine
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						MinDuration: 10,
						MaxDuration: 20,
						PodDur:      60,
						MaxSeq:      3,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Adpod config does not configured through UI",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func(ctrl *gomock.Controller, cache *gocache.Cache) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockDatabase.EXPECT().GetAdpodConfig(testPubID, testProfileID, testVersionID).Return(nil, nil)
				return mockDatabase, mockEngine
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Adpod config already present in cache",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func(ctrl *gomock.Controller, cache *gocache.Cache) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				adpod := adpodconfig.AdpodConfig{
					Dynamic: []adpodconfig.Dynamic{
						{
							MinDuration: 30,
							MaxDuration: 60,
							PodDur:      120,
							MaxSeq:      4,
						},
					},
				}
				cache.Set(key(PubAdpodConfig, testPubID, testProfileID, testVersionID), &adpod, time.Duration(10)*time.Millisecond)
				return mockDatabase, mockEngine
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						MinDuration: 30,
						MaxDuration: 60,
						PodDur:      120,
						MaxSeq:      4,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Adpod config already present in cache with nil value",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func(ctrl *gomock.Controller, cache *gocache.Cache) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				cache.Set(key(PubAdpodConfig, testPubID, testProfileID, testVersionID), nil, time.Duration(10)*time.Millisecond)
				return mockDatabase, mockEngine
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Adpod config already present in cache with different type",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func(ctrl *gomock.Controller, cache *gocache.Cache) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				adpod := map[string]interface{}{
					"dynamic": map[string]interface{}{
						"minduration": 10,
					},
				}
				cache.Set(key(PubAdpodConfig, testPubID, testProfileID, testVersionID), adpod, time.Duration(10)*time.Millisecond)
				return mockDatabase, mockEngine
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
			}
			c.db, c.metricEngine = tt.setup(ctrl, c.cache)

			got, err := c.GetAdpodConfig(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("cache.GetAdpodConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cache.GetAdpodConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
