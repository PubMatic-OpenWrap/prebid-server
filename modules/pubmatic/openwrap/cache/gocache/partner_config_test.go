package gocache

import (
	"fmt"
	"sync"
	"testing"

	mock_database "github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/database/mock"
	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func Test_cache_GetPartnerConfigMap(t *testing.T) {
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
		pubid          int
		profileid      int
		displayversion int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[int]map[string]string
		wantErr bool
		setup   func()
	}{
		{
			name: "test",
			fields: fields{
				cache: gocache.New(100, 100),
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
					VASTTagCacheExpiry: 1000,
				},
				db: mockDatabase,
			},
			args: args{
				pubid:          testPubID,
				profileid:      testProfileID,
				displayversion: testVersionID,
			},
			setup: func() {
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
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					c.GetPartnerConfigMap(tt.args.pubid, tt.args.profileid, tt.args.displayversion)
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
			//TODO: Add validation to check prometheus stat called only once
		})
	}
}

func Test_cache_getActivePartnerConfigAndPopulateWrapperMappings(t *testing.T) {
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
		pubID          int
		profileID      int
		displayVersion int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		setup   func()
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
			wantErr: true,
			setup: func() {
				mockDatabase.EXPECT().GetActivePartnerConfigurations(testPubID, testProfileID, testVersionID).Return(nil, fmt.Errorf("Error from the DB"))
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
			wantErr: false,
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
			wantErr: true,
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
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
				db:    tt.fields.db,
			}
			err := c.getActivePartnerConfigAndPopulateWrapperMappings(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)

			if tt.wantErr {
				if err == nil {
					t.Error("Error should not be nil")
				}
				return
			}

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
