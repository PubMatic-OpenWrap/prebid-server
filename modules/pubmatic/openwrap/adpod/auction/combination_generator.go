package auction

import (
	"math/big"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// generator holds all the combinations based
// on Video Ad Pod request and Bid Response Max duration
type generator struct {
	podMinDuration      uint64            // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration      uint64            // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds              uint64            // Minimum Ads value present in origin Video Ad Pod Request
	maxAds              uint64            // Maximum Ads value present in origin Video Ad Pod Request
	slotDurations       []uint64          // input slot durations for which
	slotDurationAdMap   map[uint64]uint64 // map of key = duration, value = no of creatives with given duration
	noOfSlots           int               // Number of slots to be consider (from left to right)
	combinationCountMap map[uint64]uint64 //key - number of ads, ranging from 1 to maxads given in request config value - containing no of combinations with repeatation each key can have (without validations)
	stats               stats             // metrics information
	combinations        [][]uint64        // May contains some/all combinations at given point of time
	state               snapshot          // state configurations in case of lazy loading
	order               int               // Indicates generation order e.g. maxads to min ads
}

// stats holds the metrics information for given point of time
// such as current combination count, valid combination count, repeatation count
// out of range combination
type stats struct {
	currentCombinationCount   int    // current combination count generated out of totalExpectedCombinations
	validCombinationCount     int    //
	repeatationsCount         int    // no of combinations not considered because containing some/all durations for which only single ad is present
	outOfRangeCount           int    // no of combinations out of range because not satisfied pod min and max range
	totalExpectedCombinations uint64 // indicates total number for possible combinations without validations but subtracts repeatations for duration with single ad
}

// snashot retains the state of iteration
// it is used in determing when next valid combination is requested
// using Next() method
type snapshot struct {
	start              uint64   // indicates which duration to be used to form combination
	index              int64    // indicates from which index in combination array we should fill duration given by start
	r                  uint64   // holds the current combination length ranging from minads to maxads
	lastCombination    []uint64 // holds the last combination iterated
	stateUpdated       bool     // flag indicating whether underneath search method updated the c.state values
	valueUpdated       bool     // indicates whether search method determined and updated next combination
	combinationCounter uint64   // holds the index of duration to be filled when 1 cycle of combination ends
	resetFlags         bool     // indicates whether the required flags to reset or not
}

// Init ...initializes with following
// 1. Determines the number of combinations to be generated
// 2. Intializes the c.state values required for c.Next() and iteratoor
// generationOrder indicates how combinations should be generated.
func (g *generator) Init(podMinDuration, podMaxDuration uint64, config *models.AdPod, durationAdsMap [][2]uint64, generationOrder int) {

	g.podMinDuration = podMinDuration
	g.podMaxDuration = podMaxDuration
	g.minAds = uint64(config.MinAds)
	g.maxAds = uint64(config.MaxAds)

	// map of key = duration value = number of ads(must be non zero positive number)
	g.slotDurationAdMap = make(map[uint64]uint64, len(g.slotDurations))

	// iterate and extract duration and number of ads belonging to the duration
	// split logic - :: separated

	cnt := 0
	g.slotDurations = make([]uint64, len(durationAdsMap))
	for _, durationNoOfAds := range durationAdsMap {

		g.slotDurations[cnt] = durationNoOfAds[0]
		// save duration  and no of ads info
		g.slotDurationAdMap[durationNoOfAds[0]] = durationNoOfAds[1]
		cnt++
	}

	g.noOfSlots = len(g.slotDurations)
	g.stats.currentCombinationCount = 0
	g.stats.validCombinationCount = 0
	g.state = snapshot{}

	g.combinationCountMap = make(map[uint64]uint64, g.maxAds)
	// compute no of possible combinations (without validations)
	// using configurationss
	g.stats.totalExpectedCombinations = compute(g, g.maxAds, true)
	subtractUnwantedRepeatations(g)
	// c.combinations = make([][]uint64, c.totalExpectedCombinations)
	// util.Logf("Allow Repeatation = %v", c.allowRepetitationsForEligibleDurations)
	// util.Logf("Total possible combinations (without validations) = %v ", c.totalExpectedCombinations)

	/// new states
	g.state.start = uint64(0)
	g.state.index = 0
	g.state.r = g.minAds
	g.order = generationOrder
	if g.order == MaxToMin {
		g.state.r = g.maxAds
	}
	g.state.resetFlags = true
}

// Next - Get next ad slot combination
// returns empty array if next combination is not present
func (g *generator) Next() []uint64 {
	var comb []uint64
	if len(g.slotDurations) <= 0 {
		return comb
	}
	if g.state.resetFlags {
		reset(g)
		g.state.resetFlags = false
	}
	for {
		comb = g.lazyNext()
		if len(comb) == 0 || isValidCombination(g, comb) {
			break
		}
	}
	return comb
}

func isValidCombination(gen *generator, combination []uint64) bool {
	// check if repeatations are allowed
	repeationMap := make(map[uint64]uint64, len(gen.slotDurations))
	totalAdDuration := uint64(0)
	for _, duration := range combination {
		repeationMap[uint64(duration)]++
		// check current combination contains repeating durations such that
		// count(duration) > count(no of ads aunction engine received for the duration)
		currentRepeationCnt := repeationMap[duration]
		noOfAdsPresent := gen.slotDurationAdMap[duration]
		if currentRepeationCnt > noOfAdsPresent {
			//util.Logf("count = %v :: Discarding combination '%v' as only '%v' ad is present for duration %v", c.stats.currentCombinationCount, combination, noOfAdsPresent, duration)
			gen.stats.repeatationsCount++
			return false
		}

		// check if sum of durations is withing pod min and max duration
		totalAdDuration += duration
	}

	if !(totalAdDuration >= gen.podMinDuration && totalAdDuration <= gen.podMaxDuration) {
		// totalAdDuration is not within range of Pod min and max duration
		//util.Logf("count = %v :: Discarding combination '%v' as either total Ad duration (%v) < %v (Pod min duration) or > %v (Pod Max duration)", c.stats.currentCombinationCount, combination, totalAdDuration, c.podMinDuration, c.podMaxDuration)
		gen.stats.outOfRangeCount++
		return false
	}
	gen.stats.validCombinationCount++
	return true
}

// compute - number of combinations that can be generated based on
// 1. minads
// 2. maxads
// 3. Ordering of durations not matters. i.e. 4,5,6 will not be considered again as 5,4,6 or 6,5,4
// 4. Repeatations are allowed only for those durations where multiple ads are present
// Sum ups number of combinations for each noOfAds (r) based on above criteria and returns the total
// It operates recursively
// c - algorithm config, noOfAds (r) - maxads requested (if recursion=true otherwise any valid value), recursion - whether to do recursion or not. if false then only single combination
// for given noOfAds will be computed
func compute(gen *generator, noOfAds uint64, recursion bool) uint64 {

	// can not limit till  c.minAds
	// because we want to construct
	// c.combinationCountMap required by subtractUnwantedRepeatations
	if noOfAds <= 0 || len(gen.slotDurations) <= 0 {
		return 0
	}
	var noOfCombinations *big.Int
	// Formula
	//		(r + n - 1)!
	//      ------------
	//       r! (n - 1)!
	n := uint64(len(gen.slotDurations))
	r := uint64(noOfAds)
	d1 := fact(uint64(r))
	d2 := fact(n - 1)
	d3 := d1.Mul(&d1, &d2)
	nmrt := fact(r + n - 1)

	noOfCombinations = nmrt.Div(&nmrt, d3)
	// store pure combination with repeatation in combinationCountMap
	gen.combinationCountMap[r] = noOfCombinations.Uint64()
	//util.Logf("%v", noOfCombinations)
	if recursion {

		// add only if it  is  withing limit of c.minads
		nextLevelCombinations := compute(gen, noOfAds-1, recursion)
		if noOfAds-1 >= gen.minAds {
			sumOfCombinations := noOfCombinations.Add(noOfCombinations, big.NewInt(int64(nextLevelCombinations)))
			return sumOfCombinations.Uint64()
		}

	}
	return noOfCombinations.Uint64()
}

// fact computes factorial of given number.
// It is used by compute function
func fact(no uint64) big.Int {
	if no == 0 {
		return *big.NewInt(int64(1))
	}
	var bigNo big.Int
	bigNo.SetUint64(no)

	fact := fact(no - 1)
	mult := bigNo.Mul(&bigNo, &fact)

	return *mult
}

// searchAll - searches all valid combinations
// valid combinations are those which satisifies following
// 1. sum of duration is within range of pod min and max values
// 2. Each duration within combination honours number of ads value given in the request
// 3. Number of durations in combination are within range of min and max ads
func (g *generator) searchAll() [][]uint64 {
	reset(g)
	start := uint64(0)
	index := uint64(0)

	if g.order == MinToMax {
		for r := g.minAds; r <= g.maxAds; r++ {
			data := make([]uint64, r)
			g.search(data, start, index, r, false, 0)
		}
	}
	if g.order == MaxToMin {
		for r := g.maxAds; r >= g.minAds; r-- {
			data := make([]uint64, r)
			g.search(data, start, index, r, false, 0)
		}
	}
	// util.Logf("Total combinations generated = %v", c.currentCombinationCount)
	// util.Logf("Total combinations expected = %v", c.totalExpectedCombinations)
	// result := make([][]uint64, c.totalExpectedCombinations)
	result := make([][]uint64, g.stats.validCombinationCount)
	copy(result, g.combinations)
	g.stats.currentCombinationCount = 0
	return result
}

// reset the internal counters
func reset(g *generator) {
	g.stats.currentCombinationCount = 0
	g.stats.validCombinationCount = 0
	g.stats.repeatationsCount = 0
	g.stats.outOfRangeCount = 0
}

// lazyNext performs stateful iteration. Instead of returning all valid combinations
// in one gp, it will return each combination on demand basis.
// valid combinations are those which satisifies following
// 1. sum of duration is within range of pod min and max values
// 2. Each duration within combination honours number of ads value given in the request
// 3. Number of durations in combination are within range of min and max ads
func (g *generator) lazyNext() []uint64 {
	start := g.state.start
	index := g.state.index
	r := g.state.r
	// reset last combination
	// by deleting previous values
	if g.state.lastCombination == nil {
		g.combinations = make([][]uint64, 0)
	}
	data := &g.state.lastCombination
	if *data == nil || uint64(len(*data)) != r {
		*data = make([]uint64, r)
	}
	g.state.stateUpdated = false
	g.state.valueUpdated = false
	var result []uint64
	if (g.order == MinToMax && r <= g.maxAds) || (g.order == MaxToMin && r >= g.minAds) {
		// for ;  r <= c.maxAds; r++ {
		g.search(*data, start, uint64(index), r, true, 0)
		g.state.stateUpdated = false // reset
		g.state.valueUpdated = false
		result = make([]uint64, len(*data))
		copy(result, *data)
	}
	return result
}

// search generates the combinations based on min and max number of ads
func (g *generator) search(data []uint64, start, index, r uint64, lazyLoad bool, reursionCount int) []uint64 {

	end := uint64(len(g.slotDurations) - 1)

	// Current combination is ready to be printed, print it
	if index == r {
		data1 := make([]uint64, len(data))
		for j := uint64(0); j < r; j++ {
			data1[j] = data[j]
		}
		appendComb := true
		if !lazyLoad {
			appendComb = isValidCombination(g, data1)
		}
		if appendComb {
			g.combinations = append(g.combinations, data1)
			g.stats.currentCombinationCount++
		}
		//util.Logf("%v", data1)
		g.state.valueUpdated = true
		return data1

	}

	for i := start; i <= end && end+1+g.maxAds >= r-index; i++ {
		if shouldUpdateAndReturn(g, start, index, r, lazyLoad, reursionCount, i, end) {
			return data
		}
		data[index] = g.slotDurations[i]
		currentDuration := i
		g.search(data, currentDuration, index+1, r, lazyLoad, reursionCount+1)
	}

	if lazyLoad && !g.state.stateUpdated {
		g.state.combinationCounter++
		index = uint64(g.state.index) - 1
		updateState(g, lazyLoad, r, reursionCount, end, g.state.combinationCounter, index, g.slotDurations[end])
	}
	return data
}

// getNextElement assuming arr contains unique values
// other wise next elemt will be returned when first matching value of val found
// returns nextValue and its index
func getNextElement(arr []uint64, val uint64) (uint64, uint64) {
	for i, e := range arr {
		if e == val && i+1 < len(arr) {
			return uint64(i) + 1, arr[i+1]
		}
	}
	// assuming durations will never be 0
	return 0, 0
}

// updateState - is used in case of lazy loading
// It maintains the state of iterator by updating the required flags
func updateState(gen *generator, lazyLoad bool, r uint64, reursionCount int, end uint64, i uint64, index uint64, valueAtEnd uint64) {

	if lazyLoad {
		gen.state.start = i
		// set c.state.index = 0 when
		// lastCombination contains, number X len(input) - 1 times starting from last index
		// where X = last number present in the input
		occurance := getOccurance(gen, valueAtEnd)
		//c.state.index = int64(c.state.combinationCounter)
		// c.state.index = int64(index)
		gen.state.index = int64(index)
		if occurance == r {
			gen.state.index = 0
		}

		// set c.state.combinationCounter
		//	c.state.combinationCounter++
		if gen.state.combinationCounter >= r || gen.state.combinationCounter >= uint64(len(gen.slotDurations)) {
			// LOGIC : to determine next value
			// 1. get the value P at 0th index present in lastCombination
			// 2. get the index of P
			// 3. determine the next index i.e. index(p) + 1 = q
			// 4. if q == r then set to 0
			diff := (uint64(len(gen.state.lastCombination)) - occurance)
			if diff > 0 {
				eleIndex := diff - 1
				gen.state.combinationCounter, _ = getNextElement(gen.slotDurations, gen.state.lastCombination[eleIndex])
				if gen.state.combinationCounter == r {
					//			c.state.combinationCounter = 0
				}
				gen.state.start = gen.state.combinationCounter
			} else {
				// end of r
			}
		}
		// set r
		// increament value of r if occurance == r
		if occurance == r {
			gen.state.start = 0
			gen.state.index = 0
			gen.state.combinationCounter = 0
			if gen.order == MinToMax {
				gen.state.r++
			}
			if gen.order == MaxToMin {
				gen.state.r--
			}
		}
		gen.state.stateUpdated = true
	}
}

// shouldUpdateAndReturn checks if states should be updated in case of lazy loading
// If required it updates the state
func shouldUpdateAndReturn(gen *generator, start, index, r uint64, lazyLoad bool, reursionCount int, i, end uint64) bool {
	if lazyLoad && gen.state.valueUpdated {
		if uint64(reursionCount) <= r && !gen.state.stateUpdated {
			updateState(gen, lazyLoad, r, reursionCount, end, i, index, gen.slotDurations[end])
		}
		return true
	}
	return false
}

// getOccurance checks how many time given number is occured in c.state.lastCombination
func getOccurance(gen *generator, valToCheck uint64) uint64 {
	occurance := uint64(0)
	for i := len(gen.state.lastCombination) - 1; i >= 0; i-- {
		if gen.state.lastCombination[i] == valToCheck {
			occurance++
		}
	}
	return occurance
}

// subtractUnwantedRepeatations ensures subtracting repeating combination counts
// from combinations count computed by compute fuction for each r = min and max ads range
func subtractUnwantedRepeatations(gen *generator) {

	series := getRepeatitionBreakUp(gen)

	// subtract repeatations from noOfCombinations
	// if not allowed for specific duration
	totalUnwantedRepeatitions := uint64(0)

	for _, noOfAds := range gen.slotDurationAdMap {

		// repeatation is not allowed for given duration
		// get how many repeation can have for the duration
		// at given level r = no of ads

		// Logic - to find repeatation for given duration at level r
		// 1. if r = 1 - repeatition = 0 for any duration
		// 2. if r = 2 - repeatition = 1 for any duration
		// 3. if r >= 3 - repeatition = noOfCombinations(r) - noOfCombinations(r-2)
		// 4. Using tetrahedral series determine the exact repeations w.r.t. noofads
		//    For Example, if noAds = 6  1 4 10 20 ...
		//	   1 => 1 repeatation for given number X in combination of 6
		//     4 => 4 repeatations for given number X  in combination of 5
		//    10 => 10 repeatations for given number X in combination of 4 (i.e. combination containing  ..,X,X,X....)
		/*
			4 5 8 7
																	4	5	8	7
			n = 4    r = 1      repeat = 4     no-repeat = 4        0	0	0	0
			n = 4    r = 2      repeat = 10    no-repeat = 6        1	1	1	1
			n = 4    r = 3      repeat = 20    no-repeat = 4		4	4	4	4
																1+3   1+3  1+3 1+3
			n = 4    r = 4      repeat = 35    no-repeat = 1		10	10	10	10
																1+3+6 1+3+6 1+3+6

																	4	5	8	7	18
			n = 5    r = 1      repeat = 5     no-repeat = 5        0	0	0	0	0
			n = 5    r = 2      repeat = 15    no-repeat = 10       1	1	1	1	1
			n = 5    r = 3      repeat = 35    no-repeat = 10		5	5	5	5	5
																1+4
			n = 5    r = 4      repeat = 70    no-repeat = 5		15	15	15	15	15
																1+4+10
			n = 5    r = 5      repeat = 126   no-repeat = 1		35	35	35	35	35
																1+4+10+20
			n = 5    r = 6      repeat = 210   no-repeat = xxx		70
																1+4+10+20+35


																	14	4
			n = 2    r = 1      repeat = 2            				0	0
			n = 2    r = 2      repeat = 3        					1	1

																	15
			n = 1    r = 1      repeat = 1            				0
			n = 1    r = 2      repeat = 1        					1
			n = 1    r = 3      repeat = 1            				1
			n = 1    r = 4      repeat = 1        					1
			n = 1    r = 5      repeat = 1	        				1


			if r = 1 => r1rpt = 0
			if r = 2 => r2rpt = 1

			if r >= 3

			r3rpt = comb(r3 - 2)
			r4rpt = comb(r4 - 2)
		*/

		for r := gen.minAds; r <= gen.maxAds; r++ {
			if r == 1 {
				// duration will no be repeated when noOfAds = 1
				continue // 0 to be subtracted
			}
			// if r == 2 {
			// 	// each duration will be repeated only once when noOfAds = 2
			// 	totalUnwantedRepeatitions++
			// 	// get total no of repeatations for combination of no > noOfAds
			// 	continue
			// }

			// r >= 3

			// find out how many repeatations are allowed for given duration
			// if allowedRepeatitions = 3, it means there are r = 3 ads for given duration
			// hence, we can allow duration repeated ranging from r= 1 to r= 3
			// i.e. durations can not be repeated beyong r = 3
			// so we should discard the repeations beyond r = 3 i.e. from r = 4 to r = maxads
			maxAllowedRepeatitions := noOfAds

			if maxAllowedRepeatitions > gen.maxAds {
				// maximum we can given upto c.maxads
				maxAllowedRepeatitions = gen.maxAds
			}

			// if maxAllowedRepeatitions = 2 then
			// repeatations > 2 should not be considered
			// compute not allowed repeatitions
			for i := maxAllowedRepeatitions + 1; i <= gen.maxAds; i++ {
				totalUnwantedRepeatitions += series[i]
			}

		}

	}
	// subtract all repeatations across all minads and maxads combinations count
	gen.stats.totalExpectedCombinations -= totalUnwantedRepeatitions
}

// getRepeatitionBreakUp
func getRepeatitionBreakUp(gen *generator) map[uint64]uint64 {
	series := make(map[uint64]uint64, gen.maxAds) // not using index 0
	ads := gen.maxAds
	series[ads] = 1
	seriesSum := uint64(1)
	// always generate from r = 3 where r is no of ads
	ads--
	for r := uint64(3); r <= gen.maxAds; r++ {
		// get repeations
		repeatations := gen.combinationCountMap[r-2]
		// get next series item
		nextItem := repeatations - seriesSum
		if repeatations == seriesSum {
			nextItem = repeatations
		}
		series[ads] = nextItem
		seriesSum += nextItem
		ads--
	}

	return series
}

// getInvalidCombinatioCount returns no of invalid combination due to one of the following reason
// 1. Contains repeatition of durations, which has only one ad with it
// 2. Sum of duration (combinationo) is out of Pod Min or Pod Max duration
func (g *generator) getInvalidCombinatioCount() int {
	return g.stats.repeatationsCount + g.stats.outOfRangeCount
}

// GetCurrentCombinationCount returns current combination count
// irrespective of whether it is valid combination
func (g *generator) GetCurrentCombinationCount() int {
	return g.stats.currentCombinationCount
}

// GetExpectedCombinationCount returns total number for possible combinations without validations
// but subtracts repeatations for duration with single ad
func (g *generator) GetExpectedCombinationCount() uint64 {
	return g.stats.totalExpectedCombinations
}

// GetOutOfRangeCombinationsCount returns number of combinations currently rejected because of
// not satisfying Pod Minimum and Maximum duration
func (g *generator) GetOutOfRangeCombinationsCount() int {
	return g.stats.outOfRangeCount
}

// GetRepeatedDurationCombinationCount returns number of combinations currently rejected because of containing
// one or more repeatations of duration values, for which partners returned only single ad
func (g *generator) GetRepeatedDurationCombinationCount() int {
	return g.stats.repeatationsCount
}

// GetValidCombinationCount returns the number of valid combinations
//  1. Within range of Pod min and max duration
//  2. Repeatations are inline with input no of ads
func (g *generator) GetValidCombinationCount() int {
	return g.stats.validCombinationCount
}
