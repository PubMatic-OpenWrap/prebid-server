package bidderparams

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func getTestImp(tagID string, banner bool, video bool) openrtb2.Imp {
	if banner {
		return openrtb2.Imp{
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
			TagID: tagID,
		}
	} else if video {
		return openrtb2.Imp{
			ID: "111",
			Video: &openrtb2.Video{
				W: 200,
				H: 300,
			},
			TagID: tagID,
		}
	}

	return openrtb2.Imp{
		ID: "111",
		Native: &openrtb2.Native{
			Request: "test",
			Ver:     "testVer",
		},
		TagID: tagID,
	}
}

func TestGetImpExtPubMaticKeyWords(t *testing.T) {
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
			got := getImpExtPubMaticKeyWords(tt.args.impExt, tt.args.bidderCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDealTier(t *testing.T) {
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
			got := getDealTier(tt.args.impExt, tt.args.bidderCode)
			assert.Equal(t, tt.want, got)
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
			name: "testRequest",
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
				imp:       getTestImp("/Test_Adunit1234", true, false),
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
			name: "exact_matched_slot_found_adslot_updated_from_PubMatic_secondary_flow",
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
							models.KEY_PROFILE_ID:      "1323",
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
				imp:       getTestImp("/Test_Adunit1234", true, false),
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
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234@DIV1@200x300","wrapper":{"version":0,"profile":1323},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
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
				imp:       getTestImp("/Test_Adunit1234", true, false),
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
		{
			name: "regex_matched_slot_found_adSlot_upadted_from_hashValue",
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
				imp:       getTestImp("/Test_Adunit1234", true, false),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					".*@div.*@.*": {
						PartnerId: 1,
						AdapterId: 1,
						SlotName:  "/Test_Adunit1234@Div1@200x300",
						SlotMappings: map[string]interface{}{
							"site":  "12313",
							"adtag": "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
					HashValueMap: map[string]string{
						".*@Div.*@.*": "2aa34b52a9e941c1594af7565e599c8d",
					},
				})
				mockCache.EXPECT().Get("psregex_5890_123_1_1_/Test_Adunit1234@Div1@200x300").Return(regexSlotEntry{
					SlotName:     "/Test_Adunit1234@Div1@200x300",
					RegexPattern: ".*@Div.*@.*",
				}, true)
			},
			want: want{
				matchedSlot:    "/Test_Adunit1234@Div1@200x300",
				matchedPattern: ".*@Div.*@.*",
				isRegexSlot:    true,
				params:         []byte(`{"publisherId":"5890","adSlot":"2aa34b52a9e941c1594af7565e599c8d","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		{
			name: "valid_pubmatic_native_params",
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
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
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
				imp:       getTestImp("/Test_Adunit1234", false, false),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"/test_adunit1234@1x1": {
						PartnerId: 1,
						AdapterId: 1,
						SlotName:  "/Test_Adunit1234@1x1",
						SlotMappings: map[string]interface{}{
							"site":                  "12313",
							"adtag":                 "45343",
							models.KEY_OW_SLOT_NAME: "/Test_Adunit1234@1x1",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
				})
			},
			want: want{
				matchedSlot:    "/Test_Adunit1234@1x1",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234@1x1","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		{
			name: "valid_pubmatic_video_params",
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
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
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
				imp:       getTestImp("/Test_Adunit1234", false, true),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"/test_adunit1234@0x0": {
						PartnerId: 1,
						AdapterId: 1,
						SlotName:  "/Test_Adunit1234@0x0",
						SlotMappings: map[string]interface{}{
							"site":                  "12313",
							"adtag":                 "45343",
							models.KEY_OW_SLOT_NAME: "/Test_Adunit1234@0x0",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"test", "test1"},
				})
			},
			want: want{
				matchedSlot:    "/Test_Adunit1234@0x0",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234@0x0","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		{
			name: "pubmatic_param_for_native_default",
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
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
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
				imp:       getTestImp("/Test_Adunit1234", false, false),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"random": {
						SlotName: "/Test_Adunit1234",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})

				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"random"},
					HashValueMap: map[string]string{
						"random": "2aa34b52a9e941c1594af7565e599c8d",
					},
				})
			},
			want: want{
				matchedSlot:    "",
				matchedPattern: "/Test_Adunit1234@1x1",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		{
			name: "pubmatic_param_for_banner_default",
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
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
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
				imp:       getTestImp("/Test_Adunit1234", true, false),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"random": {
						SlotName: "/Test_Adunit1234",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})

				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"random"},
					HashValueMap: map[string]string{
						"random": "2aa34b52a9e941c1594af7565e599c8d",
					},
				})
			},
			want: want{
				matchedSlot:    "",
				matchedPattern: "/Test_Adunit1234@200x300",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
		{
			name: "pubmatic_param_for_video_default",
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
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
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
				imp:       getTestImp("/Test_Adunit1234", false, true),
				partnerID: 1,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"random": {
						SlotName: "/Test_Adunit1234",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})

				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"random"},
					HashValueMap: map[string]string{
						"random": "2aa34b52a9e941c1594af7565e599c8d",
					},
				})
			},
			want: want{
				matchedSlot:    "",
				matchedPattern: "/Test_Adunit1234@0x0",
				isRegexSlot:    false,
				params:         []byte(`{"publisherId":"5890","adSlot":"/Test_Adunit1234","wrapper":{"version":1,"profile":123},"keywords":[{"key":"test_key1","value":["test_value1","test_value2"]},{"key":"test_key2","value":["test_value1","test_value2"]}]}`),
				wantErr:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			matchedSlot, matchedPattern, isRegexSlot, params, err := PreparePubMaticParamsV25(tt.args.rctx, tt.args.cache, tt.args.bidRequest, tt.args.imp, tt.args.impExt, tt.args.partnerID)
			if (err != nil) != tt.want.wantErr {
				assert.Equal(t, tt.want.wantErr, err != nil)
				return
			}
			assert.Equal(t, tt.want.matchedSlot, matchedSlot)
			assert.Equal(t, tt.want.matchedPattern, matchedPattern)
			assert.Equal(t, tt.want.isRegexSlot, isRegexSlot)
			assert.Equal(t, tt.want.params, params)
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
