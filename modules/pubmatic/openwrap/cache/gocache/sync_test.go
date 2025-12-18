package gocache

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database/mock"
)

func Test_cache_LockAndLoad(t *testing.T) {
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
		key    string
		dbFunc func() error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		setup   func()
	}{
		{
			name: "test",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
				db: mockDatabase,
			},
			args: args{
				key: "58901231",
			},
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
				db:    tt.fields.db,
			}
			tt.args.dbFunc = func() error {
				c.cache.Set("test", "test", time.Duration(100*time.Second))
				return nil
			}
			var wg sync.WaitGroup
			// var err error
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					c.LockAndLoad(tt.args.key, tt.args.dbFunc)
					wg.Done()
				}()
			}
			wg.Wait()

			obj, found := c.Get("test")
			if !found {
				t.Error("Parner Config not added in cache")
				return
			}

			value := obj.(string)
			if value != "test" {
				t.Error("Parner Config not added in map")
			}
		})
	}
}
