package featurereloader

import (
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type ampMultiformat struct {
	enabledPublishers map[int]struct{}
}

// fetch and update fsc config maps from DB
func updateAmpMutiformatConfigFromCache(publisherFeatureMap map[int]int) {

	enabledPublishers := make(map[int]struct{})
	for pubID, featureID := range publisherFeatureMap {
		if featureID == models.FeatureAMPMultiFormat {
			enabledPublishers[pubID] = struct{}{}
		}
	}

	reloaderConfig.Lock()
	reloaderConfig.ampMultiformat.enabledPublishers = enabledPublishers
	reloaderConfig.Unlock()
}

func IsAmpMultformatEnabled(pubid int) bool {
	reloaderConfig.RLock()
	defer reloaderConfig.RUnlock()

	if _, isPresent := reloaderConfig.ampMultiformat.enabledPublishers[pubid]; isPresent {
		return true
	}
	return false
}

// Exposed for test cases
func SetAndResetAmpMultiformatWithMockCache(mockCache cache.Cache, enabledPublishers map[int]struct{}) func() {
	reloaderConfig.cache = mockCache
	reloaderConfig.ampMultiformat.enabledPublishers = enabledPublishers
	return func() {
		reloaderConfig.cache = nil
		reloaderConfig.ampMultiformat.enabledPublishers = make(map[int]struct{})
	}
}
