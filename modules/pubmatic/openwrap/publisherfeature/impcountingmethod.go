package publisherfeature

import (
	"encoding/json"

	"github.com/PubMatic-OpenWrap/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/golang/glog"
)

type impCountingMethod struct {
	enabledBidders map[string]struct{}
}

func (fe *feature) updateImpCountingMethodEnabledBidders() {
	if fe.publisherFeature == nil {
		return
	}

	enabledBidders := make(map[string]struct{})
	for _, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureImpCountingMethod]; ok && val.Enabled == 1 {
			var bidders []string
			if err := json.Unmarshal([]byte(val.Value), &bidders); err != nil {
				glog.Errorf("Error in unmarshalling imp counting method enabled bidders: %v", err)
				continue
			}
			for _, bidder := range bidders {
				enabledBidders[bidder] = struct{}{}
			}
		}
	}

	fe.Lock()
	fe.impCountingMethod.enabledBidders = enabledBidders
	fe.Unlock()
}

func (fe *feature) GetImpCountingMethodEnabledBidders() map[string]struct{} {
	enabledBidders := make(map[string]struct{})
	fe.RLock()
	defer fe.RUnlock()
	for bidder := range fe.impCountingMethod.enabledBidders {
		enabledBidders[bidder] = struct{}{}
	}
	return enabledBidders
}
