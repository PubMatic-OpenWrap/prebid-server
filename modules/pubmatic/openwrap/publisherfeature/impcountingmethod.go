package publisherfeature

import (
	"strings"

	"github.com/PubMatic-OpenWrap/prebid-server/v2/modules/pubmatic/openwrap/models"
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
			bidders := strings.Split(val.Value, ",")
			for _, bidder := range bidders {
				bidder = strings.TrimSpace(bidder)
				if bidder != "" {
					enabledBidders[bidder] = struct{}{}
				}
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
