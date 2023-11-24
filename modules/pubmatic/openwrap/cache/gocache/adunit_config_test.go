package gocache

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/database"
	mock_database "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/database/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

var testAdunitConfig = &adunitconfig.AdUnitConfig{
	ConfigPattern: "_AU_",
	Regex:         true,
	Config: map[string]*adunitconfig.AdConfig{
		"default": {
			Floors: &openrtb_ext.PriceFloorRules{
				FloorMin: 15,
				Data: &openrtb_ext.PriceFloorData{
					ModelGroups: []openrtb_ext.PriceFloorModelGroup{
						{
							Schema: openrtb_ext.PriceFloorSchema{
								Delimiter: "|",
								Fields:    strings.Fields("mediaType size domain"),
							},
							Default: 5,
							Values: map[string]float64{
								"banner|300x600|*":               4,
								"banner|300x250|www.website.com": 1,
								"banner|728x90|www.website.com":  5,
								"*|728x90|www.website.com":       13,
							},
							Currency:     "USD",
							ModelWeight:  ptrutil.ToPtr(40),
							ModelVersion: "model 1 from adunit config slot level",
						},
					},
					Currency: "USD",
				},
				Enforcement: &openrtb_ext.PriceFloorEnforcement{
					EnforcePBS:  ptrutil.ToPtr(true),
					EnforceRate: 100,
					EnforceJS:   ptrutil.ToPtr(true),
				},
				Enabled: ptrutil.ToPtr(true),
			},
			Video: &adunitconfig.Video{
				Enabled: ptrutil.ToPtr(true),
				Config: &adunitconfig.VideoConfig{
					ConnectionType: []int{2},
					Video: openrtb2.Video{
						MinDuration: 10,
						MaxDuration: 50,
						BAttr: []adcom1.CreativeAttribute{
							6,
							7,
						},
						Skip:      ptrutil.ToPtr[int8](1),
						SkipMin:   10,
						SkipAfter: 15,
					},
				},
			},
			UniversalPixel: []adunitconfig.UniversalPixel{
				{
					Id:        123,
					Pixel:     "pixle",
					PixelType: "js",
					Pos:       "above",
					MediaType: "banner",
					Partners: []string{
						"pubmatic",
						"appnexus",
					},
				},
			},
		},
		"Div1": {
			Video: &adunitconfig.Video{
				Enabled: ptrutil.ToPtr(true),
				Config: &adunitconfig.VideoConfig{
					ConnectionType: []int{0, 1, 2, 4},
					Video: openrtb2.Video{
						MinDuration: 10,
						MaxDuration: 50,
						BAttr: []adcom1.CreativeAttribute{
							6,
							7,
						},
						Skip:      ptrutil.ToPtr[int8](1),
						SkipMin:   10,
						SkipAfter: 15,
					},
				},
			},
			Banner: &adunitconfig.Banner{
				Enabled: ptrutil.ToPtr(true),
				Config: &adunitconfig.BannerConfig{
					Banner: openrtb2.Banner{
						Format: []openrtb2.Format{
							{
								W: 200,
								H: 300,
							},
							{
								W: 500,
								H: 800,
							},
						},
					},
				},
			},
		},
		"Div2": {
			Video: &adunitconfig.Video{
				Enabled: ptrutil.ToPtr(true),
				Config: &adunitconfig.VideoConfig{
					ConnectionType: []int{0, 1, 2, 4},
					Video: openrtb2.Video{
						MinDuration: 10,
						MaxDuration: 50,
						BAttr: []adcom1.CreativeAttribute{
							6,
							7,
						},
						Skip:      ptrutil.ToPtr[int8](1),
						SkipMin:   10,
						SkipAfter: 15,
					},
				},
			},
		},
	},
}

func Test_cache_populateCacheWithAdunitConfig(t *testing.T) {
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
	type want struct {
		err          error
		adunitConfig *adunitconfig.AdUnitConfig
		cacheEntry   bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   want
	}{
		{
			name: "error_in_returning_adunitconfig_from_the_DB",
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
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(nil, fmt.Errorf("Invalid json"))
			},
			want: want{
				err:          fmt.Errorf("Invalid json"),
				adunitConfig: nil,
				cacheEntry:   false,
			},
		},
		{
			name: "valid_adunit_config",
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
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(testAdunitConfig, nil)
			},
			want: want{
				err:          nil,
				adunitConfig: testAdunitConfig,
				cacheEntry:   true,
			},
		},
		{
			name: "returned_nil_adunitconfig_from_the_DB",
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
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(nil, nil)
			},
			want: want{
				err:          nil,
				adunitConfig: nil,
				cacheEntry:   true,
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
			err := c.populateCacheWithAdunitConfig(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			assert.Equal(t, tt.want.err, err)
			cacheKey := key(PubAdunitConfig, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			adunitconfig, found := c.Get(cacheKey)
			if tt.want.cacheEntry {
				assert.True(t, found)
				assert.Equal(t, tt.want.adunitConfig, adunitconfig)
			} else {
				assert.False(t, found)
				assert.Nil(t, adunitconfig)
			}
		})
	}
}

func Test_cache_GetAdunitConfigFromCache(t *testing.T) {
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
		request        *openrtb2.BidRequest
		pubID          int
		profileID      int
		displayVersion int
	}
	type want struct {
		adunitConfig *adunitconfig.AdUnitConfig
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		setup  func()
	}{
		{
			name: "test_request",
			fields: fields{
				db: mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				request: &openrtb2.BidRequest{
					Test: 2,
				},
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: 1,
			},
			setup: func() {},
			want: want{
				adunitConfig: nil,
			},
		},
		{
			name: "successfully_get_value_from_cache",
			fields: fields{
				db: mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				request: &openrtb2.BidRequest{
					Test: 1,
				},
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: 2,
			},
			setup: func() {
				cacheKey := key(PubAdunitConfig, testPubID, testProfileID, 2)
				newCache.Set(cacheKey, testAdunitConfig, time.Duration(1)*time.Second)
			},
			want: want{
				adunitConfig: testAdunitConfig,
			},
		},
		{
			name: "got_empty_adunitconfig_from_cache",
			fields: fields{
				db: mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				request: &openrtb2.BidRequest{
					Test: 1,
				},
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: 3,
			},
			setup: func() {
				cacheKey := key(PubAdunitConfig, testPubID, testProfileID, 3)
				newCache.Set(cacheKey, &adunitconfig.AdUnitConfig{}, time.Duration(1*time.Second))
			},
			want: want{
				adunitConfig: &adunitconfig.AdUnitConfig{},
			},
		},
		{
			name: "cache_key_not_present_in_cache",
			fields: fields{
				cache: gocache.New(100, 100),
				db:    mockDatabase,
				cfg: config.Cache{
					CacheDefaultExpiry: 1000,
				},
			},
			args: args{
				request: &openrtb2.BidRequest{
					Test: 1,
				},
				pubID:          testPubID,
				profileID:      testProfileID,
				displayVersion: 4,
			},
			setup: func() {},
			want: want{
				adunitConfig: nil,
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
				cache: newCache,
				cfg:   tt.fields.cfg,
				db:    tt.fields.db,
			}
			adunitConfig := c.GetAdunitConfigFromCache(tt.args.request, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			assert.Equal(t, tt.want.adunitConfig, adunitConfig, "Expected: %v but got %v", tt.want.adunitConfig, adunitConfig)
		})
	}
}
