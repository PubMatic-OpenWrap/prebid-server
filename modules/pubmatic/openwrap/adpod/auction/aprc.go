package auction

import (
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/exchange"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
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
func ConvertAPRCToNBRC(bidStatus int64) *openrtb3.NoBidReason {
	var nbrCode openrtb3.NoBidReason

	switch bidStatus {
	case models.StatusOK:
		nbrCode = nbr.LossBidLostToHigherBid
	case models.StatusCategoryExclusion:
		nbrCode = exchange.ResponseRejectedCreativeCategoryExclusions
	case models.StatusDomainExclusion:
		nbrCode = exchange.ResponseRejectedCreativeAdvertiserExclusions
	case models.StatusDurationMismatch:
		nbrCode = exchange.ResponseRejectedInvalidCreative
	default:
		return nil
	}
	return &nbrCode
}
