package openwrap

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/currency"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb"
	metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/profilemetadata"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/publisherfeature"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/unwrap"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/uuidutil"
	"github.com/stretchr/testify/assert"
)

const fakeUuid = "30470a14-2949-4110-abce-b62d57304ad5"

type TestUUIDGenerator struct{}

func (TestUUIDGenerator) Generate() (string, error) {
	return fakeUuid, nil
}

func TestGetNonBRCodeFromBidRespExt(t *testing.T) {
	type args struct {
		bidder         string
		bidResponseExt openrtb_ext.ExtBidResponse
	}
	tests := []struct {
		name string
		args args
		nbr  *openrtb3.NoBidReason
	}{
		{
			name: "bidResponseExt.Errors_is_empty",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: nil,
				},
			},
			nbr: openrtb3.NoBidUnknownError.Ptr(),
		},
		{
			name: "invalid_partner_err",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage{
						"pubmatic": {
							{
								Code: 0,
							},
						},
					},
				},
			},
			nbr: exchange.ErrorGeneral.Ptr(),
		},
		{
			name: "unknown_partner_err",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage{
						"pubmatic": {
							{
								Code: errortypes.UnknownErrorCode,
							},
						},
					},
				},
			},
			nbr: exchange.ErrorGeneral.Ptr(),
		},
		{
			name: "partner_timeout_err",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage{
						"pubmatic": {
							{
								Code: errortypes.TimeoutErrorCode,
							},
						},
					},
				},
			},
			nbr: exchange.ErrorTimeout.Ptr(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nbr := getNonBRCodeFromBidRespExt(tt.args.bidder, tt.args.bidResponseExt)
			assert.Equal(t, tt.nbr, nbr, tt.name)
		})
	}
}

func TestOpenWrap_addDefaultBids(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	type fields struct {
		cfg             config.Config
		rateConvertor   *currency.RateConverter
		metricEngine    metrics.MetricsEngine
		geoInfoFetcher  geodb.Geography
		pubFeatures     publisherfeature.Feature
		unwrap          unwrap.Unwrap
		profileMetaData profilemetadata.ProfileMetaData
		uuidGenerator   uuidutil.UUIDGenerator
	}
	type args struct {
		rctx           *models.RequestCtx
		bidResponse    *openrtb2.BidResponse
		bidResponseExt openrtb_ext.ExtBidResponse
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   map[string]map[string][]openrtb2.Bid
	}{
		{
			name: "EndpointWebS2S do not add default bids for slot-not-mapped and partner-throttled",
			fields: fields{
				metricEngine:  mockEngine,
				uuidGenerator: TestUUIDGenerator{},
			},
			args: args{
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointWebS2S,
					ImpBidCtx: map[string]models.ImpCtx{
						"imp-1": {
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
								},
								"openx": {
									PrebidBidderCode: "openx",
								},
							},
							NonMapped: map[string]struct{}{
								"appnexus": {},
							},
							BidCtx: map[string]models.BidCtx{
								"pubmatic-bid-1": {
									BidExt: models.BidExt{},
								},
							},
						},
					},
					AdapterThrottleMap: map[string]struct{}{
						"rubicon": {},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "bid-1",
					SeatBid: []openrtb2.SeatBid{
						{
							Seat: "pubmatic",
							Bid: []openrtb2.Bid{
								{
									ID:    "pubmatic-bid-1",
									ImpID: "imp-1",
									Price: 1.0,
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockEngine.EXPECT().RecordPartnerResponseErrors(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			},
			want: map[string]map[string][]openrtb2.Bid{
				"imp-1": {
					"openx": {
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "imp-1",
							Ext:   []byte(`{}`),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			m := &OpenWrap{
				cfg:             tt.fields.cfg,
				metricEngine:    tt.fields.metricEngine,
				rateConvertor:   tt.fields.rateConvertor,
				geoInfoFetcher:  tt.fields.geoInfoFetcher,
				pubFeatures:     tt.fields.pubFeatures,
				unwrap:          tt.fields.unwrap,
				profileMetaData: tt.fields.profileMetaData,
				uuidGenerator:   tt.fields.uuidGenerator,
			}
			got := m.addDefaultBids(tt.args.rctx, tt.args.bidResponse, tt.args.bidResponseExt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOpenWrap_addDefaultBidsForMultiFloorsConfig(t *testing.T) {
	type fields struct {
		cfg             config.Config
		metricEngine    metrics.MetricsEngine
		rateConvertor   *currency.RateConverter
		geoInfoFetcher  geodb.Geography
		pubFeatures     publisherfeature.Feature
		unwrap          unwrap.Unwrap
		profileMetaData profilemetadata.ProfileMetaData
		uuidGenerator   uuidutil.UUIDGenerator
	}
	type args struct {
		rctx           *models.RequestCtx
		bidResponse    *openrtb2.BidResponse
		bidResponseExt openrtb_ext.ExtBidResponse
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]map[string][]openrtb2.Bid
	}{
		{
			name: "request is other than applovinmax",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint:    models.EndpointWebS2S,
					DefaultBids: map[string]map[string][]openrtb2.Bid{},
				},
				bidResponse: &openrtb2.BidResponse{
					ID:      "bid-1",
					SeatBid: []openrtb2.SeatBid{},
				},
			},
			want: map[string]map[string][]openrtb2.Bid{},
		},
		{
			name: "request is applovinmax but the multi-floors config is not enabled from DB",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint:    models.EndpointAppLovinMax,
					DefaultBids: map[string]map[string][]openrtb2.Bid{},
					AppLovinMax: models.AppLovinMax{
						MultiFloorsConfig: models.MultiFloorsConfig{
							Enabled: false,
						},
					},
				},
			},
			want: map[string]map[string][]openrtb2.Bid{},
		},
		{
			name: "mulit-floors config have three floors and no bids in the response",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"test-impID-1": {
							"pubmatic": {
								{
									ID:    "dbbsdhkldks1234",
									ImpID: "test-impID-1",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					AppLovinMax: models.AppLovinMax{
						MultiFloorsConfig: models.MultiFloorsConfig{
							Enabled: true,
							Config: models.ApplovinAdUnitFloors{
								"adunit-1": []float64{1.1, 2.1, 3.1},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"test-impID-1": {
							TagID: "adunit-1",
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
								},
							},
							BidCtx: map[string]models.BidCtx{},
						},
					},
					PrebidBidderCode: map[string]string{
						"pubmatic": "pubmatic",
					},
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "bid-1",
				},
			},
			fields: fields{
				uuidGenerator: TestUUIDGenerator{},
			},
			want: map[string]map[string][]openrtb2.Bid{
				"test-impID-1": {
					"pubmatic": {
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
					},
				},
			},
		},
		{
			name: "mulit-floors config do not have adunit configured and no bids in the response",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"test-impID-1": {
							"pubmatic": {
								{
									ID:    "dbbsdhkldks1234",
									ImpID: "test-impID-1",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					AppLovinMax: models.AppLovinMax{
						MultiFloorsConfig: models.MultiFloorsConfig{
							Enabled: true,
							Config: models.ApplovinAdUnitFloors{
								"adunit-1": []float64{1.1, 2.1, 3.1},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"test-impID-1": {
							TagID: "adunit-2",
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
								},
							},
							BidCtx: map[string]models.BidCtx{},
						},
					},
					PrebidBidderCode: map[string]string{
						"pubmatic": "pubmatic",
					},
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "bid-1",
				},
			},
			fields: fields{
				uuidGenerator: TestUUIDGenerator{},
			},
			want: map[string]map[string][]openrtb2.Bid{
				"test-impID-1": {
					"pubmatic": {
						{
							ID:    "dbbsdhkldks1234",
							ImpID: "test-impID-1",
							Ext:   []byte(`{}`),
						},
					},
				},
			},
		},
		{
			name: "mulit-floors config have three floors and only one bid in the response",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint:    models.EndpointAppLovinMax,
					DefaultBids: map[string]map[string][]openrtb2.Bid{},
					AppLovinMax: models.AppLovinMax{
						MultiFloorsConfig: models.MultiFloorsConfig{
							Enabled: true,
							Config: models.ApplovinAdUnitFloors{
								"adunit-1": []float64{1.1, 2.1, 3.1},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"test-impID-1": {
							TagID: "adunit-1",
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
								},
							},
							BidCtx: map[string]models.BidCtx{
								"pubmatic-bid-1": {
									BidExt: models.BidExt{
										MultiBidMultiFloorValue: 1.1,
									},
								},
							},
						},
					},
					PrebidBidderCode: map[string]string{
						"pubmatic": "pubmatic",
					},
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "bid-1",
					SeatBid: []openrtb2.SeatBid{
						{
							Seat: "pubmatic",
							Bid: []openrtb2.Bid{
								{
									ID:    "pubmatic-bid-1",
									ImpID: "test-impID-1",
									Price: 1.5,
									Ext:   []byte(`{"mbmfv":1.1}`),
								},
							},
						},
					},
				},
			},
			fields: fields{
				uuidGenerator: TestUUIDGenerator{},
			},
			want: map[string]map[string][]openrtb2.Bid{
				"test-impID-1": {
					"pubmatic": {
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
					},
				},
			},
		},
		{
			name: "mulit-floors config have three floors and all three bid in the response",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint:    models.EndpointAppLovinMax,
					DefaultBids: map[string]map[string][]openrtb2.Bid{},
					AppLovinMax: models.AppLovinMax{
						MultiFloorsConfig: models.MultiFloorsConfig{
							Enabled: true,
							Config: models.ApplovinAdUnitFloors{
								"adunit-1": []float64{1.1, 2.1, 3.1},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"test-impID-1": {
							TagID: "adunit-1",
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
								},
							},
							BidCtx: map[string]models.BidCtx{
								"pubmatic-bid-1": {
									BidExt: models.BidExt{
										MultiBidMultiFloorValue: 1.1,
									},
								},
								"pubmatic-bid-2": {
									BidExt: models.BidExt{
										MultiBidMultiFloorValue: 2.1,
									},
								},
								"pubmatic-bid-3": {
									BidExt: models.BidExt{
										MultiBidMultiFloorValue: 3.1,
									},
								},
							},
						},
					},
					PrebidBidderCode: map[string]string{
						"pubmatic": "pubmatic",
					},
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "bid-1",
					SeatBid: []openrtb2.SeatBid{
						{
							Seat: "pubmatic",
							Bid: []openrtb2.Bid{
								{
									ID:    "pubmatic-bid-1",
									ImpID: "test-impID-1",
									Price: 1.5,
									Ext:   []byte(`{"mbmfv":1.1}`),
								},
								{
									ID:    "pubmatic-bid-2",
									ImpID: "test-impID-1",
									Price: 2.5,
									Ext:   []byte(`{"mbmfv":2.1}`),
								},
								{
									ID:    "pubmatic-bid-3",
									ImpID: "test-impID-1",
									Price: 3.5,
								},
							},
						},
					},
				},
			},
			fields: fields{
				uuidGenerator: TestUUIDGenerator{},
			},
			want: map[string]map[string][]openrtb2.Bid{
				"test-impID-1": {},
			},
		},
		{
			name: "mulit-floors config have three floors and only one bid in the response for both partner pubmatic and pubmatic_1123",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint:    models.EndpointAppLovinMax,
					DefaultBids: map[string]map[string][]openrtb2.Bid{},
					AppLovinMax: models.AppLovinMax{
						MultiFloorsConfig: models.MultiFloorsConfig{
							Enabled: true,
							Config: models.ApplovinAdUnitFloors{
								"adunit-1": []float64{1.1, 2.1, 3.1},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"test-impID-1": {
							TagID: "adunit-1",
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
								},
								"pubmatic_1123": {
									PrebidBidderCode: "pubmatic",
								},
							},
							BidCtx: map[string]models.BidCtx{
								"pubmatic-bid-1": {
									BidExt: models.BidExt{
										MultiBidMultiFloorValue: 1.1,
									},
								},
								"pubmatic-bid-2": {
									BidExt: models.BidExt{
										MultiBidMultiFloorValue: 1.1,
									},
								},
							},
						},
					},
					PrebidBidderCode: map[string]string{
						"pubmatic_1123": "pubmatic",
						"pubmatic":      "pubmatic",
					},
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "bid-1",
					SeatBid: []openrtb2.SeatBid{
						{
							Seat: "pubmatic",
							Bid: []openrtb2.Bid{
								{
									ID:    "pubmatic-bid-1",
									ImpID: "test-impID-1",
									Price: 1.5,
									Ext:   []byte(`{"mbmfv":1.1}`),
								},
							},
						},
						{
							Seat: "pubmatic_1123",
							Bid: []openrtb2.Bid{
								{
									ID:    "pubmatic-bid-2",
									ImpID: "test-impID-1",
									Price: 1.6,
									Ext:   []byte(`{"mbmfv":1.1}`),
								},
							},
						},
					},
				},
			},
			fields: fields{
				uuidGenerator: TestUUIDGenerator{},
			},
			want: map[string]map[string][]openrtb2.Bid{
				"test-impID-1": {
					"pubmatic": {
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
					},
					"pubmatic_1123": {
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
					},
				},
			},
		},
		{
			name: "mulit-floors config have three floors and no bid in the response for both partner pubmatic and pubmatic_1123",
			args: args{
				rctx: &models.RequestCtx{
					Endpoint:    models.EndpointAppLovinMax,
					DefaultBids: map[string]map[string][]openrtb2.Bid{},
					AppLovinMax: models.AppLovinMax{
						MultiFloorsConfig: models.MultiFloorsConfig{
							Enabled: true,
							Config: models.ApplovinAdUnitFloors{
								"adunit-1": []float64{1.1, 2.1, 3.1},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"test-impID-1": {
							TagID: "adunit-1",
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
								},
								"pubmatic_1123": {
									PrebidBidderCode: "pubmatic",
								},
							},
							BidCtx: map[string]models.BidCtx{},
						},
					},
					PrebidBidderCode: map[string]string{
						"pubmatic_1123": "pubmatic",
						"pubmatic":      "pubmatic",
					},
				},
				bidResponse: &openrtb2.BidResponse{
					ID:      "bid-1",
					SeatBid: []openrtb2.SeatBid{},
				},
			},
			fields: fields{
				uuidGenerator: TestUUIDGenerator{},
			},
			want: map[string]map[string][]openrtb2.Bid{
				"test-impID-1": {
					"pubmatic": {
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
					},
					"pubmatic_1123": {
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
						{
							ID:    "30470a14-2949-4110-abce-b62d57304ad5",
							ImpID: "test-impID-1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OpenWrap{
				cfg:             tt.fields.cfg,
				metricEngine:    tt.fields.metricEngine,
				rateConvertor:   tt.fields.rateConvertor,
				geoInfoFetcher:  tt.fields.geoInfoFetcher,
				pubFeatures:     tt.fields.pubFeatures,
				unwrap:          tt.fields.unwrap,
				profileMetaData: tt.fields.profileMetaData,
				uuidGenerator:   tt.fields.uuidGenerator,
			}
			got := m.addDefaultBidsForMultiFloorsConfig(tt.args.rctx, tt.args.bidResponse, tt.args.bidResponseExt)
			assert.Equal(t, tt.want, got)
		})
	}
}
