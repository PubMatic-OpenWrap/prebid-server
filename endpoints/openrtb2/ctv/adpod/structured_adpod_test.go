package adpod

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
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
		wantImpBidMap  map[string][]*types.Bid
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
			wantImpBidMap: map[string][]*types.Bid{
				"imp1": {
					{
						Bid: &openrtb2.Bid{
							Price: 5,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
					},
					{
						Bid: &openrtb2.Bid{
							Price: 2,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 1,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
				},
				"imp2": {
					{
						Bid: &openrtb2.Bid{
							Price: 6,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
					},
					{
						Bid: &openrtb2.Bid{
							Price: 4,
						},
						DealTierSatisfied: false,
						Seat:              "index",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
			wantImpBidMap: map[string][]*types.Bid{
				"imp1": {
					{
						Bid: &openrtb2.Bid{
							Price: 6,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
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
			wantImpBidMap: map[string][]*types.Bid{
				"imp1": {
					{
						Bid: &openrtb2.Bid{
							Price: 1,
						},
						DealTierSatisfied: true,
						Seat:              "pubmatic",
					},
					{
						Bid: &openrtb2.Bid{
							Price: 6,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 4,
							Cat:   []string{"IAB-2"},
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
				},
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
			wantImpBidMap: map[string][]*types.Bid{
				"imp1": {
					{
						Bid: &openrtb2.Bid{
							Price: 6,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: true,
						Seat:              "pubmatic",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
				},
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
			wantImpBidMap: map[string][]*types.Bid{
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 1,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
			wantImpBidMap: map[string][]*types.Bid{
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 1,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 4,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: true,
						Seat:              "god",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
			wantImpBidMap: map[string][]*types.Bid{
				"imp1": {
					{
						Bid: &openrtb2.Bid{
							Price: 6,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: true,
						Seat:              "pubmatic",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 4,
							Cat:   []string{"IAB-2"},
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 4,
						},
						DealTierSatisfied: false,
						Seat:              "god",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
				},
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
			wantImpBidMap: map[string][]*types.Bid{
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 1,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 10,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: false,
						Seat:              "god",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
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
			name: "price_based_auction_with_one_slot_no_bid",
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
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "appnexux",
						},
					},
					"imp2": {
						{
							Bid: &openrtb2.Bid{
								Price: 4,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "index",
						},
						{
							Bid: &openrtb2.Bid{
								Price: 2,
								Cat:   []string{"IAB-1"},
							},
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
					},
				},
				WinningBid: make(map[string]types.Bid),
			},
			wantImpBidMap: map[string][]*types.Bid{
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
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: false,
						Seat:              "appnexux",
						Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
					},
				},
				"imp2": {
					{
						Bid: &openrtb2.Bid{
							Price: 4,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: false,
						Seat:              "index",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
					},
					{
						Bid: &openrtb2.Bid{
							Price: 2,
							Cat:   []string{"IAB-1"},
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
					},
				},
			},
			wantWinningBid: map[string]types.Bid{
				"imp1": {
					Bid: &openrtb2.Bid{
						Price: 6,
						Cat:   []string{"IAB-1"},
					},
					DealTierSatisfied: false,
					Seat:              "pubmatic",
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
			assert.Equal(t, tt.wantImpBidMap, sa.ImpBidMap)
			assert.Equal(t, tt.wantWinningBid, sa.WinningBid)
		})
	}
}

func TestStructuredAdpodCollectSeatNonBids(t *testing.T) {
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
							Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
			snb := sa.CollectSeatNonBids()
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
						Nbr:               nil,
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
