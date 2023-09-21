package bidderparams

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func TestGenerateSlotName(t *testing.T) {
	type args struct {
		h     int64
		w     int64
		kgp   string
		tagid string
		div   string
		src   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "_AU_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit",
		},
		{
			name: "_DIV_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_DIV_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "Div1",
		},
		{
			name: "_AU_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit",
		},
		{
			name: "_AU_@_W_x_H_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_@_W_x_H_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit@200x100",
		},
		{
			name: "_DIV_@_W_x_H_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_DIV_@_W_x_H_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "Div1@200x100",
		},
		{
			name: "_W_x_H_@_W_x_H_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_W_x_H_@_W_x_H_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "200x100@200x100",
		},
		{
			name: "_AU_@_DIV_@_W_x_H_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_@_DIV_@_W_x_H_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit@Div1@200x100",
		},
		{
			name: "_AU_@_SRC_@_VASTTAG_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_@_SRC_@_VASTTAG_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit@test.com@_VASTTAG_",
		},
		{
			name: "empty_kgp",
			args: args{
				h:     100,
				w:     200,
				kgp:   "",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "",
		},
		{
			name: "random_kgp",
			args: args{
				h:     100,
				w:     200,
				kgp:   "fjkdfhk",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateSlotName(tt.args.h, tt.args.w, tt.args.kgp, tt.args.tagid, tt.args.div, tt.args.src); got != tt.want {
				t.Errorf("GenerateSlotName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSlotMeta(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type args struct {
		rctx       models.RequestCtx
		cache      cache.Cache
		bidRequest openrtb2.BidRequest
		imp        openrtb2.Imp
		impExt     models.ImpExtension
		partnerID  int
	}
	type want struct {
		slots           []string
		slotMap         map[string]models.SlotMapping
		slotMappingInfo models.SlotMappingInfo
		hw              [][2]int64
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		want  want
	}{
		{
			name: "Test_value_other_than_2_and_got_slot_map_empty_from_cache",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
				},
				cache:     mockCache,
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(models.RequestCtx{
					IsTestRequest: 1,
				}, 1).Return(nil)
			},
			want: want{
				slots:           nil,
				slotMap:         nil,
				slotMappingInfo: models.SlotMappingInfo{},
				hw:              nil,
			},
		},
		{
			name: "Test_value_other_than_2_and_got_slotMappingInfo_OrderedSlotList_empty_from_cache",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
				},
				cache:     mockCache,
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(models.RequestCtx{
					IsTestRequest: 1,
				}, 1).Return(map[string]models.SlotMapping{
					"test": {
						PartnerId: 1,
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(models.RequestCtx{
					IsTestRequest: 1,
				}, 1).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{},
				})
			},
			want: want{
				slots:           nil,
				slotMap:         nil,
				slotMappingInfo: models.SlotMappingInfo{},
				hw:              nil,
			},
		},
		{
			name: "Test_value_is_2_but_partner_is_other_than_pubmatic_got_slotMappingInfo_OrderedSlotList_empty_from_cache",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 2,
					PartnerConfigMap: map[int]map[string]string{
						2: {
							"biddercode": "appnexus",
						},
					},
				},
				cache:     mockCache,
				partnerID: 2,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(models.RequestCtx{
					IsTestRequest: 2,
					PartnerConfigMap: map[int]map[string]string{
						2: {
							"biddercode": "appnexus",
						},
					},
				}, 2).Return(map[string]models.SlotMapping{
					"test": {
						PartnerId: 2,
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(models.RequestCtx{
					IsTestRequest: 2,
					PartnerConfigMap: map[int]map[string]string{
						2: {
							"biddercode": "appnexus",
						},
					},
				}, 2).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{},
				})
			},
			want: want{
				slots:           nil,
				slotMap:         nil,
				slotMappingInfo: models.SlotMappingInfo{},
				hw:              nil,
			},
		},
		{
			name: "Other_than_test_request_and_got_slot_map_and_slotMappingInfo_from_the_chche",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 0,
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"kgp": "_AU_",
						},
					},
				},
				cache:     mockCache,
				partnerID: 1,
				imp:       getTestImp("/Test_Adunti1234"),
				impExt: models.ImpExtension{
					Wrapper: &models.ExtImpWrapper{
						Div: "Div1",
					},
				},
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(models.RequestCtx{
					IsTestRequest: 0,
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"kgp": "_AU_",
						},
					},
				}, 1).Return(map[string]models.SlotMapping{
					"test": {
						PartnerId: 1,
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(models.RequestCtx{
					IsTestRequest: 0,
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"kgp": "_AU_",
						},
					},
				}, 1).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
				})
			},
			want: want{
				slots: []string{"/Test_Adunti1234", "/Test_Adunti1234", "/Test_Adunti1234", "/Test_Adunti1234"},
				slotMap: map[string]models.SlotMapping{
					"test": {
						PartnerId: 1,
					},
				},
				slotMappingInfo: models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
				},
				hw: [][2]int64{
					{300, 200},
					{500, 400},
					{0, 0},
					{1, 1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, got1, got2, got3 := getSlotMeta(tt.args.rctx, tt.args.cache, tt.args.bidRequest, tt.args.imp, tt.args.impExt, tt.args.partnerID)
			if !reflect.DeepEqual(got, tt.want.slots) {
				t.Errorf("getSlotMeta() got = %v, want %v", got, tt.want.slots)
			}
			if !reflect.DeepEqual(got1, tt.want.slotMap) {
				t.Errorf("getSlotMeta() got1 = %v, want %v", got1, tt.want.slotMap)
			}
			if !reflect.DeepEqual(got2, tt.want.slotMappingInfo) {
				t.Errorf("getSlotMeta() got2 = %v, want %v", got2, tt.want.slotMappingInfo)
			}
			if !reflect.DeepEqual(got3, tt.want.hw) {
				t.Errorf("getSlotMeta() got3 = %v, want %v", got3, tt.want.hw)
			}
		})
	}
}

func Test_getDefaultMappingKGP(t *testing.T) {
	type args struct {
		keyGenPattern string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty_keyGenPattern",
			args: args{
				keyGenPattern: "",
			},
			want: "",
		},
		{
			name: "keyGenPattern_contains_@_W_x_H_",
			args: args{
				keyGenPattern: "_AU_@_W_x_H_",
			},
			want: "_AU_",
		},
		{
			name: "keyGenPattern_contains_only_AU_",
			args: args{
				keyGenPattern: "_AU_",
			},
			want: "_AU_",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDefaultMappingKGP(tt.args.keyGenPattern); got != tt.want {
				t.Errorf("getDefaultMappingKGP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSlotMappings(t *testing.T) {
	type args struct {
		matchedSlot    string
		matchedPattern string
		slotMap        map[string]models.SlotMapping
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "found_matched_slot",
			args: args{
				matchedSlot:    "/Test_Adunit1234",
				matchedPattern: "",
				slotMap: map[string]models.SlotMapping{
					"/test_adunit1234": {
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				},
			},
			want: map[string]interface{}{
				models.SITE_CACHE_KEY: "12313",
				models.TAG_CACHE_KEY:  "45343",
			},
		},
		{
			name: "found_matched_pattern",
			args: args{
				matchedSlot:    "au123@div1@728x90",
				matchedPattern: "au1.*@div.*@.*",
				slotMap: map[string]models.SlotMapping{
					"au1.*@div.*@.*": {
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				},
			},
			want: map[string]interface{}{
				models.SITE_CACHE_KEY: "12313",
				models.TAG_CACHE_KEY:  "45343",
			},
		},
		{
			name: "not_found_matched_slot_as_well_as_matched_pattern",
			args: args{
				matchedSlot:    "",
				matchedPattern: "",
				slotMap:        map[string]models.SlotMapping{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSlotMappings(tt.args.matchedSlot, tt.args.matchedPattern, tt.args.slotMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSlotMappings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMatchingSlot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type args struct {
		rctx            models.RequestCtx
		cache           cache.Cache
		slot            string
		slotMap         map[string]models.SlotMapping
		slotMappingInfo models.SlotMappingInfo
		isRegexKGP      bool
		partnerID       int
	}
	type want struct {
		matchedSlot    string
		matchedPattern string
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		want  want
	}{
		{
			name: "Found_exact_match_slot",
			args: args{
				slotMap: map[string]models.SlotMapping{
					"/test_adunit1234": {
						PartnerId: 1,
						AdapterId: 1,
						VersionId: 1,
						SlotName:  "/Test_Adunit1234",
					},
				},
				slot: "/Test_Adunit1234",
			},
			want: want{
				matchedSlot:    "/Test_Adunit1234",
				matchedPattern: "",
			},
		},
		{
			name: "Not_found_exact_match_and_not_regex_as_well",
			args: args{
				slotMap:    map[string]models.SlotMapping{},
				isRegexKGP: false,
			},
			want: want{
				matchedSlot:    "",
				matchedPattern: "",
			},
		},
		{
			name: "found_matced_regex_slot",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
				},
				partnerID: 1,
				slot:      "AU123@Div1@728x90",
				slotMappingInfo: models.SlotMappingInfo{
					OrderedSlotList: []string{"*", ".*@.*@.*"},
					HashValueMap: map[string]string{
						".*@.*@.*": "2aa34b52a9e941c1594af7565e599c8d", // Code should match the given slot name with this regex
					},
				},
				slotMap: map[string]models.SlotMapping{
					"AU123@Div1@728x90": {
						SlotMappings: map[string]interface{}{
							"site":  "123123",
							"adtag": "45343",
						},
					},
				},
				cache:      mockCache,
				isRegexKGP: true,
			},
			setup: func() {
				mockCache.EXPECT().Get("psregex_5890_123_1_1_AU123@Div1@728x90").Return(nil, false)
				mockCache.EXPECT().Set("psregex_5890_123_1_1_AU123@Div1@728x90", regexSlotEntry{SlotName: "AU123@Div1@728x90", RegexPattern: ".*@.*@.*"}).Times(1)
			},
			want: want{
				matchedSlot:    "AU123@Div1@728x90",
				matchedPattern: ".*@.*@.*",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, got1 := GetMatchingSlot(tt.args.rctx, tt.args.cache, tt.args.slot, tt.args.slotMap, tt.args.slotMappingInfo, tt.args.isRegexKGP, tt.args.partnerID)
			if got != tt.want.matchedSlot {
				t.Errorf("GetMatchingSlot() got = %v, want %v", got, tt.want.matchedSlot)
			}
			if got1 != tt.want.matchedPattern {
				t.Errorf("GetMatchingSlot() got1 = %v, want %v", got1, tt.want.matchedPattern)
			}
		})
	}
}

func TestGetRegexMatchingSlot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		rctx            models.RequestCtx
		slot            string
		slotMap         map[string]models.SlotMapping
		slotMappingInfo models.SlotMappingInfo
		partnerID       int
	}
	type want struct {
		matchedSlot  string
		regexPattern string
	}
	tests := []struct {
		name  string
		args  args
		setup func() cache.Cache
		want  want
	}{
		{
			name: "happy_path_found_matched_regex_slot_entry_in_cahe",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
				},
				partnerID: 1,
				slot:      "/Test_Adunit1234",
			},
			setup: func() cache.Cache {
				mockCache := mock_cache.NewMockCache(ctrl)
				mockCache.EXPECT().Get("psregex_5890_123_1_1_/Test_Adunit1234").Return(interface{}(regexSlotEntry{SlotName: "/Test_Adunit1234", RegexPattern: "2aa34b52a9e941c1594af7565e599c8d"}), true)
				return mockCache
			},
			want: want{
				matchedSlot:  "/Test_Adunit1234",
				regexPattern: "2aa34b52a9e941c1594af7565e599c8d",
			},
		},
		{
			name: "not_found_matched_regex_slot_entry_in_cache",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
				},
				partnerID: 1,
				slot:      "AU123@Div1@728x90",
				slotMappingInfo: models.SlotMappingInfo{
					OrderedSlotList: []string{"AU1.*@Div.*@.*", ".*@.*@.*"},
					HashValueMap: map[string]string{
						"AU1.*@Div.*@.*": "2aa34b52a9e941c1594af7565e599c8d",
						".*@.*@.*":       "2aa34b52a9e941c1594af7565e599c8d",
					},
				},
				slotMap: map[string]models.SlotMapping{
					"AU123@Div1@728x90": {
						SlotMappings: map[string]interface{}{
							"site":  "123123",
							"adtag": "45343",
						},
					},
				},
			},
			setup: func() cache.Cache {
				mockCache := mock_cache.NewMockCache(ctrl)
				mockCache.EXPECT().Get("psregex_5890_123_1_1_AU123@Div1@728x90").Return(nil, false)
				mockCache.EXPECT().Set("psregex_5890_123_1_1_AU123@Div1@728x90", regexSlotEntry{SlotName: "AU123@Div1@728x90", RegexPattern: "AU1.*@Div.*@.*"}).Times(1)
				return mockCache
			},
			want: want{
				matchedSlot:  "AU123@Div1@728x90",
				regexPattern: "AU1.*@Div.*@.*",
			},
		},
		{
			name: "not_found_matched_regex_slot_entry_in_cache_case_Insensitive_Adslot",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
				},
				partnerID: 1,
				slot:      "au123@Div1@728x90",
				slotMappingInfo: models.SlotMappingInfo{
					OrderedSlotList: []string{"AU1.*@Div.*@.*", ".*@.*@.*"},
					HashValueMap: map[string]string{
						"AU1.*@Div.*@.*": "2aa34b52a9e941c1594af7565e599c8d",
						".*@.*@.*":       "2aa34b52a9e941c1594af7565e599c8d",
					},
				},
				slotMap: map[string]models.SlotMapping{
					"au123@Div1@728x90": {
						SlotMappings: map[string]interface{}{
							"site":  "123123",
							"adtag": "45343",
						},
					},
				},
			},
			setup: func() cache.Cache {
				mockCache := mock_cache.NewMockCache(ctrl)
				mockCache.EXPECT().Get("psregex_5890_123_1_1_au123@Div1@728x90").Return(nil, false)
				mockCache.EXPECT().Set("psregex_5890_123_1_1_au123@Div1@728x90", regexSlotEntry{SlotName: "au123@Div1@728x90", RegexPattern: "AU1.*@Div.*@.*"}).Times(1)
				return mockCache
			},
			want: want{
				matchedSlot:  "au123@Div1@728x90",
				regexPattern: "AU1.*@Div.*@.*",
			},
		},
		{
			name: "not_found_matched_regex_slot_entry_in_cache_cache_Incorrecct_regex",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
				},
				partnerID: 1,
				slot:      "au123@Div1@728x90",
				slotMappingInfo: models.SlotMappingInfo{
					OrderedSlotList: []string{"*@Div.*@*"},
					HashValueMap: map[string]string{
						"*@Div.*@*": "2aa34b52a9e941c1594af7565e599c8d",
					},
				},
				slotMap: map[string]models.SlotMapping{
					"au123@Div1@728x90": {
						SlotMappings: map[string]interface{}{
							"site":  "123123",
							"adtag": "45343",
						},
					},
				},
			},
			setup: func() cache.Cache {
				mockCache := mock_cache.NewMockCache(ctrl)
				mockCache.EXPECT().Get("psregex_5890_123_1_1_au123@Div1@728x90").Return(nil, false)
				return mockCache
			},
			want: want{
				matchedSlot:  "",
				regexPattern: "",
			},
		},
		{
			name: "not_found_matched_regex_slot_entry_in_cache_cache_Invalid_regex_pattern",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
				},
				partnerID: 1,
				slot:      "AU123@Div1@728x90",
				slotMappingInfo: models.SlotMappingInfo{
					OrderedSlotList: []string{"*", ".*@.*@.*"},
					HashValueMap: map[string]string{
						"*":        "2aa34b52a9e941c1594af7565e599c8d", // Invalid regex pattern
						".*@.*@.*": "2aa34b52a9e941c1594af7565e599c8d", // Code should match the given slot name with this regex
					},
				},
				slotMap: map[string]models.SlotMapping{
					"AU123@Div1@728x90": {
						SlotMappings: map[string]interface{}{
							"site":  "123123",
							"adtag": "45343",
						},
					},
				},
			},
			setup: func() cache.Cache {
				mockCache := mock_cache.NewMockCache(ctrl)
				mockCache.EXPECT().Get("psregex_5890_123_1_1_AU123@Div1@728x90").Return(nil, false)
				mockCache.EXPECT().Set("psregex_5890_123_1_1_AU123@Div1@728x90", regexSlotEntry{SlotName: "AU123@Div1@728x90", RegexPattern: ".*@.*@.*"}).Times(1)
				return mockCache
			},
			want: want{
				matchedSlot:  "AU123@Div1@728x90",
				regexPattern: ".*@.*@.*",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.setup()
			got, got1 := GetRegexMatchingSlot(tt.args.rctx, cache, tt.args.slot, tt.args.slotMap, tt.args.slotMappingInfo, tt.args.partnerID)
			if got != tt.want.matchedSlot {
				t.Errorf("GetRegexMatchingSlot() got = %v, want %v", got, tt.want.matchedSlot)
			}
			if got1 != tt.want.regexPattern {
				t.Errorf("GetRegexMatchingSlot() got1 = %v, want %v", got1, tt.want.regexPattern)
			}
		})
	}
}
