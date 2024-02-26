package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestPrepareSeatNonBids(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}

	tests := []struct {
		name        string
		args        args
		seatNonBids map[string][]openrtb_ext.NonBid
	}{
		{
			name: "empty_impbidctx",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: make(map[string][]openrtb_ext.NonBid),
				},
			},
			seatNonBids: make(map[string][]openrtb_ext.NonBid),
		},
		{
			name: "empty_seatnonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							ImpID: "imp1",
						},
					},
					SeatNonBids: make(map[string][]openrtb_ext.NonBid),
				},
			},
			seatNonBids: make(map[string][]openrtb_ext.NonBid),
		},
		{
			name: "partner_throttled_nonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							ImpID: "imp1",
						},
					},
					AdapterThrottleMap: map[string]struct{}{
						"pubmatic": {},
					},
					SeatNonBids: map[string][]openrtb_ext.NonBid{},
				},
			},
			seatNonBids: map[string][]openrtb_ext.NonBid{
				"pubmatic": {
					openrtb_ext.NonBid{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedPartnerThrottle),
					},
				},
			},
		},
		{
			name: "slot_not_mapped_nonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							NonMapped: map[string]struct{}{
								"pubmatic": {},
								"appnexus": {},
							},
						},
					},
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							{
								ImpId:      "imp2",
								StatusCode: 2,
							},
						},
					},
				},
			},
			seatNonBids: map[string][]openrtb_ext.NonBid{
				"pubmatic": {
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedSlotNotMapped),
					},
				},
				"appnexus": {
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedSlotNotMapped),
					},
				},
			},
		},
		{
			name: "slot_not_mapped_plus_partner_throttled_nonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							NonMapped: map[string]struct{}{
								"pubmatic": {},
							},
						},
						"imp2": {},
					},
					AdapterThrottleMap: map[string]struct{}{
						"appnexus": {},
					},
				},
			},
			seatNonBids: map[string][]openrtb_ext.NonBid{
				"pubmatic": {
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedSlotNotMapped),
					},
				},
				"appnexus": {
					{
						ImpId:      "imp2",
						StatusCode: int(exchange.RequestBlockedPartnerThrottle),
					},
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedPartnerThrottle),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seatNonBids := prepareSeatNonBids(tt.args.rctx)
			assert.Equal(t, len(seatNonBids), len(tt.seatNonBids))
			for k, v := range seatNonBids {
				// ignore order of elements in slice while comparing
				assert.ElementsMatch(t, v, tt.seatNonBids[k], tt.name)
			}
		})
	}
}

func TestAddSeatNonBidsInResponseExt(t *testing.T) {
	type args struct {
		rctx        models.RequestCtx
		responseExt *openrtb_ext.ExtBidResponse
	}

	tests := []struct {
		name string
		args args
		want *openrtb_ext.ExtBidResponse
	}{
		{
			name: "empty_rtcx_seatnonbids",
			args: args{
				rctx: models.RequestCtx{},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: nil,
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: nil,
			},
		},
		{
			name: "response_ext_prebid_is_nil",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
						},
					},
				},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: nil,
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: &openrtb_ext.ExtResponsePrebid{
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
							Seat: "pubmatic",
						},
					},
				},
			},
		},
		{
			name: "prebid_exist_but_seatnonbid_is_empty_in_ext",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
							openrtb_ext.NonBid{
								ImpId:      "imp2",
								StatusCode: 2,
							},
						},
					},
				},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						AuctionTimestamp: 100,
					},
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: &openrtb_ext.ExtResponsePrebid{
					AuctionTimestamp: 100,
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
								{
									ImpId:      "imp2",
									StatusCode: 2,
								},
							},
							Seat: "pubmatic",
						},
					},
				},
			},
		},
		{
			name: "nonbid_exist_in_rctx_and_in_ext_for_specific_bidder",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
						},
					},
				},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						AuctionTimestamp: 100,
						SeatNonBid: []openrtb_ext.SeatNonBid{
							{
								Seat: "pubmatic",
								NonBid: []openrtb_ext.NonBid{
									{
										ImpId:      "imp2",
										StatusCode: 2,
									},
								},
							},
							{
								Seat: "appnexus",
								NonBid: []openrtb_ext.NonBid{
									{
										ImpId:      "imp1",
										StatusCode: 1,
									},
								},
							},
						},
					},
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: &openrtb_ext.ExtResponsePrebid{
					AuctionTimestamp: 100,
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "pubmatic",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp2",
									StatusCode: 2,
								},
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
						},
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "nonbid_exist_in_rctx_but_not_in_ext_for_specific_bidder",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
						},
						"appnexus": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
						},
					},
				},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						AuctionTimestamp: 100,
						SeatNonBid: []openrtb_ext.SeatNonBid{
							{
								Seat: "pubmatic",
								NonBid: []openrtb_ext.NonBid{
									{
										ImpId:      "imp2",
										StatusCode: 2,
									},
								},
							},
						},
					},
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: &openrtb_ext.ExtResponsePrebid{
					AuctionTimestamp: 100,
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "pubmatic",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp2",
									StatusCode: 2,
								},
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
						},
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: 1,
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
			addSeatNonBidsInResponseExt(tt.args.rctx, tt.args.responseExt)
			assert.Equal(t, tt.want, tt.args.responseExt, tt.name)
		})
	}
}

func TestAddLostToDealBidNonBRCode(t *testing.T) {
	tests := []struct {
		name      string
		rctx      *models.RequestCtx
		impBidCtx map[string]models.ImpCtx
	}{
		{
			name: "support deal flag is false",
			rctx: &models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
								},
							},
						},
					},
				},
			},
			impBidCtx: map[string]models.ImpCtx{
				"imp1": {
					BidCtx: map[string]models.BidCtx{
						"bid-id-1": {
							BidExt: models.BidExt{
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "no winning bid for imp so dont update NonBR code",
			rctx: &models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
								},
							},
						},
					},
				},
				SupportDeals: true,
			},
			impBidCtx: map[string]models.ImpCtx{
				"imp1": {
					BidCtx: map[string]models.BidCtx{
						"bid-id-1": {
							BidExt: models.BidExt{
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "do not update LossBidLostToHigherBid NonBR code if bid satisifies dealTier",
			rctx: &models.RequestCtx{
				WinningBids: map[string]models.OwBid{
					"imp1": {
						ID: "bid-id-3",
					},
				},
				ImpBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									NetECPM: 10,
									ExtBid: openrtb_ext.ExtBid{

										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									NetECPM: 50,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									NetECPM: 100,
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
				SupportDeals: true,
			},
			impBidCtx: map[string]models.ImpCtx{
				"imp1": {
					BidCtx: map[string]models.BidCtx{
						"bid-id-1": {
							BidExt: models.BidExt{
								NetECPM: 10,
								ExtBid: openrtb_ext.ExtBid{

									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: true,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
							},
						},
						"bid-id-2": {
							BidExt: models.BidExt{
								NetECPM: 50,
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: true,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
							},
						},
						"bid-id-3": {
							BidExt: models.BidExt{
								NetECPM: 100,
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
		{
			name: "update LossBidLostToHigherBid NonBR code if bid not satisifies dealTier",
			rctx: &models.RequestCtx{
				WinningBids: map[string]models.OwBid{
					"imp1": {
						ID: "bid-id-3",
					},
				},
				ImpBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									NetECPM: 10,
									ExtBid: openrtb_ext.ExtBid{

										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									NetECPM: 100,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									NetECPM: 5,
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
				SupportDeals: true,
			},
			impBidCtx: map[string]models.ImpCtx{
				"imp1": {
					BidCtx: map[string]models.BidCtx{
						"bid-id-1": {
							BidExt: models.BidExt{
								NetECPM: 10,
								ExtBid: openrtb_ext.ExtBid{

									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
							},
						},
						"bid-id-2": {
							BidExt: models.BidExt{
								NetECPM: 100,
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
							},
						},
						"bid-id-3": {
							BidExt: models.BidExt{
								NetECPM: 5,
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
		{
			name: "test for multiple impression",
			rctx: &models.RequestCtx{
				WinningBids: map[string]models.OwBid{
					"imp1": {
						ID: "bid-id-3",
					},
					"imp2": {
						ID: "bid-id-2",
					},
				},
				ImpBidCtx: map[string]models.ImpCtx{
					"imp1": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									NetECPM: 10,
									ExtBid: openrtb_ext.ExtBid{

										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToHigherBid),
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									NetECPM: 100,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									NetECPM: 5,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
								},
							},
						},
					},
					"imp2": {
						BidCtx: map[string]models.BidCtx{
							"bid-id-1": {
								BidExt: models.BidExt{
									NetECPM: 10,
									ExtBid: openrtb_ext.ExtBid{

										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
								},
							},
							"bid-id-2": {
								BidExt: models.BidExt{
									NetECPM: 100,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: true,
										},
									},
								},
							},
							"bid-id-3": {
								BidExt: models.BidExt{
									NetECPM: 5,
									ExtBid: openrtb_ext.ExtBid{
										Prebid: &openrtb_ext.ExtBidPrebid{
											DealTierSatisfied: false,
										},
									},
									Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
								},
							},
						},
					},
				},
				SupportDeals: true,
			},
			impBidCtx: map[string]models.ImpCtx{
				"imp1": {
					BidCtx: map[string]models.BidCtx{
						"bid-id-1": {
							BidExt: models.BidExt{
								NetECPM: 10,
								ExtBid: openrtb_ext.ExtBid{

									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
							},
						},
						"bid-id-2": {
							BidExt: models.BidExt{
								NetECPM: 100,
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
							},
						},
						"bid-id-3": {
							BidExt: models.BidExt{
								NetECPM: 5,
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: true,
									},
								},
							},
						},
					},
				},
				"imp2": {
					BidCtx: map[string]models.BidCtx{
						"bid-id-1": {
							BidExt: models.BidExt{
								NetECPM: 10,
								ExtBid: openrtb_ext.ExtBid{

									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
							},
						},
						"bid-id-2": {
							BidExt: models.BidExt{
								NetECPM: 100,
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: true,
									},
								},
							},
						},
						"bid-id-3": {
							BidExt: models.BidExt{
								NetECPM: 5,
								ExtBid: openrtb_ext.ExtBid{
									Prebid: &openrtb_ext.ExtBidPrebid{
										DealTierSatisfied: false,
									},
								},
								Nbr: GetNonBidStatusCodePtr(openrtb3.LossBidLostToDealBid),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addLostToDealBidNonBRCode(tt.rctx)
			assert.Equal(t, tt.impBidCtx, tt.rctx.ImpBidCtx, tt.name)
		})
	}
}
