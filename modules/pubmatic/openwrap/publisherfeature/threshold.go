package publisherfeature

import "math/rand"

// isUnderThreshold returns 1 if the pub/dsp is applicable (not disabled and under DSP threshold), 0 otherwise.
// Shared by FSC and ACT: disabled => 0, else rand(100) < threshold => 1.
func isUnderThreshold(disabledPublishers map[int]struct{}, thresholdsPerDsp map[int]int, pubid, dspid int) int {
	if _, disabled := disabledPublishers[pubid]; disabled {
		return 0
	}
	if threshold, ok := thresholdsPerDsp[dspid]; ok && predictThresholdValue(threshold) {
		return 1
	}
	return 0
}

func predictThresholdValue(threshold int) bool {
	return rand.Intn(100) < threshold
}
