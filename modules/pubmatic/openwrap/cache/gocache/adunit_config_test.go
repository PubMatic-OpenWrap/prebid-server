package gocache

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	mock_database "github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/database/mock"
	"github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
)

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
		presentInCache    bool
		emptyAdunitConfig bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   want
	}{
		{
			name: "valid_adunit_config",
			fields: fields{
				cache: gocache.New(100, 100),
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
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(&adunitconfig.AdUnitConfig{
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
											ModelWeight:  ptrutil.ToPtr[int](40),
											ModelVersion: "model 1 from adunit config slot level",
										},
									},
									Currency: "USD",
								},
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									EnforcePBS:  ptrutil.ToPtr[bool](true),
									EnforceRate: 100,
									EnforceJS:   ptrutil.ToPtr[bool](true),
								},
								Enabled: ptrutil.ToPtr[bool](true),
							},
							Video: &adunitconfig.Video{
								Enabled: ptrutil.ToPtr[bool](true),
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
								Enabled: ptrutil.ToPtr[bool](true),
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
								Enabled: ptrutil.ToPtr[bool](true),
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
								Enabled: ptrutil.ToPtr[bool](true),
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
				}, nil)
			},
			want: want{
				presentInCache: true,
			},
		},
		{
			name: "error_in_returning_adunitconfig_from_the_DB",
			fields: fields{
				cache: gocache.New(100, 100),
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
				presentInCache: false,
			},
		},
		{
			name: "returned_empty_adunitconfig_from_the_DB",
			fields: fields{
				cache: gocache.New(100, 100),
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
				presentInCache:    true,
				emptyAdunitConfig: true,
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
			c.populateCacheWithAdunitConfig(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			cacheKey := key(PubAdunitConfig, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			obj, found := c.Get(cacheKey)
			if !tt.want.presentInCache {
				if found {
					t.Error("Adunit config should not found in cache for cache key", cacheKey)
				}
				return
			}
			if !found {
				t.Error("Adunit Config not found in cache for cache key", cacheKey)
				return
			}

			if tt.want.emptyAdunitConfig {
				return
			}

			adunitConfig := obj.(*adunitconfig.AdUnitConfig)
			config := adunitConfig.Config
			if len(config) == 0 {
				t.Error("config should be empty for cache key", cacheKey)
			}
			defaultConfig, found := config[models.AdunitConfigDefaultKey]
			if !found {
				t.Error("Adunit config not set for default")
				return
			}

			if defaultConfig.Video == nil {
				t.Error("Video config not set for default")
			}

			if defaultConfig.Floors == nil {
				t.Error("Floor JSON not set for default")
			}

			slot1Config, found := config[strings.ToLower("Div1")]
			if !found {
				t.Error("Adunit config not set for slotname Div1")
				return
			}

			if slot1Config.Video == nil || slot1Config.Banner == nil {
				t.Error("Video/banner config not set for first slot: Div1")
				return
			}

			slot2Config, found := config[strings.ToLower("Div2")]
			if !found {
				t.Error("Adunit config not set for slotname Div2")
				return
			}

			if slot2Config.Video == nil {
				t.Error("Video/banner config not set for first slot: Div2")
				return
			}

			if found := adunitConfig.Regex; !found {
				t.Error("regex config not set")
				return
			}
		})
	}
}

func Test_cache_GetAdunitConfigFromCache(t *testing.T) {
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
		request        *openrtb2.BidRequest
		pubID          int
		profileID      int
		displayVersion int
	}
	type want struct {
		wantNil         bool
		emptyAUConfig   bool
		cacheKeyPresent bool
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
				cache: gocache.New(100, 100),
				db:    mockDatabase,
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
				displayVersion: testVersionID,
			},
			setup: func() {
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(nil, fmt.Errorf("error"))
			},
			want: want{
				wantNil:         true,
				emptyAUConfig:   false,
				cacheKeyPresent: true,
			},
		},
		{
			name: "successfully_get_value_from_cache",
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
				displayVersion: testVersionID,
			},
			setup: func() {
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(&adunitconfig.AdUnitConfig{
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
											ModelWeight:  ptrutil.ToPtr[int](40),
											ModelVersion: "model 1 from adunit config slot level",
										},
									},
									Currency: "USD",
								},
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									EnforcePBS:  ptrutil.ToPtr[bool](true),
									EnforceRate: 100,
									EnforceJS:   ptrutil.ToPtr[bool](true),
								},
								Enabled: ptrutil.ToPtr[bool](true),
							},
							Video: &adunitconfig.Video{
								Enabled: ptrutil.ToPtr[bool](true),
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
								Enabled: ptrutil.ToPtr[bool](true),
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
								Enabled: ptrutil.ToPtr[bool](true),
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
								Enabled: ptrutil.ToPtr[bool](true),
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
				}, nil)
			},
			want: want{
				emptyAUConfig:   false,
				wantNil:         false,
				cacheKeyPresent: true,
			},
		},
		{
			name: "got_empty_adunitconfig_from_cache",
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
				displayVersion: testVersionID,
			},
			setup: func() {
				mockDatabase.EXPECT().GetAdunitConfig(testProfileID, testVersionID).Return(&adunitconfig.AdUnitConfig{}, nil)
			},
			want: want{
				wantNil:         false,
				emptyAUConfig:   true,
				cacheKeyPresent: true,
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
				displayVersion: testVersionID,
			},
			want: want{
				wantNil:         true,
				emptyAUConfig:   false,
				cacheKeyPresent: false,
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
			if tt.want.cacheKeyPresent {
				c.populateCacheWithAdunitConfig(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			}
			got := c.GetAdunitConfigFromCache(tt.args.request, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			if tt.want.wantNil {
				if got != nil {
					t.Error("Adunit config got from cache should be nil")
				}
				return
			}

			if tt.want.emptyAUConfig {
				if len(got.Config) != 0 {
					t.Error("Config should be empty")
				}
				return
			}

			if got.Config == nil {
				t.Errorf("config should not be empty")
				return
			}

			config := got.Config
			if config[models.AdunitConfigDefaultKey].Floors == nil || config[models.AdunitConfigDefaultKey].Video == nil {
				t.Errorf("floors/video should not be nil")
			}
		})
	}
}
