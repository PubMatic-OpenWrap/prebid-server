package gocache

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	mock_database "github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/database/mock"
	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
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
		wantErr      bool
		adunitConfig json.RawMessage
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
				wantErr:      false,
				adunitConfig: []byte(`{"configPattern":"_AU_","regex":true,"config":{"default":{"floors":{"floormin":15,"data":{"currency":"USD","modelgroups":[{"currency":"USD","modelweight":40,"modelversion":"model 1 from adunit config slot level","schema":{"fields":["mediaType","size","domain"],"delimiter":"|"},"values":{"*|728x90|www.website.com":13,"banner|300x250|www.website.com":1,"banner|300x600|*":4,"banner|728x90|www.website.com":5},"default":5}]},"enforcement":{"enforcejs":true,"enforcepbs":true,"enforcerate":100},"enabled":true},"video":{"enabled":true,"config":{"mimes":null,"minduration":10,"maxduration":50,"skip":1,"skipmin":10,"skipafter":15,"battr":[6,7],"connectiontype":[2]}},"universalpixel":[{"id":123,"pixel":"pixle","pixeltype":"js","pos":"above","mediatype":"banner","partners":["pubmatic","appnexus"]}]},"div1":{"banner":{"enabled":true,"config":{"format":[{"w":200,"h":300},{"w":500,"h":800}]}},"video":{"enabled":true,"config":{"mimes":null,"minduration":10,"maxduration":50,"skip":1,"skipmin":10,"skipafter":15,"battr":[6,7],"connectiontype":[0,1,2,4]}}},"div2":{"video":{"enabled":true,"config":{"mimes":null,"minduration":10,"maxduration":50,"skip":1,"skipmin":10,"skipafter":15,"battr":[6,7],"connectiontype":[0,1,2,4]}}}}}`),
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
				wantErr:      true,
				adunitConfig: nil,
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
				wantErr:      false,
				adunitConfig: []byte(nil),
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
			err := c.populateCacheWithAdunitConfig(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			if tt.want.wantErr && (err == nil) {
				t.Error("Error should not be nil")
				return
			}
			cacheKey := key(PubAdunitConfig, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			obj, found := c.Get(cacheKey)

			if !tt.want.wantErr && !found {
				t.Error("Adunit Config not found in cache for cache key", cacheKey)
				return
			}

			if obj != nil {
				adunitConfig := obj.(*adunitconfig.AdUnitConfig)
				if adunitConfig != nil {
					actualAdunitConfig, err := json.Marshal(adunitConfig)
					assert.NoErrorf(t, err, "failed to marshal actual actualAdunitConfig for cachekey: %v", cacheKey)
					assert.JSONEqf(t, string(tt.want.adunitConfig), string(actualAdunitConfig), "Expected adunitconfig: %v but got: %v", string(tt.want.adunitConfig), string(actualAdunitConfig))
				}
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
		adunitConfig    json.RawMessage
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
			},
			want: want{
				adunitConfig:    nil,
				cacheKeyPresent: false,
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
				cacheKeyPresent: true,
				adunitConfig:    []byte(`{"configPattern":"_AU_","regex":true,"config":{"default":{"floors":{"floormin":15,"data":{"currency":"USD","modelgroups":[{"currency":"USD","modelweight":40,"modelversion":"model 1 from adunit config slot level","schema":{"fields":["mediaType","size","domain"],"delimiter":"|"},"values":{"*|728x90|www.website.com":13,"banner|300x250|www.website.com":1,"banner|300x600|*":4,"banner|728x90|www.website.com":5},"default":5}]},"enforcement":{"enforcejs":true,"enforcepbs":true,"enforcerate":100},"enabled":true},"video":{"enabled":true,"config":{"mimes":null,"minduration":10,"maxduration":50,"skip":1,"skipmin":10,"skipafter":15,"battr":[6,7],"connectiontype":[2]}},"universalpixel":[{"id":123,"pixel":"pixle","pixeltype":"js","pos":"above","mediatype":"banner","partners":["pubmatic","appnexus"]}]},"div1":{"banner":{"enabled":true,"config":{"format":[{"w":200,"h":300},{"w":500,"h":800}]}},"video":{"enabled":true,"config":{"mimes":null,"minduration":10,"maxduration":50,"skip":1,"skipmin":10,"skipafter":15,"battr":[6,7],"connectiontype":[0,1,2,4]}}},"div2":{"video":{"enabled":true,"config":{"mimes":null,"minduration":10,"maxduration":50,"skip":1,"skipmin":10,"skipafter":15,"battr":[6,7],"connectiontype":[0,1,2,4]}}}}}`),
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
				cacheKeyPresent: true,
				adunitConfig:    []byte(`{"config":{}}`),
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
				cacheKeyPresent: false,
				adunitConfig:    nil,
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
			adunitConfig := c.GetAdunitConfigFromCache(tt.args.request, tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			if tt.want.adunitConfig == nil && adunitConfig != nil {
				t.Errorf("adunitConfig should be nil")
				return
			}
			if adunitConfig != nil {
				actualAdunitConfig, err := json.Marshal(adunitConfig)
				assert.NoErrorf(t, err, "failed to marshal actual actualAdunitConfig ")
				assert.JSONEqf(t, string(tt.want.adunitConfig), string(actualAdunitConfig), "Expected adunitconfig: %v but got: %v", string(tt.want.adunitConfig), string(actualAdunitConfig))
			}
		})
	}
}
