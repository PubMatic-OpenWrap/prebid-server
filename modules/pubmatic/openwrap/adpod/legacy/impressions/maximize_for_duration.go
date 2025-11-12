package ctvlegacy

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type ForDuration struct {
	generator
}

// newMaximizeForDuration Constucts the generator object from openrtb_ext.VideoAdPod
// It computes durations for Ad Slot and Ad Pod in multiple of X
func newMaximizeForDuration(podMinDuration, podMaxDuration int64, vPod *models.AdPod) ForDuration {
	config := newConfigWithMultipleOf(podMinDuration, podMaxDuration, vPod, multipleOf)

	// util.Logf("Computed podMinDuration = %v in multiples of %v (requestedPodMinDuration = %v)\n", config.requested.podMinDuration, multipleOf, config.requested.podMinDuration)
	// util.Logf("Computed podMaxDuration = %v in multiples of %v (requestedPodMaxDuration = %v)\n", config.requested.podMaxDuration, multipleOf, config.requested.podMaxDuration)
	// util.Logf("Computed slotMinDuration = %v in multiples of %v (requestedSlotMinDuration = %v)\n", config.internal.slotMinDuration, multipleOf, config.requested.slotMinDuration)
	// util.Logf("Computed slotMaxDuration = %v in multiples of %v (requestedSlotMaxDuration = %v)\n", config.internal.slotMaxDuration, multipleOf, *vPod.MaxDuration)
	// util.Logf("Requested minAds = %v\n", config.requested.minAds)
	// util.Logf("Requested maxAds = %v\n", config.requested.maxAds)
	return ForDuration{config}
}

// Algorithm returns MaximizeForDuration
func (fd *ForDuration) Algorithm() int {
	return models.MaximizeForDuration
}
