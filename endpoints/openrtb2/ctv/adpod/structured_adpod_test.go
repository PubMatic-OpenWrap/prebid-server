package adpod

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
)

func TestStructuredAdpodPerformAuctionAndExclusion(t *testing.T) {
	type fields struct {
		AdpodCtx          AdpodCtx
		ImpBidMap         map[string][]*types.Bid
		WinningBid        map[string]types.Bid
		CategoryExclusion bool
	}
	tests := []struct {
		name           string
		fields         fields
		wantWinningBid map[string]types.Bid
	}{
		{
			name: "only_price_based_auction_with_no_exclusion",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 2,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 5,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
							},
							DealTierSatisfied: false,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 6,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 10,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 5,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 6,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
					},
					DealTierSatisfied: false,
					Seat:              "god",
				},
			},
		},
		{
			name: "only_price_based_auction_with_exclusion",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 6,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 3,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 10,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid:        make(map[string]types.Bid),
				CategoryExclusion: true,
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 3,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: false,
					Seat:              "god",
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_no_exclusion",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 6,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: true,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 3,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 10,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid:        make(map[string]types.Bid),
				CategoryExclusion: false,
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 1,
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "index",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "god",
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 6,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 3,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 10,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid:        make(map[string]types.Bid),
				CategoryExclusion: true,
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 3,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "god",
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_1",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 6,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 3,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 8,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 9,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid:        make(map[string]types.Bid),
				CategoryExclusion: true,
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 6,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 9,
					},
					DealTierSatisfied: false,
					Seat:              "god",
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_2",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 6,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 3,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 8,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 9,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid:        make(map[string]types.Bid),
				CategoryExclusion: true,
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 6,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 9,
					},
					DealTierSatisfied: false,
					Seat:              "god",
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_3",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 6,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 3,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 8,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 9,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid:        make(map[string]types.Bid),
				CategoryExclusion: true,
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 3,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 8,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "god",
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_4",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: PodTypeStructured,
				},
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								Price: 6,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 1,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-2"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 3,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
					"imp3": {
						{
							Bid: &openrtb2.Bid{
								Price: 8,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: true,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 10,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 9,
							},
							DealTierSatisfied: false,
							Seat:              "god",
						},
					},
				},
				WinningBid:        make(map[string]types.Bid),
				CategoryExclusion: true,
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 6,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 9,
					},
					DealTierSatisfied: false,
					Seat:              "god",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &StructuredAdpod{
				AdpodCtx:          tt.fields.AdpodCtx,
				ImpBidMap:         tt.fields.ImpBidMap,
				WinningBid:        tt.fields.WinningBid,
				CategoryExclusion: tt.fields.CategoryExclusion,
			}
			sa.PerformAuctionAndExclusion()

			assert.Equal(t, sa.WinningBid, tt.wantWinningBid, "Auction failed")
		})
	}
}
