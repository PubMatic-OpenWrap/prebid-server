package gocache

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adpodconfig"
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
				mockDatabase.EXPECT().GetAdpodConfigs(testProfileID, testVersionID).Return(&adpodconfig.AdpodConfig{
					Dynamic: []adpodconfig.Dynamic{
						{
							MinDuration: 10,
							MaxDuration: 20,
							PodDur:      60,
							Maxseq:      3,
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
							Maxseq:      3,
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
		request        *openrtb2.BidRequest
		pubID          int
		profileID      int
		displayVersion int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(ctrl *gomock.Controller) *mock_database.MockDatabase
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
				request: &openrtb2.BidRequest{
					Test: 0,
				},
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func(ctrl *gomock.Controller) *mock_database.MockDatabase {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockDatabase.EXPECT().GetAdpodConfigs(testProfileID, testVersionID).Return(&adpodconfig.AdpodConfig{
					Dynamic: []adpodconfig.Dynamic{
						{
							MinDuration: 10,
							MaxDuration: 20,
							PodDur:      60,
							Maxseq:      3,
						},
					},
				}, nil)
				return mockDatabase
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						MinDuration: 10,
						MaxDuration: 20,
						PodDur:      60,
						Maxseq:      3,
					},
				},
			},
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
			c.db = tt.setup(ctrl)

			got, err := c.GetAdpodConfigs(tt.args.request, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
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
