package gocache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_cache_populatePublisherVASTTags(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatabase := mock_database.NewMockDatabase(ctrl)

	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.DBCache
		db    database.Database
	}
	type args struct {
		pubid int
	}
	type want struct {
		cacheEntry        bool
		err               error
		PublisherVASTTags models.PublisherVASTTags
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   want
	}{
		{
			name: "error_in_returning_PublisherVASTTags_from_DB",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					VASTTagCacheExpiry: 100000,
				},
			},
			args: args{
				pubid: 5890,
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherVASTTags(5890).Return(nil, fmt.Errorf("Error in returning PublisherVASTTags from the DB"))
			},
			want: want{
				cacheEntry:        false,
				err:               fmt.Errorf("Error in returning PublisherVASTTags from the DB"),
				PublisherVASTTags: nil,
			},
		},
		{
			name: "successfully_got_PublisherVASTTags_from_DB_and_not_nil",
			fields: fields{
				Map:   sync.Map{},
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					VASTTagCacheExpiry: 100000,
				},
			},
			args: args{
				pubid: 5890,
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherVASTTags(5890).Return(models.PublisherVASTTags{
					101: {ID: 101, PartnerID: 501, URL: "vast_tag_url_1", Duration: 15, Price: 2.0},
					102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
					103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
				}, nil)
			},
			want: want{
				cacheEntry: true,
				err:        nil,
				PublisherVASTTags: map[int]*models.VASTTag{
					101: {ID: 101, PartnerID: 501, URL: "vast_tag_url_1", Duration: 15, Price: 2.0},
					102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
					103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
				},
			},
		},
		{
			name: "successfully_got_PublisherVASTTags_from_DB_but_it_is_nil",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					VASTTagCacheExpiry: 100000,
				},
			},
			args: args{
				pubid: 5890,
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherVASTTags(5890).Return(nil, nil)
			},
			want: want{
				cacheEntry:        true,
				err:               nil,
				PublisherVASTTags: models.PublisherVASTTags{},
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
			err := c.populatePublisherVASTTags(tt.args.pubid)
			assert.Equal(t, tt.want.err, err)

			cacheKey := key(PubVASTTags, tt.args.pubid)
			PublisherVASTTags, found := c.cache.Get(cacheKey)
			if tt.want.cacheEntry {
				assert.True(t, found)
				assert.Equal(t, tt.want.PublisherVASTTags, PublisherVASTTags)
			} else {
				assert.False(t, found)
				assert.Nil(t, PublisherVASTTags)
			}
		})
	}
}

func Test_cache_GetPublisherVASTTagsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatabase := mock_database.NewMockDatabase(ctrl)
	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.DBCache
		db    database.Database
	}
	type args struct {
		pubID int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   models.PublisherVASTTags
		setup  func()
	}{
		{
			name: "Vast_Tags_found_in_cache_for_cache_key",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					VASTTagCacheExpiry: 100000,
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[int]*models.VASTTag{
				101: {ID: 101, PartnerID: 501, URL: "vast_tag_url_1", Duration: 15, Price: 2.0},
				102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
				103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherVASTTags(5890).Return(models.PublisherVASTTags{
					101: {ID: 101, PartnerID: 501, URL: "vast_tag_url_1", Duration: 15, Price: 2.0},
					102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
					103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
				}, nil)
			},
		},
		{
			name: "Vast_Tags_not_found_in_cache_for_cache_key",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.DBCache{
					VASTTagCacheExpiry: 100000,
				},
			},
			args: args{
				pubID: 5890,
			},
			want: models.PublisherVASTTags{},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherVASTTags(5890).Return(nil, fmt.Errorf("Error in returning PublisherVASTTags from the DB"))
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
			c.populatePublisherVASTTags(tt.args.pubID)
			cacheKey := key(PubVASTTags, tt.args)
			got := c.GetPublisherVASTTagsFromCache(tt.args.pubID)
			assert.Equal(t, tt.want, got, "Vast tags for cacheKey= %v \n Expected= %v, but got= %v", cacheKey, got, tt.want)
		})
	}
}
