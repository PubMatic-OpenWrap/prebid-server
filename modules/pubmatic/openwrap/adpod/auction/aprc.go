package auction

import (
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func collectAPRC(impAdpodBidsMap map[string]*AdPodBid, impCtxMap map[string]models.ImpCtx) {
	for impId, adpodBid := range impAdpodBidsMap {
		if len(adpodBid.Bids) == 0 {
			continue
		}

		bidIdToAprcMap := make(map[string]int64)
		for _, bid := range adpodBid.Bids {
			bidIdToAprcMap[bid.ID] = bid.Status
		}

		impCtx := impCtxMap[impId]
		impCtx.BidIDToAPRC = bidIdToAprcMap
		impCtxMap[impId] = impCtx
	}
}

// ConvertAPRCToNBRC converts the aprc to NonBidStatusCode
func ConvertAPRCToNBRC(bidStatus int64) *openrtb3.NonBidStatusCode {
	var nbrCode openrtb3.NonBidStatusCode

	switch bidStatus {
	case models.StatusOK:
		nbrCode = openrtb3.LossBidLostToHigherBid
	case models.StatusCategoryExclusion:
		nbrCode = openrtb3.LossBidCategoryExclusions
	case models.StatusDomainExclusion:
		nbrCode = openrtb3.LossBidAdvertiserExclusions
	case models.StatusDurationMismatch:
		nbrCode = openrtb3.LossBidInvalidCreative
	default:
		return nil
	}
	return &nbrCode
}
