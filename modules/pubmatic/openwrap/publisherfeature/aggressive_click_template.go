package publisherfeature

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type act struct {
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
}

/*
isUnderACTThreshold:- returns act 1/0 based on:
1. When publisher has disabled ACT in DB, return 0
2. If ACT is enabled for publisher(default), consider DSP-threshold , and predict value of act 0 or 1.
3. If dspId is not present return 0
*/
func (fe *feature) isUnderACTThreshold(pubid int, dspid int) int {
	fe.RLock()
	defer fe.RUnlock()
	return isUnderThreshold(fe.act.disabledPublishers, fe.act.thresholdsPerDsp, pubid, dspid)
}

// IsActApplicable returns true if act can be applied (act=1)
func (fe *feature) IsActApplicable(pubId int, seat string, dspId int) bool {
	return models.IsPubmaticCorePartner(seat) && (fe.isUnderACTThreshold(pubId, dspId) != 0)
}
