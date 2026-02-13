package ctvlegacy

import (
	"errors"
	"sort"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// BidsBuckets bids bucket
type BidsBuckets map[int][]*Bid

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

// Do exclusion for single adpod
func doAuctionAndExclusion(adpodBid *AdPodBid, podConfig models.AdpodConfig) (*AdPodBid, []error) {
	var errs []error

	// For dynamic adpod there will be only one slot
	podImp := podConfig.Slots[0]

	//duration wise buckets sorted
	buckets := GetDurationWiseBidsBucket(adpodBid.Bids)
	if len(buckets) == 0 {
		errs = append(errs, errors.New("prebid_ctv all bids filtered while matching lineitem duration for adpod: "+podConfig.PodID))
		return nil, errs
	}

	//combination generator
	comb := NewCombination(
		buckets,
		uint64(podImp.AdpodConfigV25.MinPodDuration),
		uint64(podImp.AdpodConfigV25.MaxPodDuration),
		podImp)

	//adpod generator
	adpodGenerator := NewAdPodGenerator(buckets, comb, podImp)

	newadpodBid := adpodGenerator.GetAdPodBids()
	if newadpodBid == nil {
		errs = append(errs, errors.New("prebid_ctv unable to generate adpod from bids combinations for adpod: "+podConfig.PodID))
		return nil, errs
	}

	newadpodBid.OriginalImpID = adpodBid.OriginalImpID
	newadpodBid.SeatName = adpodBid.SeatName

	return newadpodBid, errs
}
