package impressions

import (
	"math"

	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

func newConfig(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod) adPodConfig {
	config := adPodConfig{}

	config.requestedPodMinDuration = podMinDuration
	config.requestedPodMaxDuration = podMaxDuration

	config.requestedSlotMinDuration = int64(*vPod.MinDuration)
	config.requestedSlotMaxDuration = int64(*vPod.MaxDuration)

	// init as if multiple of 1
	config.podMinDuration = config.requestedPodMinDuration
	config.podMaxDuration = config.requestedPodMaxDuration
	config.slotMinDuration = config.requestedSlotMinDuration
	config.slotMaxDuration = config.requestedSlotMaxDuration

	config.minAds = int64(*vPod.MinAds)
	config.maxAds = int64(*vPod.MaxAds)
	config.totalSlotTime = new(int64)
	return config
}

func newConfigWithMultipleOf(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod, multipleOf int64) adPodConfig {
	config := newConfig(podMinDuration, podMaxDuration, vPod)
	if config.requestedPodMinDuration == config.requestedPodMaxDuration {
		/*TestCase 16*/
		ctv.Logf("requestedPodMinDuration = requestedPodMaxDuration = %v\n", config.requestedPodMinDuration)
		config.podMinDuration = config.requestedPodMinDuration
		config.podMaxDuration = config.podMinDuration
	} else {
		config.podMinDuration = getClosetFactorForMinDuration(config.requestedPodMinDuration, multipleOf)
		config.podMaxDuration = getClosetFactorForMaxDuration(config.requestedPodMaxDuration, multipleOf)
	}

	if config.requestedSlotMinDuration == config.requestedSlotMaxDuration {
		/*TestCase 30*/
		ctv.Logf("requestedSlotMinDuration = requestedSlotMaxDuration = %v\n", config.requestedPodMinDuration)
		config.slotMinDuration = config.requestedSlotMinDuration
		config.slotMaxDuration = config.slotMinDuration
	} else {
		config.slotMinDuration = getClosetFactorForMinDuration(int64(config.requestedSlotMinDuration), multipleOf)
		config.slotMaxDuration = getClosetFactorForMaxDuration(int64(config.requestedSlotMaxDuration), multipleOf)
	}
	return config
}

// Returns true if num is multipleof second argument. False otherwise
func isMultipleOf(num, multipleOf int64) bool {
	return math.Mod(float64(num), float64(multipleOf)) == 0
}

// Returns closet factor for num, with  respect  input multipleOf
//  Example: Closet Factor of 9, in multiples of 5 is '10'
func getClosetFactor(num, multipleOf int64) int64 {
	return int64(math.Round(float64(num)/float64(multipleOf)) * float64(multipleOf))
}

// Returns closetfactor of MinDuration, with  respect to multipleOf
// If computed factor < MinDuration then it will ensure and return
// close factor >=  MinDuration
func getClosetFactorForMinDuration(MinDuration int64, multipleOf int64) int64 {
	closedMinDuration := getClosetFactor(MinDuration, multipleOf)

	if closedMinDuration == 0 {
		return multipleOf
	}

	if closedMinDuration == MinDuration {
		return MinDuration
	}

	if closedMinDuration < MinDuration {
		return closedMinDuration + multipleOf
	}

	return closedMinDuration
}

// Returns closetfactor of maxduration, with  respect to multipleOf
// If computed factor > maxduration then it will ensure and return
// close factor <=  maxduration
func getClosetFactorForMaxDuration(maxduration, multipleOf int64) int64 {
	closedMaxDuration := getClosetFactor(maxduration, multipleOf)
	if closedMaxDuration == maxduration {
		return maxduration
	}

	// set closet maxduration closed to masduration
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

	if num1 > num2 {
		return num1
	}

	if num2 > num1 {
		return num2
	}
	// both must be equal here
	return num1
}
