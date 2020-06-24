package impressions

import (
	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
)

// generator contains Pod Minimum Duration, Pod Maximum Duration, Slot Minimum Duration and Slot Maximum Duration
// It holds additional attributes required by this algorithm for  internal computation.
// 	It contains Slots attribute. This  attribute holds the output of this algorithm
type generator struct {
	IImpressions
	Slots             [][2]int64 // Holds Minimum and Maximum duration (in seconds) for each Ad Slot. Length indicates total number of Ad Slots/ Impressions for given Ad Pod
	totalSlotMaxTime  *int64     // Total Sum of all Ad Slot Max durations (in seconds)
	totalSlotMinTime  *int64     // Total Sum of all Ad Slot Min durations (in seconds)
	freeTime          int64      // Remaining Time (in seconds) not allocated. It is compared with RequestedPodMaxDuration
	slotsWithZeroTime *int64     // Indicates number of slots with zero time (starting from 1).
	// requested holds all the requested information received
	requested pod
	// internal  holds the value closed to original value and multiples of X.
	internal pod

	// indicates how minduration should be populated for each impression break
	// NOTE: This is only gets honoured when pod min duration != pod max duration
	// i.e. range of pod duration is given
	//
	// true  - use config.requested.slotMinDuration as minDuration
	// false - user computed maxDuration as minDuration
	setMinDurationFromRequest bool
}

// pod for internal computation
// should not be used outside
type pod struct {
	minAds          int64
	maxAds          int64
	slotMinDuration int64
	slotMaxDuration int64
	podMinDuration  int64
	podMaxDuration  int64
}

// Get returns the number of Ad Slots/Impression  that input Ad Pod can have.
// It returns List 2D array containing following
//  1. Dimension 1 - Represents the minimum duration of an impression
//  2. Dimension 2 - Represents the maximum duration of an impression
func (config *generator) Get() [][2]int64 {
	ctv.Logf("Using minDurationPolicy = %v ", config.setMinDurationFromRequest)
	ctv.Logf("Pod Config with Internal Computation (using multiples of %v) = %+v\n", multipleOf, config)
	totalAds := computeTotalAds(*config)
	timeForEachSlot := computeTimeForEachAdSlot(*config, totalAds)

	config.Slots = make([][2]int64, totalAds)
	config.slotsWithZeroTime = new(int64)
	*config.slotsWithZeroTime = totalAds
	ctv.Logf("Plotted Ad Slots / Impressions of size = %v\n", len(config.Slots))
	// iterate over total time till it is < cfg.RequestedPodMaxDuration
	time := int64(0)
	ctv.Logf("Started allocating durations to each Ad Slot / Impression\n")
	fillZeroSlotsOnPriority := true
	noOfZeroSlotsFilledByLastRun := int64(0)
	*config.totalSlotMaxTime = 0
	for time < config.requested.podMaxDuration {
		adjustedTime, slotsFull := config.addTime(timeForEachSlot, fillZeroSlotsOnPriority)
		time += adjustedTime
		timeForEachSlot = computeTimeLeastValue(config.requested.podMaxDuration-time, config.requested.slotMaxDuration-timeForEachSlot)
		if slotsFull {
			ctv.Logf("All slots are full of their capacity. validating slots\n")
			break
		}

		// instruct for filling zero capacity slots on priority if
		// 1. shouldAdjustSlotWithZeroDuration returns true
		// 2. there are slots with 0 duration
		// 3. there is at least ont slot with zero duration filled by last iteration
		fillZeroSlotsOnPriority = false
		noOfZeroSlotsFilledByLastRun = *config.slotsWithZeroTime - noOfZeroSlotsFilledByLastRun
		if config.shouldAdjustSlotWithZeroDuration() && *config.slotsWithZeroTime > 0 && noOfZeroSlotsFilledByLastRun > 0 {
			fillZeroSlotsOnPriority = true
		}
	}
	ctv.Logf("Completed allocating durations to each Ad Slot / Impression\n")

	// validate slots
	config.validateSlots()

	// log free time if present to stats server
	// also check algoritm computed the no. of ads
	if config.requested.podMaxDuration-time > 0 && len(config.Slots) > 0 {
		config.freeTime = config.requested.podMaxDuration - time
		ctv.Logf("TO STATS SERVER : Free Time not allocated %v sec", config.freeTime)
	}

	ctv.Logf("\nTotal Impressions = %v, Total Allocated Time = %v sec (out of %v sec, Max Pod Duration)\n%v", len(config.Slots), *config.totalSlotMaxTime, config.requested.podMaxDuration, config.Slots)
	return config.Slots
}

// Returns total number of Ad Slots/ impressions that the Ad Pod can have
func computeTotalAds(cfg generator) int64 {
	if cfg.internal.slotMaxDuration <= 0 || cfg.internal.slotMinDuration <= 0 {
		ctv.Logf("Either cfg.slotMaxDuration or cfg.slotMinDuration or both are <= 0. Hence, totalAds = 0")
		return 0
	}
	minAds := cfg.internal.podMaxDuration / cfg.internal.slotMaxDuration
	maxAds := cfg.internal.podMaxDuration / cfg.internal.slotMinDuration

	ctv.Logf("Computed minAds = %v , maxAds = %v\n", minAds, maxAds)

	totalAds := max(minAds, maxAds)
	ctv.Logf("Computed max(minAds, maxAds) = totalAds = %v\n", totalAds)

	if totalAds < cfg.requested.minAds {
		totalAds = cfg.requested.minAds
		ctv.Logf("Computed totalAds < requested  minAds (%v). Hence, setting totalAds =  minAds = %v\n", cfg.requested.minAds, totalAds)
	}
	if totalAds > cfg.requested.maxAds {
		totalAds = cfg.requested.maxAds
		ctv.Logf("Computed totalAds > requested  maxAds (%v). Hence, setting totalAds =  maxAds = %v\n", cfg.requested.maxAds, totalAds)
	}
	ctv.Logf("Computed Final totalAds = %v  [%v <= %v <= %v]\n", totalAds, cfg.requested.minAds, totalAds, cfg.requested.maxAds)
	return totalAds
}

// Returns duration in seconds that can be allocated to each Ad Slot
// Accepts cfg containing algorithm configurations and totalAds containing Total number of
// Ad Slots / Impressions that the Ad Pod can have.
func computeTimeForEachAdSlot(cfg generator, totalAds int64) int64 {
	// Compute time for each ad
	if totalAds <= 0 {
		ctv.Logf("totalAds = 0, Hence timeForEachSlot = 0")
		return 0
	}
	timeForEachSlot := cfg.internal.podMaxDuration / totalAds

	ctv.Logf("Computed timeForEachSlot = %v (podMaxDuration/totalAds) (%v/%v)\n", timeForEachSlot, cfg.internal.podMaxDuration, totalAds)

	if timeForEachSlot < cfg.internal.slotMinDuration {
		timeForEachSlot = cfg.internal.slotMinDuration
		ctv.Logf("Computed timeForEachSlot < requested  slotMinDuration (%v). Hence, setting timeForEachSlot =  slotMinDuration = %v\n", cfg.internal.slotMinDuration, timeForEachSlot)
	}

	if timeForEachSlot > cfg.internal.slotMaxDuration {
		timeForEachSlot = cfg.internal.slotMaxDuration
		ctv.Logf("Computed timeForEachSlot > requested  slotMaxDuration (%v). Hence, setting timeForEachSlot =  slotMaxDuration = %v\n", cfg.internal.slotMaxDuration, timeForEachSlot)
	}

	// Case - Exact slot duration is given. No scope for finding multiples
	// of given number. Prefer to return computed timeForEachSlot
	// In such case timeForEachSlot no necessarily to be multiples of given number
	if cfg.requested.slotMinDuration == cfg.requested.slotMaxDuration {
		ctv.Logf("requested.slotMinDuration = requested.slotMinDuration = %v. Hence, not computing multiples of %v value.", cfg.requested.slotMaxDuration, multipleOf)
		return timeForEachSlot
	}

	// Case I- adjusted timeForEachSlot may be pushed to and fro by
	// slot min and max duration (multiples of given number)
	// Case II - timeForEachSlot*totalAds > podmaxduration
	// In such case prefer to return cfg.podMaxDuration / totalAds
	// In such case timeForEachSlot no necessarily to be multiples of given number
	if timeForEachSlot < cfg.internal.slotMinDuration || timeForEachSlot > cfg.internal.slotMaxDuration || (timeForEachSlot*totalAds) > cfg.requested.podMaxDuration {
		ctv.Logf("timeForEachSlot (%v) < cfg.internal.slotMinDuration (%v) || timeForEachSlot (%v) > cfg.internal.slotMaxDuration (%v) || timeForEachSlot*totalAds (%v) > cfg.requested.podMaxDuration (%v) ", timeForEachSlot, cfg.internal.slotMinDuration, timeForEachSlot, cfg.internal.slotMaxDuration, timeForEachSlot*totalAds, cfg.requested.podMaxDuration)
		ctv.Logf("Hence, not computing multiples of %v value.", multipleOf)
		// need that division again
		return cfg.internal.podMaxDuration / totalAds
	}

	// ensure timeForEachSlot is multipleof given number
	if !isMultipleOf(timeForEachSlot, multipleOf) {
		// get close to value of multiple
		// here we muse get either cfg.SlotMinDuration or cfg.SlotMaxDuration
		// these values are already pre-computed in multiples of given number
		timeForEachSlot = getClosestFactor(timeForEachSlot, multipleOf)
		ctv.Logf("Computed closet factor %v, in multiples of %v for timeForEachSlot\n", timeForEachSlot, multipleOf)
	}
	ctv.Logf("Computed Final timeForEachSlot = %v  [%v <= %v <= %v]\n", timeForEachSlot, cfg.requested.slotMinDuration, timeForEachSlot, cfg.requested.slotMaxDuration)
	return timeForEachSlot
}

// Checks if multipleOf can be used as least time value
// this will ensure eack slot to maximize its time if possible
// if multipleOf can not be used as least value then default input value is returned as is
// accepts time containing, which least value to be computed.
// leastTimeRequiredByEachSlot - indicates the mimimum time that any slot can accept (UOE-5268)
// Returns the least value based on multiple of X
func computeTimeLeastValue(time int64, leastTimeRequiredByEachSlot int64) int64 {
	// time if Testcase#6
	// 1. multiple of x - get smallest factor N of multiple of x for time
	// 2. not multiple of x - try to obtain smallet no N multipe of x
	// ensure N <= timeForEachSlot
	leastFactor := multipleOf
	if leastFactor < time {
		time = leastFactor
	}

	// case:  check if slots are looking for time < leastFactor
	// UOE-5268
	if leastTimeRequiredByEachSlot > 0 && leastTimeRequiredByEachSlot < time {
		time = leastTimeRequiredByEachSlot
	}

	return time
}

// Validate the algorithm computations
//  1. Verifies if 2D slice containing Min duration and Max duration values are non-zero
//  2. Idenfies the Ad Slots / Impressions with either Min Duration or Max Duration or both
//     having zero value and removes it from 2D slice
//  3. Ensures  Minimum Pod duration <= TotalSlotMaxTime <= Maximum Pod Duration
// if  any validation fails it removes all the alloated slots and  makes is of size 0
// and sets the freeTime value as RequestedPodMaxDuration
func (config *generator) validateSlots() {

	// default return value if validation fails
	emptySlots := make([][2]int64, 0)
	if len(config.Slots) == 0 {
		return
	}

	hasError := false

	// check slot with 0 values
	// remove them from config.Slots
	emptySlotCount := 0
	for index, slot := range config.Slots {
		*config.totalSlotMinTime += slot[0]
		if slot[0] == 0 || slot[1] == 0 {
			ctv.Logf("WARNING:Slot[%v][%v] is having 0 duration\n", index, slot)
			emptySlotCount++
			continue
		}

		// check slot min and max duration
		hasError = !isValidSlotDuration(*config, slot[0], index, "Min") || !isValidSlotDuration(*config, slot[1], index, "Max")
	}

	// remove empty slot
	if emptySlotCount > 0 {
		optimizedSlots := make([][2]int64, len(config.Slots)-emptySlotCount)
		for index, slot := range config.Slots {
			if slot[0] == 0 || slot[1] == 0 {
			} else {
				optimizedSlots[index][0] = slot[0]
				optimizedSlots[index][1] = slot[1]
			}
		}
		config.Slots = optimizedSlots
		ctv.Logf("Removed %v empty slots\n", emptySlotCount)
	}

	// check number of slots are within range of requested minAds and maxAds
	if int64(len(config.Slots)) < config.requested.minAds || int64(len(config.Slots)) > config.requested.maxAds {
		ctv.Logf("ERROR: slotSize %v is either less than Min Ads (%v) or greater than Max Ads (%v)\n", len(config.Slots), config.requested.minAds, config.requested.maxAds)
		hasError = true
	}

	// validate total of slot min and max duration
	hasError = hasError || !isValidTotalSlotTime(*config, *config.totalSlotMinTime, "Min") || !isValidTotalSlotTime(*config, *config.totalSlotMaxTime, "Max")

	if hasError {
		config.Slots = emptySlots
		config.freeTime = config.requested.podMaxDuration
	}
}

// Adds time to possible slots and returns total added time
//
// Checks following for each Ad Slot
//  1. Can Ad Slot adjust the input time
//  2. If addition of new time to any slot not exeeding Total Pod Max Duration
// Performs the following operations
//  1. Populates Minimum duration slot[][0] - Either Slot Minimum Duration or Actual Slot Time computed
//  2. Populates Maximum duration slot[][1] - Always actual Slot Time computed
//  3. Counts the number of Ad Slots / Impressons full with  duration  capacity. If all Ad Slots / Impressions
//     are full of capacity it returns true as second return argument, indicating all slots are full with capacity
//  4. Keeps track of TotalSlotDuration when each new time is added to the Ad Slot
//  5. Keeps track of difference between computed PodMaxDuration and RequestedPodMaxDuration (TestCase #16) and used in step #2 above
// Returns argument 1 indicating total time adusted, argument 2 whether all slots are full of duration capacity
func (config generator) addTime(timeForEachSlot int64, fillZeroSlotsOnPriority bool) (int64, bool) {
	time := int64(0)

	// iterate over each ad
	slotCountFullWithCapacity := 0
	for ad := int64(0); ad < int64(len(config.Slots)); ad++ {

		slot := &config.Slots[ad]
		// check
		// 1. time(slot(0)) <= config.SlotMaxDuration
		// 2. if adding new time  to slot0 not exeeding config.SlotMaxDuration
		// 3. if sum(slot time) +  timeForEachSlot  <= config.RequestedPodMaxDuration
		canAdjustTime := (slot[1]+timeForEachSlot) <= config.requested.slotMaxDuration && (slot[1]+timeForEachSlot) >= config.requested.slotMinDuration
		totalSlotMaxTimeWithNewTimeLessThanRequestedPodMaxDuration := *config.totalSlotMaxTime+timeForEachSlot <= config.requested.podMaxDuration

		// if fillZeroSlotsOnPriority= true ensure current slot value =  0
		allowCurrentSlot := !fillZeroSlotsOnPriority || (fillZeroSlotsOnPriority && slot[1] == 0)
		if slot[1] <= config.internal.slotMaxDuration && canAdjustTime && totalSlotMaxTimeWithNewTimeLessThanRequestedPodMaxDuration && allowCurrentSlot {
			slot[0] += timeForEachSlot

			// if we are adjusting the free time which will match up with config.RequestedPodMaxDuration
			// then set config.SlotMinDuration as min value for this slot
			// TestCase #16
			// UOE-5379 : config.setMinDurationFromRequest
			if timeForEachSlot < multipleOf || config.setMinDurationFromRequest {
				// override existing value of slot[0] here
				slot[0] = config.requested.slotMinDuration
			}

			// check if this slot duration was zero
			if slot[1] == 0 {
				// decrememt config.slotsWithZeroTime as we added some time for this slot
				*config.slotsWithZeroTime--
			}

			slot[1] += timeForEachSlot
			*config.totalSlotMaxTime += timeForEachSlot
			time += timeForEachSlot
			ctv.Logf("Slot %v = Added %v sec (New Time = %v)\n", ad, timeForEachSlot, slot[1])
		}
		// check slot capabity
		// !canAdjustTime - TestCase18
		// UOE-5268 - Check with Requested Slot Max Duration
		if slot[1] == config.requested.slotMaxDuration || !canAdjustTime {
			// slot is full
			slotCountFullWithCapacity++
		}
	}
	ctv.Logf("adjustedTime = %v\n ", time)
	return time, slotCountFullWithCapacity == len(config.Slots)
}

//shouldAdjustSlotWithZeroDuration - returns if slot with zero durations should be filled
// Currently it will return true in following condition
// cfg.minAds = cfg.maxads (i.e. Exact number of ads are required)
func (config generator) shouldAdjustSlotWithZeroDuration() bool {
	if config.requested.minAds == config.requested.maxAds {
		return true
	}
	return false
}

// isValidSlotDuration ensures duration value is within requested slot min and max duration
// returns true of duration is valid. false otherwise
func isValidSlotDuration(config generator, slotDuration int64, index int, fieldType string) bool {
	if slotDuration < config.requested.slotMinDuration || slotDuration > config.requested.slotMaxDuration {
		ctv.Logf("ERROR: Slot%v %v Duration %v sec is out of either requested.slotMinDuration (%v) or requested.slotMaxDuration (%v)\n", index, fieldType, slotDuration, config.requested.slotMinDuration, config.requested.slotMaxDuration)
		return false
	}
	return true
}

// isValidTotalSlotTime ensures totalSlotTime is within range of Pod Min Duration and Pod Max Duration
// in case PodMinDuration = PodMaxDuration then it checks totalSlotTime = PodMaxDuration
// returns true if totalSlotTime lies between min/max pod durations or equal to pod duration.
// false otherwise
func isValidTotalSlotTime(config generator, totalSlotTime int64, fieldType string) bool {
	// ensure if min pod duration = max pod duration
	// totalSlotTime = pod duration
	if config.requested.podMinDuration == config.requested.podMaxDuration && totalSlotTime != config.requested.podMaxDuration {
		ctv.Logf("ERROR: Total Slot %v Duration %v sec is not matching with Total Pod Duration %v sec\n", fieldType, totalSlotTime, config.requested.podMaxDuration)
		return false
	}

	// ensure slot duration lies between requested min pod duration and  requested max pod duration
	// Testcase #15
	if totalSlotTime < config.requested.podMinDuration || totalSlotTime > config.requested.podMaxDuration {
		ctv.Logf("ERROR: Total Slot %v Duration %v sec is either less than Requested Pod Min Duration (%v sec) or greater than Requested  Pod Max Duration (%v sec)\n", fieldType, totalSlotTime, config.requested.podMinDuration, config.requested.podMaxDuration)
		return false
	}
	return true
}
