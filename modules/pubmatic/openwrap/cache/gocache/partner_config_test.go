package gocache

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/database/mock"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/stretchr/testify/assert"
)

func TestCacheGetPartnerConfigMap(t *testing.T) {
	type fields struct {
		Map   sync.Map
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
		want    map[int]map[string]string
		wantErr bool
		setup   func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine)
	}{
		{
			name: "get_valid_partnerConfig_map",
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
			setup: func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(formTestPartnerConfig(), nil)
				mockDatabase.EXPECT().GetPublisherSlotNameHash(testPubID).Return(map[string]string{"adunit@728x90": "2aa34b52a9e941c1594af7565e599c8d"}, nil)
				mockDatabase.EXPECT().GetPublisherVASTTags(testPubID).Return(nil, nil)
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(&adunitconfig.AdUnitConfig{
					Config: map[string]*adunitconfig.AdConfig{
						"default": {
							BidderFilter: &adunitconfig.BidderFilter{
								Filters: []adunitconfig.Filter{
									{
										Bidders: []string{
											"pubmatic",
										},
										BiddingConditions: `{"in":[{"var":"country"},["IND"]]}`,
									},
								},
							},
						},
					},
				}, nil)
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
				mockEngine.EXPECT().RecordGetProfileDataTime(gomock.Any()).Return().Times(1)
				return mockDatabase, mockEngine
			},
			wantErr: false,
			want: map[int]map[string]string{
				1: {
					"partnerId":         "1",
					"prebidPartnerName": "pubmatic",
					"serverSideEnabled": "1",
					"level":             "multi",
					"kgp":               "_AU_@_W_x_H",
					"timeout":           "220",
					"bidderCode":        "pubmatic",
					"bidderFilters":     `{"in":[{"var":"country"},["IND"]]}`,
				},
			},
		},
		// {
		// 	name: "db_queries_failed_getting_partnerConfig_map",
		// 	fields: fields{
		// 		cache: gocache.New(100, 100),
		// 		cfg: config.Cache{
		// 			CacheDefaultExpiry: 1000,
		// 			VASTTagCacheExpiry: 1000,
		// 		},
		// 	},
		// 	args: args{
		// 		pubID:          testPubID,
		// 		profileID:      testProfileID,
		// 		displayVersion: testVersionID,
		// 	},
		// 	setup: func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
		// 		mockDatabase := mock_database.NewMockDatabase(ctrl)
		// 		mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
		// 		mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(nil, fmt.Errorf("Error from the DB"))
		// 		mockDatabase.EXPECT().GetPublisherSlotNameHash(testPubID).Return(nil, fmt.Errorf("Error from the DB"))
		// 		mockDatabase.EXPECT().GetPublisherVASTTags(testPubID).Return(nil, fmt.Errorf("Error from the DB"))
		// 		mockEngine.EXPECT().RecordGetProfileDataTime(gomock.Any()).Return()
		// 		mockEngine.EXPECT().RecordDBQueryFailure(models.SlotNameHash, "5890", "123").Return()
		// 		mockEngine.EXPECT().RecordDBQueryFailure(models.PartnerConfigQuery, "5890", "123").Return()
		// 		mockEngine.EXPECT().RecordDBQueryFailure(models.PublisherVASTTagsQuery, "5890", "123").Return()
		// 		return mockDatabase, mockEngine
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "error_in_adunitconfig_unmarshal",
		// 	fields: fields{
		// 		cache: gocache.New(100, 100),
		// 		cfg: config.Cache{
		// 			CacheDefaultExpiry: 1000,
		// 			VASTTagCacheExpiry: 1000,
		// 		},
		// 	},
		// 	args: args{
		// 		pubID:          testPubID,
		// 		profileID:      testProfileID,
		// 		displayVersion: 0,
		// 	},
		// 	setup: func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
		// 		mockDatabase := mock_database.NewMockDatabase(ctrl)
		// 		mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
		// 		mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, 0).Return(formTestPartnerConfig(), nil)
		// 		mockDatabase.EXPECT().GetPublisherSlotNameHash(testPubID).Return(map[string]string{"adunit@728x90": "2aa34b52a9e941c1594af7565e599c8d"}, nil)
		// 		mockDatabase.EXPECT().GetPublisherVASTTags(testPubID).Return(nil, nil)
		// 		mockDatabase.EXPECT().GetAdunitConfig(testProfileID, 0).Return(nil, adunitconfig.ErrAdUnitUnmarshal)
		// 		mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, 0).Return(nil, nil)
		// 		mockEngine.EXPECT().RecordGetProfileDataTime(gomock.Any()).Return().Times(1)
		// 		mockEngine.EXPECT().RecordDBQueryFailure(models.AdUnitFailUnmarshal, "5890", "123").Return().Times(1)
		// 		return mockDatabase, mockEngine
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "db_queries_failed_getting_adunitconfig",
		// 	fields: fields{
		// 		cache: gocache.New(100, 100),
		// 		cfg: config.Cache{
		// 			CacheDefaultExpiry: 1000,
		// 			VASTTagCacheExpiry: 1000,
		// 		},
		// 	},
		// 	args: args{
		// 		pubID:          testPubID,
		// 		profileID:      testProfileID,
		// 		displayVersion: 0,
		// 	},
		// 	setup: func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
		// 		mockDatabase := mock_database.NewMockDatabase(ctrl)
		// 		mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
		// 		mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, 0).Return(formTestPartnerConfig(), nil)
		// 		mockDatabase.EXPECT().GetPublisherSlotNameHash(testPubID).Return(map[string]string{"adunit@728x90": "2aa34b52a9e941c1594af7565e599c8d"}, nil)
		// 		mockDatabase.EXPECT().GetPublisherVASTTags(testPubID).Return(nil, nil)
		// 		mockDatabase.EXPECT().GetAdunitConfig(testProfileID, 0).Return(nil, errors.New("Failed to connect DB"))
		// 		mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, 0).Return(nil, nil)
		// 		mockEngine.EXPECT().RecordGetProfileDataTime(gomock.Any()).Return().Times(1)
		// 		mockEngine.EXPECT().RecordDBQueryFailure(models.AdunitConfigForLiveVersion, "5890", "123").Return().Times(1)
		// 		return mockDatabase, mockEngine
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "db_queries_failed_getting_wrapper_slotmappings",
		// 	fields: fields{
		// 		cache: gocache.New(100, 100),
		// 		cfg: config.Cache{
		// 			CacheDefaultExpiry: 1000,
		// 			VASTTagCacheExpiry: 1000,
		// 		},
		// 	},
		// 	args: args{
		// 		pubID:          testPubID,
		// 		profileID:      testProfileID,
		// 		displayVersion: 0,
		// 	},
		// 	setup: func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
		// 		mockDatabase := mock_database.NewMockDatabase(ctrl)
		// 		mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
		// 		mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, 0).Return(formTestPartnerConfig(), nil)
		// 		mockDatabase.EXPECT().GetPublisherSlotNameHash(testPubID).Return(map[string]string{"adunit@728x90": "2aa34b52a9e941c1594af7565e599c8d"}, nil)
		// 		mockDatabase.EXPECT().GetPublisherVASTTags(testPubID).Return(nil, nil)
		// 		mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, 0).Return(nil, fmt.Errorf("Error from the DB"))
		// 		mockEngine.EXPECT().RecordGetProfileDataTime(gomock.Any()).Return().Times(1)
		// 		mockEngine.EXPECT().RecordDBQueryFailure(models.WrapperLiveVersionSlotMappings, "5890", "123").Return().Times(1)
		// 		return mockDatabase, mockEngine
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
			}
			c.db, c.metricEngine = tt.setup(ctrl)

			got, err := c.GetPartnerConfigMap(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("cache.GetPartnerConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_cache_GetPartnerConfigMap_LockandLoad(t *testing.T) {
	type fields struct {
		Map   sync.Map
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
		want    map[int]map[string]string
		wantErr bool
		setup   func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine)
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
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: testVersionID,
			},
			setup: func(ctrl *gomock.Controller) (*mock_database.MockDatabase, *mock_metrics.MockMetricsEngine) {
				mockDatabase := mock_database.NewMockDatabase(ctrl)
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
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
				mockEngine.EXPECT().RecordGetProfileDataTime(gomock.Any()).Return().Times(1)
				return mockDatabase, mockEngine
			},
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := &cache{
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
			}
			c.db, c.metricEngine = tt.setup(ctrl)

			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					c.GetPartnerConfigMap(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
					wg.Done()
				}()
			}
			wg.Wait()
			cacheKey := key(PUB_HB_PARTNER, testPubID, testProfileID, testVersionID)
			obj, found := c.Get(cacheKey)
			if !found {
				t.Error("Parner Config not added in cache")
				return
			}

			partnerConfigMap := obj.(map[int]map[string]string)
			if _, found := partnerConfigMap[testAdapterID]; !found {
				t.Error("Parner Config not added in map")
			}
		})
	}
}

func Test_cache_getActivePartnerConfigAndPopulateWrapperMappings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)

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
				cacheEntry: false,
				err:        fmt.Errorf("there are no active partners for pubId:%d, profileId:%d, displayVersion:%d", testPubID, testProfileID, testVersionID),
			},
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(nil, nil)
			},
		},
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
				cacheEntry: false,
				err:        fmt.Errorf("Error from the DB"),
			},
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(nil, fmt.Errorf("Error from the DB"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.PartnerConfigQuery, "5890", "123").Return()
			},
		},
		{
			name: "No partner config in case of error in GetWrapperSlotMappings",
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
				cacheEntry: false,
				err:        fmt.Errorf("Error from the DB"),
			},
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(formTestPartnerConfig(), nil)

				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(nil, errors.New("Error from the DB"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.WrapperSlotMappingsQuery, "5890", "123").Return()
			},
		},
		{
			name: "No partner config in case of error in GetAdunitConfig",
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
				cacheEntry: false,
				err:        fmt.Errorf("Error from the DB"),
			},
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(formTestPartnerConfig(), nil)
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(nil, errors.New("Error from the DB"))
				mockEngine.EXPECT().RecordDBQueryFailure(models.AdunitConfigQuery, "5890", "123").Return()
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
			name: "Partner config in case of empty wrapperSlotMappings",
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
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(nil, nil)
			},
		},
		{
			name: "Partner config in case of empty adunitConfig",
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
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(nil, nil)
			},
		},

		{
			name: "Partner config in case of empty adunitConfig and wrapperSlotMappings",
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
				mockDatabase.EXPECT().GetWrapperSlotMappings(formTestPartnerConfig(), testProfileID, testVersionID).Return(nil, nil)
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
				cache:        tt.fields.cache,
				cfg:          tt.fields.cfg,
				db:           tt.fields.db,
				metricEngine: mockEngine,
			}
			err := c.getActivePartnerConfigAndPopulateWrapperMappings(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			cacheKey := key(PUB_HB_PARTNER, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			partnerConfigMap, found := c.Get(cacheKey)
			if tt.want.cacheEntry {
				assert.True(t, found)
				assert.Equal(t, tt.want.partnerConfigMap, partnerConfigMap)
			} else {
				assert.Equal(t, tt.want.err.Error(), err.Error())
				assert.False(t, found)
				assert.Nil(t, partnerConfigMap)
			}
		})
	}
}
