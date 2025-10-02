package ctvlegacy

import (
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
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

func getAprc(adpodBids *AdPodBid) map[string]int64 {
	bidIdToAprcMap := make(map[string]int64)
	for _, bid := range adpodBids.Bids {
		bidIdToAprcMap[bid.ID] = bid.Status
	}
	return bidIdToAprcMap
}

// ConvertAPRCToNBRC converts the aprc to NonBidStatusCode
func ConvertAPRCToNBR(bidStatus int64) *openrtb3.NoBidReason {
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
