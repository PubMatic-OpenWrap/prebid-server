package publisherfeature

import "github.com/PubMatic-OpenWrap/prebid-server/v2/modules/pubmatic/openwrap/models"

type bidRecovery struct {
	enabledPublishers map[int]struct{}
}

func (fe *feature) updateBidRecoveryEnabledPublishers() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublishers := make(map[int]struct{})
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureBidRecovery]; ok && val.Enabled == 1 {
			enabledPublishers[pubID] = struct{}{}
		}
	}

	fe.Lock()
	fe.bidRecovery.enabledPublishers = enabledPublishers
	fe.Unlock()
}

// IsBidRecoveryEnabled returns true if bid recovery is enabled for the given publisher
func (fe *feature) IsBidRecoveryEnabled(pubID int) bool {
	fe.RLock()
	defer fe.RUnlock()
	_, isPresent := fe.bidRecovery.enabledPublishers[pubID]
	return isPresent
}
