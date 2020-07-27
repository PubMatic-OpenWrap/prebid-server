package ctv

import (
	"fmt"
	"time"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/PubMatic-OpenWrap/prebid-server/pbsmetrics"
)

/********************* AdPodGenerator Functions *********************/

//IAdPodGenerator interface for generating AdPod from Ads
type IAdPodGenerator interface {
	GetAdPodBids() *AdPodBid
}
type filteredBid struct {
	bid        *Bid
	reasonCode FilterReasonCode
}
type highestCombination struct {
	bids          []*Bid
	bidIDs        []string
	durations     []int
	price         float64
	categoryScore map[string]int
	domainScore   map[string]int
	filteredBids  map[string]*filteredBid
}

//AdPodGenerator AdPodGenerator
type AdPodGenerator struct {
	IAdPodGenerator
	request  *openrtb.BidRequest
	impIndex int
	buckets  BidsBuckets
	comb     ICombination
	adpod    *openrtb_ext.VideoAdPod
	met      pbsmetrics.MetricsEngine
}

//NewAdPodGenerator will generate adpod based on configuration
func NewAdPodGenerator(request *openrtb.BidRequest, impIndex int, buckets BidsBuckets, comb ICombination, adpod *openrtb_ext.VideoAdPod, met pbsmetrics.MetricsEngine) *AdPodGenerator {
	return &AdPodGenerator{
		request:  request,
		impIndex: impIndex,
		buckets:  buckets,
		comb:     comb,
		adpod:    adpod,
		met:      met,
	}
}

//GetAdPodBids will return Adpod based on configurations
func (o *AdPodGenerator) GetAdPodBids() *AdPodBid {
	defer TimeTrack(time.Now(), fmt.Sprintf("Tid:%v ImpId:%v adpodgenerator", o.request.ID, o.request.Imp[o.impIndex].ID))
	isTimedOutORReceivedAllResponses := false
	responseCount := 0
	totalRequest := 0
	maxRequests := 5
	responseCh := make(chan *highestCombination, maxRequests)
	var results []*highestCombination

	timeout := 50 * time.Millisecond
	ticker := time.NewTicker(timeout)

	// monitor combination generator execution time
	start := time.Now()

	for totalRequest < maxRequests {
		durations := o.comb.Get()
		if len(durations) == 0 {
			break
		}

		totalRequest++
		go o.getUniqueBids(responseCh, durations)
	}

	labels := pbsmetrics.PodLabels{
		AlgorithmName:    "comb_gen",
		NoOfCombinations: new(int),
	}
	// defer o.met.RecordPodCombGenTime(labels, start)
	*labels.NoOfCombinations = totalRequest
	o.met.RecordPodCombGenTime(labels, start)

	for totalRequest > 0 && !isTimedOutORReceivedAllResponses {
		select {
		case hbc := <-responseCh:
			responseCount++
			if nil != hbc {
				results = append(results, hbc)
			}
			if responseCount == totalRequest {
				// monitor
				labels := pbsmetrics.PodLabels{
					AlgorithmName:    "comp_exclusion",
					NoOfResponseBids: new(int),
				}
				*labels.NoOfResponseBids = totalRequest
				o.met.RecordPodCompititveExclusionTime(labels, start)
				isTimedOutORReceivedAllResponses = true
			}
		case <-ticker.C:
			isTimedOutORReceivedAllResponses = true
			Logf("Tid:%v ImpId:%v GetAdPodBids Timeout Reached %v", o.request.ID, o.request.Imp[o.impIndex].ID, timeout)
		}
	}

	defer ticker.Stop()
	defer o.cleanupResponseChannel(responseCh, totalRequest-responseCount)

	if 0 == len(results) {
		Logf("Tid:%v ImpId:%v NoBid", o.request.ID, o.request.Imp[o.impIndex].ID)
		return nil
	}

	//Get Max Response
	var maxResult *highestCombination
	for _, result := range results {
		for _, rc := range result.filteredBids {
			if CTVRCDidNotGetChance == rc.bid.FilterReasonCode {
				rc.bid.FilterReasonCode = rc.reasonCode
			}
		}
		if len(result.bids) != 0 || nil == maxResult || maxResult.price < result.price {
			maxResult = result
		}
	}

	adpodBid := &AdPodBid{
		Bids:    maxResult.bids[:],
		Price:   maxResult.price,
		ADomain: make([]string, 0),
		Cat:     make([]string, 0),
	}

	//Get Unique Domains
	for domain := range maxResult.domainScore {
		adpodBid.ADomain = append(adpodBid.ADomain, domain)
	}

	//Get Unique Categories
	for cat := range maxResult.categoryScore {
		adpodBid.Cat = append(adpodBid.Cat, cat)
	}

	Logf("Tid:%v ImpId:%v Selected Durations:%v Bids:%v", o.request.ID, o.request.Imp[o.impIndex].ID, maxResult.durations[:], maxResult.bidIDs[:])
	return adpodBid
}

func (o *AdPodGenerator) cleanupResponseChannel(responseCh <-chan *highestCombination, responseCount int) {
	for responseCount > 0 {
		extra := <-responseCh
		Logf("Tid:%v ImpId:%v Delayed Response Durations:%v Bids:%v", o.request.ID, o.request.Imp[o.impIndex].ID, extra.durations, extra.bidIDs)
		responseCount--
	}
}

func (o *AdPodGenerator) getUniqueBids(responseCh chan<- *highestCombination, durationSequence []int) {
	defer TimeTrack(time.Now(), fmt.Sprintf("Tid:%v ImpId:%v getUniqueBids:%v", o.request.ID, o.request.Imp[o.impIndex].ID, durationSequence))

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
	hbc := findUniqueCombinations(data[:], combinations[:], *o.adpod.IABCategoryExclusionPercent, *o.adpod.AdvertiserExclusionPercent)
	hbc.durations = durationSequence[:]
	responseCh <- hbc
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

	hc := &highestCombination{}
	var ehc *highestCombination
	var rc FilterReasonCode
	inext, jnext := n-1, 0
	filterBids := map[string]*filteredBid{}

	// maintain highest price combination
	for true {

		ehc, inext, jnext, rc = evaluate(data[:], indices[:], totalBids, maxCategoryScore, maxDomainScore)
		if nil != ehc {
			if nil == hc || hc.price < ehc.price {
				hc = ehc
			} else {
				// if you see current combination price lower than the highest one then break the loop
				break
			}
		} else {
			//Filtered Bid
			for i := 0; i <= inext; i++ {
				for j := 0; j < combination[i] && !(i == inext && j > jnext); j++ {
					bid := data[i][indices[i][j]]
					if _, ok := filterBids[bid.ID]; !ok {
						filterBids[bid.ID] = &filteredBid{bid: bid, reasonCode: rc}
					}
				}
			}
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
			break
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
	if nil != filterBids {
		hc.filteredBids = filterBids
	}
	return hc
}

func evaluate(bids [][]*Bid, indices [][]int, totalBids int, maxCategoryScore, maxDomainScore int) (*highestCombination, int, int, FilterReasonCode) {

	hbc := &highestCombination{
		bids:          make([]*Bid, totalBids),
		bidIDs:        make([]string, totalBids),
		price:         0,
		categoryScore: make(map[string]int),
		domainScore:   make(map[string]int),
	}
	pos := 0

	for inext := range indices {
		for jnext := range indices[inext] {
			bid := bids[inext][indices[inext][jnext]]

			hbc.bids[pos] = bid
			hbc.bidIDs[pos] = bid.ID
			pos++

			//Price
			hbc.price = hbc.price + bid.Price

			//Categories
			for _, cat := range bid.Cat {
				hbc.categoryScore[cat]++
				if hbc.categoryScore[cat] > 1 && (hbc.categoryScore[cat]*100/totalBids) > maxCategoryScore {
					return nil, inext, jnext, CTVRCCategoryExclusion
				}
			}

			//Domain
			for _, domain := range bid.ADomain {
				hbc.domainScore[domain]++
				if hbc.domainScore[domain] > 1 && (hbc.domainScore[domain]*100/totalBids) > maxDomainScore {
					return nil, inext, jnext, CTVRCDomainExclusion
				}
			}
		}
	}

	return hbc, -1, -1, CTVRCWinningBid
}
