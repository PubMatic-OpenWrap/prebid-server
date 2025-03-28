package gocache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

const (
	testPubID     = 5890
	testVersionID = 1
	testProfileID = 123
	testAdapterID = 1
	testPartnerID = 1
	testSlotName  = "adunit@300x250"
	testTimeout   = 200
	testHashValue = "2aa34b52a9e941c1594af7565e599c8d"
)

func Test_cache_populateCacheWithPubSlotNameHash(t *testing.T) {
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
		pubid int
	}
	type want struct {
		publisherSlotNameHashMap interface{}
		err                      error
		foundCacheKey            bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   want
	}{
		{
			name: "returned_error_from_DB",
			args: args{
				pubid: 5890,
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherSlotNameHash(5890).Return(nil, fmt.Errorf("Error from the DB"))
			},
			want: want{
				publisherSlotNameHashMap: nil,
				err:                      fmt.Errorf("Error from the DB"),
				foundCacheKey:            false,
			},
		},
		{
			name: "returned_non_nil_PublisherSlotNameHash_from_DB",
			args: args{
				pubid: 5890,
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherSlotNameHash(5890).Return(map[string]string{
					testSlotName: testHashValue,
				}, nil)
			},
			want: want{
				err: nil,
				publisherSlotNameHashMap: map[string]string{
					testSlotName: testHashValue,
				},
				foundCacheKey: true,
			},
		},
		{
			name: "returned_nil_PublisherSlotNameHash_from_DB",
			args: args{
				pubid: 5890,
			},
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
			},
			setup: func() {
				mockDatabase.EXPECT().GetPublisherSlotNameHash(5890).Return(nil, nil)
			},
			want: want{
				publisherSlotNameHashMap: map[string]string(nil),
				err:                      nil,
				foundCacheKey:            true,
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
			err := c.populateCacheWithPubSlotNameHash(tt.args.pubid)
			assert.Equal(t, tt.want.err, err)
			cacheKey := key(PubSlotNameHash, tt.args.pubid)
			publisherSlotNameHashMap, found := c.cache.Get(cacheKey)
			assert.Equal(t, tt.want.foundCacheKey, found)
			assert.Equal(t, tt.want.publisherSlotNameHashMap, publisherSlotNameHashMap)
		})
	}
}

func Test_cache_populateCacheWithWrapperSlotMappings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)

	newCache := gocache.New(10, 10)

	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.Cache
		db    database.Database
	}
	type args struct {
		pubid            int
		partnerConfigMap map[int]map[string]string
		profileId        int
		displayVersion   int
	}
	type want struct {
		partnerSlotMapping interface{}
		err                error
		foundCacheKey      bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   want
	}{
		{
			name: "Error from the DB",
			fields: fields{
				cache: newCache,
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
				db: mockDatabase,
			},
			args: args{
				pubid:            58901,
				partnerConfigMap: formTestPartnerConfig(),
				profileId:        testProfileID,
				displayVersion:   testVersionID,
			},
			setup: func() {
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(nil, fmt.Errorf("Error from the DB"))
			},
			want: want{
				partnerSlotMapping: nil,
				err:                fmt.Errorf("Error from the DB"),
				foundCacheKey:      false,
			},
		},
		{
			name: "empty_mappings",
			fields: fields{
				cache: newCache,
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
				db: mockDatabase,
			},
			args: args{
				pubid:            58902,
				partnerConfigMap: formTestPartnerConfig(),
				profileId:        testProfileID,
				displayVersion:   testVersionID,
			},
			setup: func() {
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(nil, nil)
			},
			want: want{
				partnerSlotMapping: map[string]models.SlotMapping{},
				err:                nil,
				foundCacheKey:      true,
			},
		},
		{
			name: "valid_mappings",
			fields: fields{
				cache: newCache,
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
				db: mockDatabase,
			},
			args: args{
				pubid:            58903,
				partnerConfigMap: formTestPartnerConfig(),
				profileId:        testProfileID,
				displayVersion:   testVersionID,
			},
			setup: func() {
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(map[int][]models.SlotMapping{
					1: {
						{
							PartnerId:   testPartnerID,
							AdapterId:   testAdapterID,
							VersionId:   testVersionID,
							SlotName:    testSlotName,
							MappingJson: "{\"adtag\":\"1405192\",\"site\":\"47124\",\"video\":{\"skippable\":\"TRUE\"}}",
						},
					},
				}, nil)
			},
			want: want{
				partnerSlotMapping: map[string]models.SlotMapping{
					"adunit@300x250": {
						PartnerId:   testPartnerID,
						AdapterId:   testAdapterID,
						VersionId:   testVersionID,
						SlotName:    testSlotName,
						MappingJson: "{\"adtag\":\"1405192\",\"site\":\"47124\",\"video\":{\"skippable\":\"TRUE\"}}",
						SlotMappings: map[string]interface{}{
							"adtag": "1405192",
							"site":  "47124",
							"video": map[string]interface{}{
								"skippable": "TRUE",
							},
							"owSlotName": "adunit@300x250",
						},
						Hash:    "",
						OrderID: 0,
					},
				},
				err:           nil,
				foundCacheKey: true,
			},
		},
		{
			name: "HashValues",
			fields: fields{
				cache: newCache,
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
				db: mockDatabase,
			},
			args: args{
				pubid:            58904,
				partnerConfigMap: formTestPartnerConfig(),
				profileId:        testProfileID,
				displayVersion:   testVersionID,
			},
			setup: func() {
				cacheKey := key(PubSlotNameHash, 58904)
				newCache.Set(cacheKey, map[string]string{testSlotName: testHashValue}, time.Duration(1)*time.Second)
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(map[int][]models.SlotMapping{
					1: {
						{
							PartnerId:   testPartnerID,
							AdapterId:   testAdapterID,
							VersionId:   testVersionID,
							SlotName:    testSlotName,
							MappingJson: "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						},
					},
				}, nil)
			},
			want: want{
				err:           nil,
				foundCacheKey: true,
				partnerSlotMapping: map[string]models.SlotMapping{
					"adunit@300x250": {
						PartnerId:   testPartnerID,
						AdapterId:   testAdapterID,
						VersionId:   testVersionID,
						SlotName:    testSlotName,
						MappingJson: "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: map[string]interface{}{
							"adtag":      "1405192",
							"site":       "47124",
							"owSlotName": "adunit@300x250",
						},
						Hash:    testHashValue,
						OrderID: 0,
					},
				},
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
			err := c.populateCacheWithWrapperSlotMappings(tt.args.pubid, tt.args.partnerConfigMap, tt.args.profileId, tt.args.displayVersion)
			assert.Equal(t, tt.want.err, err)

			cacheKey := key(PUB_SLOT_INFO, tt.args.pubid, tt.args.profileId, tt.args.displayVersion, testPartnerID)
			partnerSlotMapping, found := c.cache.Get(cacheKey)
			assert.Equal(t, tt.want.foundCacheKey, found)
			assert.Equal(t, tt.want.partnerSlotMapping, partnerSlotMapping)

		})
	}
}

func Test_cache_GetMappingsFromCacheV25(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)

	newCache := gocache.New(10, 10)

	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.Cache
		db    database.Database
	}
	type args struct {
		rctx      models.RequestCtx
		partnerID int
	}
	type want struct {
		mappings map[string]models.SlotMapping
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		setup  func()
	}{
		{
			name: "non_nil_partnerConf_map",
			fields: fields{
				cache: newCache,
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				rctx: models.RequestCtx{
					PubID:     testPubID,
					ProfileID: testProfileID,
					DisplayID: 1,
				},
				partnerID: testAdapterID,
			},
			setup: func() {
				cacheKey := key(PUB_SLOT_INFO, testPubID, testProfileID, testVersionID, testAdapterID)
				newCache.Set(cacheKey, map[string]models.SlotMapping{
					"adunit@300x250": {
						PartnerId:   10,
						AdapterId:   1,
						VersionId:   1,
						SlotName:    "adunit@300x250",
						MappingJson: "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: map[string]interface{}{
							"adtag":      "1405192",
							"site":       "47124",
							"owSlotName": "adunit@300x250",
						},
						Hash:    "",
						OrderID: 0,
					},
				}, time.Duration(1)*time.Second)
			},
			want: want{
				mappings: map[string]models.SlotMapping{
					"adunit@300x250": {
						PartnerId:   10,
						AdapterId:   1,
						VersionId:   1,
						SlotName:    "adunit@300x250",
						MappingJson: "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: map[string]interface{}{
							"adtag":      "1405192",
							"site":       "47124",
							"owSlotName": "adunit@300x250",
						},
						Hash:    "",
						OrderID: 0,
					},
				},
			},
		},
		{
			name: "nil_partnerConf_map",
			fields: fields{
				cache: newCache,
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				rctx: models.RequestCtx{
					PubID:     testPubID,
					ProfileID: testProfileID,
					DisplayID: 2,
				},
				partnerID: 1,
			},
			want: want{
				mappings: nil,
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
			got := c.GetMappingsFromCacheV25(tt.args.rctx, tt.args.partnerID)
			assert.Equal(t, tt.want.mappings, got)
		})
	}
}

func Test_cache_GetSlotToHashValueMapFromCacheV25(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)

	newCache := gocache.New(10, 10)

	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.Cache
		db    database.Database
	}
	type args struct {
		rctx      models.RequestCtx
		partnerID int
	}
	type want struct {
		mappinInfo models.SlotMappingInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		setup  func()
	}{
		{
			name: "non_empty_SlotMappingInfo",
			fields: fields{
				cache: newCache,
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				rctx: models.RequestCtx{
					PubID:     testPubID,
					ProfileID: testProfileID,
					DisplayID: testVersionID,
				},
				partnerID: 1,
			},
			setup: func() {
				cacheKey := key(PubSlotHashInfo, testPubID, testProfileID, testVersionID, testAdapterID)
				newCache.Set(cacheKey, models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@300x250"},
					HashValueMap: map[string]string{
						"adunit@300x250": "2aa34b52a9e941c1594af7565e599c8d",
					},
				}, time.Duration(1)*time.Second)
			},
			want: want{
				mappinInfo: models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@300x250"},
					HashValueMap: map[string]string{
						"adunit@300x250": "2aa34b52a9e941c1594af7565e599c8d",
					},
				},
			},
		},
		{
			name: "empty_SlotMappingInfo",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
				},
				partnerID: 1,
			},
			want: want{
				mappinInfo: models.SlotMappingInfo{},
			},
			setup: func() {},
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
			got := c.GetSlotToHashValueMapFromCacheV25(tt.args.rctx, tt.args.partnerID)
			assert.Equal(t, tt.want.mappinInfo, got)
		})
	}
}

func formTestPartnerConfig() map[int]map[string]string {

	partnerConfigMap := make(map[int]map[string]string)

	partnerConfigMap[testAdapterID] = map[string]string{
		models.PARTNER_ID:          "1",
		models.PREBID_PARTNER_NAME: "pubmatic",
		models.SERVER_SIDE_FLAG:    "1",
		models.LEVEL:               "multi",
		models.KEY_GEN_PATTERN:     "_AU_@_W_x_H",
		models.TIMEOUT:             "220",
		models.BidderCode:          "pubmatic",
	}

	return partnerConfigMap
}
