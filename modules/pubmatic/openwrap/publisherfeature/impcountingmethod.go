package publisherfeature

import (
	"strings"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type impCountingMethod struct {
	enabledBidders [2]map[string]struct{}
	index          int32
}

func newImpCountingMethod() impCountingMethod {
	return impCountingMethod{
		enabledBidders: [2]map[string]struct{}{
			make(map[string]struct{}),
			make(map[string]struct{}),
		},
		index: 0,
	}
}

func (fe *feature) updateImpCountingMethodEnabledBidders() {
	if fe.publisherFeature == nil {
		return
	}

	enabledBidders := make(map[string]struct{})
	for pubID, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureImpCountingMethod]; ok && pubID == 0 && val.Enabled == 1 {
			bidders := strings.Split(val.Value, ",")
			for _, bidder := range bidders {
				bidder = strings.TrimSpace(bidder)
				if bidder != "" {
					enabledBidders[bidder] = struct{}{}
				}
			}
		}
	}

	fe.impCountingMethod.enabledBidders[fe.impCountingMethod.index^1] = enabledBidders
	fe.impCountingMethod.index ^= 1
}

func (fe *feature) GetImpCountingMethodEnabledBidders() map[string]struct{} {
	return fe.impCountingMethod.enabledBidders[fe.impCountingMethod.index]
}
