package auction

import (
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

// minduration/maxduration OR rqddurs
func durationOK(dur int64, s models.SlotConfig) bool {
	if len(s.RqdDurs) > 0 {
		for _, d := range s.RqdDurs {
			if d == dur {
				return true
			}
		}
		return false
	}
	if s.MinDuration > 0 && dur < s.MinDuration {
		return false
	}
	if s.MaxDuration > 0 && dur > s.MaxDuration {
		return false
	}
	return true
}

func exclusionSatisfied(excl models.ExclusionConfig, c *podBid, usedDom, usedCat map[string]struct{}) bool {
	if c == nil {
		return false
	}
	if excl.AdvertiserDomainExclusion {
		for _, d := range c.ADomain {
			if _, ok := usedDom[d]; ok {
				c.Nbr = ptrutil.ToPtr(exchange.ResponseRejectedCreativeAdvertiserExclusions)
				return false
			}
		}
	}
	if excl.IABCategoryExclusion {
		for _, cat := range c.Cat {
			if _, ok := usedCat[cat]; ok {
				c.Nbr = ptrutil.ToPtr(exchange.ResponseRejectedCreativeCategoryExclusions)
				return false
			}
		}
	}
	return true
}

func deepCloneMap(m map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(m))
	for k := range m {
		out[k] = struct{}{}
	}
	return out
}
