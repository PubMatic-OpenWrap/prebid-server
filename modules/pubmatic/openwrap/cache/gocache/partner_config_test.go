package gocache

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_cache_GetPartnerConfigMap(t *testing.T) {
	var queryFailCase bool
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type fields struct {
		Map   sync.Map
		cache *gocache.Cache
		cfg   config.Cache
	}
	type args struct {
		pubid          int
		profileid      int
		displayversion int
		endpoint       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[int]map[string]string
		wantErr bool
		setup   func() (*mock_database.MockDatabase, *mock.MockMetricsEngine)
	}{
		{
			name: "get_partnerConfig_map",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubid:          testPubID,
				profileid:      testProfileID,
				displayversion: testVersionID,
				endpoint:       models.EndpointV25,
			},
			setup: func() (*mock_database.MockDatabase, *mock.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock.NewMockMetricsEngine(ctrl)
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(formTestPartnerConfig(), nil)
				mockDatabase.EXPECT().GetPublisherSlotNameHash(testPubID).Return(map[string]string{"adunit@728x90": "2aa34b52a9e941c1594af7565e599c8d"}, nil)
				mockDatabase.EXPECT().GetPublisherVASTTags(testPubID).Return(nil, nil)
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(nil, nil)
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
				mockEngine.EXPECT().RecordGetProfileDataTime(models.EndpointV25, "123", gomock.Any()).Return().Times(1)
				return mockDatabase, mockEngine
			},
		},
		{
			name: "db_queries_failed_getting_partnerConfig_map",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubid:          1234,
				profileid:      testProfileID,
				displayversion: 5,
				endpoint:       models.EndpointV25,
			},
			setup: func() (*mock_database.MockDatabase, *mock.MockMetricsEngine) {
				queryFailCase = true
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock.NewMockMetricsEngine(ctrl)
				// Only populateSlotNameHash populates empty map[string]string{} in cache, others expects multiple db calls
				mockDatabase.EXPECT().GetActivePartnerConfigurations(1234, testProfileID, 5).Return(nil, fmt.Errorf("Error from the DB")).AnyTimes()
				mockDatabase.EXPECT().GetPublisherSlotNameHash(1234).Return(nil, fmt.Errorf("Error from the DB"))
				mockDatabase.EXPECT().GetPublisherVASTTags(1234).Return(nil, fmt.Errorf("Error from the DB")).AnyTimes()
				mockEngine.EXPECT().RecordGetProfileDataTime(models.EndpointV25, "123", gomock.Any()).Return().AnyTimes()
				mockEngine.EXPECT().RecordDBQueryFailure(models.SlotNameHash, "1234", "123").Return().Times(1)
				mockEngine.EXPECT().RecordDBQueryFailure(models.PartnerConfigQuery, "1234", "123").Return().AnyTimes()
				mockEngine.EXPECT().RecordDBQueryFailure(models.PublisherVASTTagsQuery, "1234", "123").Return().AnyTimes()
				return mockDatabase, mockEngine
			},
		},
		{
			name: "db_queries_failed_getting_adunitconfig_and_wrapper_slotmappings",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
			},
			args: args{
				pubid:          5234,
				profileid:      testProfileID,
				displayversion: 0,
				endpoint:       models.EndpointAMP,
			},
			setup: func() (*mock_database.MockDatabase, *mock.MockMetricsEngine) {
				queryFailCase = true
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock.NewMockMetricsEngine(ctrl)
				mockDatabase.EXPECT().GetActivePartnerConfigurations(5234, testProfileID, 0).Return(formTestPartnerConfig(), nil)
				mockDatabase.EXPECT().GetPublisherSlotNameHash(5234).Return(map[string]string{"adunit@728x90": "2aa34b52a9e941c1594af7565e599c8d"}, nil)
				mockDatabase.EXPECT().GetPublisherVASTTags(5234).Return(nil, nil)
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, 0).Return(nil, errors.New("unmarshal error adunitconfig"))
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, 0).Return(nil, fmt.Errorf("Error from the DB"))
				mockEngine.EXPECT().RecordGetProfileDataTime(models.EndpointAMP, "123", gomock.Any()).Return().Times(1)
				mockEngine.EXPECT().RecordDBQueryFailure(models.AdUnitFailUnmarshal, "5234", "123").Return().Times(1)
				mockEngine.EXPECT().RecordDBQueryFailure(models.WrapperLiveVersionSlotMappings, "5234", "123").Return().Times(1)
				return mockDatabase, mockEngine
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryFailCase = false
			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
			}
			c.db, c.metricEngine = tt.setup()

			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					c.GetPartnerConfigMap(tt.args.pubid, tt.args.profileid, tt.args.displayversion, tt.args.endpoint)
					wg.Done()
				}()
			}
			wg.Wait()
			cacheKey := key(PUB_HB_PARTNER, testPubID, testProfileID, testVersionID)
			obj, found := c.Get(cacheKey)
			if !found && !queryFailCase {
				t.Error("Parner Config not added in cache")
				return
			}

			var partnerConfigMap map[int]map[string]string
			if obj != nil {
				partnerConfigMap = obj.(map[int]map[string]string)
			}
			if _, found := partnerConfigMap[testAdapterID]; !found && !queryFailCase {
				t.Error("Parner Config not added in map")
			}
		})
	}
}

func Test_cache_getActivePartnerConfigAndPopulateWrapperMappings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock.NewMockMetricsEngine(ctrl)

	type fields struct {
		Map   sync.Map
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
		cacheEntry       bool
		err              error
		partnerConfigMap map[int]map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		setup  func()
	}{
		{
			name: "error_returning_Active_partner_configuration_from_DB",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
				db: mockDatabase,
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			want: want{
				cacheEntry:       false,
				err:              fmt.Errorf("Error from the DB"),
				partnerConfigMap: nil,
			},
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(nil, fmt.Errorf("Error from the DB"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.PartnerConfigQuery, "5890", "123").Return()
			},
		},
		{
			name: "non_nil_partnerConfigMap_from_DB",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
				db: mockDatabase,
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			want: want{
				cacheEntry: true,
				err:        nil,
				partnerConfigMap: map[int]map[string]string{
					1: {
						"bidderCode":        "pubmatic",
						"kgp":               "_AU_@_W_x_H",
						"level":             "multi",
						"partnerId":         "1",
						"prebidPartnerName": "pubmatic",
						"serverSideEnabled": "1",
						"timeout":           "220",
					},
				},
			},
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(formTestPartnerConfig(), nil)
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(nil, nil)
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
		},
		{
			name: "empty_partnerConfigMap_from_DB",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 100,
				},
				db: mockDatabase,
			},
			args: args{
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			want: want{
				cacheEntry:       false,
				err:              fmt.Errorf("there are no active partners for pubId:%d, profileId:%d, displayVersion:%d", testPubID, testProfileID, testVersionID),
				partnerConfigMap: nil,
			},
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(nil, nil)
			},
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
				metricEngine: mockEngine,
			}
			err := c.getActivePartnerConfigAndPopulateWrapperMappings(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			assert.Equal(t, tt.want.err, err)
			cacheKey := key(PUB_HB_PARTNER, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			partnerConfigMap, found := c.Get(cacheKey)
			if tt.want.cacheEntry {
				assert.True(t, found)
				assert.Equal(t, tt.want.partnerConfigMap, partnerConfigMap)
			} else {
				assert.False(t, found)
				assert.Nil(t, partnerConfigMap)
			}
		})
	}
}
