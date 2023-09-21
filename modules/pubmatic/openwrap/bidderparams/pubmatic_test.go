package bidderparams

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func getTestImp(tagID string) openrtb2.Imp {
	imp := openrtb2.Imp{
		ID: "111",
		Banner: &openrtb2.Banner{
			W: ptrutil.ToPtr[int64](200),
			H: ptrutil.ToPtr[int64](300),
			Format: []openrtb2.Format{
				{
					W: 400,
					H: 500,
				},
			},
		},
		Video:  &openrtb2.Video{},
		Native: &openrtb2.Native{},
		TagID:  tagID,
	}
	return imp
}

func Test_getImpExtPubMaticKeyWords(t *testing.T) {
	type args struct {
		impExt     models.ImpExtension
		bidderCode string
	}
	tests := []struct {
		name string
		args args
		want []*openrtb_ext.ExtImpPubmaticKeyVal
	}{
		{
			name: "empty_impExt_bidder",
			args: args{
				impExt: models.ImpExtension{
					Bidder: nil,
				},
			},
			want: nil,
		},
		{
			name: "bidder_code_is_not_present_in_impExt_bidder",
			args: args{
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"appnexus": {
							KeyWords: []models.KeyVal{
								{
									Key:    "test_key1",
									Values: []string{"test_value1", "test_value2"},
								},
							},
						},
					},
				},
				bidderCode: "pubmatic",
			},
			want: nil,
		},
		{
			name: "impExt_bidder_contains_key_value_pair_for_bidder_code",
			args: args{
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"pubmatic": {
							KeyWords: []models.KeyVal{
								{
									Key:    "test_key1",
									Values: []string{"test_value1", "test_value2"},
								},
								{
									Key:    "test_key2",
									Values: []string{"test_value1", "test_value2"},
								},
							},
						},
					},
				},
				bidderCode: "pubmatic",
			},
			want: []*openrtb_ext.ExtImpPubmaticKeyVal{
				{
					Key:    "test_key1",
					Values: []string{"test_value1", "test_value2"},
				},
				{
					Key:    "test_key2",
					Values: []string{"test_value1", "test_value2"},
				},
			},
		},
		{
			name: "impExt_bidder_contains_key_value_pair_for_bidder_code_ignore_key_value_pair_with_no_values",
			args: args{
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"pubmatic": {
							KeyWords: []models.KeyVal{
								{
									Key:    "test_key1",
									Values: []string{"test_value1", "test_value2"},
								},
								{
									Key:    "test_key2",
									Values: []string{},
								},
							},
						},
					},
				},
				bidderCode: "pubmatic",
			},
			want: []*openrtb_ext.ExtImpPubmaticKeyVal{
				{
					Key:    "test_key1",
					Values: []string{"test_value1", "test_value2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getImpExtPubMaticKeyWords(tt.args.impExt, tt.args.bidderCode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getImpExtPubMaticKeyWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDealTier(t *testing.T) {
	type args struct {
		impExt     models.ImpExtension
		bidderCode string
	}
	tests := []struct {
		name string
		args args
		want *openrtb_ext.DealTier
	}{
		{
			name: "impExt_bidder_is_empty",
			args: args{
				impExt: models.ImpExtension{
					Bidder: nil,
				},
			},
			want: nil,
		},
		{
			name: "bidder_code_is_not_present_in_impExt_bidder",
			args: args{
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"appnexus": {
							DealTier: &openrtb_ext.DealTier{
								Prefix:      "test",
								MinDealTier: 10,
							},
						},
					},
				},
				bidderCode: "pubmatic",
			},
			want: nil,
		},
		{
			name: "bidder_code_is_present_in_impExt_bidder",
			args: args{
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"pubmatic": {
							DealTier: &openrtb_ext.DealTier{
								Prefix:      "test",
								MinDealTier: 10,
							},
						},
					},
				},
				bidderCode: "pubmatic",
			},
			want: &openrtb_ext.DealTier{
				Prefix:      "test",
				MinDealTier: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDealTier(tt.args.impExt, tt.args.bidderCode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDealTier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreparePubMaticParamsV25(t *testing.T) {
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
		matchedSlot    string
		matchedPattern string
		isRegexSlot    bool
		params         []byte
		wantErr        bool
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		want  want
	}{
		{
			name: "request_with_test_value_1",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
					PubID:         5890,
					ProfileID:     123,
					DisplayID:     1,
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.PREBID_PARTNER_NAME: "pubmatic",
							models.BidderCode:          "pubmatic",
							models.TIMEOUT:             "200",
							models.KEY_GEN_PATTERN:     "_AU_@_DIV_@_W_x_H_",
							models.SERVER_SIDE_FLAG:    "1",
						},
					},
				},
				cache: mockCache,
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"pubmatic": {
							KeyWords: []models.KeyVal{
								{
									Key:    "test_key1",
									Values: []string{"test_value1", "test_value2"},
								},
								{
									Key:    "test_key2",
									Values: []string{"test_value1", "test_value2"},
								},
							},
						},
					},
					Wrapper: &models.ExtImpWrapper{
						Div: "Div1",
					},
				},
				imp:       getTestImp("/Test_Adunit1234"),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"test": {
						PartnerId: 1,
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
				})
			},
			want: want{
				matchedSlot:    "/Test_Adunit1234@Div1@200x300",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234@Div1@200x300","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		{
			name: "exact_matched_slot_found_adslot_updated_from_slotName",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 0,
					PubID:         5890,
					ProfileID:     123,
					DisplayID:     1,
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.PREBID_PARTNER_NAME: "pubmatic",
							models.BidderCode:          "pubmatic",
							models.TIMEOUT:             "200",
							models.KEY_GEN_PATTERN:     "_AU_@_DIV_@_W_x_H_",
							models.SERVER_SIDE_FLAG:    "1",
						},
					},
				},
				cache: mockCache,
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"pubmatic": {
							KeyWords: []models.KeyVal{
								{
									Key:    "test_key1",
									Values: []string{"test_value1", "test_value2"},
								},
								{
									Key:    "test_key2",
									Values: []string{"test_value1", "test_value2"},
								},
							},
						},
					},
					Wrapper: &models.ExtImpWrapper{
						Div: "Div1",
					},
				},
				imp:       getTestImp("/Test_Adunit1234"),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"/test_adunit1234@div1@200x300": {
						PartnerId: 1,
						AdapterId: 1,
						SlotName:  "/Test_Adunit1234@Div1@200x300",
						SlotMappings: map[string]interface{}{
							"site":     "12313",
							"adtag":    "45343",
							"slotName": "/Test_Adunit1234@DIV1@200x300",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
				})
			},
			want: want{
				matchedSlot:    "/Test_Adunit1234@Div1@200x300",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234@DIV1@200x300","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		{
			name: "exact_matched_slot_found_adSlot_upadted_from_owSlotName",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 0,
					PubID:         5890,
					ProfileID:     123,
					DisplayID:     1,
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.PREBID_PARTNER_NAME: "pubmatic",
							models.BidderCode:          "pubmatic",
							models.TIMEOUT:             "200",
							models.KEY_GEN_PATTERN:     "_AU_@_DIV_@_W_x_H_",
							models.SERVER_SIDE_FLAG:    "1",
						},
					},
				},
				cache: mockCache,
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"pubmatic": {
							KeyWords: []models.KeyVal{
								{
									Key:    "test_key1",
									Values: []string{"test_value1", "test_value2"},
								},
								{
									Key:    "test_key2",
									Values: []string{"test_value1", "test_value2"},
								},
							},
						},
					},
					Wrapper: &models.ExtImpWrapper{
						Div: "Div1",
					},
				},
				imp:       getTestImp("/Test_Adunit1234"),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"/test_adunit1234@div1@200x300": {
						PartnerId: 1,
						AdapterId: 1,
						SlotName:  "/Test_Adunit1234@Div1@200x300",
						SlotMappings: map[string]interface{}{
							"site":                  "12313",
							"adtag":                 "45343",
							models.KEY_OW_SLOT_NAME: "/Test_Adunit1234@DIV1@200x300",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
				})
			},
			want: want{
				matchedSlot:    "/Test_Adunit1234@Div1@200x300",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234@DIV1@200x300","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		// {
		// 	name: "custom_profileid_publisherid",
		// 	args: args{
		// 		rctx: models.RequestCtx{
		// 			IsTestRequest: 0,
		// 			PubID:         101,
		// 			ProfileID:     201,
		// 			DisplayID:     0,
		// 			PartnerConfigMap: map[int]map[string]string{
		// 				1: {
		// 					models.PREBID_PARTNER_NAME: "pubmatic",
		// 					models.BidderCode:          "pubmatic",
		// 					models.TIMEOUT:             "200",
		// 					models.KEY_GEN_PATTERN:     "_AU_@_DIV_@_W_x_H_",
		// 					models.SERVER_SIDE_FLAG:    "1",
		// 					models.KEY_PUBLISHER_ID:    "101",
		// 					models.KEY_PROFILE_ID:      "201",
		// 				},
		// 			},
		// 		},
		// 		cache: mockCache,
		// 		impExt: models.ImpExtension{
		// 			Bidder: map[string]*models.BidderExtension{
		// 				"pubmatic": {
		// 					KeyWords: []models.KeyVal{
		// 						{
		// 							Key:    "pmzoneid",
		// 							Values: []string{"val1", "val2"},
		// 						},
		// 					},
		// 				},
		// 			},
		// 			Wrapper: &models.ExtImpWrapper{
		// 				Div: "Div1",
		// 			},
		// 		},
		// 		imp:       getTestImp("/Test_Adunit1234"),
		// 		partnerID: 1,
		// 	},
		// 	setup: func() {
		// 		mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
		// 			".*@.*@.*": createSlotMapping(".*@.*@.*", map[string]interface{}{
		// 				models.SITE_CACHE_KEY: "12313",
		// 				models.TAG_CACHE_KEY:  "45343",
		// 				models.KEY_SLOT_NAME:  "NewSlotName",
		// 			}),
		// 		})
		// 		mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
		// 			OrderedSlotList: []string{".*@.*@.*"},
		// 			HashValueMap: map[string]string{
		// 				".*@.*@.*": "2aa34b52a9e941c1594af7565e599c8d",
		// 			},
		// 		})
		// 		mockCache.EXPECT().Get("psregex_101_201_0_1_/Test_Adunit1234@Div1@200x300").Return(
		// 			regexSlotEntry{
		// 				SlotName:     "Test_Adunit1234@Div1@200x300",
		// 				RegexPattern: "2aa34b52a9e941c1594af7565e599c8d",
		// 			},
		// 		)
		// 	},
		// 	want: want{
		// 		matchedSlot:    "/Test_Adunit1234@Div1@200x300",
		// 		matchedPattern: "",
		// 		isRegexSlot:    false,
		// 		params:         []byte(`{"publisherId":"101","adSlot":"NewSlotName","wrapper":{"version":0,"profile":201},"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}`),
		// 		wantErr:        false,
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, got1, got2, got3, err := PreparePubMaticParamsV25(tt.args.rctx, tt.args.cache, tt.args.bidRequest, tt.args.imp, tt.args.impExt, tt.args.partnerID)
			fmt.Println(string(got3))
			if (err != nil) != tt.want.wantErr {
				t.Errorf("PreparePubMaticParamsV25() error = %v, wantErr %v", err, tt.want.wantErr)
				return
			}
			if got != tt.want.matchedSlot {
				t.Errorf("PreparePubMaticParamsV25() got = %v, want %v", got, tt.want.matchedSlot)
			}
			if got1 != tt.want.matchedPattern {
				t.Errorf("PreparePubMaticParamsV25() got1 = %v, want %v", got1, tt.want.matchedPattern)
			}
			if got2 != tt.want.isRegexSlot {
				t.Errorf("PreparePubMaticParamsV25() got2 = %v, want %v", got2, tt.want.isRegexSlot)
			}
			if !reflect.DeepEqual(got3, tt.want.params) {
				t.Errorf("PreparePubMaticParamsV25() got3 = %v, want %v", got3, tt.want.params)
			}
		})
	}
}

func createSlotMapping(slotName string, mappings map[string]interface{}) models.SlotMapping {
	return models.SlotMapping{
		PartnerId:    0,
		AdapterId:    0,
		VersionId:    0,
		SlotName:     slotName,
		SlotMappings: mappings,
		Hash:         "",
		OrderID:      0,
	}
}
