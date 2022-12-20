package response

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
)

func Test_findUniqueCombinations(t *testing.T) {
	type args struct {
		data             [][]*types.Bid
		combination      []int
		maxCategoryScore int
		maxDomainScore   int
	}
	tests := []struct {
		name string
		args args
		want *highestCombination
	}{
		{
			name: "sample",
			args: args{
				data: [][]*types.Bid{
					{
						{
							Bid:               &openrtb2.Bid{ID: "3-ed72b572-ba62-4220-abba-c19c0bf6346b", Price: 6.339115524232314},
							DealTierSatisfied: true,
						},
						{
							Bid:               &openrtb2.Bid{ID: "4-ed72b572-ba62-4220-abba-c19c0bf6346b", Price: 3.532468782358357},
							DealTierSatisfied: true,
						},
						{
							Bid:               &openrtb2.Bid{ID: "7-VIDEO12-89A1-41F1-8708-978FD3C0912A", Price: 5},
							DealTierSatisfied: false,
						},
						{
							Bid:               &openrtb2.Bid{ID: "8-VIDEO12-89A1-41F1-8708-978FD3C0912A", Price: 5},
							DealTierSatisfied: false,
						},
					}, //20

					{
						{
							Bid:               &openrtb2.Bid{ID: "2-ed72b572-ba62-4220-abba-c19c0bf6346b", Price: 3.4502433547413878},
							DealTierSatisfied: true,
						},
						{
							Bid:               &openrtb2.Bid{ID: "1-ed72b572-ba62-4220-abba-c19c0bf6346b", Price: 3.329644588311827},
							DealTierSatisfied: true,
						},
						{
							Bid:               &openrtb2.Bid{ID: "5-VIDEO12-89A1-41F1-8708-978FD3C0912A", Price: 5},
							DealTierSatisfied: false,
						},
						{
							Bid:               &openrtb2.Bid{ID: "6-VIDEO12-89A1-41F1-8708-978FD3C0912A", Price: 5},
							DealTierSatisfied: false,
						},
					}, //25
				},

				combination:      []int{2, 2},
				maxCategoryScore: 100,
				maxDomainScore:   100,
			},
			want: &highestCombination{
				bidIDs:    []string{"3-ed72b572-ba62-4220-abba-c19c0bf6346b", "4-ed72b572-ba62-4220-abba-c19c0bf6346b", "2-ed72b572-ba62-4220-abba-c19c0bf6346b", "1-ed72b572-ba62-4220-abba-c19c0bf6346b"},
				price:     16.651472249643884,
				nDealBids: 4,
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findUniqueCombinations(tt.args.data, tt.args.combination, tt.args.maxCategoryScore, tt.args.maxDomainScore)
			assert.Equal(t, tt.want.bidIDs, got.bidIDs, "bidIDs")
			assert.Equal(t, tt.want.nDealBids, got.nDealBids, "nDealBids")
			assert.Equal(t, tt.want.price, got.price, "price")
		})
	}
}

func TestAdPodGenerator_getMaxAdPodBid(t *testing.T) {
	type fields struct {
		request  *openrtb2.BidRequest
		impIndex int
	}
	type args struct {
		results         []*highestCombination
		getFilteredBids func(results []*highestCombination) []*types.Bid
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		want                 *types.AdPodBid
		expectedFilteredBids []*types.Bid
	}{
		{
			name: `EmptyResults`,
			fields: fields{
				request:  &openrtb2.BidRequest{ID: `req-1`, Imp: []openrtb2.Imp{{ID: `imp-1`}}},
				impIndex: 0,
			},
			args: args{
				results: nil,
				getFilteredBids: func(results []*highestCombination) []*types.Bid {
					filteredIds := []string{}
					bids := []*types.Bid{}

					for _, result := range results {
						for _, fb := range filteredIds {
							bids = append(bids, result.filteredBids[fb].bid)
						}
					}
					return bids
				},
			},
			want:                 nil,
			expectedFilteredBids: []*types.Bid{},
		},
		{
			name: `AllBidsFiltered`,
			fields: fields{
				request:  &openrtb2.BidRequest{ID: `req-1`, Imp: []openrtb2.Imp{{ID: `imp-1`}}},
				impIndex: 0,
			},
			args: args{
				results: []*highestCombination{
					{
						filteredBids: map[string]*filteredBid{
							`bid-1`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-1`}}, status: constant.StatusCategoryExclusion},
							`bid-2`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-2`}}, status: constant.StatusCategoryExclusion},
							`bid-3`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-3`}}, status: constant.StatusCategoryExclusion},
						},
					},
				},
				getFilteredBids: func(results []*highestCombination) []*types.Bid {
					filteredIds := []string{"bid-1", "bid-2", "bid-3"}
					bids := []*types.Bid{}
					for _, result := range results {
						for _, fb := range filteredIds {
							bids = append(bids, result.filteredBids[fb].bid)
						}
					}
					return bids
				},
			},
			want:                 nil,
			expectedFilteredBids: []*types.Bid{{Bid: &openrtb2.Bid{ID: `bid-1`}, Status: constant.StatusCategoryExclusion}, {Bid: &openrtb2.Bid{ID: `bid-2`}, Status: constant.StatusCategoryExclusion}, {Bid: &openrtb2.Bid{ID: `bid-3`}, Status: constant.StatusCategoryExclusion}},
		},
		{
			name: `SingleResponse`,
			fields: fields{
				request:  &openrtb2.BidRequest{ID: `req-1`, Imp: []openrtb2.Imp{{ID: `imp-1`}}},
				impIndex: 0,
			},
			args: args{
				results: []*highestCombination{
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-1`}},
							{Bid: &openrtb2.Bid{ID: `bid-2`}},
							{Bid: &openrtb2.Bid{ID: `bid-3`}},
						},
						bidIDs:    []string{`bid-1`, `bid-2`, `bid-3`},
						price:     20,
						nDealBids: 0,
						categoryScore: map[string]int{
							`cat-1`: 1,
							`cat-2`: 1,
						},
						domainScore: map[string]int{
							`domain-1`: 1,
							`domain-2`: 1,
						},
						filteredBids: map[string]*filteredBid{
							`bid-4`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-4`}}, status: constant.StatusCategoryExclusion},
						},
					},
				},
				getFilteredBids: func(results []*highestCombination) []*types.Bid {
					filteredIds := []string{"bid-4"}
					bids := []*types.Bid{}
					for _, result := range results {
						for _, fb := range filteredIds {
							bids = append(bids, result.filteredBids[fb].bid)
						}
					}
					return bids
				},
			},
			want: &types.AdPodBid{
				Bids: []*types.Bid{
					{Bid: &openrtb2.Bid{ID: `bid-1`}},
					{Bid: &openrtb2.Bid{ID: `bid-2`}},
					{Bid: &openrtb2.Bid{ID: `bid-3`}},
				},
				Cat:     []string{`cat-1`, `cat-2`},
				ADomain: []string{`domain-1`, `domain-2`},
				Price:   20,
			},
			expectedFilteredBids: []*types.Bid{{Bid: &openrtb2.Bid{ID: `bid-4`}, Status: constant.StatusCategoryExclusion}},
		},
		{
			name: `MultiResponse-AllNonDealBids`,
			fields: fields{
				request:  &openrtb2.BidRequest{ID: `req-1`, Imp: []openrtb2.Imp{{ID: `imp-1`}}},
				impIndex: 0,
			},
			args: args{
				results: []*highestCombination{
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-11`}},
						},
						bidIDs:    []string{`bid-11`},
						price:     10,
						nDealBids: 0,
						filteredBids: map[string]*filteredBid{
							`bid-12`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-12`}}, status: constant.StatusCategoryExclusion},
						},
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-21`}},
						},
						bidIDs:    []string{`bid-21`},
						price:     20,
						nDealBids: 0,
						filteredBids: map[string]*filteredBid{
							`bid-22`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-22`}}, status: constant.StatusCategoryExclusion},
						},
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-31`}},
						},
						bidIDs:    []string{`bid-31`},
						price:     10,
						nDealBids: 0,
						filteredBids: map[string]*filteredBid{
							`bid-32`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-32`}}, status: constant.StatusCategoryExclusion},
						},
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-41`}},
						},
						bidIDs:    []string{`bid-41`},
						price:     15,
						nDealBids: 0,
						filteredBids: map[string]*filteredBid{
							`bid-42`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-42`}}, status: constant.StatusCategoryExclusion},
						},
					},
				},
				getFilteredBids: func(results []*highestCombination) []*types.Bid {
					filteredIds := []string{"bid-22"}
					bids := []*types.Bid{}
					for _, result := range results {
						for _, fb := range filteredIds {
							if _, ok := result.filteredBids[fb]; ok {
								bids = append(bids, result.filteredBids[fb].bid)
							}
						}
					}
					return bids
				},
			},
			want: &types.AdPodBid{
				Bids: []*types.Bid{
					{Bid: &openrtb2.Bid{ID: `bid-21`}},
				},
				Cat:     []string{},
				ADomain: []string{},
				Price:   20,
			},
			expectedFilteredBids: []*types.Bid{{Bid: &openrtb2.Bid{ID: `bid-22`}, Status: constant.StatusCategoryExclusion}},
		},
		{
			name: `MultiResponse-AllDealBids-SameCount`,
			fields: fields{
				request:  &openrtb2.BidRequest{ID: `req-1`, Imp: []openrtb2.Imp{{ID: `imp-1`}}},
				impIndex: 0,
			},
			args: args{
				results: []*highestCombination{
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-11`}},
						},
						bidIDs:    []string{`bid-11`},
						price:     10,
						nDealBids: 1,
						filteredBids: map[string]*filteredBid{
							`bid-12`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-12`}}, status: constant.StatusDomainExclusion},
						},
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-21`}},
						},
						bidIDs:    []string{`bid-21`},
						price:     20,
						nDealBids: 1,
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-31`}},
						},
						bidIDs:    []string{`bid-31`},
						price:     10,
						nDealBids: 1,
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-41`}},
						},
						bidIDs:    []string{`bid-41`},
						price:     15,
						nDealBids: 1,
					},
				},
				getFilteredBids: func(results []*highestCombination) []*types.Bid {
					filteredIds := []string{"bid-22"}
					bids := []*types.Bid{}
					for _, result := range results {
						for _, fb := range filteredIds {
							if _, ok := result.filteredBids[fb]; ok {
								bids = append(bids, result.filteredBids[fb].bid)
							}
						}
					}
					return bids
				},
			},
			want: &types.AdPodBid{
				Bids: []*types.Bid{
					{Bid: &openrtb2.Bid{ID: `bid-21`}},
				},
				Cat:     []string{},
				ADomain: []string{},
				Price:   20,
			},
			expectedFilteredBids: []*types.Bid{},
		},
		{
			name: `MultiResponse-AllDealBids-DifferentCount`,
			fields: fields{
				request:  &openrtb2.BidRequest{ID: `req-1`, Imp: []openrtb2.Imp{{ID: `imp-1`}}},
				impIndex: 0,
			},
			args: args{
				results: []*highestCombination{
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-11`}},
						},
						bidIDs:    []string{`bid-11`},
						price:     10,
						nDealBids: 2,
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-21`}},
						},
						bidIDs:    []string{`bid-21`},
						price:     20,
						nDealBids: 1,
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-31`}},
						},
						bidIDs:    []string{`bid-31`},
						price:     10,
						nDealBids: 3,
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-41`}},
						},
						bidIDs:    []string{`bid-41`},
						price:     15,
						nDealBids: 2,
					},
				},
				getFilteredBids: func(results []*highestCombination) []*types.Bid {
					filteredIds := []string{"bid-22"}
					bids := []*types.Bid{}
					for _, result := range results {
						for _, fb := range filteredIds {
							if _, ok := result.filteredBids[fb]; ok {
								bids = append(bids, result.filteredBids[fb].bid)
							}
						}
					}
					return bids
				},
			},
			want: &types.AdPodBid{
				Bids: []*types.Bid{
					{Bid: &openrtb2.Bid{ID: `bid-31`}},
				},
				Cat:     []string{},
				ADomain: []string{},
				Price:   10,
			},
			expectedFilteredBids: []*types.Bid{},
		},
		{
			name: `MultiResponse-Mixed-DealandNonDealBids`,
			fields: fields{
				request:  &openrtb2.BidRequest{ID: `req-1`, Imp: []openrtb2.Imp{{ID: `imp-1`}}},
				impIndex: 0,
			},
			args: args{
				results: []*highestCombination{
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-11`}},
						},
						bidIDs:    []string{`bid-11`},
						price:     10,
						nDealBids: 2,
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-21`}},
						},
						bidIDs:    []string{`bid-21`},
						price:     20,
						nDealBids: 0,
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-31`}},
						},
						bidIDs:    []string{`bid-31`},
						price:     10,
						nDealBids: 3,
						filteredBids: map[string]*filteredBid{
							`bid-32`: {bid: &types.Bid{Bid: &openrtb2.Bid{ID: `bid-32`}}, status: constant.StatusDomainExclusion},
						},
					},
					{
						bids: []*types.Bid{
							{Bid: &openrtb2.Bid{ID: `bid-41`}},
						},
						bidIDs:    []string{`bid-41`},
						price:     15,
						nDealBids: 0,
					},
				},
				getFilteredBids: func(results []*highestCombination) []*types.Bid {
					filteredIds := []string{"bid-32"}
					bids := []*types.Bid{}
					for _, result := range results {
						for _, fb := range filteredIds {
							if _, ok := result.filteredBids[fb]; ok {
								bids = append(bids, result.filteredBids[fb].bid)
							}
						}
					}
					return bids
				},
			},
			want: &types.AdPodBid{
				Bids: []*types.Bid{
					{Bid: &openrtb2.Bid{ID: `bid-31`}},
				},
				Cat:     []string{},
				ADomain: []string{},
				Price:   10,
			},
			expectedFilteredBids: []*types.Bid{{Bid: &openrtb2.Bid{ID: `bid-32`}, Status: constant.StatusDomainExclusion}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &AdPodGenerator{
				request:  tt.fields.request,
				impIndex: tt.fields.impIndex,
			}
			got := o.getMaxAdPodBid(tt.args.results)
			if nil != got {
				sort.Strings(got.ADomain)
				sort.Strings(got.Cat)
			}
			assert.Equal(t, tt.want, got)
			filteredBids := tt.args.getFilteredBids(tt.args.results)
			assert.Equal(t, tt.expectedFilteredBids, filteredBids, "Filtered Bids mismatch")
		})
	}
}
