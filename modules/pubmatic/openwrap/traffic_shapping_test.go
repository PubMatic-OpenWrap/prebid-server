package openwrap

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb"
	mock_geodb "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func TestEvaluateBiddingCondition(t *testing.T) {
	type args struct {
		data  string
		logic string
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
	}{
		{
			name: "No data present",
			args: args{
				data:  "{}",
				logic: `{ "in": [{ "var": "country"}, ["IND"]]}`,
			},
			wantResult: false,
		},
		{
			name: "Invalid data present",
			args: args{
				data:  `{"country":a}`,
				logic: `{ "in": [{ "var": "country"}, ["IND"]]}`,
			},
			wantResult: false,
		},
		{
			name: "No logic present",
			args: args{
				data:  `{"country": "IND"}`,
				logic: "{}",
			},
			wantResult: false,
		},
		{
			name: "Logic data present and evaluation returns true",
			args: args{
				data:  `{"country": "IND"}`,
				logic: `{ "in": [{ "var": "country"}, ["IND"]]}`,
			},
			wantResult: true,
		},
		{
			name: "Logic data present and evaluation returns false",
			args: args{
				data:  `{"country": "IND"}`,
				logic: `{ "in": [{ "var": "country"}, ["USA"]]}`,
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
	testCases := []struct {
		name           string
		requestCtx     models.RequestCtx
		bidRequest     *openrtb2.BidRequest
		cache          cache.Cache
		expectedResult map[string]struct{}
		expectedFlag   bool
	}{
		{
			name: "No bidder filter present",
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
						models.BidderFilters:    `{ "in": [{ "var": "country"}, ["USA"]]}`,
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
				DeviceCtx: models.DeviceCtx{Country: "IND"},
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner1",
						models.BidderFilters:    `{ "in": [{ "var": "country"}, ["IND"]]}`,
					},
					2: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner2",
					},
				},
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
		{
			name: "Bidder filter found in cache - All partner filtered",
			requestCtx: models.RequestCtx{
				PubID:     1,
				ProfileID: 2,
				DisplayID: 3,
				DeviceCtx: models.DeviceCtx{Country: "IND"},
				PartnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner1",
						models.BidderFilters:    `{ "in": [{ "var": "country"}, ["USA"]]}`,
					},
					2: {
						models.SERVER_SIDE_FLAG: "1",
						models.BidderCode:       "partner2",
						models.BidderFilters:    `{ "in": [{ "var": "country"}, ["USA"]]}`,
					},
				},
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
				"partner2": {},
			},
			expectedFlag: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := OpenWrap{}
			result, flag := m.getFilteredBidders(tc.requestCtx, tc.bidRequest)
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
			name: "getting country from request ",
			args: args{
				rCtx: models.RequestCtx{
					DeviceCtx: models.DeviceCtx{Country: "IND"},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			setup: func() {
			},
			want: "IND",
		},
		{
			name: "detecting_country_from_request_ip",
			args: args{
				rCtx: models.RequestCtx{
					DeviceCtx: models.DeviceCtx{IP: "101.143.255.255"},
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
			name: "both_ip_and_country_are_missing_in_request",
			args: args{
				rCtx: models.RequestCtx{
					DeviceCtx: models.DeviceCtx{IP: ""},
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
			m := OpenWrap{
				geoInfoFetcher: mockGeoDb,
			}
			tt.setup()
			got := m.getCountryFromRequest(tt.args.rCtx)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestGetCountryCodes(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGeoDb := mock_geodb.NewMockGeography(ctrl)

	type args struct {
		ip             string
		geoInfoFetcher geodb.Geography
	}
	tests := []struct {
		name                  string
		args                  args
		setup                 func()
		wantISOCountryCode    string
		wantAlpha3CountryCode string
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
			wantAlpha3CountryCode: "AUS",
			wantISOCountryCode:    "AU",
		},
		{
			name: "geoDB_instance_missing",
			args: args{
				ip:             "1.179.71.255",
				geoInfoFetcher: nil,
			},
			setup:                 func() {},
			wantAlpha3CountryCode: "",
			wantISOCountryCode:    "",
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
			wantAlpha3CountryCode: "",
			wantISOCountryCode:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			m := OpenWrap{
				geoInfoFetcher: tt.args.geoInfoFetcher,
			}
			got1, got2 := m.getCountryCodes(tt.args.ip)
			assert.Equal(t, got1, tt.wantISOCountryCode)
			assert.Equal(t, got2, tt.wantAlpha3CountryCode)
		})
	}
}
