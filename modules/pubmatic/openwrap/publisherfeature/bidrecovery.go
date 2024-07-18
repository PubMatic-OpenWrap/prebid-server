package publisherfeature

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type bidRecovery struct {
	enabledPublisherProfile map[int]map[int]struct{}
}

func (fe *feature) updateBidRecoveryEnabledPublishers() {
	if fe.publisherFeature == nil {
		return
	}

	enabledPublisherProfile := make(map[int]map[int]struct{})
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureBidRecovery]; ok && val.Enabled == 1 && len(val.Value) > 0 {
			var profiles []int
			if err := json.Unmarshal([]byte(val.Value), &profiles); err != nil {
				glog.Errorf("ErrJSONUnmarshalFailed BidRecovery pubid: %d profiles: %s err: %s", pubID, val.Value, err.Error())
				continue
			}
			enabledProfiles := make(map[int]struct{})
			for _, profileID := range profiles {
				enabledProfiles[profileID] = struct{}{}
			}
			enabledPublisherProfile[pubID] = enabledProfiles
		}
	}
	fe.Lock()
	fe.bidRecovery.enabledPublisherProfile = enabledPublisherProfile
	fe.Unlock()
}

// IsBidRecoveryEnabled returns true if bid recovery is enabled for the given publisher
func (fe *feature) IsBidRecoveryEnabled(pubID int, profileID int) bool {
	fe.RLock()
	defer fe.RUnlock()
	_, isPresent := fe.bidRecovery.enabledPublisherProfile[pubID][profileID]
	return isPresent
}
