package ctv

import (
	"context"
	"time"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

/********************* AdPodGenerator Functions *********************/

//IAdPodGenerator interface for generating AdPod from Ads
type IAdPodGenerator interface {
	GetAdPodBids() *AdPodBid
}
type filteredBids struct {
	bid        *Bid
	reasonCode FilterReasonCode
}
type highestCombination struct {
	bids          []*Bid
	price         float64
	categoryScore map[string]int
	domainScore   map[string]int
	filteredBids  []filteredBids
}

//AdPodGenerator AdPodGenerator
type AdPodGenerator struct {
	IAdPodGenerator
	buckets BidsBuckets
	comb    ICombination
	adpod   *openrtb_ext.VideoAdPod
}

//NewAdPodGenerator will generate adpod based on configuration
func NewAdPodGenerator(buckets BidsBuckets, comb ICombination, adpod *openrtb_ext.VideoAdPod) *AdPodGenerator {
	return &AdPodGenerator{
		buckets: buckets,
		comb:    comb,
		adpod:   adpod,
	}
}

//GetAdPodBids will return Adpod based on configurations
func (o *AdPodGenerator) GetAdPodBids() *AdPodBid {

	isTimedOutORReceivedAllResponses := false
	responseCount := 0
	totalRequest := 0
	maxRequests := 5
	results := make([]*highestCombination, maxRequests)
	responseCh := make(chan *highestCombination, maxRequests)

	timeout := 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for totalRequest < maxRequests {
		durations := o.comb.Get()
		if len(durations) == 0 {
			break
		}

		totalRequest++
		go o.getUniqueBids(responseCh, durations)
	}

	for !isTimedOutORReceivedAllResponses {
		select {
		case <-ctx.Done():
			isTimedOutORReceivedAllResponses = true
		case hbc := <-responseCh:
			responseCount++
			if nil != hbc { //&& (nil == maxResult || maxResult.price < hbc.price) {
				results = append(results, hbc)
			}
			if responseCount == totalRequest {
				isTimedOutORReceivedAllResponses = true
			}
		}
	}

	go cleanupResponseChannel(responseCh, totalRequest-responseCount)

	if 0 == len(results) {
		return nil
	}

	var maxResult *highestCombination
	for _, result := range results {
		if nil == maxResult || maxResult.price < result.price {
			maxResult = result
		}

		for _, rc := range result.filteredBids {
			if CTVRCDidNotGetChance == rc.bid.FilterReasonCode {
				rc.bid.FilterReasonCode = rc.reasonCode
			}
		}
	}

	adpodBid := &AdPodBid{
		Bids:    maxResult.bids[:],
		Price:   maxResult.price,
		ADomain: make([]string, len(maxResult.domainScore)),
		Cat:     make([]string, len(maxResult.categoryScore)),
	}

	//Get Unique Domains
	for domain := range maxResult.domainScore {
		adpodBid.ADomain = append(adpodBid.ADomain, domain)
	}

	//Get Unique Categories
	for cat := range maxResult.categoryScore {
		adpodBid.Cat = append(adpodBid.Cat, cat)
	}

	return adpodBid
}

func cleanupResponseChannel(responseCh <-chan *highestCombination, responseCount int) {
	for responseCount > 0 {
		<-responseCh
		responseCount--
	}
}

func (o *AdPodGenerator) getUniqueBids(responseCh chan<- *highestCombination, durationSequence []int) {
	data := [][]*Bid{}
	combinations := []int{}

	uniqueDuration := 0
	for index, duration := range durationSequence {
		if 0 != index && durationSequence[index-1] == duration {
			combinations[uniqueDuration-1]++
			continue
		}
		data = append(data, o.buckets[duration][:])
		combinations = append(combinations, 1)
		uniqueDuration++
	}

	responseCh <- findUniqueCombinations(data[:], combinations[:], *o.adpod.IABCategoryExclusionPercent, *o.adpod.AdvertiserExclusionPercent)
}

func findUniqueCombinations(data [][]*Bid, combination []int, maxCategoryScore, maxDomainScore int) *highestCombination {
	// number of arrays
	n := len(combination)
	totalBids := 0
	//  to keep track of next element in each of the n arrays
	// indices is initialized
	indices := make([][]int, len(combination))
	for i := 0; i < len(combination); i++ {
		indices[i] = make([]int, combination[i])
		for j := 0; j < combination[i]; j++ {
			indices[i][j] = j
			totalBids++
		}
	}

	hc := &highestCombination{price: 0}
	var ehc *highestCombination
	var rc FilterReasonCode
	inext, jnext := n-1, 0
	var filterBids []filteredBids

	// maintain highest price combination
	for true {

		ehc, inext, jnext, rc = evaluate(data[:], indices[:], totalBids, maxCategoryScore, maxDomainScore)
		if nil != ehc {
			if nil == hc || hc.price < ehc.price {
				hc = ehc
			} else {
				// if you see current combination price lower than the highest one then break the loop
				hc.filteredBids = filterBids[:]
				return hc
			}
		} else {
			//Filtered Bid
			filterBids = append(filterBids, filteredBids{bid: data[inext][indices[inext][jnext]], reasonCode: rc})
		}

		if -1 == inext {
			inext, jnext = n-1, 0
		}

		// find the rightmost array that has more
		// elements left after the current element
		// in that array
		inext, jnext := n-1, 0

		for inext >= 0 {
			jnext = len(indices[inext]) - 1
			for jnext >= 0 && (indices[inext][jnext]+1 > (len(data[inext]) - len(indices[inext]) + jnext)) {
				jnext--
			}
			if jnext >= 0 {
				break
			}
			inext--
		}

		// no such array is found so no more combinations left
		if inext < 0 {
			// return output
			return nil
		}

		// if found move to next element in that array
		indices[inext][jnext]++

		// for all arrays to the right of this
		// array current index again points to
		// first element
		jnext++
		for i := inext; i < len(combination); i++ {
			for j := jnext; j < combination[i]; j++ {
				if i == inext {
					indices[i][j] = indices[i][j-1] + 1
				} else {
					indices[i][j] = j
				}
			}
			jnext = 0
		}
	}
	//setting filteredBids
	hc.filteredBids = filterBids[:]
	return hc
}

func evaluate(bids [][]*Bid, indices [][]int, totalBids int, maxCategoryScore, maxDomainScore int) (*highestCombination, int, int, FilterReasonCode) {

	hbc := &highestCombination{
		bids:          make([]*Bid, totalBids),
		price:         0,
		categoryScore: make(map[string]int),
		domainScore:   make(map[string]int),
	}
	pos := 0

	for inext := range indices {
		for jnext := range indices[inext] {
			bid := bids[inext][indices[inext][jnext]]
			hbc.bids[pos] = bid

			//Price
			hbc.price = hbc.price + bid.Price

			//Categories
			for _, cat := range bid.Cat {
				hbc.categoryScore[cat]++
				if (hbc.categoryScore[cat] * 100 / totalBids) > maxCategoryScore {
					return nil, inext, jnext, CTVRCCategoryExclusion
				}
			}

			//Domain
			for _, domain := range bid.ADomain {
				hbc.domainScore[domain]++
				if (hbc.domainScore[domain] * 100 / totalBids) > maxDomainScore {
					return nil, inext, jnext, CTVRCDomainExclusion
				}
			}
		}
	}

	return hbc, -1, -1, CTVRCWinningBid
}
