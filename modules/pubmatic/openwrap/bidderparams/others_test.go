package bidderparams

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestPrepareAdapterParamsV25(t *testing.T) {
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
			name: "AdUnit,Size slot matched",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
					PartnerConfigMap: map[int]map[string]string{
						19323: {
							models.PREBID_PARTNER_NAME: "appnexus",
							models.BidderCode:          "appnexus",
							models.TIMEOUT:             "200",
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
							models.SERVER_SIDE_FLAG:    "1",
						},
					},
				},
				cache: mockCache,
				impExt: models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/43743431/DMDEMO1",
					},
				},
				imp:       getTestImp("/43743431/DMDEMO1", true, false),
				partnerID: 19323,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"/43743431/dmdemo1@200x300": {
						PartnerId:   19323,
						AdapterId:   2,
						VersionId:   92588,
						SlotName:    "/43743431/DMDemo1@200x300",
						MappingJson: "{\"placementId\":\"9880618\"}",
						SlotMappings: map[string]interface{}{
							"placementId": "9880618",
						},
						OrderID: 1,
						Hash:    "",
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"/43743431/DMDemo1@200x300"},
					HashValueMap:    map[string]string{"/43743431/DMDemo1@200x300": ""},
				})
			},
			want: want{
				matchedSlot:    "/43743431/DMDEMO1@200x300",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         []byte("{\"placementId\":9880618}"),
				wantErr:        false,
			},
		},
		{
			name: "partnerconfig not found for partnerId",
			args: args{
				rctx: models.RequestCtx{
					PubID:            5890,
					ProfileID:        123,
					DisplayID:        1,
					PartnerConfigMap: nil,
				},
				cache:     mockCache,
				partnerID: 1,
			},
			setup: func() {},
			want: want{
				matchedSlot:    "",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         nil,
				wantErr:        true,
			},
		},
		{
			name: "slots not founds",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
					PubID:         5890,
					ProfileID:     123,
					DisplayID:     1,
					PartnerConfigMap: map[int]map[string]string{
						256: {
							models.PREBID_PARTNER_NAME: "appnexus",
							models.BidderCode:          "appnexus",
							models.TIMEOUT:             "200",
							models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
							models.SERVER_SIDE_FLAG:    "1",
						},
					},
				},
				cache: mockCache,
				impExt: models.ImpExtension{
					Bidder: map[string]*models.BidderExtension{
						"appnexus": {
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
				partnerID: 256,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				matchedSlot:    "",
				matchedPattern: "",
				isRegexSlot:    false,
				params:         nil,
				wantErr:        false,
			},
		},
		{
			name: "regex mapping slot found",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
					PartnerConfigMap: map[int]map[string]string{
						19323: {
							models.PREBID_PARTNER_NAME: "appnexus",
							models.BidderCode:          "appnexus",
							models.TIMEOUT:             "200",
							models.KEY_GEN_PATTERN:     "_AU_@_DIV_@_W_x_H_",
							models.SERVER_SIDE_FLAG:    "1",
						},
					},
				},
				cache: mockCache,
				impExt: models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/43743431/DMDEMO1",
					},
				},
				imp:       getTestImp("/43743431/DMDEMO1", true, false),
				partnerID: 19323,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"^/43743431/dmdemo[0-9]*@div[12]@^200x300$": {
						PartnerId:   19323,
						AdapterId:   2,
						VersionId:   92588,
						SlotName:    "^/43743431/DMDemo[0-9]*@Div[12]@^200x300$",
						MappingJson: "{\"placementId\":\"9880618\"}",
						SlotMappings: map[string]interface{}{
							"placementId": "9880618",
						},
						OrderID: 1,
						Hash:    "",
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"^/43743431/DMDemo[0-9]*@Div[12]@^200x300$"},
					HashValueMap:    map[string]string{"^/43743431/DMDemo[0-9]*@Div[12]@^200x300$": ""},
				})
				mockCache.EXPECT().Get("psregex_5890_123_1_19323_/43743431/DMDEMO1@@200x300").Return(regexSlotEntry{
					SlotName:     "/43743431/DMDEMO1@@200x300",
					RegexPattern: "^/43743431/DMDemo[0-9]*@Div[12]@^200x300$",
				}, true)
			},
			want: want{
				matchedSlot:    "/43743431/DMDEMO1@@200x300",
				matchedPattern: "^/43743431/DMDemo[0-9]*@Div[12]@^200x300$",
				isRegexSlot:    true,
				params:         []byte("{\"placementId\":9880618}"),
				wantErr:        false,
			},
		},
		{
			name: "prebid s2s regex mapping slot found",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 123,
					DisplayID: 1,
					PartnerConfigMap: map[int]map[string]string{
						19323: {
							models.PREBID_PARTNER_NAME: "appnexus",
							models.BidderCode:          "appnexus",
							models.TIMEOUT:             "200",
							models.KEY_GEN_PATTERN:     "_RE_@_W_x_H_",
							models.SERVER_SIDE_FLAG:    "1",
						},
					},
				},
				cache: mockCache,
				impExt: models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/43743431/DMDEMO1",
					},
				},
				imp:       getTestImp("/43743431/DMDEMO1", true, false),
				partnerID: 19323,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"^/43743431/dmdemo[0-9]*@^200x300$": {
						PartnerId:   19323,
						AdapterId:   2,
						VersionId:   92588,
						SlotName:    "^/43743431/DMDemo[0-9]*@^200x300$",
						MappingJson: "{\"placementId\":\"9880618\"}",
						SlotMappings: map[string]interface{}{
							"placementId": "9880618",
						},
						OrderID: 1,
						Hash:    "",
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"^/43743431/DMDemo[0-9]*@^200x300$"},
					HashValueMap:    map[string]string{"^/43743431/DMDemo[0-9]*@^200x300$": ""},
				})
				mockCache.EXPECT().Get("psregex_5890_123_1_19323_/43743431/DMDEMO1@200x300").Return(regexSlotEntry{
					SlotName:     "/43743431/DMDEMO1@200x300",
					RegexPattern: "^/43743431/DMDemo[0-9]*@^200x300$",
				}, true)
			},
			want: want{
				matchedSlot:    "/43743431/DMDEMO1@200x300",
				matchedPattern: "^/43743431/DMDemo[0-9]*@^200x300$",
				isRegexSlot:    true,
				params:         []byte("{\"placementId\":9880618}"),
				wantErr:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			adapters.InitBidders("./static/bidder-params/")
			matchedSlot, matchedPattern, isRegexSlot, params, err := PrepareAdapterParamsV25(tt.args.rctx, tt.args.cache, tt.args.imp, tt.args.impExt, tt.args.partnerID)
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
