package auction

// Package combination generates possible ad pod response
// based on bid response durations. It ensures that generated
// combination is satifying ad pod request configurations like
// Min Pod Duation, Maximum Pod Duration, Minimum number of ads, Maximum number of Ads.
// It also considers number of bids received for given duration
// For Example, if for 60 second duration we have 2 bids then
// then it will ensure combination contains at most 2 repeatations of 60 sec; not more than that

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const (
	// MinToMax tells combination generator to generate combinations
	// starting from Min Ads to Max Ads
	MinToMax = iota

	// MaxToMin tells combination generator to generate combinations
	// starting from Max Ads to Min Ads
	MaxToMin
)

// ICombination ...
type CombinationGenerator interface {
	Get() []int
}

// Combination ...
type Combination struct {
	data      []int
	generator generator
	config    *models.AdPod
	order     int // order of combination generator
}

// NewCombination ... Generates on demand valid combinations
// Valid combinations are those who satisifies
//  1. Pod Min Max duration
//  2. minAds <= size(combination) <= maxads
//  3. If  Combination contains repeatition for given duration then
//     repeatitions are <= no of ads received for the duration
//
// Use Get method to start getting valid combinations
func NewCombination(buckets BidsBuckets, podMinDuration, podMaxDuration uint64, config *models.AdPod) CombinationGenerator {
	generator := new(generator)
	durationBidsCnts := make([][2]uint64, 0)
	for duration, bids := range buckets {
		durationBidsCnts = append(durationBidsCnts, [2]uint64{uint64(duration), uint64(len(bids))})
	}
	generator.Init(podMinDuration, podMaxDuration, config, durationBidsCnts, MaxToMin)
	return &Combination{
		generator: *generator,
		config:    config,
	}
}

// Get next valid combination
// Retuns empty slice if all combinations are generated
func (c *Combination) Get() []int {
	nextComb := c.generator.Next()
	nextCombInt := make([]int, len(nextComb))
	cnt := 0
	for _, duration := range nextComb {
		nextCombInt[cnt] = int(duration)
		cnt++
	}
	return nextCombInt
}
