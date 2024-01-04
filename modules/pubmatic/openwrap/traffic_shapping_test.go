package openwrap

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb"
	mock_geodb "github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb/mock"
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
				AdapterThrottleMap: map[string]struct{}{
					"partner3": {},
				},
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner1",
					},
					2: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner2",
					},
					3: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner3",
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
											"IND",
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

func TestGetCountryFromRequest(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGeoDb := mock_geodb.NewMockGeography(ctrl)

	type args struct {
		rCtx       models.RequestCtx
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		want  string
	}{
		{
			name: "country_present_in_device_object",
			args: args{
				rCtx: models.RequestCtx{},
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Geo: &openrtb2.Geo{Country: "IND"},
					},
				},
			},
			setup: func() {},
			want:  "IND",
		},
		{
			name: "contry_present_in_user_object",
			args: args{
				rCtx: models.RequestCtx{},
				bidRequest: &openrtb2.BidRequest{
					User: &openrtb2.User{
						Geo: &openrtb2.Geo{
							Country: "JPN",
						},
					},
				},
			},
			setup: func() {},
			want:  "JPN",
		},
		{
			name: "detecting_country_from_request_ip",
			args: args{
				rCtx: models.RequestCtx{
					IP:             "101.143.255.255",
					GeoInfoFetcher: mockGeoDb,
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			setup: func() {
				mockGeoDb.EXPECT().LookUp("101.143.255.255").Return(&geodb.GeoInfo{
					CountryCode: "jp", ISOCountryCode: "JP", RegionCode: "13", City: "tokyo", PostalCode: "", DmaCode: 392001, Latitude: 35.68000030517578, Longitude: 139.75, AreaCode: "", AlphaThreeCountryCode: "JPN",
				}, nil)
			},
			want: "JPN",
		},
		{
			name: "detecting_country_from_device_ip",
			args: args{
				rCtx: models.RequestCtx{
					GeoInfoFetcher: mockGeoDb,
				},
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{IP: "100.43.128.0"},
				},
			},
			setup: func() {
				mockGeoDb.EXPECT().LookUp("100.43.128.0").Return(&geodb.GeoInfo{
					CountryCode: "us", ISOCountryCode: "US", RegionCode: "13", City: "abc", PostalCode: "", DmaCode: 392001, Latitude: 35.68000030517578, Longitude: 139.75, AreaCode: "", AlphaThreeCountryCode: "USA",
				}, nil)
			},
			want: "USA",
		},
		{
			name: "detecting_country_from_device_ipv6",
			args: args{
				rCtx: models.RequestCtx{
					GeoInfoFetcher: mockGeoDb,
				},
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{IPv6: "1.179.71.255"},
				},
			},
			setup: func() {
				mockGeoDb.EXPECT().LookUp("1.179.71.255").Return(&geodb.GeoInfo{
					CountryCode: "au", ISOCountryCode: "AU", RegionCode: "nsw", City: "brookvale", PostalCode: "", DmaCode: 36122, Latitude: -33.77000045776367, Longitude: 151.27000427246094, AreaCode: "", AlphaThreeCountryCode: "AUS",
				}, nil)
			},
			want: "AUS",
		},
		{
			name: "both_ip_and_country_are_missing_in_request",
			args: args{
				rCtx: models.RequestCtx{
					IP:             "",
					GeoInfoFetcher: mockGeoDb,
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			setup: func() {

			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := getCountryFromRequest(tt.args.rCtx, tt.args.bidRequest)
			assert.Equal(t, got, tt.want)

		})
	}
}

func TestGetCountryFromIP(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGeoDb := mock_geodb.NewMockGeography(ctrl)

	type args struct {
		ip             string
		geoInfoFetcher geodb.Geography
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		want  string
		err   error
	}{
		{
			name: "valid_ip",
			args: args{
				ip:             "1.179.71.255",
				geoInfoFetcher: mockGeoDb,
			},
			setup: func() {
				mockGeoDb.EXPECT().LookUp("1.179.71.255").Return(&geodb.GeoInfo{
					CountryCode: "au", ISOCountryCode: "AU", RegionCode: "nsw", City: "brookvale", PostalCode: "", DmaCode: 36122, Latitude: -33.77000045776367, Longitude: 151.27000427246094, AreaCode: "", AlphaThreeCountryCode: "AUS",
				}, nil)
			},
			want: "AUS",
			err:  nil,
		},
		{
			name: "geoDB_instance_missing",
			args: args{
				ip:             "1.179.71.255",
				geoInfoFetcher: nil,
			},
			setup: func() {},
			want:  "",
			err:   errors.New("geoDB instance is missing"),
		},
		{
			name: "invalid_ip",
			args: args{
				ip:             "1.179.71.255.123",
				geoInfoFetcher: mockGeoDb,
			},
			setup: func() {
				mockGeoDb.EXPECT().LookUp("1.179.71.255.123").Return(&geodb.GeoInfo{
					CountryCode: "", ISOCountryCode: "", RegionCode: "", City: "", PostalCode: "", DmaCode: 0, Latitude: 0, Longitude: 0, AreaCode: "", AlphaThreeCountryCode: "",
				}, nil)
			},
			want: "",
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := getCountryFromIP(tt.args.ip, tt.args.geoInfoFetcher)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, err, tt.err)

		})
	}
}
