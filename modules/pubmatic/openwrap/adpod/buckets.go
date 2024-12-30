package adpod

import (
	"sort"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type BidsBuckets map[int][]*models.Bid

func GetDurationWiseBidsBucket(bids []*models.Bid) BidsBuckets {
	result := BidsBuckets{}

	for i, bid := range bids {
		if bid.Status == models.StatusOK {
			result[bid.Duration] = append(result[bid.Duration], bids[i])
		}
	}

	for k, v := range result {
		//sort.Slice(v[:], func(i, j int) bool { return v[i].Price > v[j].Price })
		sortBids(v)
		result[k] = v
	}

	return result
}

func sortBids(bids []*models.Bid) {
	sort.Slice(bids, func(i, j int) bool {
		if bids[i].DealTierSatisfied == bids[j].DealTierSatisfied {
			return bids[i].Price > bids[j].Price
		}
		return bids[i].DealTierSatisfied
	})
}
