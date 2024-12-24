package openwrap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
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
		seatNonBids openrtb_ext.NonBidCollection
	}{
		{
			name: "empty_impbidctx",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: make(map[string][]openrtb_ext.NonBid),
				},
			},
			seatNonBids: openrtb_ext.NonBidCollection{},
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
			seatNonBids: openrtb_ext.NonBidCollection{},
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
			seatNonBids: getNonBids(map[string][]openrtb_ext.NonBidParams{"pubmatic": {{Bid: &openrtb2.Bid{ImpID: "imp1"}, NonBidReason: int(nbr.RequestBlockedPartnerThrottle)}}}),
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
			seatNonBids: getNonBids(map[string][]openrtb_ext.NonBidParams{
				"pubmatic": {
					{
						Bid: &openrtb2.Bid{
							ImpID: "imp1",
						},
						NonBidReason: int(nbr.RequestBlockedSlotNotMapped),
					},
				},
				"appnexus": {
					{
						Bid:          &openrtb2.Bid{ImpID: "imp1"},
						NonBidReason: int(nbr.RequestBlockedSlotNotMapped),
					},
				},
			}),
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
					},
					AdapterThrottleMap: map[string]struct{}{
						"appnexus": {},
					},
				},
			},
			seatNonBids: getNonBids(map[string][]openrtb_ext.NonBidParams{
				"pubmatic": {
					{
						Bid:          &openrtb2.Bid{ImpID: "imp1"},
						NonBidReason: int(nbr.RequestBlockedSlotNotMapped),
					},
				},
				"appnexus": {
					{
						Bid:          &openrtb2.Bid{ImpID: "imp1"},
						NonBidReason: int(nbr.RequestBlockedPartnerThrottle),
					},
				},
			}),
		},
		{
			name: "seatnonbid_should_be_updated_from_defaultbids_from_webs2s_endpoint",
			args: args{
				rctx: models.RequestCtx{
					Endpoint: models.EndpointWebS2S,
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"imp1": {
							"pubmatic": {
								{
									ID:    "bid-id-1",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorGeneral.Ptr(),
									},
								},
							},
						},
					},
				},
			},
			seatNonBids: getNonBids(map[string][]openrtb_ext.NonBidParams{
				"pubmatic": {
					{
						Bid:          &openrtb2.Bid{ImpID: "imp1"},
						NonBidReason: int(exchange.ErrorGeneral),
					},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareSeatNonBids(tt.args.rctx)
			assert.Equal(t, tt.seatNonBids, got, "mismatched seatnonbids")
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
				WinningBids: models.WinningBids{
					"imp1": []*models.OwBid{
						{
							ID: "bid-id-3",
						},
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
									Nbr: nbr.LossBidLostToHigherBid.Ptr(),
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
									Nbr: nbr.LossBidLostToHigherBid.Ptr(),
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
								Nbr: nbr.LossBidLostToHigherBid.Ptr(),
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
								Nbr: nbr.LossBidLostToHigherBid.Ptr(),
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
				WinningBids: models.WinningBids{
					"imp1": []*models.OwBid{
						{
							ID: "bid-id-3",
						},
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
									Nbr: nbr.LossBidLostToHigherBid.Ptr(),
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
									Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
								Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
								Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
				WinningBids: models.WinningBids{
					"imp1": []*models.OwBid{
						{
							ID: "bid-id-3",
						},
					},
					"imp2": []*models.OwBid{
						{
							ID: "bid-id-2",
						},
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
									Nbr: nbr.LossBidLostToHigherBid.Ptr(),
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
									Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
									Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
									Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
								Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
								Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
								Nbr: nbr.LossBidLostToDealBid.Ptr(),
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
								Nbr: nbr.LossBidLostToDealBid.Ptr(),
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

func getNonBids(bidParamsMap map[string][]openrtb_ext.NonBidParams) openrtb_ext.NonBidCollection {
	nonBids := openrtb_ext.NonBidCollection{}
	for bidder, bidParams := range bidParamsMap {
		for _, bidParam := range bidParams {
			nonBid := openrtb_ext.NewNonBid(bidParam)
			nonBids.AddBid(nonBid, bidder)
		}
	}
	return nonBids
}

func TestUpdateSeatNonBidsFromDefaultBids(t *testing.T) {
	type args struct {
		rctx       models.RequestCtx
		seatNonBid *openrtb_ext.NonBidCollection
	}
	tests := []struct {
		name           string
		args           args
		wantSeatNonBid []openrtb_ext.SeatNonBid
	}{
		{
			name: "no default bids",
			args: args{
				rctx: models.RequestCtx{
					DefaultBids: nil,
				},
				seatNonBid: &openrtb_ext.NonBidCollection{},
			},
			wantSeatNonBid: nil,
		},
		{
			name: "imp not present in impbidctx for default bid",
			args: args{
				rctx: models.RequestCtx{
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"imp1": {
							"pubmatic": {
								{
									ID:    "bid-id-1",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp2": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorGeneral.Ptr(),
									},
								},
							},
						},
					},
				},
				seatNonBid: &openrtb_ext.NonBidCollection{},
			},
		},
		{
			name: "bid absent in impbidctx for default bid",
			args: args{
				rctx: models.RequestCtx{
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"imp1": {
							"pubmatic": {
								{
									ID:    "bid-id-1",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-2": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorGeneral.Ptr(),
									},
								},
							},
						},
					},
				},
				seatNonBid: &openrtb_ext.NonBidCollection{},
			},
		},
		{
			name: "default bid with no non-bid reason",
			args: args{
				rctx: models.RequestCtx{
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"imp1": {
							"pubmatic": {
								{
									ID:    "bid-id-1",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										Nbr: nil,
									},
								},
							},
						},
					},
				},
				seatNonBid: &openrtb_ext.NonBidCollection{},
			},
			wantSeatNonBid: nil,
		},
		{
			name: "singal default bid",
			args: args{
				rctx: models.RequestCtx{
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"imp1": {
							"pubmatic": {
								{
									ID:    "bid-id-1",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorGeneral.Ptr(),
									},
								},
							},
						},
					},
				},
				seatNonBid: &openrtb_ext.NonBidCollection{},
			},
			wantSeatNonBid: []openrtb_ext.SeatNonBid{
				{
					Seat: "pubmatic",
					NonBid: []openrtb_ext.NonBid{
						{
							ImpId:      "imp1",
							StatusCode: 100,
						},
					},
				},
			},
		},
		{
			name: "multiple default bids for same imp",
			args: args{
				rctx: models.RequestCtx{
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"imp1": {
							"pubmatic": {
								{
									ID:    "bid-id-1",
									ImpID: "imp1",
								},
							},
							"appnexus": {
								{
									ID:    "bid-id-2",
									ImpID: "imp1",
								},
							},
							"rubicon": {
								{
									ID:    "bid-id-3",
									ImpID: "imp1",
								},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorGeneral.Ptr(),
									},
								},
								"bid-id-2": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorTimeout.Ptr(),
									},
								},
								"bid-id-3": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorBidderUnreachable.Ptr(),
									},
								},
							},
						},
					},
				},
				seatNonBid: &openrtb_ext.NonBidCollection{},
			},
			wantSeatNonBid: []openrtb_ext.SeatNonBid{
				{
					Seat: "pubmatic",
					NonBid: []openrtb_ext.NonBid{
						{
							ImpId:      "imp1",
							StatusCode: 100,
						},
					},
				},
				{
					Seat: "appnexus",
					NonBid: []openrtb_ext.NonBid{
						{
							ImpId:      "imp1",
							StatusCode: 101,
						},
					},
				},
				{
					Seat: "rubicon",
					NonBid: []openrtb_ext.NonBid{
						{
							ImpId:      "imp1",
							StatusCode: 103,
						},
					},
				},
			},
		},
		{
			name: "multiple default bids for different imp",
			args: args{
				rctx: models.RequestCtx{
					DefaultBids: map[string]map[string][]openrtb2.Bid{
						"imp1": {
							"pubmatic": {
								{
									ID:    "bid-id-1",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
							"appnexus": {
								{
									ID:    "bid-id-2",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
							"rubicon": {
								{
									ID:    "bid-id-3",
									ImpID: "imp1",
									Ext:   []byte(`{}`),
								},
							},
						},
						"imp2": {
							"pubmatic": {
								{
									ID:    "bid-id-4",
									ImpID: "imp2",
									Ext:   []byte(`{}`),
								},
							},
							"appnexus": {
								{
									ID:    "bid-id-5",
									ImpID: "imp2",
									Ext:   []byte(`{}`),
								},
							},
							"rubicon": {
								{
									ID:    "bid-id-6",
									ImpID: "imp2",
									Ext:   []byte(`{}`),
								},
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorGeneral.Ptr(),
									},
								},
								"bid-id-2": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorTimeout.Ptr(),
									},
								},
								"bid-id-3": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorBidderUnreachable.Ptr(),
									},
								},
							},
						},
						"imp2": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-4": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorGeneral.Ptr(),
									},
								},
								"bid-id-5": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorTimeout.Ptr(),
									},
								},
								"bid-id-6": {
									BidExt: models.BidExt{
										Nbr: exchange.ErrorBidderUnreachable.Ptr(),
									},
								},
							},
						},
					},
				},
				seatNonBid: &openrtb_ext.NonBidCollection{},
			},
			wantSeatNonBid: []openrtb_ext.SeatNonBid{
				{
					Seat: "pubmatic",
					NonBid: []openrtb_ext.NonBid{
						{
							ImpId:      "imp1",
							StatusCode: 100,
						},
						{
							ImpId:      "imp2",
							StatusCode: 100,
						},
					},
				},
				{
					Seat: "appnexus",
					NonBid: []openrtb_ext.NonBid{
						{
							ImpId:      "imp1",
							StatusCode: 101,
						},
						{
							ImpId:      "imp2",
							StatusCode: 101,
						},
					},
				},
				{
					Seat: "rubicon",
					NonBid: []openrtb_ext.NonBid{
						{
							ImpId:      "imp1",
							StatusCode: 103,
						},
						{
							ImpId:      "imp2",
							StatusCode: 103,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateSeatNonBidsFromDefaultBids(tt.args.rctx, tt.args.seatNonBid)
			gotSetaNonBid := tt.args.seatNonBid.Get()

			cmp.Equal(tt.wantSeatNonBid, gotSetaNonBid,
				cmpopts.SortSlices(func(a, b openrtb_ext.SeatNonBid) bool {
					return a.Seat < b.Seat
				}),
				cmpopts.SortSlices(sortNonBids),
			)
		})
	}
}

func sortNonBids(a, b openrtb_ext.NonBid) bool {
	return a.ImpId < b.ImpId
}
