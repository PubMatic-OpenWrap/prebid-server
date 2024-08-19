package auction

import (
	"errors"
	"sort"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// BidsBuckets bids bucket
type BidsBuckets map[int][]*Bid

// doAdPodExclusions
func doAdPodExclusions(impBidMap map[string]*AdPodBid, impCtx map[string]models.ImpCtx) ([]*AdPodBid, []error) {

	result := []*AdPodBid{}
	var errs []error
	for impId, bid := range impBidMap {
		if bid != nil && len(bid.Bids) > 0 {
			eachImpCtx := impCtx[impId]
			//TODO: MULTI ADPOD IMPRESSIONS
			//duration wise buckets sorted
			buckets := GetDurationWiseBidsBucket(bid.Bids)

			if len(buckets) == 0 {
				errs = append(errs, errors.New("prebid_ctv all bids filtered while matching lineitem duration"))
				continue
			}

			//combination generator
			comb := NewCombination(
				buckets,
				uint64(eachImpCtx.Video.MinDuration),
				uint64(eachImpCtx.Video.MaxDuration),
				eachImpCtx.AdpodConfig)

			//adpod generator
			adpodGenerator := NewAdPodGenerator(buckets, comb, eachImpCtx.AdpodConfig)

			adpodBids := adpodGenerator.GetAdPodBids()
			if adpodBids == nil {
				errs = append(errs, errors.New("prebid_ctv unable to generate adpod from bids combinations"))
				continue
			}

			adpodBids.OriginalImpID = bid.OriginalImpID
			adpodBids.SeatName = bid.SeatName
			result = append(result, adpodBids)
		}
	}
	return result, errs
}

func GetDurationWiseBidsBucket(bids []*Bid) BidsBuckets {
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

func sortBids(bids []*Bid) {
	sort.Slice(bids, func(i, j int) bool {
		if bids[i].DealTierSatisfied == bids[j].DealTierSatisfied {
			return bids[i].Price > bids[j].Price
		}
		return bids[i].DealTierSatisfied
	})
}
