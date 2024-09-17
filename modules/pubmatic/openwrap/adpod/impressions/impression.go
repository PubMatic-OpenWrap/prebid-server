package impressions

import (
	"math"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// Value use to compute Ad Slot Durations and Pod Durations for internal computation
// Right now this value is set to 5, based on passed data observations
// Observed that typically video impression contains contains minimum and maximum duration in multiples of  5
const (
	MultipleOf         = 5
	ImpressionIDFormat = `%v` + models.ImpressionIDSeparator + `%v`
)

// ImpGenerator ...
type ImpGenerator interface {
	Get() [][2]int64
	// Algorithm() int // returns algorithm used for computing number of impressions
}

// SelectAlgorithm is factory function which will return valid Algorithm based on adpod parameters
// Return Value:
//   - MinMaxAlgorithm (default)
//   - ByDurationRanges: if reqAdPod extension has VideoAdDuration and VideoAdDurationMatchingPolicy is "exact" algorithm
func SelectAlgorithm(reqAdPod *models.AdPod, adpodProfileCfg *models.AdpodProfileConfig) int {
	if reqAdPod != nil {
		if adpodProfileCfg != nil && len(adpodProfileCfg.AdserverCreativeDurations) > 0 &&
			(models.OWExactVideoAdDurationMatching == adpodProfileCfg.AdserverCreativeDurationMatchingPolicy || models.OWRoundupVideoAdDurationMatching == adpodProfileCfg.AdserverCreativeDurationMatchingPolicy) {
			return models.ByDurationRanges
		}
	}
	return models.MinMaxAlgorithm
}

// NewImpressions generate object of impression generator
// based on input algorithm type
// if invalid algorithm type is passed, it returns default algorithm which will compute
// impressions based on minimum ad slot duration
func NewImpressions(podMinDuration, podMaxDuration int64, adpod *models.AdPod, adpodProfileCfg *models.AdpodProfileConfig, algorithm int) ImpGenerator {
	switch algorithm {
	case models.MaximizeForDuration:
		g := newMaximizeForDuration(podMinDuration, podMaxDuration, adpod)
		return &g

	case models.MinMaxAlgorithm:
		g := newMinMaxAlgorithm(podMinDuration, podMaxDuration, adpod)
		return &g

	case models.ByDurationRanges:
		g := newByDurationRanges(adpodProfileCfg, int(adpod.MaxAds), adpod.MinDuration, adpod.MaxDuration)
		return &g
	}

	// return default algorithm with slot durations set to minimum slot duration
	// util.Logf("Selected 'DefaultAlgorithm'")
	defaultGenerator := newConfig(podMinDuration, podMinDuration, adpod)
	return &defaultGenerator
}

// newConfigWithMultipleOf initializes the generator instance
// it internally calls newConfig to obtain the generator instance
// then it computes closed to factor basedon 'multipleOf' parameter value
// and accordingly determines the Pod Min/Max and Slot Min/Max values for internal
// computation only.
func newConfigWithMultipleOf(podMinDuration, podMaxDuration int64, vPod *models.AdPod, multipleOf int) generator {
	config := newConfig(podMinDuration, podMaxDuration, vPod)

	// try to compute slot level min and max duration values in multiple of
	// given number. If computed values are overlapping then prefer requested
	if config.requested.slotMinDuration == config.requested.slotMaxDuration {
		config.internal.slotMinDuration = config.requested.slotMinDuration
		config.internal.slotMaxDuration = config.requested.slotMaxDuration
	} else {
		config.internal.slotMinDuration = getClosestFactorForMinDuration(config.requested.slotMinDuration, int64(multipleOf))
		config.internal.slotMaxDuration = getClosestFactorForMaxDuration(config.requested.slotMaxDuration, int64(multipleOf))
		config.internal.slotDurationComputed = true
		if config.internal.slotMinDuration > config.internal.slotMaxDuration {
			// computed slot min duration > computed slot max duration
			// avoid overlap and prefer requested values
			config.internal.slotMinDuration = config.requested.slotMinDuration
			config.internal.slotMaxDuration = config.requested.slotMaxDuration
			// update marker indicated slot duation values are not computed
			// this required by algorithm in computeTimeForEachAdSlot function
			config.internal.slotDurationComputed = false
		}
	}
	return config
}

// newConfig initializes the generator instance
func newConfig(podMinDuration, podMaxDuration int64, vPod *models.AdPod) generator {
	config := generator{}
	config.totalSlotTime = new(int64)
	// configure requested pod
	config.requested = pod{
		podMinDuration:  podMinDuration,
		podMaxDuration:  podMaxDuration,
		slotMinDuration: int64(vPod.MinDuration),
		slotMaxDuration: int64(vPod.MaxDuration),
		minAds:          int64(vPod.MinAds),
		maxAds:          int64(vPod.MaxAds),
	}

	// configure internal object (FOR INTERNAL USE ONLY)
	// this  is used for internal computation and may contains modified values of
	// slotMinDuration and slotMaxDuration in multiples of multipleOf factor
	// This function will by deault intialize this pod with same values
	// as of requestedPod
	// There is another function newConfigWithMultipleOf, which computes and assigns
	// values to this object
	config.internal = internal{
		slotMinDuration: config.requested.slotMinDuration,
		slotMaxDuration: config.requested.slotMaxDuration,
	}
	return config
}

// Returns closest factor for num, with  respect  input multipleOf
//
//	Example: Closest Factor of 9, in multiples of 5 is '10'
func getClosestFactor(num, multipleOf int64) int64 {
	return int64(math.Round(float64(num)/float64(multipleOf)) * float64(multipleOf))
}

// Returns closestfactor of MinDuration, with  respect to multipleOf
// If computed factor < MinDuration then it will ensure and return
// close factor >=  MinDuration
func getClosestFactorForMinDuration(MinDuration, multipleOf int64) int64 {
	closedMinDuration := getClosestFactor(MinDuration, multipleOf)

	if closedMinDuration == 0 {
		return multipleOf
	}

	if closedMinDuration < MinDuration {
		return closedMinDuration + multipleOf
	}

	return closedMinDuration
}

// Returns closestfactor of maxduration, with  respect to multipleOf
// If computed factor > maxduration then it will ensure and return
// close factor <=  maxduration
func getClosestFactorForMaxDuration(maxduration, multipleOf int64) int64 {
	closedMaxDuration := getClosestFactor(maxduration, multipleOf)
	if closedMaxDuration == maxduration {
		return maxduration
	}

	// set closest maxduration closed to maxduration
	for i := closedMaxDuration; i <= maxduration; {
		if closedMaxDuration < maxduration {
			closedMaxDuration = i + multipleOf
			i = closedMaxDuration
		}
	}

	if closedMaxDuration > maxduration {
		duration := closedMaxDuration - multipleOf
		if duration == 0 {
			// return input value as is instead of zero to avoid NPE
			return maxduration
		}
		return duration
	}

	return closedMaxDuration
}

// Returns Maximum number out off 2 input numbers
func max(num1, num2 int64) int64 {
	if num1 >= num2 {
		return num1
	}

	return num2
}

// Returns true if num is multipleof second argument. False otherwise
func isMultipleOf(num, multipleOf int64) bool {
	return math.Mod(float64(num), float64(multipleOf)) == 0
}
