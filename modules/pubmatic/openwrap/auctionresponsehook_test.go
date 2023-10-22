package openwrap

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/hooks/hookstage"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestSeatNonBidsInHandleAuctionResponseHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		moduleCtx hookstage.ModuleInvocationContext
		payload   hookstage.AuctionResponsePayload
	}

	type want struct {
		bidResponseExt json.RawMessage
		err            error
	}

	tests := []struct {
		name             string
		args             args
		want             want
		getMetricsEngine func() *mock_metrics.MockMetricsEngine
	}{
		{
			name: "returnallbidstatus_true",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime:          time.Now().UnixMilli(),
							ReturnAllBidStatus: true,
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							AdapterThrottleMap: map[string]struct{}{
								"pubmatic": {},
							},
							PubIDStr: "5890",
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordNobidErrPrebidServerResponse("5890")
				return mockEngine
			},
			want: want{
				bidResponseExt: json.RawMessage(`{"prebid":{"seatnonbid":[{"nonbid":[{"impid":"imp1","statuscode":504,"ext":{"prebid":{"bid":{"id":""}}}}],"seat":"pubmatic","ext":null}]},"matchedimpression":{}}`),
			},
		},
		{
			name: "returnallbidstatus_false",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime:          time.Now().UnixMilli(),
							ReturnAllBidStatus: false,
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							AdapterThrottleMap: map[string]struct{}{
								"pubmatic": {},
							},
							PubIDStr: "5890",
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordNobidErrPrebidServerResponse("5890")
				return mockEngine
			},
			want: want{
				bidResponseExt: json.RawMessage(`{"matchedimpression":{}}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OpenWrap{
				metricEngine: tt.getMetricsEngine(),
			}
			hookResult, err := o.handleAuctionResponseHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.want.err, err, tt.name)
			mutations := hookResult.ChangeSet.Mutations()
			assert.NotEmpty(t, mutations, tt.name)
			for _, mut := range mutations {
				result, err := mut.Apply(tt.args.payload)
				assert.Nil(t, err, tt.name)
				assert.Equal(t, tt.want.bidResponseExt, result.BidResponse.Ext, tt.name)
			}
		})
	}
}

func TestNonBRCodesInHandleAuctionResponseHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		moduleCtx hookstage.ModuleInvocationContext
		payload   hookstage.AuctionResponsePayload
	}
	type want struct {
		impBidCtx map[string]models.ImpCtx
	}
	tests := []struct {
		name             string
		args             args
		want             want
		getMetricsEngine func() *mock_metrics.MockMetricsEngine
	}{
		{
			name: "single bid and supportdeal is false",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr: "5890",
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "pubmatic",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      5,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test auction between 3 bids when supportdeal is false and no bid satisfies dealTier",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr: "5890",
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "pubmatic",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-2",
										ImpID: "imp1",
										Price: 20,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "appnexus",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-3",
										ImpID: "imp1",
										Price: 10,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "rubicon",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "rubicon")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
									CreativeType: models.Banner,
									NetECPM:      5,
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      20,
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
									CreativeType: models.Banner,
									NetECPM:      10,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test auction between 3 bids when supportdeal is false and all bids satisfies dealTier",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr: "5890",
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "pubmatic",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-2",
										ImpID: "imp1",
										Price: 20,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "appnexus",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-3",
										ImpID: "imp1",
										Price: 10,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "rubicon",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "rubicon")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
									CreativeType: models.Banner,
									NetECPM:      5,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      20,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
									CreativeType: models.Banner,
									NetECPM:      10,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
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
			name: "single bid and supportdeal is true",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr:     "5890",
							SupportDeals: true,
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "pubmatic",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      5,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "auction between 3 bids when supportdeal is true and no bid satisfies dealTier",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr:     "5890",
							SupportDeals: true,
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 20,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "pubmatic",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-2",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "appnexus",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-3",
										ImpID: "imp1",
										Price: 10,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "rubicon",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "rubicon")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      20,
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
									CreativeType: models.Banner,
									NetECPM:      5,
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
									CreativeType: models.Banner,
									NetECPM:      10,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "auction between 3 bids when supportdeal is true and only middle bid satisfies dealTier",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr:     "5890",
							SupportDeals: true,
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 20,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "pubmatic",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-2",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "appnexus",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-3",
										ImpID: "imp1",
										Price: 10,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "rubicon",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "rubicon")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
									CreativeType: models.Banner,
									NetECPM:      20,
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      5,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
									CreativeType: models.Banner,
									NetECPM:      10,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "auction between 3 bids when supportdeal is true and only last bid satisfies dealTier",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr:     "5890",
							SupportDeals: true,
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 20,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "pubmatic",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-2",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "appnexus",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-3",
										ImpID: "imp1",
										Price: 10,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "rubicon",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "rubicon")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
									CreativeType: models.Banner,
									NetECPM:      20,
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      5,
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
									CreativeType: models.Banner,
									NetECPM:      10,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "auction between 3 bids when supportdeal is true and only first bid satisfies dealTier",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr:     "5890",
							SupportDeals: true,
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 20,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "pubmatic",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-2",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "appnexus",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-3",
										ImpID: "imp1",
										Price: 10,
										Ext:   json.RawMessage(`{}`),
									},
								},
								Seat: "rubicon",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "rubicon")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
									CreativeType: models.Banner,
									NetECPM:      20,
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									CreativeType: models.Banner,
									NetECPM:      5,
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
									CreativeType: models.Banner,
									NetECPM:      10,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "auction between 3 bids when supportdeal is true and all bids satisfies dealTier",
			args: args{
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime: time.Now().UnixMilli(),
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							PubIDStr:     "5890",
							SupportDeals: true,
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						Cur: "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 20,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "pubmatic",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-2",
										ImpID: "imp1",
										Price: 5,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "appnexus",
							},
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-3",
										ImpID: "imp1",
										Price: 10,
										Ext:   json.RawMessage(`{"prebid":{"dealtiersatisfied":true}}`),
									},
								},
								Seat: "rubicon",
							},
						},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "pubmatic")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats("", "5890", "rubicon")
				return mockEngine
			},
			want: want{
				impBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
									CreativeType: models.Banner,
									NetECPM:      20,
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
									CreativeType: models.Banner,
									NetECPM:      5,
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
									Nbr:          GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
									CreativeType: models.Banner,
									NetECPM:      10,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OpenWrap{
				metricEngine: tt.getMetricsEngine(),
			}
			hookResult, _ := o.handleAuctionResponseHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			mutations := hookResult.ChangeSet.Mutations()
			assert.NotEmpty(t, mutations, tt.name)
			rctxInterface := hookResult.AnalyticsTags.Activities[0].Results[0].Values["request-ctx"]
			rctx := rctxInterface.(*models.RequestCtx)
			assert.Equal(t, tt.want.impBidCtx, rctx.ImpBidCtx, tt.name)
		})
	}
}
