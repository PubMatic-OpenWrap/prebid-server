package gocache

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/stretchr/testify/assert"
)

func Test_key(t *testing.T) {
	type args struct {
		format string
		v      []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "get_key",
			args: args{
				format: PUB_SLOT_INFO,
				v:      []interface{}{5890, 123, 1, 10},
			},
			want: "pslot_5890_123_1_10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := key(tt.args.format, tt.args.v...); got != tt.want {
				t.Errorf("key() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock.NewMockMetricsEngine(ctrl)

	type args struct {
		goCache       *gocache.Cache
		database      database.Database
		cfg           config.Cache
		metricsEngine metrics.MetricsEngine
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "new_cache_instance",
			args: args{
				goCache:  gocache.New(100, 100),
				database: mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
				metricsEngine: mockEngine,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := New(tt.args.goCache, tt.args.database, tt.args.cfg, tt.args.metricsEngine)
			assert.NotNil(t, cache, "chache object should not be nl")
		})
	}
}

func Test_getSeconds(t *testing.T) {
	type args struct {
		duration int
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "get_to_seconds",
			args: args{
				duration: 10,
			},
			want: time.Duration(10000000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSeconds(tt.args.duration); got != tt.want {
				t.Errorf("getSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cache_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)

	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.Cache
		db    database.Database
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "set_to_cache",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				key:   "test_key",
				value: "test_value",
			},
			want: "test_value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
				db:    tt.fields.db,
			}
			c.Set(tt.args.key, tt.args.value)
			value, found := c.Get(tt.args.key)
			if !found {
				t.Errorf("key should be present")
				return
			}
			if value != tt.want {
				t.Errorf("Expected= %v but got= %v", tt.want, value)
			}

		})
	}
}

func Test_cache_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)

	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.Cache
		db    database.Database
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
		want1  bool
	}{
		{
			name: "get_from_cache",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				key: "test_key",
			},
			want:  "test_value",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
				db:    tt.fields.db,
			}
			c.Set(tt.args.key, "test_value")
			got, got1 := c.Get(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cache.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("cache.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
