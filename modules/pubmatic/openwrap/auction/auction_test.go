package auction

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

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
