package adpod

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestStructuredAdpodPerformAuctionAndExclusion(t *testing.T) {
	type fields struct {
		AdpodCtx   AdpodCtx
		ImpBidMap  map[string][]*types.Bid
		WinningBid map[string]types.Bid
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
					Type: Structured,
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
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 6,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
					},
					DealTierSatisfied: false,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
		{
			name: "only_price_based_auction_with_exclusion",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: Structured,
					Exclusion: Exclusion{
						IABCategoryExclusion:      true,
						AdvertiserDomainExclusion: true,
					},
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
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 3,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: false,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_no_exclusion",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: Structured,
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
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 1,
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "index",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: Structured,
					Exclusion: Exclusion{
						IABCategoryExclusion:      true,
						AdvertiserDomainExclusion: true,
					},
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
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 3,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 10,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_1",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: Structured,
					Exclusion: Exclusion{
						IABCategoryExclusion:      true,
						AdvertiserDomainExclusion: true,
					},
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
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 6,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 9,
					},
					DealTierSatisfied: false,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_2",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: Structured,
					Exclusion: Exclusion{
						IABCategoryExclusion:      true,
						AdvertiserDomainExclusion: true,
					},
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
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 6,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 9,
					},
					DealTierSatisfied: false,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_3",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: Structured,
					Exclusion: Exclusion{
						IABCategoryExclusion:      true,
						AdvertiserDomainExclusion: true,
					},
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
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 3,
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 8,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
		{
			name: "price_and_deal_based_auction_with_exclusion_better_price_4",
			fields: fields{
				AdpodCtx: AdpodCtx{
					Type: Structured,
					Exclusion: Exclusion{
						IABCategoryExclusion:      true,
						AdvertiserDomainExclusion: true,
					},
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
				WinningBid: make(map[string]types.Bid),
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 6,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: true,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp2": {
					Bid: &openrtb2.Bid{
						Price: 4,
						Cat:   []string{"IAB-2"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
					Status:            constant.StatusWinningBid,
				},
				"imp3": {
					Bid: &openrtb2.Bid{
						Price: 9,
					},
					DealTierSatisfied: false,
					Seat:              "god",
					Status:            constant.StatusWinningBid,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &structuredAdpod{
				AdpodCtx:   tt.fields.AdpodCtx,
				ImpBidMap:  tt.fields.ImpBidMap,
				WinningBid: tt.fields.WinningBid,
			}
			sa.HoldAuction()

			assert.Equal(t, sa.WinningBid, tt.wantWinningBid, "Auction failed")
		})
	}
}

func TestStructuredAdpodGetSeatNonBid(t *testing.T) {
	type fields struct {
		ImpBidMap map[string][]*types.Bid
	}
	tests := []struct {
		name   string
		fields fields
		want   openrtb_ext.NonBidCollection
	}{
		{
			name: "Test Get Seat Non Bid",
			fields: fields{
				ImpBidMap: map[string][]*types.Bid{
					"imp1": {
						{
							Bid: &openrtb2.Bid{
								ID:    "BID-1",
								Price: 10,
							},
							ExtBid: openrtb_ext.ExtBid{
								Prebid: &openrtb_ext.ExtBidPrebid{
									Meta: &openrtb_ext.ExtBidPrebidMeta{
										AdapterCode: "pubmatic",
									},
									Type: "video",
								},
								OriginalBidCPM:    10,
								OriginalBidCur:    "USD",
								OriginalBidCPMUSD: 10,
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								ID:    "BID-2",
								Price: 15,
							},
							ExtBid: openrtb_ext.ExtBid{
								Prebid: &openrtb_ext.ExtBidPrebid{
									Meta: &openrtb_ext.ExtBidPrebidMeta{
										AdapterCode: "pubmatic",
									},
								},
							},
							Status:            constant.StatusWinningBid,
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
				},
			},
			want: func() openrtb_ext.NonBidCollection {
				seatNonBid := openrtb_ext.NonBidCollection{}
				nonBid := openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
					Bid:               &openrtb2.Bid{ID: "BID-1", Price: 10},
					NonBidReason:      501,
					OriginalBidCPM:    10,
					OriginalBidCur:    "USD",
					BidType:           "video",
					OriginalBidCPMUSD: 10,
					BidMeta: &openrtb_ext.ExtBidPrebidMeta{
						AdapterCode: "pubmatic",
					},
				})
				seatNonBid.AddBid(nonBid, "pubmatic")
				return seatNonBid
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &structuredAdpod{
				ImpBidMap: tt.fields.ImpBidMap,
			}
			snb := sa.GetSeatNonBid()
			assert.Equal(t, snb, tt.want)
		})
	}
}

func TestStructuredAdpodGetAdpodSeatBids(t *testing.T) {
	type fields struct {
		WinningBid map[string]types.Bid
	}
	tests := []struct {
		name   string
		fields fields
		want   []openrtb2.SeatBid
	}{
		{
			name: "Test Empty Bids in WinningBids",
			fields: fields{
				WinningBid: map[string]types.Bid{},
			},
			want: nil,
		},
		{
			name: "Test get adpod seat Bids",
			fields: fields{
				WinningBid: map[string]types.Bid{
					"imp1": {
						Bid: &openrtb2.Bid{
							Price: 5,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Status:            constant.StatusWinningBid,
					},
				},
			},
			want: []openrtb2.SeatBid{
				{
					Seat: "pubmatic",
					Bid: []openrtb2.Bid{
						{
							Price: 5,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &structuredAdpod{
				WinningBid: tt.fields.WinningBid,
			}
			if got := sa.GetAdpodSeatBids(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("structuredAdpod.GetAdpodSeatBids() = %v, want %v", got, tt.want)
			}
		})
	}
}
