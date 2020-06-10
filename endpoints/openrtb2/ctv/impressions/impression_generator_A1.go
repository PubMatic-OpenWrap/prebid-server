// Package impressions provides various algorithms to get the number of impressions
// along with minimum and maximum duration of each impression.
// It uses Ad pod request for it
package impressions

import (
	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

// Constucts the adPodConfig object from openrtb_ext.VideoAdPod
// It computes durations for Ad Slot and Ad Pod in multiple of X
func newImpGenA1(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod) adPodConfig {
	config := newConfigWithMultipleOf(podMinDuration, podMaxDuration, vPod, multipleOf)

	ctv.Logf("Computed podMinDuration = %v in multiples of %v (requestedPodMinDuration = %v)\n", config.podMinDuration, multipleOf, config.requestedPodMinDuration)
	ctv.Logf("Computed podMaxDuration = %v in multiples of %v (requestedPodMaxDuration = %v)\n", config.podMaxDuration, multipleOf, config.requestedPodMaxDuration)
	ctv.Logf("Computed slotMinDuration = %v in multiples of %v (requestedSlotMinDuration = %v)\n", config.slotMinDuration, multipleOf, config.requestedSlotMinDuration)
	ctv.Logf("Computed slotMaxDuration = %v in multiples of %v (requestedSlotMaxDuration = %v)\n", config.slotMaxDuration, multipleOf, *vPod.MaxDuration)
	ctv.Logf("Requested minAds = %v\n", config.minAds)
	ctv.Logf("Requested maxAds = %v\n", config.maxAds)
	return config
}

// Algorithm returns Algorithm1
func (config adPodConfig) Algorithm() int {
	return Algorithm1
}
