package ctv

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

//AdSlotDurationCombinations holds all the combinations based
//on Video Ad Pod request and Bid Response Max duration
type AdSlotDurationCombinations struct {
	podMinDuration uint64 // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration uint64 // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds         uint64 // Minimum Ads value present in origin Video Ad Pod Request
	maxAds         uint64 // Maximum Ads value present in origin Video Ad Pod Request

	slotDurations     []uint64          // input slot durations for which
	slotDurationAdMap map[uint64]uint64 // map of key = duration, value = no of creatives with given duration
	noOfSlots         int               // Number of slots to be consider (from left to right)

	// cursors
	currentCombinationCount int
	validCombinationCount   int

	totalExpectedCombinations uint64     // indicates total number for possible combinations
	combinations              [][]uint64 // May contains some/all combinations at given point of time

	state snapshot

	// configurations

	// Indicates whether this algorithm should consider repetitations
	// For Example: Input durations are 10 23 40 56. For duration 23 there are
	// multiple ads present. In such case if this value is true, algorithm will generate
	// repetitations only for 23 duration.
	// NOTE: Repetitations will be of consecative durations only.
	// It means 10,23,23,23  10,23,23,56 will be generated
	// But 10,23,40,23  23, 10, 23, 23 will not be generated
	allowRepetitationsForEligibleDurations bool
}

type snapshot struct {
	/// new states
	start              uint64
	index              int64
	r                  uint64
	lastCombination    []uint64
	stateUpdated       bool
	valueUpdated       bool
	combinationCounter uint64
	// indicates how many repeating combinations skipped
	repeatingCombinationsSkipped uint64

	resetFlags bool
}

// Init ...
func (c *AdSlotDurationCombinations) Init(podMindDuration, podMaxDuration, minAds, maxAds int64, durationAdsMap []string, allowRepetitationsForEligibleDurations bool) {

	c.podMinDuration = uint64(podMindDuration)
	c.podMaxDuration = uint64(podMaxDuration)
	c.minAds = uint64(minAds)
	c.maxAds = uint64(maxAds)

	// map of key = duration value = number of ads(must be non zero positive number)
	c.slotDurationAdMap = make(map[uint64]uint64, len(c.slotDurations))

	// iterate and extract duration and number of ads belonging to the duration
	// split logic - :: separated

	cnt := 0
	c.slotDurations = make([]uint64, len(durationAdsMap))
	for _, durationAd := range durationAdsMap {
		info := strings.Split(strings.Trim(durationAd, " "), "::")
		// save durations
		duration, err := strconv.Atoi(info[0])
		if err != nil {
			print("Error in determining duration")
			return
		}

		c.slotDurations[cnt] = uint64(duration)
		// save duration  and no of ads info
		noOfAds, err := strconv.Atoi(info[1])
		if err != nil {
			print("Error in determining duration")
			return
		}
		c.slotDurationAdMap[uint64(duration)] = uint64(noOfAds)
		cnt++
	}

	c.noOfSlots = len(c.slotDurations)
	c.currentCombinationCount = 0
	c.validCombinationCount = 0
	c.state = snapshot{}

	// default configurations
	c.allowRepetitationsForEligibleDurations = allowRepetitationsForEligibleDurations

	// compute no of possible combinations (without validations)
	// using configurationss
	c.totalExpectedCombinations = compute(c, c.maxAds, true)
	// c.combinations = make([][]uint64, c.totalExpectedCombinations)
	// print("Allow Repeatation = %v", c.allowRepetitationsForEligibleDurations)
	// print("Total possible combinations (without validations) = %v ", c.totalExpectedCombinations)

	/// new states
	c.state.start = uint64(0)
	c.state.index = 0
	c.state.r = c.minAds
	c.state.resetFlags = true
}

//Next - Get next ad slot combination
//returns empty array if next combination is not present
func (c *AdSlotDurationCombinations) Next() []uint64 {
	if c.state.resetFlags {
		reset(c)
		c.state.resetFlags = false
	}
	validComb := true
	comb := make([]uint64, 0)
	for c.HasNext() && validComb {
		comb = c.search1trlazy()
		if isValidCombination(c, comb) {
			break
		}
	}
	return comb
}

func isValidCombination(c *AdSlotDurationCombinations, combination []uint64) bool {
	// check if repeatations are allowed
	repeationMap := make(map[uint64]uint64, len(c.slotDurations))
	totalAdDuration := uint64(0)
	for _, duration := range combination {
		repeationMap[uint64(duration)]++
		// check if this duration value is greater than 1 and also only 1 ad is present for
		// this duration
		if repeationMap[uint64(duration)] > 1 && c.slotDurationAdMap[uint64(duration)] == 1 {
			//print("count = %v :: Discarding combination '%v' as only 1 ad is present for duration %v", c.currentCombinationCount, combination, duration)
			return false
		}

		// check if sum of durations is withing pod min and max duration
		totalAdDuration += duration
	}

	if !(totalAdDuration >= c.podMinDuration && totalAdDuration <= c.podMaxDuration) {
		// totalAdDuration is not within range of Pod min and max duration
		//print("count = %v :: Discarding combination '%v' as either total Ad duration (%v) < %v (Pod min duration) or > %v (Pod Max duration)", c.currentCombinationCount, combination, totalAdDuration, c.podMinDuration, c.podMaxDuration)
		return false
	}
	c.validCombinationCount++
	return true
}

// HasNext - true if next combination is present
// false if not
func (c AdSlotDurationCombinations) HasNext() bool {
	return uint64(c.currentCombinationCount) < c.totalExpectedCombinations
	//return uint64(c.currentCombinationCount) < uint64(c.validCombinationCount)
}

func compute(c *AdSlotDurationCombinations, computeCombinationForTotalAds uint64, recursion bool) uint64 {
	if computeCombinationForTotalAds < c.minAds {
		return 0
	}

	var noOfCombinations *big.Int

	if c.allowRepetitationsForEligibleDurations {
		// Formula
		//		(r + n - 1)!
		//      ------------
		//       r! (n - 1)!
		n := uint64(len(c.slotDurations))
		r := uint64(computeCombinationForTotalAds)
		d1 := fact(uint64(r))
		d2 := fact(n - 1)
		d3 := d1.Mul(&d1, &d2)
		nmrt := fact(r + n - 1)

		noOfCombinations = nmrt.Div(&nmrt, d3)
	} else {
		// compute combintations without repeatation
		// Formula (Pure combination Formula)
		//			 n!
		//      ------------
		//       r! (n - r)!
		n := uint64(len(c.slotDurations))
		r := computeCombinationForTotalAds
		if r > n {
			noOfCombinations = big.NewInt(0)
			print("Can not generate combination for maxads = %v, with  %v input bid response durations and repeatations allowed", r, n)
			return noOfCombinations.Uint64()
		}
		numerator := fact(n)
		d1 := fact(r)
		d2 := fact(n - r)
		denominator := d1.Mul(&d1, &d2)
		noOfCombinations = numerator.Div(&numerator, denominator)
	}

	//print("%v", noOfCombinations)
	if recursion {
		return noOfCombinations.Uint64() + compute(c, computeCombinationForTotalAds-1, recursion)
	}
	return noOfCombinations.Uint64()
}

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

func print(format string, v ...interface{}) {
	// log.Printf(format, v...)
	fmt.Printf(format+"\n", v...)
}

func (c *AdSlotDurationCombinations) search1tr() [][]uint64 {
	reset(c)
	start := uint64(0)
	index := uint64(0)

	merged := c.combinations
	for r := c.minAds; r <= c.maxAds; r++ {
		data := make([]uint64, r)
		c.search1(data, start, index, r, merged, false, 0)
	}
	// print("Total combinations generated = %v", c.currentCombinationCount)
	// print("Total combinations expected = %v", c.totalExpectedCombinations)
	// result := make([][]uint64, c.totalExpectedCombinations)
	result := make([][]uint64, c.validCombinationCount)
	copy(result, c.combinations)
	c.currentCombinationCount = 0
	return result
}

func reset(c *AdSlotDurationCombinations) {
	c.currentCombinationCount = 0
	c.validCombinationCount = 0
}

func (c *AdSlotDurationCombinations) search1trlazy() []uint64 {

	//	merged := make([][]uint64, c.totalExpectedCombinations)
	start := c.state.start
	index := c.state.index
	r := c.state.r

	// reset last combination
	// by deleting previous values
	if c.state.lastCombination == nil {
		c.combinations = make([][]uint64, 0)
	}
	merged := c.combinations
	// data := make([]uint64, r)
	data := new([]uint64)
	data = &c.state.lastCombination
	if *data == nil || uint64(len(*data)) != r {
		*data = make([]uint64, r)
	}
	c.state.stateUpdated = false
	c.state.valueUpdated = false
	c.state.repeatingCombinationsSkipped = 0
	for ; r <= c.maxAds; r++ {

		//common change
		//index = index + 1
		if c.allowRepetitationsForEligibleDurations {
		} else {
			// start = start + 1
		}
		c.search1(*data, start, uint64(index), r, merged, true, 0)
		c.state.stateUpdated = false // reset
		c.state.valueUpdated = false
		c.state.repeatingCombinationsSkipped = 0
		break
	}

	result := make([]uint64, len(*data))
	copy(result, *data)
	return result
}

//
//static void combinationUtil(
//	int arr[], int data[], int start, int end, int index, int r)
func (c *AdSlotDurationCombinations) search1(data []uint64, start, index, r uint64, merged [][]uint64, lazyLoad bool, reursionCount int) []uint64 {

	end := uint64(len(c.slotDurations) - 1)

	// Current combination is ready to be printed, print it
	if index == r {
		data1 := make([]uint64, len(data))
		for j := uint64(0); j < r; j++ {
			// fmt.Print(data[j])
			// fmt.Print(" ")
			data1[j] = data[j]
		}
		appendComb := true
		if !lazyLoad {
			appendComb = isValidCombination(c, data1)
		}
		if appendComb {
			c.combinations = append(c.combinations, data1)
			c.currentCombinationCount++
		}
		// println("")
		// fmt.Println(c.currentCombinationCount, " :: index	=", index, ", i=", start, " :: slot = ", data)
		//fmt.Println(data1)
		c.state.valueUpdated = true
		return data1

	}

	if c.allowRepetitationsForEligibleDurations {
		for i := start; i <= end && end+1+c.maxAds >= r-index; i++ {
			if shouldUpdateAndReturn(c, start, index, r, merged, lazyLoad, reursionCount, i, end) {
				return data
			}
			data[index] = c.slotDurations[i]
			//fmt.Println(c.currentCombinationCount, " :: index =", index, ", recursioncnt = ", reursionCount, ", i=", start, " :: slot = ", data)

			currentDuration := i
			// increment duration when
			// 1. duration has only single ad
			// 2. data[index] contains that ad and now going to set same duration on next index
			// if c.slotDurationAdMap[c.slotDurations[i]] == 1 && index+1 < r && data[index] == c.slotDurations[i] {
			// 	// no repeations
			// 	currentDuration = i + 1
			// 	c.state.repeatingCombinationsSkipped++
			// }

			c.search1(data, currentDuration, index+1, r, merged, lazyLoad, reursionCount+1)
		}
	} else {

		// replace index with all possible elements. The condition
		// "end-i+1 >= r-index" makes sure that including one element
		// at index will make a combination with remaining elements
		// at remaining positions

		for i := start; i <= end && end-i+1 >= r-index; i++ {
			if shouldUpdateAndReturn(c, start, index, r, merged, lazyLoad, reursionCount, i, end) {
				return data
			}
			data[index] = c.slotDurations[i]
			//fmt.Println(c.currentCombinationCount, " :: index =", index, ", recursioncnt = ", reursionCount, ", i=", start, " :: slot = ", data)
			c.search1(data, i+1, index+1, r, merged, lazyLoad, reursionCount+1)

		}
	}

	if lazyLoad && !c.state.stateUpdated {
		c.state.combinationCounter++
		index = uint64(c.state.index) - 1 + c.state.repeatingCombinationsSkipped
		//index = uint64(c.state.index) - 1
		updateState(c, lazyLoad, r, reursionCount, end, c.state.combinationCounter, index, c.slotDurations[end])

	}
	return data
}

// assuming arr contains unique values
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

func updateState(c *AdSlotDurationCombinations, lazyLoad bool, r uint64, reursionCount int, end uint64, i uint64, index uint64, valueAtEnd uint64) {

	//valueAtEnd := c.slotDurations[end]

	// if lazyLoad && uint64(reursionCount) == r-1 {
	if lazyLoad {

		// set index

		// c.state.start = c.state.combinationCounter

		c.state.start += c.state.repeatingCombinationsSkipped

		c.state.start = i
		// set c.state.index = 0 when
		// lastCombination contains, number X len(input) - 1 times starting from last index
		// where X = last number present in the input
		occurance := getOccurance(c, valueAtEnd)
		//c.state.index = int64(c.state.combinationCounter)
		// c.state.index = int64(index)
		c.state.index = int64(index)
		if occurance == r {
			c.state.index = 0
		}

		// set c.state.combinationCounter
		//	c.state.combinationCounter++
		if c.state.combinationCounter >= r || c.state.combinationCounter >= uint64(len(c.slotDurations)) {
			// LOGIC : to determine next value
			// 1. get the value P at 0th index present in lastCombination
			// 2. get the index of P
			// 3. determine the next index i.e. index(p) + 1 = q
			// 4. if q == r then set to 0
			diff := (uint64(len(c.state.lastCombination)) - occurance)
			if diff > 0 {
				eleIndex := diff - 1
				c.state.combinationCounter, _ = getNextElement(c.slotDurations, c.state.lastCombination[eleIndex])
				if c.state.combinationCounter == r {
					//			c.state.combinationCounter = 0
				}
				c.state.start = c.state.combinationCounter
			} else {
				// fmt.Println("Must be end of r ", r)
			}
		}

		// Use case: lastCombination (To be given outside) contains duplicate durations
		// for which only single ad is present
		// when above  for loop inside search1 not able to detect repeatations
		// for the durations, which contains only 1 ad
		// While developing, it is typically observed at the end of each  combination

		// check if the duration excepts only single ad
		/*if occurance > 1 && c.slotDurationAdMap[valueAtEnd] == 1 {
			//adjust lastduration
			i := int64(occurance)
			for ; i < int64(len(c.state.lastCombination)); i++ {
				c.state.lastCombination[i] = c.slotDurations[c.state.start]
			}
			c.state.start++
			c.state.index = i - 1
			c.state.combinationCounter++

			// check if slot value w.r.t. c.state.start has only single ad
			// in such case call update state to get next combination
			valueAtEnd := c.slotDurations[c.state.start-1]
			for c.slotDurationAdMap[valueAtEnd] == 1 {
				updateState(c, lazyLoad, r, reursionCount, end, c.state.start, uint64(c.state.index), valueAtEnd)
				break
			}

		}*/

		// set r
		// increament value of r if occurance == r
		if occurance == r {
			c.state.start = 0
			c.state.index = 0
			c.state.combinationCounter = 0
			c.state.r++
		}

		c.state.stateUpdated = true
	}
}

func shouldUpdateAndReturn(c *AdSlotDurationCombinations, start, index, r uint64, merged [][]uint64, lazyLoad bool, reursionCount int, i, end uint64) bool {
	if lazyLoad && c.state.valueUpdated {
		if uint64(reursionCount) <= r && !c.state.stateUpdated {
			updateState(c, lazyLoad, r, reursionCount, end, i, index, c.slotDurations[end])
		}
		// if reursionCount == 1 {
		// 	break
		// }

		// if i <= end && end+1+c.maxAds >= r-index {
		// 	_index = uint64(index)
		// }
		return true
	}
	return false
}

func getOccurance(c *AdSlotDurationCombinations, valToCheck uint64) uint64 {
	occurance := uint64(0)
	for i := len(c.state.lastCombination) - 1; i >= 0; i-- {
		if c.state.lastCombination[i] == valToCheck {
			occurance++
		}
	}
	return occurance
}
