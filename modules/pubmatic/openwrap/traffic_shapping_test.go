package openwrap

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func TestEvaluateBiddingCondition(t *testing.T) {
	type args struct {
		data  interface{}
		logic interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
	}{
		{
			name: "No data present",
			args: args{
				data: nil,
				logic: map[string]interface{}{
					"or": []interface{}{
						map[string]interface{}{
							"and": []interface{}{
								map[string]interface{}{
									"in": []interface{}{
										map[string]interface{}{
											"var": "country",
										},
										[]interface{}{
											"JPN",
											"KOR",
										},
									},
								},
								map[string]interface{}{
									"==": []interface{}{
										map[string]interface{}{
											"var": "buyeruidAvailable",
										},
										true,
									},
								},
							},
						},
						map[string]interface{}{
							"and": []interface{}{
								map[string]interface{}{
									"==": []interface{}{
										map[string]interface{}{
											"var": "testScenario",
										},
										"a-jpn-kor-no-uid",
									},
								},
								map[string]interface{}{
									"in": []interface{}{
										map[string]interface{}{
											"var": "country",
										},
										[]interface{}{
											"JPN",
											"KOR",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "No logic present",
			args: args{
				data: map[string]interface{}{
					"country":           "IND",
					"buyeruidAvailable": true,
				},
				logic: nil,
			},
		},
		{
			name: "Logic data present and evaluation returns true",
			args: args{
				data: map[string]interface{}{
					"country":           "JPN",
					"buyeruidAvailable": true,
				},
				logic: map[string]interface{}{
					"or": []interface{}{
						map[string]interface{}{
							"and": []interface{}{
								map[string]interface{}{
									"in": []interface{}{
										map[string]interface{}{
											"var": "country",
										},
										[]interface{}{
											"JPN",
											"KOR",
										},
									},
								},
								map[string]interface{}{
									"==": []interface{}{
										map[string]interface{}{
											"var": "buyeruidAvailable",
										},
										true,
									},
								},
							},
						},
						map[string]interface{}{
							"and": []interface{}{
								map[string]interface{}{
									"==": []interface{}{
										map[string]interface{}{
											"var": "testScenario",
										},
										"a-jpn-kor-no-uid",
									},
								},
								map[string]interface{}{
									"in": []interface{}{
										map[string]interface{}{
											"var": "country",
										},
										[]interface{}{
											"JPN",
											"KOR",
										},
									},
								},
							},
						},
					},
				},
			},
			wantResult: true,
		},
		{
			name: "Logic data present and evaluation returns false",
			args: args{
				data: map[string]interface{}{
					"country":           "IND",
					"buyeruidAvailable": true,
				},
				logic: map[string]interface{}{
					"or": []interface{}{
						map[string]interface{}{
							"and": []interface{}{
								map[string]interface{}{
									"in": []interface{}{
										map[string]interface{}{
											"var": "country",
										},
										[]interface{}{
											"JPN",
											"KOR",
										},
									},
								},
								map[string]interface{}{
									"==": []interface{}{
										map[string]interface{}{
											"var": "buyeruidAvailable",
										},
										true,
									},
								},
							},
						},
						map[string]interface{}{
							"and": []interface{}{
								map[string]interface{}{
									"==": []interface{}{
										map[string]interface{}{
											"var": "testScenario",
										},
										"a-jpn-kor-no-uid",
									},
								},
								map[string]interface{}{
									"in": []interface{}{
										map[string]interface{}{
											"var": "country",
										},
										[]interface{}{
											"JPN",
											"KOR",
										},
									},
								},
							},
						},
					},
				},
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := evaluateBiddingCondition(tt.args.data, tt.args.logic)
			assert.Equal(t, gotResult, tt.wantResult, tt.name)
		})
	}
}

func TestGetFilteredBidders(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	testCases := []struct {
		name           string
		requestCtx     models.RequestCtx
		bidRequest     *openrtb2.BidRequest
		cache          cache.Cache
		expectedResult map[string]struct{}
		expectedFlag   bool
		setup          func()
	}{
		{
			name: "Bidder filter not found in cache",
			requestCtx: models.RequestCtx{
				PubID:     1,
				ProfileID: 2,
				DisplayID: 3,
			},
			bidRequest: &openrtb2.BidRequest{},
			setup: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(nil, false)
			},
			expectedResult: map[string]struct{}{},
			expectedFlag:   false,
		},
		{
			name: "Invalid data in cache",
			requestCtx: models.RequestCtx{
				PubID:     1,
				ProfileID: 2,
				DisplayID: 3,
			},
			bidRequest: &openrtb2.BidRequest{},
			setup: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return("abc", false)
			},
			expectedResult: map[string]struct{}{},
			expectedFlag:   false,
		},
		{
			name: "Bidder filter found in cache - All partners filtered",
			requestCtx: models.RequestCtx{
				PubID:     1,
				ProfileID: 2,
				DisplayID: 3,
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner1",
					},
					2: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner2",
					},
				},
			},
			setup: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(map[string]interface{}{
					"partner1": map[string]interface{}{
						"or": []interface{}{
							map[string]interface{}{
								"and": []interface{}{
									map[string]interface{}{
										"in": []interface{}{
											map[string]interface{}{
												"var": "country",
											},
											[]interface{}{
												"JPN",
												"KOR",
											},
										},
									},
									map[string]interface{}{
										"==": []interface{}{
											map[string]interface{}{
												"var": "buyeruidAvailable",
											},
											true,
										},
									},
								},
							},
							map[string]interface{}{
								"and": []interface{}{
									map[string]interface{}{
										"==": []interface{}{
											map[string]interface{}{
												"var": "testScenario",
											},
											"a-jpn-kor-no-uid",
										},
									},
									map[string]interface{}{
										"in": []interface{}{
											map[string]interface{}{
												"var": "country",
											},
											[]interface{}{
												"JPN",
												"KOR",
											},
										},
									},
								},
							},
						},
					},
				}, true)
			},
			bidRequest: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					Geo: &openrtb2.Geo{
						Country: "IND",
					},
				},
			},
			expectedResult: map[string]struct{}{
				"partner1": {},
			},
			expectedFlag: false,
		},
		{
			name: "Bidder filter found in cache - No partner filtered",
			requestCtx: models.RequestCtx{
				PubID:     1,
				ProfileID: 2,
				DisplayID: 3,
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner1",
					},
					2: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner2",
					},
				},
			},
			setup: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(
					map[string]interface{}{
						"partner1": map[string]interface{}{
							"or": []interface{}{
								map[string]interface{}{
									"in": []interface{}{
										map[string]interface{}{
											"var": "country",
										},
										[]interface{}{
											"IN",
											"KOR",
										},
									},
								},
								map[string]interface{}{
									"and": []interface{}{
										map[string]interface{}{
											"==": []interface{}{
												map[string]interface{}{
													"var": "testScenario",
												},
												"a-jpn-kor-no-uid",
											},
										},
										map[string]interface{}{
											"in": []interface{}{
												map[string]interface{}{
													"var": "country",
												},
												[]interface{}{
													"JPN",
													"KOR",
												},
											},
										},
									},
								},
							},
						},
					}, true)
			},
			bidRequest: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					Geo: &openrtb2.Geo{
						Country: "IND",
					},
				},
			},
			expectedResult: map[string]struct{}{},
			expectedFlag:   false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			result, flag := getFilteredBidders(tc.requestCtx, tc.bidRequest, mockCache)
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedFlag, flag)
		})
	}
}
