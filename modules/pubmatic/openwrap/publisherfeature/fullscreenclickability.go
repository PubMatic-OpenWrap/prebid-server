package publisherfeature

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type fsc struct {
	disabledPublishers map[int]struct{}
	thresholdsPerDsp   map[int]int
}

/*
IsUnderFSCThreshold:- returns fsc 1/0 based on:
1. When publisher has disabled FSC in DB, return 0
2. If FSC is enabled for publisher(default), consider DSP-threshold , and predict value of fsc 0 or 1.
3. If dspId is not present return 0
*/
func (fe *feature) isUnderFSCThreshold(pubid int, dspid int) int {
	fe.RLock()
	defer fe.RUnlock()
	return isUnderThreshold(fe.fsc.disabledPublishers, fe.fsc.thresholdsPerDsp, pubid, dspid)
}

// IsFscApplicable returns true if fsc can be applied (fsc=1)
func (fe *feature) IsFscApplicable(pubId int, seat string, dspId int) bool {
	return models.IsPubmaticCorePartner(seat) && (fe.isUnderFSCThreshold(pubId, dspId) != 0)
}
